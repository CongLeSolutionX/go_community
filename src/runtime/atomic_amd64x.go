// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32

package runtime

import (
	"unsafe"
)

//go:noescape
func xchg(ptr *uint32, new uint32) uint32

// NO go:noescape annotation; see atomic_pointer.go.
func xchgp1(ptr unsafe.Pointer, new unsafe.Pointer) unsafe.Pointer

//go:noescape
func xchguintptr(ptr *uintptr, new uintptr) uintptr
