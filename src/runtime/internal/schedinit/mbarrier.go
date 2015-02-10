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

package schedinit

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func wbshadowinit() {
	// Initialize write barrier shadow heap if we were asked for it
	// and we have enough address space (not on 32-bit).
	if _lock.Debug.Wbshadow == 0 {
		return
	}
	if _core.PtrSize != 8 {
		print("runtime: GODEBUG=wbshadow=1 disabled on 32-bit system\n")
		return
	}

	var reserved bool
	p1 := sysReserveHigh(_lock.Mheap_.Arena_end-_lock.Mheap_.Arena_start, &reserved)
	if p1 == nil {
		_lock.Throw("cannot map shadow heap")
	}
	_lock.Mheap_.Shadow_heap = uintptr(p1) - _lock.Mheap_.Arena_start
	_sched.SysMap(p1, _lock.Mheap_.Arena_used-_lock.Mheap_.Arena_start, reserved, &_lock.Memstats.Other_sys)
	_sched.Memmove(p1, unsafe.Pointer(_lock.Mheap_.Arena_start), _lock.Mheap_.Arena_used-_lock.Mheap_.Arena_start)

	_lock.Mheap_.Shadow_reserved = reserved
	start := ^uintptr(0)
	end := uintptr(0)
	if start > uintptr(unsafe.Pointer(&Noptrdata)) {
		start = uintptr(unsafe.Pointer(&Noptrdata))
	}
	if start > uintptr(unsafe.Pointer(&_gc.Data)) {
		start = uintptr(unsafe.Pointer(&_gc.Data))
	}
	if start > uintptr(unsafe.Pointer(&Noptrbss)) {
		start = uintptr(unsafe.Pointer(&Noptrbss))
	}
	if start > uintptr(unsafe.Pointer(&_gc.Bss)) {
		start = uintptr(unsafe.Pointer(&_gc.Bss))
	}
	if end < uintptr(unsafe.Pointer(&Enoptrdata)) {
		end = uintptr(unsafe.Pointer(&Enoptrdata))
	}
	if end < uintptr(unsafe.Pointer(&_gc.Edata)) {
		end = uintptr(unsafe.Pointer(&_gc.Edata))
	}
	if end < uintptr(unsafe.Pointer(&Enoptrbss)) {
		end = uintptr(unsafe.Pointer(&Enoptrbss))
	}
	if end < uintptr(unsafe.Pointer(&_gc.Ebss)) {
		end = uintptr(unsafe.Pointer(&_gc.Ebss))
	}
	start &^= _lock.PhysPageSize - 1
	end = _lock.Round(end, _lock.PhysPageSize)
	_lock.Mheap_.Data_start = start
	_lock.Mheap_.Data_end = end
	reserved = false
	p1 = sysReserveHigh(end-start, &reserved)
	if p1 == nil {
		_lock.Throw("cannot map shadow data")
	}
	_lock.Mheap_.Shadow_data = uintptr(p1) - start
	_sched.SysMap(p1, end-start, reserved, &_lock.Memstats.Other_sys)
	_sched.Memmove(p1, unsafe.Pointer(start), end-start)

	_lock.Mheap_.Shadow_enabled = true
}
