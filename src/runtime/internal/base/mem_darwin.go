// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//go:nosplit
func SysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	v := (unsafe.Pointer)(Mmap(nil, n, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE, -1, 0))
	if uintptr(v) < 4096 {
		return nil
	}
	mSysStatInc(sysStat, n)
	return v
}

func sysUsed(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//go:nosplit
func SysFree(v unsafe.Pointer, n uintptr, sysStat *uint64) {
	mSysStatDec(sysStat, n)
	munmap(v, n)
}

func SysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	*reserved = true
	p := (unsafe.Pointer)(Mmap(v, n, PROT_NONE, MAP_ANON|MAP_PRIVATE, -1, 0))
	if uintptr(p) < 4096 {
		return nil
	}
	return p
}

const (
	_ENOMEM = 12
)

func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {
	mSysStatInc(sysStat, n)
	p := (unsafe.Pointer)(Mmap(v, n, PROT_READ|PROT_WRITE, MAP_ANON|MAP_FIXED|MAP_PRIVATE, -1, 0))
	if uintptr(p) == _ENOMEM {
		Throw("runtime: out of memory")
	}
	if p != v {
		Throw("runtime: cannot map pages in arena address space")
	}
}
