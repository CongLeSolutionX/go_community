// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

import (
	"unsafe"
)

type TypeAlg struct {
	// function for hashing objects of this type
	// (ptr to object, size, seed) -> hash
	Hash func(unsafe.Pointer, uintptr, uintptr) uintptr
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B, size) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer, uintptr) bool
}
