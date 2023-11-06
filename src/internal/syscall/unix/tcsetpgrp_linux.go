// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"syscall"
)

func Tcsetpgrp(fd int, pgid int32) (err error) {
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCSPGRP), uintptr(pgid), 0, 0, 0)
	if errno != 0 {
		err = errno
	}
	return
}
