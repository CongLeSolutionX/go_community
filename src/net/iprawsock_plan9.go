// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"syscall"
	"time"
)

func (c *IPConn) readFromIP(b []byte) (int, *IPAddr, error) {
	return 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *IPConn) readMsgIP(b, oob []byte) (n, oobn, flags int, addr *IPAddr, err error) {
	return 0, 0, 0, nil, &OpError{Op: "read", Net: c.fd.dir, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EPLAN9}
}

func (c *IPConn) writeToIP(b []byte, addr *IPAddr) (int, error) {
	return 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: syscall.EPLAN9}
}

func (c *IPConn) writeMsgIP(b, oob []byte, addr *IPAddr) (n, oobn int, err error) {
	return 0, 0, &OpError{Op: "write", Net: c.fd.dir, Source: c.fd.laddr, Addr: addr, Err: syscall.EPLAN9}
}

func dialIP(netProto string, laddr, raddr *IPAddr, deadline time.Time) (*IPConn, error) {
	return nil, &OpError{Op: "dial", Net: netProto, Source: laddr, Addr: raddr, Err: syscall.EPLAN9}
}

func listenIP(netProto string, laddr *IPAddr) (*IPConn, error) {
	return nil, &OpError{Op: "listen", Net: netProto, Source: nil, Addr: laddr, Err: syscall.EPLAN9}
}
