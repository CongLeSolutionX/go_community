// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

func Pcdatavalue(f *_base.Func, table int32, targetpc uintptr) int32 {
	if table < 0 || table >= f.Npcdata {
		return -1
	}
	off := *(*int32)(_base.Add(unsafe.Pointer(&f.Nfuncdata), unsafe.Sizeof(f.Nfuncdata)+uintptr(table)*4))
	return _base.Pcvalue(f, off, targetpc, true)
}

func Funcdata(f *_base.Func, i int32) unsafe.Pointer {
	if i < 0 || i >= f.Nfuncdata {
		return nil
	}
	p := _base.Add(unsafe.Pointer(&f.Nfuncdata), unsafe.Sizeof(f.Nfuncdata)+uintptr(f.Npcdata)*4)
	if _base.PtrSize == 8 && uintptr(p)&4 != 0 {
		if uintptr(unsafe.Pointer(f))&4 != 0 {
			println("runtime: misaligned func", f)
		}
		p = _base.Add(p, 4)
	}
	return *(*unsafe.Pointer)(_base.Add(p, uintptr(i)*_base.PtrSize))
}

type Stackmap struct {
	N        int32   // number of bitmaps
	nbit     int32   // number of bits in each bitmap
	bytedata [1]byte // bitmaps, each starting on a 32-bit boundary
}

//go:nowritebarrier
func Stackmapdata(stkmap *Stackmap, n int32) _base.Bitvector {
	if n < 0 || n >= stkmap.N {
		_base.Throw("stackmapdata: index out of range")
	}
	return _base.Bitvector{stkmap.nbit, (*byte)(_base.Add(unsafe.Pointer(&stkmap.bytedata), uintptr(n*((stkmap.nbit+31)/32*4))))}
}
