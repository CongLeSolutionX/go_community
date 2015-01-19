// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stackwb

import (
	_core "runtime/internal/core"
	_hash "runtime/internal/hash"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_prof "runtime/internal/prof"
	_sched "runtime/internal/sched"
	"unsafe"
)

func loadPtrMask(typ *_core.Type) []uint8 {
	var ptrmask *uint8
	nptr := (uintptr(typ.Size) + _core.PtrSize - 1) / _core.PtrSize
	if typ.Kind&_hash.KindGCProg != 0 {
		masksize := nptr
		if masksize%2 != 0 {
			masksize *= 2 // repeated
		}
		masksize = masksize * _sched.XPointersPerByte / 8 // 4 bits per word
		masksize++                                        // unroll flag in the beginning
		if masksize > _sched.XMaxGCMask && typ.Gc[1] != 0 {
			// write barriers have not been updated to deal with this case yet.
			_lock.Throw("maxGCMask too small for now")
		}
		ptrmask = (*uint8)(unsafe.Pointer(uintptr(typ.Gc[0])))
		// Check whether the program is already unrolled
		// by checking if the unroll flag byte is set
		maskword := uintptr(_prof.Atomicloadp(unsafe.Pointer(ptrmask)))
		if *(*uint8)(unsafe.Pointer(&maskword)) == 0 {
			_lock.Systemstack(func() {
				_maps.Unrollgcprog_m(typ)
			})
		}
		ptrmask = (*uint8)(_core.Add(unsafe.Pointer(ptrmask), 1)) // skip the unroll flag byte
	} else {
		ptrmask = (*uint8)(unsafe.Pointer(typ.Gc[0])) // pointer to unrolled mask
	}
	return (*[1 << 30]byte)(unsafe.Pointer(ptrmask))[:(nptr+1)/2]
}
