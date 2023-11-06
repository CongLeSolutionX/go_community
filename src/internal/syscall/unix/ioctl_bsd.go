// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package unix

import (
	"unsafe"
)

func ioctlPtr(fd int, req uint, arg unsafe.Pointer) (err error)

//go:linkname ioctlPtr syscall.ioctlPtr

func Ioctl(fd int, cmd int, args unsafe.Pointer) (err error) {
	return ioctlPtr(fd, uint(cmd), args)
}
