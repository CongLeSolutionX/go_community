// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"io"
	"os"
	"syscall"
	"time"
)

func newTCPConn(fd *netFD) *TCPConn {
	return &TCPConn{conn{fd}}
}

func (c *TCPConn) readFrom(r io.Reader) (int64, error) {
	n, err := genericReadFrom(c, r)
	if err != nil && err != io.EOF {
		err = &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}

func (c *TCPConn) setLinger(sec int) error {
	return &OpError{Op: "set", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *TCPConn) setNoDelay(noDelay bool) error {
	return &OpError{Op: "set", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func dialTCP(net string, laddr, raddr *TCPAddr, deadline time.Time) (*TCPConn, error) {
	if !deadline.IsZero() {
		panic("net.dialTCP: deadline not implemented on Plan 9")
	}
	fd, err := dialPlan9(net, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return newTCPConn(fd), nil
}

func (l *TCPListener) ok() bool { return l != nil && l.fd != nil && l.fd.ctl != nil }

func (l *TCPListener) acceptTCP() (*TCPConn, error) {
	fd, err := l.fd.acceptPlan9()
	if err != nil {
		return nil, err
	}
	return newTCPConn(fd), nil
}

func (l *TCPListener) close() error {
	if _, err := l.fd.ctl.WriteString("hangup"); err != nil {
		l.fd.ctl.Close()
		return &OpError{Op: "close", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	if err := l.fd.ctl.Close(); err != nil {
		return &OpError{Op: "close", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return nil
}

func (l *TCPListener) file() (*os.File, error) {
	f, err := l.dup()
	if err != nil {
		return nil, &OpError{Op: "file", Net: l.fd.dir, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return f, nil
}

func listenTCP(net string, laddr *TCPAddr) (*TCPListener, error) {
	fd, err := listenPlan9(net, laddr)
	if err != nil {
		return nil, err
	}
	return &TCPListener{fd}, nil
}
