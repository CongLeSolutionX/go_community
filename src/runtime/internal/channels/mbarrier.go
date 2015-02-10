// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: write barriers.
//
// For the concurrent garbage collector, the Go compiler implements
// updates to pointer-valued fields that may be in heap objects by
// emitting calls to write barriers. This file contains the actual write barrier
// implementation, markwb, and the various wrappers called by the
// compiler to implement pointer assignment, slice assignment,
// typed memmove, and so on.
//
// To check for missed write barriers, the GODEBUG=wbshadow debugging
// mode allocates a second copy of the heap. Write barrier-based pointer
// updates make changes to both the real heap and the shadow, and both
// the pointer updates and the GC look for inconsistencies between the two,
// indicating pointer writes that bypassed the barrier.

package channels

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// NOTE: Really dst *unsafe.Pointer, src unsafe.Pointer,
// but if we do that, Go inserts a write barrier on *dst = src.
//go:nosplit
func Writebarrierptr(dst *uintptr, src uintptr) {
	if !_sched.Needwb() {
		*dst = src
		return
	}

	if src != 0 && (src < _lock.PhysPageSize || src == _lock.PoisonStack) {
		_lock.Systemstack(func() { _lock.Throw("bad pointer in write barrier") })
	}

	if _lock.Mheap_.Shadow_enabled {
		writebarrierptr_shadow(dst, src)
	}

	*dst = src
	_sched.Writebarrierptr_nostore1(dst, src)
}

//go:nosplit
func writebarrierptr_shadow(dst *uintptr, src uintptr) {
	_lock.Systemstack(func() {
		addr := uintptr(unsafe.Pointer(dst))
		shadow := _sched.Shadowptr(addr)
		if shadow == nil {
			return
		}
		// There is a race here but only if the program is using
		// racy writes instead of sync/atomic. In that case we
		// don't mind crashing.
		if *shadow != *dst && *shadow != _sched.NoShadow && _sched.Istrackedptr(*dst) {
			_lock.Mheap_.Shadow_enabled = false
			print("runtime: write barrier dst=", dst, " old=", _core.Hex(*dst), " shadow=", shadow, " old=", _core.Hex(*shadow), " new=", _core.Hex(src), "\n")
			_lock.Throw("missed write barrier")
		}
		*shadow = src
	})
}

//go:generate go run wbfat_gen.go -- wbfat.go
//
// The above line generates multiword write barriers for
// all the combinations of ptr+scalar up to four words.
// The implementations are written to wbfat.go.

// typedmemmove copies a value of type t to dst from src.
//go:nosplit
func Typedmemmove(typ *_core.Type, dst, src unsafe.Pointer) {
	if !_sched.Needwb() || (typ.Kind&KindNoPointers) != 0 {
		_sched.Memmove(dst, src, typ.Size)
		return
	}

	_lock.Systemstack(func() {
		mask := TypeBitmapInHeapBitmapFormat(typ)
		nptr := typ.Size / _core.PtrSize
		for i := uintptr(0); i < nptr; i += 2 {
			bits := mask[i/2]
			if (bits>>2)&_sched.TypeMask == _sched.TypePointer {
				Writebarrierptr((*uintptr)(dst), *(*uintptr)(src))
			} else {
				*(*uintptr)(dst) = *(*uintptr)(src)
			}
			// TODO(rsc): The noescape calls should be unnecessary.
			dst = _core.Add(_core.Noescape(dst), _core.PtrSize)
			src = _core.Add(_core.Noescape(src), _core.PtrSize)
			if i+1 == nptr {
				break
			}
			bits >>= 4
			if (bits>>2)&_sched.TypeMask == _sched.TypePointer {
				Writebarrierptr((*uintptr)(dst), *(*uintptr)(src))
			} else {
				*(*uintptr)(dst) = *(*uintptr)(src)
			}
			dst = _core.Add(_core.Noescape(dst), _core.PtrSize)
			src = _core.Add(_core.Noescape(src), _core.PtrSize)
		}
	})
}
