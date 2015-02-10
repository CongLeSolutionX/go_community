// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

//go:nosplit
func xchgp(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	old := xchgp1(_core.Noescape(ptr), new)
	_sched.Writebarrierptr_nostore((*uintptr)(ptr), uintptr(new))
	if _lock.Mheap_.Shadow_enabled {
		_sched.Writebarrierptr_noshadow((*uintptr)(_core.Noescape(ptr)))
	}
	return old
}

// Like above, but implement in terms of sync/atomic's uintptr operations.
// We cannot just call the runtime routines, because the race detector expects
// to be able to intercept the sync/atomic forms but not the runtime forms.

//go:linkname sync_atomic_StoreUintptr sync/atomic.StoreUintptr
func sync_atomic_StoreUintptr(ptr *uintptr, new uintptr)

//go:linkname sync_atomic_StorePointer sync/atomic.StorePointer
//go:nosplit
func sync_atomic_StorePointer(ptr *unsafe.Pointer, new unsafe.Pointer) {
	sync_atomic_StoreUintptr((*uintptr)(unsafe.Pointer(ptr)), uintptr(new))
	_sched.Atomicstorep1(_core.Noescape(unsafe.Pointer(ptr)), new)
	_sched.Writebarrierptr_nostore((*uintptr)(unsafe.Pointer(ptr)), uintptr(new))
	if _lock.Mheap_.Shadow_enabled {
		_sched.Writebarrierptr_noshadow((*uintptr)(_core.Noescape(unsafe.Pointer(ptr))))
	}
}

//go:linkname sync_atomic_SwapUintptr sync/atomic.SwapUintptr
func sync_atomic_SwapUintptr(ptr *uintptr, new uintptr) uintptr

//go:linkname sync_atomic_SwapPointer sync/atomic.SwapPointer
//go:nosplit
func sync_atomic_SwapPointer(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer {
	old := unsafe.Pointer(sync_atomic_SwapUintptr((*uintptr)(_core.Noescape(ptr)), uintptr(new)))
	_sched.Writebarrierptr_nostore((*uintptr)(ptr), uintptr(new))
	if _lock.Mheap_.Shadow_enabled {
		_sched.Writebarrierptr_noshadow((*uintptr)(_core.Noescape(ptr)))
	}
	return old
}

//go:linkname sync_atomic_CompareAndSwapUintptr sync/atomic.CompareAndSwapUintptr
func sync_atomic_CompareAndSwapUintptr(ptr *uintptr, old, new uintptr) bool

//go:linkname sync_atomic_CompareAndSwapPointer sync/atomic.CompareAndSwapPointer
//go:nosplit
func sync_atomic_CompareAndSwapPointer(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	if !sync_atomic_CompareAndSwapUintptr((*uintptr)(_core.Noescape(unsafe.Pointer(ptr))), uintptr(old), uintptr(new)) {
		return false
	}
	_sched.Writebarrierptr_nostore((*uintptr)(unsafe.Pointer(ptr)), uintptr(new))
	if _lock.Mheap_.Shadow_enabled {
		_sched.Writebarrierptr_noshadow((*uintptr)(_core.Noescape(unsafe.Pointer(ptr))))
	}
	return true
}
