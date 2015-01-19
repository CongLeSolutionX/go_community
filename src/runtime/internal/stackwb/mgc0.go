// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stackwb

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

const (
	_PoisonGC    = 0xf969696969696969 & (1<<(8*_core.PtrSize) - 1)
	_PoisonStack = 0x6868686868686868 & (1<<(8*_core.PtrSize) - 1)
)

func needwb() bool {
	return _sched.Gcphase == _sched.GCmark || _sched.Gcphase == _sched.GCmarktermination
}

// NOTE: Really dst *unsafe.Pointer, src unsafe.Pointer,
// but if we do that, Go inserts a write barrier on *dst = src.
//go:nosplit
func writebarrierptr(dst *uintptr, src uintptr) {
	*dst = src
	if needwb() {
		Writebarrierptr_nostore(dst, src)
	}
}

// Like writebarrierptr, but the store has already been applied.
// Do not reapply.
//go:nosplit
func Writebarrierptr_nostore(dst *uintptr, src uintptr) {
	if _core.Getg() == nil || !needwb() { // very low-level startup
		return
	}

	if src != 0 && (src < _core.PageSize || src == _PoisonGC || src == _PoisonStack) {
		_lock.Systemstack(func() { _lock.Gothrow("bad pointer in write barrier") })
	}

	mp := _sched.Acquirem()
	if mp.Inwb || mp.Dying > 0 {
		_sched.Releasem(mp)
		return
	}
	mp.Inwb = true
	_lock.Systemstack(func() {
		gcmarkwb_m(dst, src)
	})
	mp.Inwb = false
	_sched.Releasem(mp)
}

//go:nosplit
func writebarrierstring(dst *[2]uintptr, src [2]uintptr) {
	writebarrierptr(&dst[0], src[0])
	dst[1] = src[1]
}

//go:nosplit
func writebarrierslice(dst *[3]uintptr, src [3]uintptr) {
	writebarrierptr(&dst[0], src[0])
	dst[1] = src[1]
	dst[2] = src[2]
}

//go:nosplit
func writebarrieriface(dst *[2]uintptr, src [2]uintptr) {
	writebarrierptr(&dst[0], src[0])
	writebarrierptr(&dst[1], src[1])
}

//go:generate go run wbfat_gen.go -- wbfat.go
//
// The above line generates multiword write barriers for
// all the combinations of ptr+scalar up to four words.
// The implementations are written to wbfat.go.

//go:nosplit
func writebarrierfat(typ *_core.Type, dst, src unsafe.Pointer) {
	if !needwb() {
		_sched.Memmove(dst, src, typ.Size)
		return
	}

	_lock.Systemstack(func() {
		mask := loadPtrMask(typ)
		nptr := typ.Size / _core.PtrSize
		for i := uintptr(0); i < nptr; i += 2 {
			bits := mask[i/2]
			if (bits>>2)&_sched.BitsMask == _sched.BitsPointer {
				writebarrierptr((*uintptr)(dst), *(*uintptr)(src))
			} else {
				*(*uintptr)(dst) = *(*uintptr)(src)
			}
			dst = _core.Add(dst, _core.PtrSize)
			src = _core.Add(src, _core.PtrSize)
			if i+1 == nptr {
				break
			}
			bits >>= 4
			if (bits>>2)&_sched.BitsMask == _sched.BitsPointer {
				writebarrierptr((*uintptr)(dst), *(*uintptr)(src))
			} else {
				*(*uintptr)(dst) = *(*uintptr)(src)
			}
			dst = _core.Add(dst, _core.PtrSize)
			src = _core.Add(src, _core.PtrSize)
		}
	})
}

//go:nosplit
func writebarriercopy(typ *_core.Type, dst, src _core.Slice) int {
	n := dst.Len
	if n > src.Len {
		n = src.Len
	}
	if n == 0 {
		return 0
	}
	dstp := unsafe.Pointer(dst.Array)
	srcp := unsafe.Pointer(src.Array)

	if !needwb() {
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
				writebarrierfat(typ, dstp, srcp)
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
				writebarrierfat(typ, dstp, srcp)
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
