// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

const (
	// stackDebug == 0: no logging
	//            == 1: logging of per-stack operations
	//            == 2: logging of per-frame operations
	//            == 3: logging of per-word updates
	//            == 4: logging of per-word reads
	StackDebug       = 0
	StackFromSystem  = 0 // allocate stacks from system memory instead of the heap
	StackFaultOnFree = 0 // old stacks are mapped noaccess to detect use after free
	StackPoisonCopy  = 0 // fill stack that should not be accessed with garbage, to detect bad dereferences during copy

	StackCache = 1
)

const (
	UintptrMask = 1<<(8*PtrSize) - 1
	PoisonStack = UintptrMask & 0x6868686868686868

	// Goroutine preemption request.
	// Stored into g->stackguard0 to cause split stack check failure.
	// Must be greater than any real sp.
	// 0xfffffade in hex.
	StackPreempt = UintptrMask & -1314

	// Thread is forking.
	// Stored into g->stackguard0 to cause split stack check failure.
	// Must be greater than any real sp.
	StackFork = UintptrMask & -1234
)

// Global pool of spans that have free stacks.
// Stacks are assigned an order according to size.
//     order = log_2(size/FixedStack)
// There is a free list for each order.
// TODO: one lock per order?
var Stackpool [NumStackOrders]Mspan
var Stackpoolmu Mutex

// Cached value of haveexperiment("framepointer")
var Framepointer_enabled bool

// Allocates a stack from the free pool.  Must be called with
// stackpoolmu held.
func stackpoolalloc(order uint8) Gclinkptr {
	list := &Stackpool[order]
	s := list.Next
	if s == list {
		// no free stacks.  Allocate another span worth.
		s = mHeap_AllocStack(&Mheap_, StackCacheSize>>PageShift)
		if s == nil {
			Throw("out of memory")
		}
		if s.Ref != 0 {
			Throw("bad ref")
		}
		if s.Freelist.Ptr() != nil {
			Throw("bad freelist")
		}
		for i := uintptr(0); i < StackCacheSize; i += FixedStack << order {
			x := Gclinkptr(uintptr(s.Start)<<PageShift + i)
			x.Ptr().Next = s.Freelist
			s.Freelist = x
		}
		MSpanList_Insert(list, s)
	}
	x := s.Freelist
	if x.Ptr() == nil {
		Throw("span has no free stacks")
	}
	s.Freelist = x.Ptr().Next
	s.Ref++
	if s.Freelist.Ptr() == nil {
		// all stacks in s are allocated.
		MSpanList_Remove(s)
	}
	return x
}

// stackcacherefill/stackcacherelease implement a global pool of stack segments.
// The pool is required to prevent unlimited growth of per-thread caches.
func stackcacherefill(c *Mcache, order uint8) {
	if StackDebug >= 1 {
		print("stackcacherefill order=", order, "\n")
	}

	// Grab some stacks from the global cache.
	// Grab half of the allowed capacity (to prevent thrashing).
	var list Gclinkptr
	var size uintptr
	Lock(&Stackpoolmu)
	for size < StackCacheSize/2 {
		x := stackpoolalloc(order)
		x.Ptr().Next = list
		list = x
		size += FixedStack << order
	}
	Unlock(&Stackpoolmu)
	c.Stackcache[order].List = list
	c.Stackcache[order].Size = size
}

func Stackalloc(n uint32) (Stack, []Stkbar) {
	// Stackalloc must be called on scheduler stack, so that we
	// never try to grow the stack during the code that stackalloc runs.
	// Doing so would cause a deadlock (issue 1547).
	thisg := Getg()
	if thisg != thisg.M.G0 {
		Throw("stackalloc not on scheduler stack")
	}
	if n&(n-1) != 0 {
		Throw("stack size not a power of 2")
	}
	if StackDebug >= 1 {
		print("stackalloc ", n, "\n")
	}

	// Compute the size of stack barrier array.
	maxstkbar := gcMaxStackBarriers(int(n))
	nstkbar := unsafe.Sizeof(Stkbar{}) * uintptr(maxstkbar)

	if Debug.Efence != 0 || StackFromSystem != 0 {
		v := SysAlloc(Round(uintptr(n), PageSize), &Memstats.Stacks_sys)
		if v == nil {
			Throw("out of memory (stackalloc)")
		}
		top := uintptr(n) - nstkbar
		stkbarSlice := Slice{Add(v, top), 0, maxstkbar}
		return Stack{uintptr(v), uintptr(v) + top}, *(*[]Stkbar)(unsafe.Pointer(&stkbarSlice))
	}

	// Small stacks are allocated with a fixed-size free-list allocator.
	// If we need a stack of a bigger size, we fall back on allocating
	// a dedicated span.
	var v unsafe.Pointer
	if StackCache != 0 && n < FixedStack<<NumStackOrders && n < StackCacheSize {
		order := uint8(0)
		n2 := n
		for n2 > FixedStack {
			order++
			n2 >>= 1
		}
		var x Gclinkptr
		c := thisg.M.Mcache
		if c == nil || thisg.M.Preemptoff != "" || thisg.M.Helpgc != 0 {
			// c == nil can happen in the guts of exitsyscall or
			// procresize. Just get a stack from the global pool.
			// Also don't touch stackcache during gc
			// as it's flushed concurrently.
			Lock(&Stackpoolmu)
			x = stackpoolalloc(order)
			Unlock(&Stackpoolmu)
		} else {
			x = c.Stackcache[order].List
			if x.Ptr() == nil {
				stackcacherefill(c, order)
				x = c.Stackcache[order].List
			}
			c.Stackcache[order].List = x.Ptr().Next
			c.Stackcache[order].Size -= uintptr(n)
		}
		v = (unsafe.Pointer)(x)
	} else {
		s := mHeap_AllocStack(&Mheap_, Round(uintptr(n), PageSize)>>PageShift)
		if s == nil {
			Throw("out of memory")
		}
		v = (unsafe.Pointer)(s.Start << PageShift)
	}

	if Raceenabled {
		Racemalloc(v, uintptr(n))
	}
	if StackDebug >= 1 {
		print("  allocated ", v, "\n")
	}
	top := uintptr(n) - nstkbar
	stkbarSlice := Slice{Add(v, top), 0, maxstkbar}
	return Stack{uintptr(v), uintptr(v) + top}, *(*[]Stkbar)(unsafe.Pointer(&stkbarSlice))
}

// Information from the compiler about the layout of stack frames.
type Bitvector struct {
	N        int32 // # of bits
	Bytedata *uint8
}

// round x up to a power of 2.
func Round2(x int32) int32 {
	s := uint(0)
	for 1<<s < x {
		s++
	}
	return 1 << s
}
