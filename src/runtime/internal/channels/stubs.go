// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package channels

import (
	"unsafe"
)

//go:noescape
func Atomicloaduint(ptr *uint) uint

//go:noescape
func setcallerpc(argp unsafe.Pointer, pc uintptr)
