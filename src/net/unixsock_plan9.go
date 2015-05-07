// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"syscall"
	"time"
)

func (c *UnixConn) readFromUnix(b []byte) (int, *UnixAddr, error) {
	return 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *UnixConn) readMsgUnix(b, oob []byte) (n, oobn, flags int, addr *UnixAddr, err error) {
	return 0, 0, 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *UnixConn) writeToUnix(b []byte, addr *UnixAddr) (int, error) {
	return 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: syscall.EPLAN9}
}

func (c *UnixConn) writeMsgUnix(b, oob []byte, addr *UnixAddr) (n, oobn int, err error) {
	return 0, 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: syscall.EPLAN9}
}

func dialUnix(net string, laddr, raddr *UnixAddr, deadline time.Time) (*UnixConn, error) {
	return nil, &OpError{Op: "dial", Net: net, Source: laddr, Addr: raddr, Err: syscall.EPLAN9}
}

func (l *UnixListener) ok() bool { return l != nil && l.fd != nil && l.fd.ctl != nil }

func (l *UnixListener) acceptUnix() (*UnixConn, error) {
	return nil, &OpError{Op: "accept", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: syscall.EPLAN9}
}

func (l *UnixListener) close() error {
	return &OpError{Op: "close", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: syscall.EPLAN9}
}

func (l *UnixListener) file() (*os.File, error) {
	return nil, &OpError{Op: "file", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: syscall.EPLAN9}
}

func listenUnix(net string, laddr *UnixAddr) (*UnixListener, error) {
	return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: syscall.EPLAN9}
}

func listenUnixgram(net string, laddr *UnixAddr) (*UnixConn, error) {
	return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: syscall.EPLAN9}
}
