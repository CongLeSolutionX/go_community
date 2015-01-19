// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9
// +build !solaris
// +build !windows
// +build !nacl

package sched

import (
	"unsafe"
)

func munmap(addr unsafe.Pointer, n uintptr)

func Memmove(addr1, addr2 unsafe.Pointer, n uintptr) {
	memmove(addr1, addr2, n)
}
