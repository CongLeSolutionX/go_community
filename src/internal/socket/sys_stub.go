// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package socket

import "errors"

func getsockname(s uintptr) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func getpeername(s uintptr) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func getsockopt(s uintptr, level, name int, b []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func setsockopt(s uintptr, level, name int, b []byte) error {
	return errors.New("not implemented")
}
