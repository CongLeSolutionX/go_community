// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_hash "runtime/internal/hash"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// base address for all 0-byte allocations
var Zerobase uintptr

// Allocate an object of size bytes.
// Small objects are allocated from the per-P cache's free lists.
// Large objects (> 32 kB) are allocated straight from the heap.
func Mallocgc(size uintptr, typ *_core.Type, flags uint32) unsafe.Pointer {
	if size == 0 {
		return unsafe.Pointer(&Zerobase)
	}
	size0 := size

	if flags&_sched.XFlagNoScan == 0 && typ == nil {
		_lock.Throw("malloc missing type")
	}

	// This function must be atomic wrt GC, but for performance reasons
	// we don't acquirem/releasem on fast path. The code below does not have
	// split stack checks, so it can't be preempted by GC.
	// Functions like roundup/add are inlined. And systemstack/racemalloc are nosplit.
	// If debugMalloc = true, these assumptions are checked below.
	if _sched.DebugMalloc {
		mp := _sched.Acquirem()
		if mp.Mallocing != 0 {
			_lock.Throw("malloc deadlock")
		}
		mp.Mallocing = 1
		if mp.Curg != nil {
			mp.Curg.Stackguard0 = ^uintptr(0xfff) | 0xbad
		}
	}

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
			tinysize := uintptr(c.Tinysize)
			if size <= tinysize {
				tiny := unsafe.Pointer(c.Tiny)
				// Align tiny pointer for required (conservative) alignment.
				if size&7 == 0 {
					tiny = _lock.Roundup(tiny, 8)
				} else if size&3 == 0 {
					tiny = _lock.Roundup(tiny, 4)
				} else if size&1 == 0 {
					tiny = _lock.Roundup(tiny, 2)
				}
				size1 := size + (uintptr(tiny) - uintptr(unsafe.Pointer(c.Tiny)))
				if size1 <= tinysize {
					// The object fits into existing tiny block.
					x = tiny
					c.Tiny = (*byte)(_core.Add(x, size))
					c.Tinysize -= uintptr(size1)
					c.Local_tinyallocs++
					if _sched.DebugMalloc {
						mp := _sched.Acquirem()
						if mp.Mallocing == 0 {
							_lock.Throw("bad malloc")
						}
						mp.Mallocing = 0
						if mp.Curg != nil {
							mp.Curg.Stackguard0 = mp.Curg.Stack.Lo + _core.StackGuard
						}
						// Note: one releasem for the acquirem just above.
						// The other for the acquirem at start of malloc.
						_sched.Releasem(mp)
						_sched.Releasem(mp)
					}
					return x
				}
			}
			// Allocate a new maxTinySize block.
			s = c.Alloc[_sched.TinySizeClass]
			v := s.Freelist
			if v.Ptr() == nil {
				_lock.Systemstack(func() {
					mCache_Refill(c, _sched.TinySizeClass)
				})
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
			if _sched.MaxTinySize-size > tinysize {
				c.Tiny = (*byte)(_core.Add(x, size))
				c.Tinysize = uintptr(_sched.MaxTinySize - size)
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
		_lock.Systemstack(func() {
			s = largeAlloc(size, uint32(flags))
		})
		x = unsafe.Pointer(uintptr(s.Start << _sched.PageShift))
		size = uintptr(s.Elemsize)
	}

	if flags&_sched.XFlagNoScan != 0 {
		// All objects are pre-marked as noscan.
		goto marked
	}

	// If allocating a defer+arg block, now that we've picked a malloc size
	// large enough to hold everything, cut the "asked for" size down to
	// just the defer header, so that the GC bitmap will record the arg block
	// as containing nothing at all (as if it were unused space at the end of
	// a malloc block caused by size rounding).
	// The defer arg areas are scanned as part of scanstack.
	if typ == DeferType {
		size0 = unsafe.Sizeof(_core.Defer{})
	}

	// From here till marked label marking the object as allocated
	// and storing type info in the GC bitmap.
	{
		arena_start := uintptr(unsafe.Pointer(_lock.Mheap_.Arena_start))
		off := (uintptr(x) - arena_start) / _core.PtrSize
		xbits := (*uint8)(unsafe.Pointer(arena_start - off/_sched.WordsPerBitmapByte - 1))
		shift := (off % _sched.WordsPerBitmapByte) * _sched.GcBits
		if _sched.DebugMalloc && ((*xbits>>shift)&(_sched.BitMask|_sched.BitPtrMask)) != _sched.BitBoundary {
			println("runtime: bits =", (*xbits>>shift)&(_sched.BitMask|_sched.BitPtrMask))
			_lock.Throw("bad bits in markallocated")
		}

		var ti, te uintptr
		var ptrmask *uint8
		if size == _core.PtrSize {
			// It's one word and it has pointers, it must be a pointer.
			*xbits |= (_sched.XBitsPointer << 2) << shift
			goto marked
		}
		if typ.Kind&_hash.KindGCProg != 0 {
			nptr := (uintptr(typ.Size) + _core.PtrSize - 1) / _core.PtrSize
			masksize := nptr
			if masksize%2 != 0 {
				masksize *= 2 // repeated
			}
			masksize = masksize * _sched.XPointersPerByte / 8 // 4 bits per word
			masksize++                                        // unroll flag in the beginning
			if masksize > _sched.XMaxGCMask && typ.Gc[1] != 0 {
				// write barriers have not been updated to deal with this case yet.
				_lock.Throw("maxGCMask too small for now")
				// If the mask is too large, unroll the program directly
				// into the GC bitmap. It's 7 times slower than copying
				// from the pre-unrolled mask, but saves 1/16 of type size
				// memory for the mask.
				_lock.Systemstack(func() {
					unrollgcproginplace_m(x, typ, size, size0)
				})
				goto marked
			}
			ptrmask = (*uint8)(unsafe.Pointer(uintptr(typ.Gc[0])))
			// Check whether the program is already unrolled
			// by checking if the unroll flag byte is set
			maskword := uintptr(_prof.Atomicloadp(unsafe.Pointer(ptrmask)))
			if *(*uint8)(unsafe.Pointer(&maskword)) == 0 {
				_lock.Systemstack(func() {
					Unrollgcprog_m(typ)
				})
			}
			ptrmask = (*uint8)(_core.Add(unsafe.Pointer(ptrmask), 1)) // skip the unroll flag byte
		} else {
			ptrmask = (*uint8)(unsafe.Pointer(typ.Gc[0])) // pointer to unrolled mask
		}
		if size == 2*_core.PtrSize {
			*xbits = *ptrmask | _sched.BitBoundary
			goto marked
		}
		te = uintptr(typ.Size) / _core.PtrSize
		// If the type occupies odd number of words, its mask is repeated.
		if te%2 == 0 {
			te /= 2
		}
		// Copy pointer bitmask into the bitmap.
		for i := uintptr(0); i < size0; i += 2 * _core.PtrSize {
			v := *(*uint8)(_core.Add(unsafe.Pointer(ptrmask), ti))
			ti++
			if ti == te {
				ti = 0
			}
			if i == 0 {
				v |= _sched.BitBoundary
			}
			if i+_core.PtrSize == size0 {
				v &^= uint8(_sched.BitPtrMask << 4)
			}

			*xbits = v
			xbits = (*byte)(_core.Add(unsafe.Pointer(xbits), ^uintptr(0)))
		}
		if size0%(2*_core.PtrSize) == 0 && size0 < size {
			// Mark the word after last object's word as bitsDead.
			*xbits = _sched.XBitsDead << 2
		}
	}
marked:

	// GCmarkterminate allocates black
	// All slots hold nil so no scanning is needed.
	// This may be racing with GC so do it atomically if there can be
	// a race marking the bit.
	if _sched.Gcphase == _sched.GCmarktermination {
		_lock.Systemstack(func() {
			gcmarknewobject_m(uintptr(x))
		})
	}

	if _sched.Raceenabled {
		_sched.Racemalloc(x, size)
	}

	if _sched.DebugMalloc {
		mp := _sched.Acquirem()
		if mp.Mallocing == 0 {
			_lock.Throw("bad malloc")
		}
		mp.Mallocing = 0
		if mp.Curg != nil {
			mp.Curg.Stackguard0 = mp.Curg.Stack.Lo + _core.StackGuard
		}
		// Note: one releasem for the acquirem just above.
		// The other for the acquirem at start of malloc.
		_sched.Releasem(mp)
		_sched.Releasem(mp)
	}

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

	if _lock.Memstats.Heap_alloc >= _lock.Memstats.Next_gc/2 {
		_gc.Gogc(0)
	}

	return x
}

func Newobject(typ *_core.Type) unsafe.Pointer {
	return newobject(typ)
}

// implementation of new builtin
func newobject(typ *_core.Type) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&_hash.KindNoPointers != 0 {
		flags |= _sched.XFlagNoScan
	}
	return Mallocgc(uintptr(typ.Size), typ, flags)
}

// implementation of make builtin for slices
func Newarray(typ *_core.Type, n uintptr) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&_hash.KindNoPointers != 0 {
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
