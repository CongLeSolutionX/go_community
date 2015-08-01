// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9
// +build !solaris
// +build !windows
// +build !nacl

package base

import (
	"unsafe"
)

func Exit(code int32)
func Nanotime() int64
func Usleep(usec uint32)

func Mmap(addr unsafe.Pointer, n uintptr, prot, flags, fd int32, off uint32) unsafe.Pointer
func munmap(addr unsafe.Pointer, n uintptr)
