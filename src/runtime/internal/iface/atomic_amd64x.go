// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32

package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

//go:nosplit
func Atomicloadp(ptr unsafe.Pointer) unsafe.Pointer {
	_base.Nop()
	return *(*unsafe.Pointer)(ptr)
}

//go:noescape
func atomicand8(ptr *uint8, val uint8)
