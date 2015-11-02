// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic

import "unsafe"

//go:noescape
func Cas(ptr *uint32, old, new uint32) bool

// NO go:noescape annotation; see atomic_pointer.go.
func Casp1(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool

func nop() // call to prevent inlining of function body

//go:noescape
func Casuintptr(ptr *uintptr, old, new uintptr) bool

//go:noescape
func Storeuintptr(ptr *uintptr, new uintptr)

//go:noescape
func Loaduintptr(ptr *uintptr) uintptr

//go:noescape
func Loaduint(ptr *uint) uint

// TODO: Write native implementations of int64 atomic ops (or improve
// inliner). These portable ones can't be inlined right now, so we're
// taking an extra function call hit.

func Atomicstoreint64(ptr *int64, new int64) {
	Store64((*uint64)(unsafe.Pointer(ptr)), uint64(new))
}

func Atomicloadint64(ptr *int64) int64 {
	return int64(Load64((*uint64)(unsafe.Pointer(ptr))))
}

func Xaddint64(ptr *int64, delta int64) int64 {
	return int64(Xadd64((*uint64)(unsafe.Pointer(ptr)), delta))
}
