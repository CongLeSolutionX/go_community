// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9
// +build !solaris
// +build !windows
// +build !nacl

package core

import (
	"unsafe"
)

func Exit(code int32)
func Usleep(usec uint32)

//go:noescape
func Write(fd uintptr, p unsafe.Pointer, n int32) int32
