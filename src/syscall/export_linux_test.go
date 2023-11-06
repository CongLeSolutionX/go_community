// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import (
	"unsafe"
)

var (
	RawSyscallNoError = rawSyscallNoError
	ForceClone3       = &forceClone3
)

const (
	Sys_GETEUID = sys_GETEUID
)

// Note that pgid should really be pid_t, however _C_int (aka int32) is
// generally equivalent.

func Tcgetpgrp(fd int) (pgid int32, err error) {
	_, _, errno := Syscall6(SYS_IOCTL, uintptr(fd), uintptr(TIOCGPGRP), uintptr(unsafe.Pointer(&pgid)), 0, 0, 0)
	if errno != 0 {
		return -1, errno
	}
	return pgid, nil
}

func Tcsetpgrp(fd int, pgid int32) (err error) {
	_, _, errno := Syscall6(SYS_IOCTL, uintptr(fd), uintptr(TIOCSPGRP), uintptr(unsafe.Pointer(&pgid)), 0, 0, 0)
	if errno != 0 {
		return errno
	}
	return nil
}
