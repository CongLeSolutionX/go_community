// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	_core "runtime/internal/core"
	"unsafe"
)

// casp cannot have a go:noescape annotation, because
// while ptr and old do not escape, new does. If new is marked as
// not escaping, the compiler will make incorrect escape analysis
// decisions about the value being xchg'ed.
// Instead, make casp a wrapper around the actual atomic.
// When calling the wrapper we mark ptr as noescape explicitly.

//go:nosplit
func casp(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool {
	return casp1((*unsafe.Pointer)(_core.Noescape(unsafe.Pointer(ptr))), _core.Noescape(old), new)
}

func casp1(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool

func prefetcht0(addr uintptr)
func prefetcht1(addr uintptr)
func prefetcht2(addr uintptr)
func prefetchnta(addr uintptr)
