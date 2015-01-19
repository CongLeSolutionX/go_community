// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

//go:nosplit
func Atomicload64(ptr *uint64) uint64 {
	_lock.Nop()
	return *ptr
}

//go:noescape
func Xchg64(ptr *uint64, new uint64) uint64

//go:noescape
func Atomicor8(ptr *uint8, val uint8)

//go:noescape
func Cas64(ptr *uint64, old, new uint64) bool

//go:noescape
func Atomicstore64(ptr *uint64, val uint64)

// atomicstorep cannot have a go:noescape annotation.
// See comment above for xchgp.

//go:nosplit
func Atomicstorep(ptr unsafe.Pointer, new unsafe.Pointer) {
	atomicstorep1(_core.Noescape(ptr), new)
}

func atomicstorep1(ptr unsafe.Pointer, val unsafe.Pointer)
