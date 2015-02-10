// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

//go:nosplit
func casp(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	if !casp1((*unsafe.Pointer)(_core.Noescape(unsafe.Pointer(ptr))), _core.Noescape(old), new) {
		return false
	}
	_sched.Writebarrierptr_nostore((*uintptr)(unsafe.Pointer(ptr)), uintptr(new))
	if _lock.Mheap_.Shadow_enabled {
		_sched.Writebarrierptr_noshadow((*uintptr)(_core.Noescape(unsafe.Pointer(ptr))))
	}
	return true
}
