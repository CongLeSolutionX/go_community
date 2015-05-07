// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package net

import (
	"io"
	"os"
	"syscall"
	"time"
)

func sockaddrToTCP(sa syscall.Sockaddr) Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return &TCPAddr{IP: sa.Addr[0:], Port: sa.Port}
	case *syscall.SockaddrInet6:
		return &TCPAddr{IP: sa.Addr[0:], Port: sa.Port, Zone: zoneToString(int(sa.ZoneId))}
	}
	return nil
}

func (a *TCPAddr) family() int {
	if a == nil || len(a.IP) <= IPv4len {
		return syscall.AF_INET
	}
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

func (a *TCPAddr) sockaddr(family int) (syscall.Sockaddr, error) {
	if a == nil {
		return nil, nil
	}
	return ipToSockaddr(family, a.IP, a.Port, a.Zone)
}

func newTCPConn(fd *netFD) *TCPConn {
	c := &TCPConn{conn{fd}}
	setNoDelay(c.fd, true)
	return c
}

func (c *TCPConn) readFrom(r io.Reader) (int64, error) {
	if n, err, handled := sendFile(c.fd, r); handled {
		if err != nil && err != io.EOF {
			err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		}
		return n, err
	}
	n, err := genericReadFrom(c, r)
	if err != nil && err != io.EOF {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, err
}

func (c *TCPConn) setLinger(sec int) error {
	if err := setLinger(c.fd, sec); err != nil {
		return &OpError{Op: "set", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return nil
}

func (c *TCPConn) setNoDelay(noDelay bool) error {
	if err := setNoDelay(c.fd, noDelay); err != nil {
		return &OpError{Op: "set", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return nil
}

func dialTCP(net string, laddr, raddr *TCPAddr, deadline time.Time) (*TCPConn, error) {
	fd, err := internetSocket(net, laddr, raddr, deadline, syscall.SOCK_STREAM, 0, "dial")

	// TCP has a rarely used mechanism called a 'simultaneous connection' in
	// which Dial("tcp", addr1, addr2) run on the machine at addr1 can
	// connect to a simultaneous Dial("tcp", addr2, addr1) run on the machine
	// at addr2, without either machine executing Listen.  If laddr == nil,
	// it means we want the kernel to pick an appropriate originating local
	// address.  Some Linux kernels cycle blindly through a fixed range of
	// local ports, regardless of destination port.  If a kernel happens to
	// pick local port 50001 as the source for a Dial("tcp", "", "localhost:50001"),
	// then the Dial will succeed, having simultaneously connected to itself.
	// This can only happen when we are letting the kernel pick a port (laddr == nil)
	// and when there is no listener for the destination address.
	// It's hard to argue this is anything other than a kernel bug.  If we
	// see this happen, rather than expose the buggy effect to users, we
	// close the fd and try again.  If it happens twice more, we relent and
	// use the result.  See also:
	//	http://golang.org/issue/2690
	//	http://stackoverflow.com/questions/4949858/
	//
	// The opposite can also happen: if we ask the kernel to pick an appropriate
	// originating local address, sometimes it picks one that is already in use.
	// So if the error is EADDRNOTAVAIL, we have to try again too, just for
	// a different reason.
	//
	// The kernel socket code is no doubt enjoying watching us squirm.
	for i := 0; i < 2 && (laddr == nil || laddr.Port == 0) && (selfConnect(fd, err) || spuriousENOTAVAIL(err)); i++ {
		if err == nil {
			fd.Close()
		}
		fd, err = internetSocket(net, laddr, raddr, deadline, syscall.SOCK_STREAM, 0, "dial")
	}

	if err != nil {
		return nil, &OpError{Op: "dial", Net: net, Source: laddr, Addr: raddr, Err: err}
	}
	return newTCPConn(fd), nil
}

func selfConnect(fd *netFD, err error) bool {
	// If the connect failed, we clearly didn't connect to ourselves.
	if err != nil {
		return false
	}

	// The socket constructor can return an fd with raddr nil under certain
	// unknown conditions. The errors in the calls there to Getpeername
	// are discarded, but we can't catch the problem there because those
	// calls are sometimes legally erroneous with a "socket not connected".
	// Since this code (selfConnect) is already trying to work around
	// a problem, we make sure if this happens we recognize trouble and
	// ask the DialTCP routine to try again.
	// TODO: try to understand what's really going on.
	if fd.laddr == nil || fd.raddr == nil {
		return true
	}
	l := fd.laddr.(*TCPAddr)
	r := fd.raddr.(*TCPAddr)
	return l.Port == r.Port && l.IP.Equal(r.IP)
}

func spuriousENOTAVAIL(err error) bool {
	e, ok := err.(*OpError)
	return ok && e.Err == syscall.EADDRNOTAVAIL
}

func (l *TCPListener) ok() bool { return l != nil && l.fd != nil }

func (l *TCPListener) acceptTCP() (*TCPConn, error) {
	fd, err := l.fd.accept()
	if err != nil {
		return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return newTCPConn(fd), nil
}

func (l *TCPListener) close() error {
	if err := l.fd.Close(); err != nil {
		return &OpError{Op: "close", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return nil
}

func (l *TCPListener) file() (*os.File, error) {
	f, err := l.fd.dup()
	if err != nil {
		return nil, &OpError{Op: "file", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return f, nil
}

func listenTCP(net string, laddr *TCPAddr) (*TCPListener, error) {
	fd, err := internetSocket(net, laddr, nil, noDeadline, syscall.SOCK_STREAM, 0, "listen")
	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: err}
	}
	return &TCPListener{fd}, nil
}
