// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

type fakeError struct{ error }

func (e fakeError) Error() string { return "fake " + e.error.Error() }
func (e fakeError) Timeout() bool { return e == errFakeTimedout }

var (
	errFakeTimedout       = fakeError{error: syscall.ETIMEDOUT}
	errFakeOpNotSupported = fakeError{error: syscall.EPLAN9}
)

var abortedConnRequestErrors []error

func isPlatformError(err error) bool {
	if _, ok := err.(fakeError); ok {
		return true
	}
	_, ok := err.(syscall.ErrorString)
	return ok
}
