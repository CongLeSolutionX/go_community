// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux freebsd netbsd openbsd

package syscall

import (
	stdsyscall "syscall"
	"unsafe"
)

// Fstatat calls the Fstatat syscall for the platform.
func Fstatat(fd int, path string, stat *stdsyscall.Stat_t, flags int) (err error) {
	var _p0 *byte
	_p0, err = stdsyscall.BytePtrFromString(path)
	if err != nil {
		return
	}
	_, _, e1 := stdsyscall.Syscall6(fstatatNum, uintptr(fd), uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(stat)), uintptr(flags), 0, 0)
	use(unsafe.Pointer(_p0))
	if e1 != 0 {
		err = e1
	}
	return
}
