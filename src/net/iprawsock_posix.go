// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package net

import (
	"syscall"
	"time"
)

func sockaddrToIP(sa syscall.Sockaddr) Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		return &IPAddr{IP: sa.Addr[0:]}
	case *syscall.SockaddrInet6:
		return &IPAddr{IP: sa.Addr[0:], Zone: zoneToString(int(sa.ZoneId))}
	}
	return nil
}

func (a *IPAddr) family() int {
	if a == nil || len(a.IP) <= IPv4len {
		return syscall.AF_INET
	}
	if a.IP.To4() != nil {
		return syscall.AF_INET
	}
	return syscall.AF_INET6
}

func (a *IPAddr) sockaddr(family int) (syscall.Sockaddr, error) {
	if a == nil {
		return nil, nil
	}
	return ipToSockaddr(family, a.IP, 0, a.Zone)
}

func newIPConn(fd *netFD) *IPConn { return &IPConn{conn{fd}} }

func (c *IPConn) readFromIP(b []byte) (int, *IPAddr, error) {
	// TODO(cw,rsc): consider using readv if we know the family
	// type to avoid the header trim/copy
	var addr *IPAddr
	n, sa, err := c.fd.readFrom(b)
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = &IPAddr{IP: sa.Addr[0:]}
		n = stripIPv4Header(n, b)
	case *syscall.SockaddrInet6:
		addr = &IPAddr{IP: sa.Addr[0:], Zone: zoneToString(int(sa.ZoneId))}
	}
	if err != nil {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, addr, err
}

func stripIPv4Header(n int, b []byte) int {
	if len(b) < 20 {
		return n
	}
	l := int(b[0]&0x0f) << 2
	if 20 > l || l > len(b) {
		return n
	}
	if b[0]>>4 != 4 {
		return n
	}
	copy(b, b[l:])
	return n - l
}

func (c *IPConn) readMsgIP(b, oob []byte) (n, oobn, flags int, addr *IPAddr, err error) {
	var sa syscall.Sockaddr
	n, oobn, flags, sa, err = c.fd.readMsg(b, oob)
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		addr = &IPAddr{IP: sa.Addr[0:]}
	case *syscall.SockaddrInet6:
		addr = &IPAddr{IP: sa.Addr[0:], Zone: zoneToString(int(sa.ZoneId))}
	}
	if err != nil {
		err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return
}

func (c *IPConn) writeToIP(b []byte, addr *IPAddr) (int, error) {
	if c.fd.isConnected {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: ErrWriteToConnected}
	}
	if addr == nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: nil, Err: errMissingAddress}
	}
	sa, err := addr.sockaddr(c.fd.family)
	if err != nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: err}
	}
	n, err := c.fd.writeTo(b, sa)
	if err != nil {
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: err}
	}
	return n, err
}

func (c *IPConn) writeMsgIP(b, oob []byte, addr *IPAddr) (n, oobn int, err error) {
	if c.fd.isConnected {
		return 0, 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: ErrWriteToConnected}
	}
	if addr == nil {
		return 0, 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: nil, Err: errMissingAddress}
	}
	var sa syscall.Sockaddr
	sa, err = addr.sockaddr(c.fd.family)
	if err != nil {
		return 0, 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: err}
	}
	n, oobn, err = c.fd.writeMsg(b, oob, sa)
	if err != nil {
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr.opAddr(), Err: err}
	}
	return
}

func dialIP(netProto string, laddr, raddr *IPAddr, deadline time.Time) (*IPConn, error) {
	network, proto, err := parseNetwork(netProto)
	if err != nil {
		return nil, &OpError{Op: "dial", Net: netProto, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
	}
	switch network {
	case "ip", "ip4", "ip6":
	default:
		return nil, &OpError{Op: "dial", Net: netProto, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: UnknownNetworkError(netProto)}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: netProto, Source: laddr.opAddr(), Addr: nil, Err: errMissingAddress}
	}
	fd, err := internetSocket(network, laddr, raddr, deadline, syscall.SOCK_RAW, proto, "dial")
	if err != nil {
		return nil, &OpError{Op: "dial", Net: netProto, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
	}
	return newIPConn(fd), nil
}

func listenIP(netProto string, laddr *IPAddr) (*IPConn, error) {
	network, proto, err := parseNetwork(netProto)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: netProto, Source: nil, Addr: laddr.opAddr(), Err: err}
	}
	switch network {
	case "ip", "ip4", "ip6":
	default:
		return nil, &OpError{Op: "listen", Net: netProto, Source: nil, Addr: laddr.opAddr(), Err: UnknownNetworkError(netProto)}
	}
	fd, err := internetSocket(network, laddr, nil, noDeadline, syscall.SOCK_RAW, proto, "listen")
	if err != nil {
		return nil, &OpError{Op: "listen", Net: netProto, Source: nil, Addr: laddr.opAddr(), Err: err}
	}
	return newIPConn(fd), nil
}
