// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package socket

// This file contains duplicates of syscall package.
//
// This package is supposed to be used by the net package of standard
// library. Therefore a package set used in the package must be the
// same as net package.

import "syscall"

var (
	errEAGAIN error = syscall.EAGAIN
	errEINVAL error = syscall.EINVAL
	errENOENT error = syscall.ENOENT
)

// errnoErr returns common boxed Errno values, to prevent allocations
// at runtime.
func errnoErr(errno syscall.Errno) error {
	switch errno {
	case 0:
		return nil
	case syscall.EAGAIN:
		return errEAGAIN
	case syscall.EINVAL:
		return errEINVAL
	case syscall.ENOENT:
		return errENOENT
	}
	return errno
}
