// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

const (
	// StackDebug == 0: no logging
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

// Global pool of spans that have free stacks.
// Stacks are assigned an order according to size.
//     order = log_2(size/FixedStack)
// There is a free list for each order.
// TODO: one lock per order?
var Stackpool [_core.NumStackOrders]_core.Mspan
var Stackpoolmu _core.Mutex

// Allocates a stack from the free pool.  Must be called with
// stackpoolmu held.
func stackpoolalloc(order uint8) _core.Gclinkptr {
	list := &Stackpool[order]
	s := list.Next
	if s == list {
		// no free stacks.  Allocate another span worth.
		s = mHeap_AllocStack(&_lock.Mheap_, _core.StackCacheSize>>_core.PageShift)
		if s == nil {
			_lock.Throw("out of memory")
		}
		if s.Ref != 0 {
			_lock.Throw("bad ref")
		}
		if s.Freelist.Ptr() != nil {
			_lock.Throw("bad freelist")
		}
		for i := uintptr(0); i < _core.StackCacheSize; i += _core.FixedStack << order {
			x := _core.Gclinkptr(uintptr(s.Start)<<_core.PageShift + i)
			x.Ptr().Next = s.Freelist
			s.Freelist = x
		}
		MSpanList_Insert(list, s)
	}
	x := s.Freelist
	if x.Ptr() == nil {
		_lock.Throw("span has no free stacks")
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
func stackcacherefill(c *_core.Mcache, order uint8) {
	if StackDebug >= 1 {
		print("stackcacherefill order=", order, "\n")
	}

	// Grab some stacks from the global cache.
	// Grab half of the allowed capacity (to prevent thrashing).
	var list _core.Gclinkptr
	var size uintptr
	_lock.Lock(&Stackpoolmu)
	for size < _core.StackCacheSize/2 {
		x := stackpoolalloc(order)
		x.Ptr().Next = list
		list = x
		size += _core.FixedStack << order
	}
	_lock.Unlock(&Stackpoolmu)
	c.Stackcache[order].List = list
	c.Stackcache[order].Size = size
}

func Stackalloc(n uint32) _core.Stack {
	// Stackalloc must be called on scheduler stack, so that we
	// never try to grow the stack during the code that stackalloc runs.
	// Doing so would cause a deadlock (issue 1547).
	thisg := _core.Getg()
	if thisg != thisg.M.G0 {
		_lock.Throw("stackalloc not on scheduler stack")
	}
	if n&(n-1) != 0 {
		_lock.Throw("stack size not a power of 2")
	}
	if StackDebug >= 1 {
		print("stackalloc ", n, "\n")
	}

	if _lock.Debug.Efence != 0 || StackFromSystem != 0 {
		v := _lock.SysAlloc(Round(uintptr(n), _core.PageSize), &_lock.Memstats.Stacks_sys)
		if v == nil {
			_lock.Throw("out of memory (stackalloc)")
		}
		return _core.Stack{uintptr(v), uintptr(v) + uintptr(n)}
	}

	// Small stacks are allocated with a fixed-size free-list allocator.
	// If we need a stack of a bigger size, we fall back on allocating
	// a dedicated span.
	var v unsafe.Pointer
	if StackCache != 0 && n < _core.FixedStack<<_core.NumStackOrders && n < _core.StackCacheSize {
		order := uint8(0)
		n2 := n
		for n2 > _core.FixedStack {
			order++
			n2 >>= 1
		}
		var x _core.Gclinkptr
		c := thisg.M.Mcache
		if c == nil || thisg.M.Gcing != 0 || thisg.M.Helpgc != 0 {
			// c == nil can happen in the guts of exitsyscall or
			// procresize. Just get a stack from the global pool.
			// Also don't touch stackcache during gc
			// as it's flushed concurrently.
			_lock.Lock(&Stackpoolmu)
			x = stackpoolalloc(order)
			_lock.Unlock(&Stackpoolmu)
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
		s := mHeap_AllocStack(&_lock.Mheap_, Round(uintptr(n), _core.PageSize)>>_core.PageShift)
		if s == nil {
			_lock.Throw("out of memory")
		}
		v = (unsafe.Pointer)(s.Start << _core.PageShift)
	}

	if Raceenabled {
		Racemalloc(v, uintptr(n))
	}
	if StackDebug >= 1 {
		print("  allocated ", v, "\n")
	}
	return _core.Stack{uintptr(v), uintptr(v) + uintptr(n)}
}

// round x up to a power of 2.
func Round2(x int32) int32 {
	s := uint(0)
	for 1<<s < x {
		s++
	}
	return 1 << s
}
