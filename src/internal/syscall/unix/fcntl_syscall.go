// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build dragonfly || freebsd || linux || netbsd || (openbsd && mips64)

package unix

import "syscall"

func Fcntl(fd int, cmd int, arg int) (int, error) {
	r, _, e := syscall.Syscall(FcntlSyscall, uintptr(fd), uintptr(cmd), uintptr(arg))
	if e != 0 {
		return int(r), syscall.Errno(e)
	}
	return int(r), nil
}
