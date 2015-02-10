// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// base address for all 0-byte allocations
var Zerobase uintptr

// Trigger the concurrent GC when 1/triggerratio memory is available to allocate.
// Adjust this ratio as part of a scheme to ensure that mutators have enough
// memory to allocate in durring a concurrent GC cycle.
var triggerratio = int64(8)

// Determine whether to initiate a GC.
// If the GC is already working no need to trigger another one.
// This should establish a feedback loop where if the GC does not
// have sufficient time to complete then more memory will be
// requested from the OS increasing heap size thus allow future
// GCs more time to complete.
// memstat.heap_alloc and memstat.next_gc reads have benign races
// A false negative simple does not start a GC, a false positive
// will start a GC needlessly. Neither have correctness issues.
func shouldtriggergc() bool {
	return triggerratio*(int64(_lock.Memstats.Next_gc)-int64(_lock.Memstats.Heap_alloc)) <= int64(_lock.Memstats.Next_gc) && _channels.Atomicloaduint(&_gc.Bggc.Working) == 0
}

// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// Large objects (> 32 kB) are allocated straight from the heap.
func Mallocgc(size uintptr, typ *_core.Type, flags uint32) unsafe.Pointer {
	shouldhelpgc := false
	if size == 0 {
		return unsafe.Pointer(&Zerobase)
	}
	dataSize := size

	if flags&_sched.XFlagNoScan == 0 && typ == nil {
		_lock.Throw("malloc missing type")
	}

	// Set mp.mallocing to keep from being preempted by GC.
	mp := _sched.Acquirem()
	if mp.Mallocing != 0 {
		_lock.Throw("malloc deadlock")
	}
	mp.Mallocing = 1

	c := _sem.Gomcache()
	var s *_core.Mspan
	var x unsafe.Pointer
	if size <= _sched.MaxSmallSize {
		if flags&_sched.XFlagNoScan != 0 && size < _sched.MaxTinySize {
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
				off = _lock.Round(off, 8)
			} else if size&3 == 0 {
				off = _lock.Round(off, 4)
			} else if size&1 == 0 {
				off = _lock.Round(off, 2)
			}
			if off+size <= _sched.MaxTinySize && c.Tiny != nil {
				// The object fits into existing tiny block.
				x = _core.Add(c.Tiny, off)
				c.Tinyoffset = off + size
				c.Local_tinyallocs++
				mp.Mallocing = 0
				_sched.Releasem(mp)
				return x
			}
			// Allocate a new maxTinySize block.
			s = c.Alloc[_sched.TinySizeClass]
			v := s.Freelist
			if v.Ptr() == nil {
				_lock.Systemstack(func() {
					mCache_Refill(c, _sched.TinySizeClass)
				})
				shouldhelpgc = true
				s = c.Alloc[_sched.TinySizeClass]
				v = s.Freelist
			}
			s.Freelist = v.Ptr().Next
			s.Ref++
			//TODO: prefetch v.next
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if size < c.Tinyoffset {
				c.Tiny = x
				c.Tinyoffset = size
			}
			size = _sched.MaxTinySize
		} else {
			var sizeclass int8
			if size <= 1024-8 {
				sizeclass = Size_to_class8[(size+7)>>3]
			} else {
				sizeclass = Size_to_class128[(size-1024+127)>>7]
			}
			size = uintptr(_gc.Class_to_size[sizeclass])
			s = c.Alloc[sizeclass]
			v := s.Freelist
			if v.Ptr() == nil {
				_lock.Systemstack(func() {
					mCache_Refill(c, int32(sizeclass))
				})
				shouldhelpgc = true
				s = c.Alloc[sizeclass]
				v = s.Freelist
			}
			s.Freelist = v.Ptr().Next
			s.Ref++
			//TODO: prefetch
			x = unsafe.Pointer(v)
			if flags&_sched.XFlagNoZero == 0 {
				v.Ptr().Next = 0
				if size > 2*_core.PtrSize && ((*[2]uintptr)(x))[1] != 0 {
					_core.Memclr(unsafe.Pointer(v), size)
				}
			}
		}
		c.Local_cachealloc += _core.Intptr(size)
	} else {
		var s *_core.Mspan
		shouldhelpgc = true
		_lock.Systemstack(func() {
			s = largeAlloc(size, uint32(flags))
		})
		x = unsafe.Pointer(uintptr(s.Start << _sched.PageShift))
		size = uintptr(s.Elemsize)
	}

	if flags&_sched.XFlagNoScan != 0 {
		// All objects are pre-marked as noscan. Nothing to do.
	} else {
		// If allocating a defer+arg block, now that we've picked a malloc size
		// large enough to hold everything, cut the "asked for" size down to
		// just the defer header, so that the GC bitmap will record the arg block
		// as containing nothing at all (as if it were unused space at the end of
		// a malloc block caused by size rounding).
		// The defer arg areas are scanned as part of scanstack.
		if typ == DeferType {
			dataSize = unsafe.Sizeof(_core.Defer{})
		}
		heapBitsSetType(uintptr(x), size, dataSize, typ)
	}

	// GCmarkterminate allocates black
	// All slots hold nil so no scanning is needed.
	// This may be racing with GC so do it atomically if there can be
	// a race marking the bit.
	if _sched.Gcphase == _sched.GCmarktermination {
		_lock.Systemstack(func() {
			gcmarknewobject_m(uintptr(x))
		})
	}

	if _lock.Mheap_.Shadow_enabled {
		Clearshadow(uintptr(x), size)
	}

	if _sched.Raceenabled {
		_sched.Racemalloc(x, size)
	}

	mp.Mallocing = 0
	_sched.Releasem(mp)

	if _lock.Debug.Allocfreetrace != 0 {
		tracealloc(x, size, typ)
	}

	if rate := _lock.MemProfileRate; rate > 0 {
		if size < uintptr(rate) && int32(size) < c.Next_sample {
			c.Next_sample -= int32(size)
		} else {
			mp := _sched.Acquirem()
			profilealloc(mp, x, size)
			_sched.Releasem(mp)
		}
	}

	if shouldtriggergc() {
		_gc.Gogc(0)
	} else if shouldhelpgc && _channels.Atomicloaduint(&_gc.Bggc.Working) == 1 {
		// bggc.lock not taken since race on bggc.working is benign.
		// At worse we don't call gchelpwork.
		// Delay the gchelpwork until the epilogue so that it doesn't
		// interfere with the inner working of malloc such as
		// mcache refills that might happen while doing the gchelpwork
		_lock.Systemstack(gchelpwork)
	}

	return x
}

// implementation of new builtin
func Newobject(typ *_core.Type) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&_channels.KindNoPointers != 0 {
		flags |= _sched.XFlagNoScan
	}
	return Mallocgc(uintptr(typ.Size), typ, flags)
}

// implementation of make builtin for slices
func Newarray(typ *_core.Type, n uintptr) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&_channels.KindNoPointers != 0 {
		flags |= _sched.XFlagNoScan
	}
	if int(n) < 0 || (typ.Size > 0 && n > _core.MaxMem/uintptr(typ.Size)) {
		panic("runtime: allocation size out of range")
	}
	return Mallocgc(uintptr(typ.Size)*n, typ, flags)
}

func profilealloc(mp *_core.M, x unsafe.Pointer, size uintptr) {
	c := mp.Mcache
	rate := _lock.MemProfileRate
	if size < uintptr(rate) {
		// pick next profile time
		// If you change this, also change allocmcache.
		if rate > 0x3fffffff { // make 2*rate not overflow
			rate = 0x3fffffff
		}
		next := int32(_lock.Fastrand1()) % (2 * int32(rate))
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
