// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/sys"
	"unsafe"
)

// Don't split the stack as this method may be invoked without a valid G, which
// prevents us from allocating more stack.
//go:nosplit
func sysAlloc(n uintptr, sysStat *uint64) unsafe.Pointer {
	p, err := mmap(nil, n, _PROT_READ|_PROT_WRITE, _MAP_ANONYMOUS|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		if err == _EACCES {
			print("runtime: mmap: access denied\n")
			exit(2)
		}
		if err == _EAGAIN {
			print("runtime: mmap: too much locked memory (check 'ulimit -l').\n")
			exit(2)
		}
		println("sysAlloc failed: ", err)
		return nil
	}
	// mSysStatInc(sysStat, n)
	return p
}

func sysUnused(v unsafe.Pointer, n uintptr) {
	madvise(v, n, _MADV_DONTNEED)
}

func sysUsed(v unsafe.Pointer, n uintptr) {
}

// Don't split the stack as this function may be invoked without a valid G,
// which prevents us from allocating more stack.
//go:nosplit
func sysFree(v unsafe.Pointer, n uintptr, sysStat *uint64) {
	// mSysStatDec(sysStat, n)
	munmap(v, n)

}

func sysFault(v unsafe.Pointer, n uintptr) {
	mmap(v, n, _PROT_NONE, _MAP_ANONYMOUS|_MAP_PRIVATE|_MAP_FIXED, -1, 0)
}

func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	// On 64-bit, people with ulimit -v set complain if we reserve too
	// much address space. Instead, assume that the reservation is okay
	// if we can reserve at least 64K and check the assumption in SysMap.
	// Only user-mode Linux (UML) rejects these requests.
	if sys.PtrSize == 8 && uint64(n) > 1<<32 {
		p, err := mmap(v, 64<<10, _PROT_NONE, _MAP_ANONYMOUS|_MAP_PRIVATE|_MAP_FIXED, -1, 0)
		if p != v || err != 0 {
			println("sysAlloc failed: ", err)
			if err == 0 {
				munmap(p, 64<<10)
			}
			return nil
		}
		munmap(p, 64<<10)
		*reserved = false
		return v
	}
	p, err := mmap(v, n, _PROT_NONE, _MAP_ANONYMOUS|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		println("sysAlloc failed: ", err)
		return nil
	}
	*reserved = true
	return p
}

func sysMap(v unsafe.Pointer, n uintptr, reserved bool, sysStat *uint64) {
	// mSysStatInc(sysStat, n)

	// On 64-bit, we don't actually have v reserved, so tread carefully.
	if !reserved {
		p, err := mmap(v, n, _PROT_READ|_PROT_WRITE, _MAP_ANONYMOUS|_MAP_PRIVATE|_MAP_FIXED, -1, 0)
		if err == _ENOMEM {
			throw("runtime: out of memory")
		}
		if p != v || err != 0 {
			print("runtime: address space conflict: map(", v, ") = ", p, " (err ", err, ")\n")
			throw("runtime: address space conflict")
		}
		return
	}

	// AIX does not allow mapping a range that is already mapped.
	// So always unmap first even if it is already unmapped.
	munmap(v, n)
	p, err := mmap(v, n, _PROT_READ|_PROT_WRITE, _MAP_ANONYMOUS|_MAP_FIXED|_MAP_PRIVATE, -1, 0)

	if err == _ENOMEM {
		throw("runtime: out of memory")
	}
	if p != v || err != 0 {
		throw("runtime: cannot map pages in arena address space")
	}
}
