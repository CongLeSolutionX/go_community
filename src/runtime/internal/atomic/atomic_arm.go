// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm

package atomic

import "unsafe"

// copied from runtime2.go. Is that ok?
type mutex struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
}

var locktab [57]struct {
	l   mutex
	pad [_CacheLineSize - unsafe.Sizeof(mutex{})]byte
}

// copied from stubs.go. Is that ok?
// Should be a built-in for unsafe.Pointer?
//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func addrLock(addr *uint64) *mutex {
	return &locktab[(uintptr(unsafe.Pointer(addr))>>3)%uintptr(len(locktab))].l
}

// Atomic add and return new value.
//go:nosplit
func Xadd(val *uint32, delta int32) uint32 {
	for {
		oval := *val
		nval := oval + uint32(delta)
		if Cas(val, oval, nval) {
			return nval
		}
	}
}

//go:noescape
//go:linkname Xadduintptr runtime/internal/atomic.Xadd
func Xadduintptr(ptr *uintptr, delta uintptr) uintptr

//go:nosplit
func Xchg(addr *uint32, v uint32) uint32 {
	for {
		old := *addr
		if Cas(addr, old, v) {
			return old
		}
	}
}

//go:nosplit
func Xchguintptr(addr *uintptr, v uintptr) uintptr {
	return uintptr(Xchg((*uint32)(unsafe.Pointer(addr)), uint32(v)))
}

//go:nosplit
func Load(addr *uint32) uint32 {
	return add(addr, 0)
}

//go:nosplit
func Loadp(addr unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(uintptr(Xadd((*uint32)(addr), 0)))
}

//go:nosplit
func Storep1(addr unsafe.Pointer, v unsafe.Pointer) {
	for {
		old := *(*unsafe.Pointer)(addr)
		if Casp1((*unsafe.Pointer)(addr), old, v) {
			return
		}
	}
}

//go:nosplit
func Store(addr *uint32, v uint32) {
	for {
		old := *addr
		if Cas(addr, old, v) {
			return
		}
	}
}

//go:nosplit
func Cas64(addr *uint64, old, new uint64) bool {
	var ok bool
	systemstack(func() {
		lock(addrLock(addr))
		if *addr == old {
			*addr = new
			ok = true
		}
		unlock(addrLock(addr))
	})
	return ok
}

//go:nosplit
func Xadd64(addr *uint64, delta int64) uint64 {
	var r uint64
	systemstack(func() {
		lock(addrLock(addr))
		r = *addr + uint64(delta)
		*addr = r
		unlock(addrLock(addr))
	})
	return r
}

//go:nosplit
func Xchg64(addr *uint64, v uint64) uint64 {
	var r uint64
	systemstack(func() {
		lock(addrLock(addr))
		r = *addr
		*addr = v
		unlock(addrLock(addr))
	})
	return r
}

//go:nosplit
func Load64(addr *uint64) uint64 {
	var r uint64
	systemstack(func() {
		lock(addrLock(addr))
		r = *addr
		unlock(addrLock(addr))
	})
	return r
}

//go:nosplit
func Store64(addr *uint64, v uint64) {
	systemstack(func() {
		lock(addrLock(addr))
		*addr = v
		unlock(addrLock(addr))
	})
}

//go:nosplit
func Or8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8) // little endian
	for {
		old := *addr32
		if Cas(addr32, old, old|word) {
			return
		}
	}
}

//go:nosplit
func And8(addr *uint8, v uint8) {
	// Align down to 4 bytes and use 32-bit CAS.
	uaddr := uintptr(unsafe.Pointer(addr))
	addr32 := (*uint32)(unsafe.Pointer(uaddr &^ 3))
	word := uint32(v) << ((uaddr & 3) * 8)    // little endian
	mask := uint32(0xFF) << ((uaddr & 3) * 8) // little endian
	word |= ^mask
	for {
		old := *addr32
		if Cas(addr32, old, old&word) {
			return
		}
	}
}
