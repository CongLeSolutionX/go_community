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

package runtime

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

//go:linkname reflect_typedmemmove reflect.typedmemmove
func reflect_typedmemmove(typ *_core.Type, dst, src unsafe.Pointer) {
	_channels.Typedmemmove(typ, dst, src)
}

// typedmemmovepartial is like typedmemmove but assumes that
// dst and src point off bytes into the value and only copies size bytes.
//go:linkname reflect_typedmemmovepartial reflect.typedmemmovepartial
func reflect_typedmemmovepartial(typ *_core.Type, dst, src unsafe.Pointer, off, size uintptr) {
	if !_sched.Needwb() || (typ.Kind&_channels.KindNoPointers) != 0 || size < _core.PtrSize {
		_sched.Memmove(dst, src, size)
		return
	}

	if off&(_core.PtrSize-1) != 0 {
		frag := -off & (_core.PtrSize - 1)
		// frag < size, because size >= ptrSize, checked above.
		_sched.Memmove(dst, src, frag)
		size -= frag
		dst = _core.Add(_core.Noescape(dst), frag)
		src = _core.Add(_core.Noescape(src), frag)
		off += frag
	}

	mask := _channels.TypeBitmapInHeapBitmapFormat(typ)
	nptr := (off + size) / _core.PtrSize
	for i := uintptr(off / _core.PtrSize); i < nptr; i++ {
		bits := mask[i/2] >> ((i & 1) << 2)
		if (bits>>2)&_sched.TypeMask == _sched.TypePointer {
			_channels.Writebarrierptr((*uintptr)(dst), *(*uintptr)(src))
		} else {
			*(*uintptr)(dst) = *(*uintptr)(src)
		}
		// TODO(rsc): The noescape calls should be unnecessary.
		dst = _core.Add(_core.Noescape(dst), _core.PtrSize)
		src = _core.Add(_core.Noescape(src), _core.PtrSize)
	}
	size &= _core.PtrSize - 1
	if size > 0 {
		_sched.Memmove(dst, src, size)
	}
}

// callwritebarrier is invoked at the end of reflectcall, to execute
// write barrier operations to record the fact that a call's return
// values have just been copied to frame, starting at retoffset
// and continuing to framesize. The entire frame (not just the return
// values) is described by typ. Because the copy has already
// happened, we call writebarrierptr_nostore, and we must be careful
// not to be preempted before the write barriers have been run.
//go:nosplit
func callwritebarrier(typ *_core.Type, frame unsafe.Pointer, framesize, retoffset uintptr) {
	if !_sched.Needwb() || typ == nil || (typ.Kind&_channels.KindNoPointers) != 0 || framesize-retoffset < _core.PtrSize {
		return
	}

	_lock.Systemstack(func() {
		mask := _channels.TypeBitmapInHeapBitmapFormat(typ)
		// retoffset is known to be pointer-aligned (at least).
		// TODO(rsc): The noescape call should be unnecessary.
		dst := _core.Add(_core.Noescape(frame), retoffset)
		nptr := framesize / _core.PtrSize
		for i := uintptr(retoffset / _core.PtrSize); i < nptr; i++ {
			bits := mask[i/2] >> ((i & 1) << 2)
			if (bits>>2)&_sched.TypeMask == _sched.TypePointer {
				_sched.Writebarrierptr_nostore((*uintptr)(dst), *(*uintptr)(dst))
			}
			// TODO(rsc): The noescape call should be unnecessary.
			dst = _core.Add(_core.Noescape(dst), _core.PtrSize)
		}
	})
}

//go:nosplit
func typedslicecopy(typ *_core.Type, dst, src _core.Slice) int {
	n := dst.Len
	if n > src.Len {
		n = src.Len
	}
	if n == 0 {
		return 0
	}
	dstp := unsafe.Pointer(dst.Array)
	srcp := unsafe.Pointer(src.Array)

	if !_sched.Needwb() {
		_sched.Memmove(dstp, srcp, uintptr(n)*typ.Size)
		return int(n)
	}

	_lock.Systemstack(func() {
		if uintptr(srcp) < uintptr(dstp) && uintptr(srcp)+uintptr(n)*typ.Size > uintptr(dstp) {
			// Overlap with src before dst.
			// Copy backward, being careful not to move dstp/srcp
			// out of the array they point into.
			dstp = _core.Add(dstp, uintptr(n-1)*typ.Size)
			srcp = _core.Add(srcp, uintptr(n-1)*typ.Size)
			i := uint(0)
			for {
				_channels.Typedmemmove(typ, dstp, srcp)
				if i++; i >= n {
					break
				}
				dstp = _core.Add(dstp, -typ.Size)
				srcp = _core.Add(srcp, -typ.Size)
			}
		} else {
			// Copy forward, being careful not to move dstp/srcp
			// out of the array they point into.
			i := uint(0)
			for {
				_channels.Typedmemmove(typ, dstp, srcp)
				if i++; i >= n {
					break
				}
				dstp = _core.Add(dstp, typ.Size)
				srcp = _core.Add(srcp, typ.Size)
			}
		}
	})
	return int(n)
}

//go:linkname reflect_typedslicecopy reflect.typedslicecopy
func reflect_typedslicecopy(elemType *_core.Type, dst, src _core.Slice) int {
	return typedslicecopy(elemType, dst, src)
}
