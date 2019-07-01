// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package os

import (
	"runtime"
)

// rawConn implements syscall.RawConn.
type rawConn struct {
	file *File
}

func (c *rawConn) Control(f func(uintptr)) error {
	try(c.file.checkValid("SyscallConn.Control"))
	err := c.file.pfd.RawControl(f)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Read(f func(uintptr) bool) error {
	try(c.file.checkValid("SyscallConn.Read"))
	err := c.file.pfd.RawRead(f)
	runtime.KeepAlive(c.file)
	return err
}

func (c *rawConn) Write(f func(uintptr) bool) error {
	try(c.file.checkValid("SyscallConn.Write"))
	err := c.file.pfd.RawWrite(f)
	runtime.KeepAlive(c.file)
	return err
}

func newRawConn(file *File) (*rawConn, error) {
	return &rawConn{file: file}, nil
}
