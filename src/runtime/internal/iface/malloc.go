// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory allocator, based on tcmalloc.
// http://goog-perftools.sourceforge.net/doc/tcmalloc.html

// The main allocator works in runs of pages.
// Small allocation sizes (up to and including 32 kB) are
// rounded to one of about 100 size classes, each of which
// has its own free list of objects of exactly that size.
// Any free page of memory can be split into a set of objects
// of one size class, which are then managed using free list
// allocators.
//
// The allocator's data structures are:
//
//	FixAlloc: a free-list allocator for fixed-size objects,
//		used to manage storage used by the allocator.
//	MHeap: the malloc heap, managed at page (4096-byte) granularity.
//	MSpan: a run of pages managed by the MHeap.
//	MCentral: a shared free list for a given size class.
//	MCache: a per-thread (in Go, per-P) cache for small objects.
//	MStats: allocation statistics.
//
// Allocating a small object proceeds up a hierarchy of caches:
//
//	1. Round the size up to one of the small size classes
//	   and look in the corresponding MCache free list.
//	   If the list is not empty, allocate an object from it.
//	   This can all be done without acquiring a lock.
//
//	2. If the MCache free list is empty, replenish it by
//	   taking a bunch of objects from the MCentral free list.
//	   Moving a bunch amortizes the cost of acquiring the MCentral lock.
//
//	3. If the MCentral free list is empty, replenish it by
//	   allocating a run of pages from the MHeap and then
//	   chopping that memory into objects of the given size.
//	   Allocating many objects amortizes the cost of locking
//	   the heap.
//
//	4. If the MHeap is empty or has no page runs large enough,
//	   allocate a new group of pages (at least 1MB) from the
//	   operating system.  Allocating a large run of pages
//	   amortizes the cost of talking to the operating system.
//
// Freeing a small object proceeds up the same hierarchy:
//
//	1. Look up the size class for the object and add it to
//	   the MCache free list.
//
//	2. If the MCache free list is too long or the MCache has
//	   too much memory, return some to the MCentral free lists.
//
//	3. If all the objects in a given span have returned to
//	   the MCentral list, return that span to the page heap.
//
//	4. If the heap has too much memory, return some to the
//	   operating system.
//
//	TODO(rsc): Step 4 is not implemented.
//
// Allocating and freeing a large object uses the page heap
// directly, bypassing the MCache and MCentral free lists.
//
// The small objects on the MCache and MCentral free lists
// may or may not be zeroed.  They are zeroed if and only if
// the second word of the object is zero.  A span in the
// page heap is zeroed unless s->needzero is set. When a span
// is allocated to break into small objects, it is zeroed if needed
// and s->needzero is set. There are two main benefits to delaying the
// zeroing this way:
//
//	1. stack frames allocated from the small object lists
//	   or the page heap can avoid zeroing altogether.
//	2. the cost of zeroing when reusing a small object is
//	   charged to the mutator, not the garbage collector.
//
// This code was written with an eye toward translating to Go
// in the future.  Methods have the form Type_Method(Type *t, ...).

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

// base address for all 0-byte allocations
var Zerobase uintptr

// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// Large objects (> 32 kB) are allocated straight from the heap.
func Mallocgc(size uintptr, typ *_base.Type, flags uint32) unsafe.Pointer {
	if _base.Gcphase == _base.GCmarktermination {
		_base.Throw("mallocgc called with gcphase == _GCmarktermination")
	}

	if size == 0 {
		return unsafe.Pointer(&Zerobase)
	}

	if flags&_base.XFlagNoScan == 0 && typ == nil {
		_base.Throw("malloc missing type")
	}

	if _base.Debug.Sbrk != 0 {
		align := uintptr(16)
		if typ != nil {
			align = uintptr(typ.Align)
		}
		return _base.Persistentalloc(size, align, &_base.Memstats.Other_sys)
	}

	// Set mp.mallocing to keep from being preempted by GC.
	mp := _base.Acquirem()
	if mp.Mallocing != 0 {
		_base.Throw("malloc deadlock")
	}
	if mp.Gsignal == _base.Getg() {
		_base.Throw("malloc during signal")
	}
	mp.Mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := Gomcache()
	var s *_base.Mspan
	var x unsafe.Pointer
	if size <= _base.XMaxSmallSize {
		if flags&_base.XFlagNoScan != 0 && size < _base.MaxTinySize {
			// Tiny allocator.
			//
			// Tiny allocator combines several tiny allocation requests
			// into a single memory block. The resulting memory block
			// is freed when all subobjects are unreachable. The subobjects
			// must be FlagNoScan (don't have pointers), this ensures that
			// the amount of potentially wasted memory is bounded.
			//
			// Size of the memory block used for combining (maxTinySize) is tunable.
			// Current setting is 16 bytes, which relates to 2x worst case memory
			// wastage (when all but one subobjects are unreachable).
			// 8 bytes would result in no wastage at all, but provides less
			// opportunities for combining.
			// 32 bytes provides more opportunities for combining,
			// but can lead to 4x worst case wastage.
			// The best case winning is 8x regardless of block size.
			//
			// Objects obtained from tiny allocator must not be freed explicitly.
			// So when an object will be freed explicitly, we ensure that
			// its size >= maxTinySize.
			//
			// SetFinalizer has a special case for objects potentially coming
			// from tiny allocator, it such case it allows to set finalizers
			// for an inner byte of a memory block.
			//
			// The main targets of tiny allocator are small strings and
			// standalone escaping variables. On a json benchmark
			// the allocator reduces number of allocations by ~12% and
			// reduces heap size by ~20%.
			off := c.Tinyoffset
			// Align tiny pointer for required (conservative) alignment.
			if size&7 == 0 {
				off = _base.Round(off, 8)
			} else if size&3 == 0 {
				off = _base.Round(off, 4)
			} else if size&1 == 0 {
				off = _base.Round(off, 2)
			}
			if off+size <= _base.MaxTinySize && c.Tiny != nil {
				// The object fits into existing tiny block.
				x = _base.Add(c.Tiny, off)
				c.Tinyoffset = off + size
				c.Local_tinyallocs++
				mp.Mallocing = 0
				_base.Releasem(mp)
				return x
			}
			// Allocate a new maxTinySize block.
			s = c.Alloc[_base.XTinySizeClass]
			v := s.Freelist
			if v.Ptr() == nil {
				_base.Systemstack(func() {
					mCache_Refill(c, _base.XTinySizeClass)
				})
				shouldhelpgc = true
				s = c.Alloc[_base.XTinySizeClass]
				v = s.Freelist
			}
			s.Freelist = v.Ptr().Next
			s.Ref++
			// prefetchnta offers best performance, see change list message.
			_base.Prefetchnta(uintptr(v.Ptr().Next))
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if size < c.Tinyoffset {
				c.Tiny = x
				c.Tinyoffset = size
			}
			size = _base.MaxTinySize
		} else {
			var sizeclass int8
			if size <= 1024-8 {
				sizeclass = Size_to_class8[(size+7)>>3]
			} else {
				sizeclass = Size_to_class128[(size-1024+127)>>7]
			}
			size = uintptr(Class_to_size[sizeclass])
			s = c.Alloc[sizeclass]
			v := s.Freelist
			if v.Ptr() == nil {
				_base.Systemstack(func() {
					mCache_Refill(c, int32(sizeclass))
				})
				shouldhelpgc = true
				s = c.Alloc[sizeclass]
				v = s.Freelist
			}
			s.Freelist = v.Ptr().Next
			s.Ref++
			// prefetchnta offers best performance, see change list message.
			_base.Prefetchnta(uintptr(v.Ptr().Next))
			x = unsafe.Pointer(v)
			if flags&_base.XFlagNoZero == 0 {
				v.Ptr().Next = 0
				if size > 2*_base.PtrSize && ((*[2]uintptr)(x))[1] != 0 {
					_base.Memclr(unsafe.Pointer(v), size)
				}
			}
		}
		c.Local_cachealloc += size
	} else {
		var s *_base.Mspan
		shouldhelpgc = true
		_base.Systemstack(func() {
			s = largeAlloc(size, uint32(flags))
		})
		x = unsafe.Pointer(uintptr(s.Start << _base.XPageShift))
		size = uintptr(s.Elemsize)
	}

	if flags&_base.XFlagNoScan != 0 {
		// All objects are pre-marked as noscan. Nothing to do.
	} else {
		// If allocating a defer+arg block, now that we've picked a malloc size
		// large enough to hold everything, cut the "asked for" size down to
		// just the defer header, so that the GC bitmap will record the arg block
		// as containing nothing at all (as if it were unused space at the end of
		// a malloc block caused by size rounding).
		// The defer arg areas are scanned as part of scanstack.
		if typ == DeferType {
			dataSize = unsafe.Sizeof(_base.Defer{})
		}
		heapBitsSetType(uintptr(x), size, dataSize, typ)
		if dataSize > typ.Size {
			// Array allocation. If there are any
			// pointers, GC has to scan to the last
			// element.
			if typ.Ptrdata != 0 {
				c.Local_scan += dataSize - typ.Size + typ.Ptrdata
			}
		} else {
			c.Local_scan += typ.Ptrdata
		}

		// Ensure that the stores above that initialize x to
		// type-safe memory and set the heap bits occur before
		// the caller can make x observable to the garbage
		// collector. Otherwise, on weakly ordered machines,
		// the garbage collector could follow a pointer to x,
		// but see uninitialized memory or stale heap bits.
		publicationBarrier()
	}

	// GCmarkterminate allocates black
	// All slots hold nil so no scanning is needed.
	// This may be racing with GC so do it atomically if there can be
	// a race marking the bit.
	if _base.Gcphase == _base.GCmarktermination || _base.GcBlackenPromptly {
		_base.Systemstack(func() {
			gcmarknewobject_m(uintptr(x), size)
		})
	}

	if _base.Raceenabled {
		_base.Racemalloc(x, size)
	}

	mp.Mallocing = 0
	_base.Releasem(mp)

	if _base.Debug.Allocfreetrace != 0 {
		tracealloc(x, size, typ)
	}

	if rate := _base.MemProfileRate; rate > 0 {
		if size < uintptr(rate) && int32(size) < c.Next_sample {
			c.Next_sample -= int32(size)
		} else {
			mp := _base.Acquirem()
			profilealloc(mp, x, size)
			_base.Releasem(mp)
		}
	}

	if shouldhelpgc && shouldtriggergc() {
		StartGC(_gc.GcBackgroundMode, false)
	} else if _base.GcBlackenEnabled != 0 {
		// Assist garbage collector. We delay this until the
		// epilogue so that it doesn't interfere with the
		// inner working of malloc such as mcache refills that
		// might happen while doing the gcAssistAlloc.
		gcAssistAlloc(size, shouldhelpgc)
	} else if shouldhelpgc && Bggc.Working != 0 {
		// The GC is starting up or shutting down, so we can't
		// assist, but we also can't allocate unabated. Slow
		// down this G's allocation and help the GC stay
		// scheduled by yielding.
		//
		// TODO: This is a workaround. Either help the GC make
		// the transition or block.
		gp := _base.Getg()
		if gp != gp.M.G0 && gp.M.Locks == 0 && gp.M.Preemptoff == "" {
			_gc.Gosched()
		}
	}

	return x
}

