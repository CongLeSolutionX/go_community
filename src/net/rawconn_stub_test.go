// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !aix,!darwin,!dragonfly,!freebsd,!linux,!netbsd,!openbsd,!solaris,!windows

package net

import (
	"errors"
	"os"
	"syscall"
)

func readRawConn(c syscall.RawConn, b []byte) (int, error) {
	return 0, errors.New("not supported")
}

func writeRawConn(c syscall.RawConn, b []byte) error {
	return errors.New("not supported")
}

func controlRawConn(c syscall.RawConn, addr Addr) error {
	return errors.New("not supported")
}

func controlOnConnSetup(network string, address string, c syscall.RawConn) error {
	return nil
}

func newSyscallConnFile() (string, *os.File, error) {
	return "", nil, errors.New("not supported")
}
