// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

type fakeErrno syscall.Errno

func (e fakeErrno) Error() string { return "fake " + syscall.Errno(e).Error() }
func (e fakeErrno) Timeout() bool { return e == errFakeTimedout }

const (
	errFakeTimedout       = fakeErrno(syscall.ETIMEDOUT)
	errFakeOpNotSupported = fakeErrno(syscall.EOPNOTSUPP)
)

var abortedConnRequestErrors = []error{syscall.ERROR_NETNAME_DELETED, syscall.WSAECONNRESET} // see accept in fd_windows.go

func isPlatformError(err error) bool {
	if _, ok := err.(fakeErrno); ok {
		return true
	}
	_, ok := err.(syscall.Errno)
	return ok
}