func largeAlloc(size uintptr, flag uint32) *_base.Mspan {
	// print("largeAlloc size=", size, "\n")

	if size+_base.PageSize < size {
		_base.Throw("out of memory")
	}
	npages := size >> _base.PageShift
	if size&_base.PageMask != 0 {
		npages++
	}

	// Deduct credit for this span allocation and sweep if
	// necessary. mHeap_Alloc will also sweep npages, so this only
	// pays the debt down to npage pages.
	deductSweepCredit(npages*_base.PageSize, npages)

	s := mHeap_Alloc(&_base.Mheap_, npages, 0, true, flag&_base.FlagNoZero == 0)
	if s == nil {
		_base.Throw("out of memory")
	}
	s.Limit = uintptr(s.Start)<<_base.PageShift + size
	_gc.HeapBitsForSpan(s.Base()).InitSpan(s.Layout())
	return s
}

// implementation of new builtin
func Newobject(typ *_base.Type) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&KindNoPointers != 0 {
		flags |= _base.XFlagNoScan
	}
	return Mallocgc(uintptr(typ.Size), typ, flags)
}

func profilealloc(mp *_base.M, x unsafe.Pointer, size uintptr) {
	c := mp.Mcache
	rate := _base.MemProfileRate
	if size < uintptr(rate) {
		// pick next profile time
		// If you change this, also change allocmcache.
		if rate > 0x3fffffff { // make 2*rate not overflow
			rate = 0x3fffffff
		}
		next := int32(_base.Fastrand1()) % (2 * int32(rate))
		// Subtract the "remainder" of the current allocation.
		// Otherwise objects that are close in size to sampling rate
		// will be under-sampled, because we consistently discard this remainder.
		next -= (int32(size) - c.Next_sample)
		if next < 0 {
			next = 0
		}
		c.Next_sample = next
	}

	mProf_Malloc(x, size)
}
