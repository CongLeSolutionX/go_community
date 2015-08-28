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

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

//go:linkname reflect_typedmemmove reflect.typedmemmove
func reflect_typedmemmove(typ *_base.Type, dst, src unsafe.Pointer) {
	_iface.Typedmemmove(typ, dst, src)
}

// typedmemmovepartial is like typedmemmove but assumes that
// dst and src point off bytes into the value and only copies size bytes.
//go:linkname reflect_typedmemmovepartial reflect.typedmemmovepartial
func reflect_typedmemmovepartial(typ *_base.Type, dst, src unsafe.Pointer, off, size uintptr) {
	_base.Memmove(dst, src, size)
	if !_base.WriteBarrierEnabled || typ.Kind&_iface.KindNoPointers != 0 || size < _base.PtrSize || !_base.Inheap(uintptr(dst)) {
		return
	}

	if frag := -off & (_base.PtrSize - 1); frag != 0 {
		dst = _base.Add(dst, frag)
		size -= frag
	}
	_iface.HeapBitsBulkBarrier(uintptr(dst), size&^(_base.PtrSize-1))
}

// callwritebarrier is invoked at the end of reflectcall, to execute
// write barrier operations to record the fact that a call's return
// values have just been copied to frame, starting at retoffset
// and continuing to framesize. The entire frame (not just the return
// values) is described by typ. Because the copy has already
// happened, we call writebarrierptr_nostore, and we must be careful
// not to be preempted before the write barriers have been run.
//go:nosplit
func callwritebarrier(typ *_base.Type, frame unsafe.Pointer, framesize, retoffset uintptr) {
	if !_base.WriteBarrierEnabled || typ == nil || typ.Kind&_iface.KindNoPointers != 0 || framesize-retoffset < _base.PtrSize || !_base.Inheap(uintptr(frame)) {
		return
	}
	_iface.HeapBitsBulkBarrier(uintptr(_base.Add(frame, retoffset)), framesize-retoffset)
}

//go:nosplit
func typedslicecopy(typ *_base.Type, dst, src _base.Slice) int {
	// TODO(rsc): If typedslicecopy becomes faster than calling
	// typedmemmove repeatedly, consider using during func growslice.
	n := dst.Len
	if n > src.Len {
		n = src.Len
	}
	if n == 0 {
		return 0
	}
	dstp := unsafe.Pointer(dst.Array)
	srcp := unsafe.Pointer(src.Array)

	if _base.Raceenabled {
		callerpc := _base.Getcallerpc(unsafe.Pointer(&typ))
		pc := _base.FuncPC(slicecopy)
		_race.Racewriterangepc(dstp, uintptr(n)*typ.Size, callerpc, pc)
		_race.Racereadrangepc(srcp, uintptr(n)*typ.Size, callerpc, pc)
	}

	// Note: No point in checking typ.kind&kindNoPointers here:
	// compiler only emits calls to typedslicecopy for types with pointers,
	// and growslice and reflect_typedslicecopy check for pointers
	// before calling typedslicecopy.
	if !_base.WriteBarrierEnabled {
		_base.Memmove(dstp, srcp, uintptr(n)*typ.Size)
		return n
	}

	_base.Systemstack(func() {
		if uintptr(srcp) < uintptr(dstp) && uintptr(srcp)+uintptr(n)*typ.Size > uintptr(dstp) {
			// Overlap with src before dst.
			// Copy backward, being careful not to move dstp/srcp
			// out of the array they point into.
			dstp = _base.Add(dstp, uintptr(n-1)*typ.Size)
			srcp = _base.Add(srcp, uintptr(n-1)*typ.Size)
			i := 0
			for {
				_iface.Typedmemmove(typ, dstp, srcp)
				if i++; i >= n {
					break
				}
				dstp = _base.Add(dstp, -typ.Size)
				srcp = _base.Add(srcp, -typ.Size)
			}
		} else {
			// Copy forward, being careful not to move dstp/srcp
			// out of the array they point into.
			i := 0
			for {
				_iface.Typedmemmove(typ, dstp, srcp)
				if i++; i >= n {
					break
				}
				dstp = _base.Add(dstp, typ.Size)
				srcp = _base.Add(srcp, typ.Size)
			}
		}
	})
	return int(n)
}

//go:linkname reflect_typedslicecopy reflect.typedslicecopy
func reflect_typedslicecopy(elemType *_base.Type, dst, src _base.Slice) int {
	if elemType.Kind&_iface.KindNoPointers != 0 {
		n := dst.Len
		if n > src.Len {
			n = src.Len
		}
		_base.Memmove(dst.Array, src.Array, uintptr(n)*elemType.Size)
		return n
	}
	return typedslicecopy(elemType, dst, src)
}
