// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package net

import "syscall"

// Read implements the Conn Read method.
func (c *TCPConn) Read(b []byte) (int, error) {
	if !c.ok() {
		return 0, &OpError{Op: "read", Net: c.fd.net, Addr: c.fd.laddr, Err: syscall.EINVAL}
	}
	if !c.fastOpen || !supportsTCPActiveFastOpen {
		return c.fd.Read(b)
	}
	return c.fd.openRead(b, syscall.MSG_FASTOPEN)
}

// Write implements the Conn Write method.
func (c *TCPConn) Write(b []byte) (int, error) {
	if !c.ok() {
		return 0, &OpError{Op: "write", Net: c.fd.net, Addr: c.fd.raddr, Err: syscall.EINVAL}
	}
	if !c.fastOpen || !supportsTCPActiveFastOpen {
		return c.fd.Write(b)
	}
	return c.fd.openWrite(b, syscall.MSG_FASTOPEN)
}
