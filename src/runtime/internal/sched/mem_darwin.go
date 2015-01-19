// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func sysUsed(v unsafe.Pointer, n uintptr) {
}

func SysFree(v unsafe.Pointer, n uintptr, stat *uint64) {
	_lock.Xadd64(stat, -int64(n))
	munmap(v, n)
}

func SysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	*reserved = true
	p := (unsafe.Pointer)(_lock.Mmap(v, n, _lock.PROT_NONE, _lock.MAP_ANON|_lock.MAP_PRIVATE, -1, 0))
	if uintptr(p) < 4096 {
		return nil
	}
	return p
}

const (
	_ENOMEM = 12
)

func sysMap(v unsafe.Pointer, n uintptr, reserved bool, stat *uint64) {
	_lock.Xadd64(stat, int64(n))
	p := (unsafe.Pointer)(_lock.Mmap(v, n, _lock.PROT_READ|_lock.PROT_WRITE, _lock.MAP_ANON|_lock.MAP_FIXED|_lock.MAP_PRIVATE, -1, 0))
	if uintptr(p) == _ENOMEM {
		_lock.Gothrow("runtime: out of memory")
	}
	if p != v {
		_lock.Gothrow("runtime: cannot map pages in arena address space")
	}
}
