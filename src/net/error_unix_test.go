// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9,!windows

package net

import (
	"os"
	"syscall"
)

type fakeErrno syscall.Errno

func (e fakeErrno) Error() string { return "fake " + syscall.Errno(e).Error() }
func (e fakeErrno) Timeout() bool { return e == errFakeTimedout }

const (
	errFakeTimedout       = fakeErrno(syscall.ETIMEDOUT)
	errFakeOpNotSupported = fakeErrno(syscall.EOPNOTSUPP)
)

var abortedConnRequestErrors = []error{syscall.ECONNABORTED} // see accept in fd_unix.go

func isPlatformError(err error) bool {
	if _, ok := err.(fakeErrno); ok {
		return true
	}
	_, ok := err.(syscall.Errno)
	return ok
}

func samePlatformError(err, want error) bool {
	if op, ok := err.(*OpError); ok {
		err = op.Err
	}
	if sys, ok := err.(*os.SyscallError); ok {
		err = sys.Err
	}
	return err == want
}
