// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

func Pcdatavalue(f *_lock.Func, table int32, targetpc uintptr) int32 {
	if table < 0 || table >= f.Npcdata {
		return -1
	}
	off := *(*int32)(_core.Add(unsafe.Pointer(&f.Nfuncdata), unsafe.Sizeof(f.Nfuncdata)+uintptr(table)*4))
	return _lock.Pcvalue(f, off, targetpc, true)
}

func Funcdata(f *_lock.Func, i int32) unsafe.Pointer {
	if i < 0 || i >= f.Nfuncdata {
		return nil
	}
	p := _core.Add(unsafe.Pointer(&f.Nfuncdata), unsafe.Sizeof(f.Nfuncdata)+uintptr(f.Npcdata)*4)
	if _core.PtrSize == 8 && uintptr(p)&4 != 0 {
		if uintptr(unsafe.Pointer(f))&4 != 0 {
			println("runtime: misaligned func", f)
		}
		p = _core.Add(p, 4)
	}
	return *(*unsafe.Pointer)(_core.Add(p, uintptr(i)*_core.PtrSize))
}
