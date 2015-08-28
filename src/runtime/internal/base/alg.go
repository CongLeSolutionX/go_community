// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// typeAlg is also copied/used in reflect/type.go.
// keep them in sync.
type TypeAlg struct {
	// function for hashing objects of this type
	// (ptr to object, seed) -> hash
	Hash func(unsafe.Pointer, uintptr) uintptr
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
}

var UseAeshash bool

// in asm_*.s
func aeshash(p unsafe.Pointer, h, s uintptr) uintptr

// used in hash{32,64}.go to seed the hash function
var Hashkey [4]uintptr
