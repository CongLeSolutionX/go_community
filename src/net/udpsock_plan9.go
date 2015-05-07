// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"errors"
	"os"
	"syscall"
	"time"
)

func newUDPConn(fd *netFD) *UDPConn { return &UDPConn{conn{fd}} }

func (c *UDPConn) readFromUDP(b []byte) (n int, addr *UDPAddr, err error) {
	if c.fd.data == nil {
		return 0, nil, syscall.EINVAL
	}
	buf := make([]byte, udpHeaderSize+len(b))
	m, err := c.fd.data.Read(buf)
	if err != nil {
		return 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	if m < udpHeaderSize {
		return 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: errors.New("short read reading UDP header")}
	}
	buf = buf[:m]

	h, buf := unmarshalUDPHeader(buf)
	n = copy(b, buf)
	return n, &UDPAddr{IP: h.raddr, Port: int(h.rport)}, nil
}

func (c *UDPConn) readMsgUDP(b, oob []byte) (n, oobn, flags int, addr *UDPAddr, err error) {
	return 0, 0, 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *UDPConn) writeToUDP(b []byte, addr *UDPAddr) (int, error) {
	if c.fd.data == nil {
		return 0, syscall.EINVAL
	}
	if addr == nil {
		return 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: errMissingAddress}
	}
	h := new(udpHeader)
	h.raddr = addr.IP.To16()
	h.laddr = c.fd.laddr.(*UDPAddr).IP.To16()
	h.ifcaddr = IPv6zero // ignored (receive only)
	h.rport = uint16(addr.Port)
	h.lport = uint16(c.fd.laddr.(*UDPAddr).Port)

	buf := make([]byte, udpHeaderSize+len(b))
	i := copy(buf, h.Bytes())
	copy(buf[i:], b)
	if _, err := c.fd.data.Write(buf); err != nil {
		return 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: err}
	}
	return len(b), nil
}

func (c *UDPConn) writeMsgUDP(b, oob []byte, addr *UDPAddr) (n, oobn int, err error) {
	return 0, 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: syscall.EPLAN9}
}

func dialUDP(net string, laddr, raddr *UDPAddr, deadline time.Time) (*UDPConn, error) {
	if !deadline.IsZero() {
		panic("net.dialUDP: deadline not implemented on Plan 9")
	}
	fd, err := dialPlan9(net, laddr, raddr)
	if err != nil {
		return nil, err
	}
	return newUDPConn(fd), nil
}

const udpHeaderSize = 16*3 + 2*2

type udpHeader struct {
	raddr, laddr, ifcaddr IP
	rport, lport          uint16
}

func (h *udpHeader) Bytes() []byte {
	b := make([]byte, udpHeaderSize)
	i := 0
	i += copy(b[i:i+16], h.raddr)
	i += copy(b[i:i+16], h.laddr)
	i += copy(b[i:i+16], h.ifcaddr)
	b[i], b[i+1], i = byte(h.rport>>8), byte(h.rport), i+2
	b[i], b[i+1], i = byte(h.lport>>8), byte(h.lport), i+2
	return b
}

func unmarshalUDPHeader(b []byte) (*udpHeader, []byte) {
	h := new(udpHeader)
	h.raddr, b = IP(b[:16]), b[16:]
	h.laddr, b = IP(b[:16]), b[16:]
	h.ifcaddr, b = IP(b[:16]), b[16:]
	h.rport, b = uint16(b[0])<<8|uint16(b[1]), b[2:]
	h.lport, b = uint16(b[0])<<8|uint16(b[1]), b[2:]
	return h, b
}

func listenUDP(net string, laddr *UDPAddr) (*UDPConn, error) {
	l, err := listenPlan9(net, laddr)
	if err != nil {
		return nil, err
	}
	_, err = l.ctl.WriteString("headers")
	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: err}
	}
	l.data, err = os.OpenFile(l.dir+"/data", os.O_RDWR, 0)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: err}
	}
	fd, err := l.netFD()
	return newUDPConn(fd), err
}

func listenMulticastUDP(net string, ifi *Interface, gaddr *UDPAddr) (*UDPConn, error) {
	return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: gaddr, Err: syscall.EPLAN9}
}
