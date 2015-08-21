// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Page heap.
//
// See malloc.go for overview.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

var H_allspans []*_base.Mspan // TODO: make this h.allspans once mheap can be defined in Go

// Look up the span at the given address.
// Address is guaranteed to be in map
// and is guaranteed to be start or end of span.
func mHeap_Lookup(h *_base.Mheap, v unsafe.Pointer) *_base.Mspan {
	p := uintptr(v)
	p -= uintptr(unsafe.Pointer(h.Arena_start))
	return _base.H_spans[p>>_base.XPageShift]
}

// Free the span back into the heap.
func mHeap_Free(h *_base.Mheap, s *_base.Mspan, acct int32) {
	_base.Systemstack(func() {
		mp := _base.Getg().M
		_base.Lock(&h.Lock)
		_base.Memstats.Heap_live += uint64(mp.Mcache.Local_cachealloc)
		mp.Mcache.Local_cachealloc = 0
		_base.Memstats.Heap_scan += uint64(mp.Mcache.Local_scan)
		mp.Mcache.Local_scan = 0
		_base.Memstats.Tinyallocs += uint64(mp.Mcache.Local_tinyallocs)
		mp.Mcache.Local_tinyallocs = 0
		if acct != 0 {
			_base.Memstats.Heap_objects--
		}
		_base.GcController.Revise()
		_base.MHeap_FreeSpanLocked(h, s, true, true, 0)
		if _base.Trace.Enabled {
			TraceHeapAlloc()
		}
		_base.Unlock(&h.Lock)
	})
}

func mHeap_FreeStack(h *_base.Mheap, s *_base.Mspan) {
	_g_ := _base.Getg()
	if _g_ != _g_.M.G0 {
		_base.Throw("mheap_freestack not on g0 stack")
	}
	s.Needzero = 1
	_base.Lock(&h.Lock)
	_base.Memstats.Stacks_inuse -= uint64(s.Npages << _base.XPageShift)
	_base.MHeap_FreeSpanLocked(h, s, true, true, 0)
	_base.Unlock(&h.Lock)
}

const (
	KindSpecialFinalizer = 1
	KindSpecialProfile   = 2
	// Note: The finalizer special must be first because if we're freeing
	// an object, a finalizer special will cause the freeing operation
	// to abort, and we want to keep the other special records around
	// if that happens.
)

// The described object has a finalizer set for it.
type Specialfinalizer struct {
	Special _base.Special
	Fn      *_base.Funcval
	Nret    uintptr
	Fint    *_base.Type
	Ot      *Ptrtype
}

// The described object is being heap profiled.
type Specialprofile struct {
	Special _base.Special
	B       *Bucket
}

// Do whatever cleanup needs to be done to deallocate s.  It has
// already been unlinked from the MSpan specials list.
// Returns true if we should keep working on deallocating p.
func freespecial(s *_base.Special, p unsafe.Pointer, size uintptr, freed bool) bool {
	switch s.Kind {
	case KindSpecialFinalizer:
		sf := (*Specialfinalizer)(unsafe.Pointer(s))
		queuefinalizer(p, sf.Fn, sf.Nret, sf.Fint, sf.Ot)
		_base.Lock(&_base.Mheap_.Speciallock)
		_base.FixAlloc_Free(&_base.Mheap_.Specialfinalizeralloc, (unsafe.Pointer)(sf))
		_base.Unlock(&_base.Mheap_.Speciallock)
		return false // don't free p until finalizer is done
	case KindSpecialProfile:
		sp := (*Specialprofile)(unsafe.Pointer(s))
		mProf_Free(sp.B, size, freed)
		_base.Lock(&_base.Mheap_.Speciallock)
		_base.FixAlloc_Free(&_base.Mheap_.Specialprofilealloc, (unsafe.Pointer)(sp))
		_base.Unlock(&_base.Mheap_.Speciallock)
		return true
	default:
		_base.Throw("bad special kind")
		panic("not reached")
	}
}
