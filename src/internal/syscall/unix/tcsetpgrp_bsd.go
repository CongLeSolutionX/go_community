// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package unix

import (
	"syscall"
	"unsafe"
)

func ioctlPtr(fd int, req uint, arg unsafe.Pointer) (err error)

//go:linkname ioctlPtr syscall.ioctlPtr

func Tcsetpgrp(fd int, pgid int32) (err error) {
	return ioctlPtr(fd, syscall.TIOCSPGRP, unsafe.Pointer(&pgid))
}
