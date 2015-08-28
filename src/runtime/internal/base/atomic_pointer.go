// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// These functions cannot have go:noescape annotations,
// because while ptr does not escape, new does.
// If new is marked as not escaping, the compiler will make incorrect
// escape analysis decisions about the pointer value being stored.
// Instead, these are wrappers around the actual atomics (xchgp1 and so on)
// that use noescape to convey which arguments do not escape.
//
// Additionally, these functions must update the shadow heap for
// write barrier checking.

//go:nosplit
func Atomicstorep(ptr unsafe.Pointer, new unsafe.Pointer) {
	Atomicstorep1(Noescape(ptr), new)
	Writebarrierptr_nostore((*uintptr)(ptr), uintptr(new))
}
