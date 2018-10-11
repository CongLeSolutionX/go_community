// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"runtime"
	"syscall"
	"time"
)

// BUG(tmm1): On Windows, the Write method of syscall.RawConn
// does not integrate with the runtime's network poller. It cannot
// wait for the connection to become writeable, and does not respect
// deadlines. If the user-provided callback returns false, the Write
// method will fail immediately.

// BUG(mikio): On AIX, JS, NaCl and Plan 9, the Control, Read and
// Write methods of syscall.RawConn are not implemented.

// BUG(mikio): On AIX, JS, NaCl and Plan 9, methods and functions
// related to SyscallConn are not implemented.

type rawConn struct {
	fd *netFD
}

func (c *rawConn) ok() bool { return c != nil && c.fd != nil }

func (c *rawConn) Control(f func(uintptr)) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	err := c.fd.pfd.RawControl(f)
	runtime.KeepAlive(c.fd)
	if err != nil {
		err = &OpError{Op: "raw-control", Net: c.fd.net, Source: nil, Addr: c.fd.laddr, Err: err}
	}
	return err
}

func (c *rawConn) Read(f func(uintptr) bool) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	err := c.fd.pfd.RawRead(f)
	runtime.KeepAlive(c.fd)
	if err != nil {
		err = &OpError{Op: "raw-read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return err
}

func (c *rawConn) Write(f func(uintptr) bool) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	err := c.fd.pfd.RawWrite(f)
	runtime.KeepAlive(c.fd)
	if err != nil {
		err = &OpError{Op: "raw-write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return err
}

func newRawConn(fd *netFD) (*rawConn, error) {
	return &rawConn{fd: fd}, nil
}

type rawListener struct {
	rawConn
}

func (l *rawListener) Read(func(uintptr) bool) error {
	return syscall.EINVAL
}

func (l *rawListener) Write(func(uintptr) bool) error {
	return syscall.EINVAL
}

func newRawListener(fd *netFD) (*rawListener, error) {
	return &rawListener{rawConn{fd: fd}}, nil
}

// SyscallConn is an implementation of the syscall.Conn interface.
type SyscallConn struct {
	c rawConn
}

func (c *SyscallConn) ok() bool { return c != nil && c.c.fd != nil }

func (c *SyscallConn) Close() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.c.fd.Close()
}

// SyscallConn returns a raw network connection.
func (c *SyscallConn) SyscallConn() (syscall.RawConn, error) {
	if !c.ok() {
		return nil, syscall.EINVAL
	}
	return &c.c, nil
}

// SetDeadline implements the SetDeadline method of Conn or PacketConn
// interface.
func (c *SyscallConn) SetDeadline(t time.Time) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.c.fd.pfd.SetDeadline(t)
}

// SetReadDeadline implements the SetReadDeadline method of Conn or
// PacketConn interface.
func (c *SyscallConn) SetReadDeadline(t time.Time) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.c.fd.pfd.SetReadDeadline(t)
}

// SetWriteDeadline implements the SetWriteDeadline method of Conn or
// PacketConn interface.
func (c *SyscallConn) SetWriteDeadline(t time.Time) error {
	if !c.ok() {
		return syscall.EINVAL
	}
	return c.c.fd.pfd.SetWriteDeadline(t)
}

// NewSyscallConn returns a new raw network connection.
func NewSyscallConn(network string, f *os.File) (*SyscallConn, error) {
	c, err := newSyscallConn(network, f)
	if err != nil {
		return nil, syscall.EINVAL
	}
	return c, nil
}
