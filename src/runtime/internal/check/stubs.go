// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package check

import (
	"unsafe"
)

// NO go:noescape annotation; see atomic_pointer.go.
func casp1(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool

func prefetcht0(addr uintptr)
func prefetcht1(addr uintptr)
func prefetcht2(addr uintptr)
func prefetchnta(addr uintptr)
