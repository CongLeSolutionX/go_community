// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package net

import (
	"errors"
	"os"
	"syscall"
	"time"
)

func unixSocket(net string, laddr, raddr sockaddr, mode string, deadline time.Time) (*netFD, error) {
	var sotype int
	switch net {
	case "unix":
		sotype = syscall.SOCK_STREAM
	case "unixgram":
		sotype = syscall.SOCK_DGRAM
	case "unixpacket":
		sotype = syscall.SOCK_SEQPACKET
	default:
		return nil, UnknownNetworkError(net)
	}

	switch mode {
	case "dial":
		if laddr != nil && laddr.isWildcard() {
			laddr = nil
		}
		if raddr != nil && raddr.isWildcard() {
			raddr = nil
		}
		if raddr == nil && (sotype != syscall.SOCK_DGRAM || laddr == nil) {
			return nil, errMissingAddress
		}
	case "listen":
	default:
		return nil, errors.New("unknown mode: " + mode)
	}

	fd, err := socket(net, syscall.AF_UNIX, sotype, 0, false, laddr, raddr, deadline)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func sockaddrToUnix(sa syscall.Sockaddr) Addr {
	if s, ok := sa.(*syscall.SockaddrUnix); ok {
		return &UnixAddr{Name: s.Name, Net: "unix"}
	}
	return nil
}

func sockaddrToUnixgram(sa syscall.Sockaddr) Addr {
	if s, ok := sa.(*syscall.SockaddrUnix); ok {
		return &UnixAddr{Name: s.Name, Net: "unixgram"}
	}
	return nil
}

func sockaddrToUnixpacket(sa syscall.Sockaddr) Addr {
	if s, ok := sa.(*syscall.SockaddrUnix); ok {
		return &UnixAddr{Name: s.Name, Net: "unixpacket"}
	}
	return nil
}

func sotypeToNet(sotype int) string {
	switch sotype {
	case syscall.SOCK_STREAM:
		return "unix"
	case syscall.SOCK_DGRAM:
		return "unixgram"
	case syscall.SOCK_SEQPACKET:
		return "unixpacket"
	default:
		panic("sotypeToNet unknown socket type")
	}
}

func (a *UnixAddr) family() int {
	return syscall.AF_UNIX
}

func (a *UnixAddr) isWildcard() bool {
	return a == nil || a.Name == ""
}

func (a *UnixAddr) sockaddr(family int) (syscall.Sockaddr, error) {
	if a == nil {
		return nil, nil
	}
	return &syscall.SockaddrUnix{Name: a.Name}, nil
}

func newUnixConn(fd *netFD) *UnixConn { return &UnixConn{conn{fd}} }

func (c *UnixConn) readFromUnix(b []byte) (int, *UnixAddr, error) {
	var addr *UnixAddr
	n, sa, err := c.fd.readFrom(b)
	switch sa := sa.(type) {
	case *syscall.SockaddrUnix:
		if sa.Name != "" {
			addr = &UnixAddr{Name: sa.Name, Net: sotypeToNet(c.fd.sotype)}
		}
	}
	if err != nil {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, addr, err
}

func (c *UnixConn) readMsgUnix(b, oob []byte) (n, oobn, flags int, addr *UnixAddr, err error) {
	n, oobn, flags, sa, err := c.fd.readMsg(b, oob)
	switch sa := sa.(type) {
	case *syscall.SockaddrUnix:
		if sa.Name != "" {
			addr = &UnixAddr{Name: sa.Name, Net: sotypeToNet(c.fd.sotype)}
		}
	}
	if err != nil {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return
}

func (c *UnixConn) writeToUnix(b []byte, addr *UnixAddr) (int, error) {
	if c.fd.isConnected {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: ErrWriteToConnected}
	}
	if addr == nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: errMissingAddress}
	}
	if addr.Net != sotypeToNet(c.fd.sotype) {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: syscall.EAFNOSUPPORT}
	}
	sa := &syscall.SockaddrUnix{Name: addr.Name}
	n, err := c.fd.writeTo(b, sa)
	if err != nil {
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: err}
	}
	return n, err
}

func (c *UnixConn) writeMsgUnix(b, oob []byte, addr *UnixAddr) (n, oobn int, err error) {
	if c.fd.sotype == syscall.SOCK_DGRAM && c.fd.isConnected {
		return 0, 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: ErrWriteToConnected}
	}
	var sa syscall.Sockaddr
	if addr != nil {
		if addr.Net != sotypeToNet(c.fd.sotype) {
			return 0, 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: syscall.EAFNOSUPPORT}
		}
		sa = &syscall.SockaddrUnix{Name: addr.Name}
	}
	n, oobn, err = c.fd.writeMsg(b, oob, sa)
	if err != nil {
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: err}
	}
	return
}

func dialUnix(net string, laddr, raddr *UnixAddr, deadline time.Time) (*UnixConn, error) {
	fd, err := unixSocket(net, laddr, raddr, "dial", deadline)
	if err != nil {
		return nil, &OpError{Op: "dial", Net: net, Source: laddr, Addr: raddr, Err: err}
	}
	return newUnixConn(fd), nil
}

func (l *UnixListener) ok() bool { return l != nil && l.fd != nil }

func (l *UnixListener) acceptUnix() (*UnixConn, error) {
	fd, err := l.fd.accept()
	if err != nil {
		return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return newUnixConn(fd), nil
}

func (l *UnixListener) close() error {
	// The operating system doesn't clean up
	// the file that announcing created, so
	// we have to clean it up ourselves.
	// There's a race here--we can't know for
	// sure whether someone else has come along
	// and replaced our socket name already--
	// but this sequence (remove then close)
	// is at least compatible with the auto-remove
	// sequence in ListenUnix.  It's only non-Go
	// programs that can mess us up.
	if l.path[0] != '@' {
		syscall.Unlink(l.path)
	}
	err := l.fd.Close()
	if err != nil {
		err = &OpError{Op: "close", Net: l.fd.net, Source: l.fd.laddr, Addr: l.fd.raddr, Err: err}
	}
	return err
}

func (l *UnixListener) file() (*os.File, error) {
	f, err := l.fd.dup()
	if err != nil {
		return nil, &OpError{Op: "file", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return f, nil
}

func listenUnix(net string, laddr *UnixAddr) (*UnixListener, error) {
	fd, err := unixSocket(net, laddr, nil, "listen", noDeadline)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: err}
	}
	return &UnixListener{fd: fd, path: fd.laddr.String()}, nil
}

func listenUnixgram(net string, laddr *UnixAddr) (*UnixConn, error) {
	fd, err := unixSocket(net, laddr, nil, "listen", noDeadline)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: err}
	}
	return newUnixConn(fd), nil
}
