// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9
// +build !solaris
// +build !windows
// +build !nacl

package hash

import (
	"unsafe"
)

func read(fd int32, p unsafe.Pointer, n int32) int32
func close(fd int32) int32

//go:noescape
func open(name *byte, mode, perm int32) int32
