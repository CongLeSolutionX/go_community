// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"io"
	"os"
	"syscall"
)

func probeTCPStack() (supportsPassiveTCPFastOpen, supportsActiveTCPFastOpen bool) {
	// See Documentation/networking/ip-sysctl.txt.
	fd, err := open("/proc/sys/net/ipv4/tcp_fastopen")
	if err != nil {
		return
	}
	defer fd.close()
	l, ok := fd.readLine()
	if !ok {
		return
	}
	f := getFields(l)
	n, _, ok := dtoi(f[0], 0)
	if !ok {
		return
	}
	if n&0x1 != 0 {
		supportsActiveTCPFastOpen = true
	}
	if n&0x2 != 0 {
		supportsPassiveTCPFastOpen = true
	}
	return
}

// Write implements the Conn Write method.
func (c *TCPConn) Write(b []byte) (n int, err error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	if err := c.fd.writeLock(); err != nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	defer c.fd.writeUnlock()
	if err := c.fd.pd.PrepareWrite(); err != nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	if !c.fastOpen || c.fd.isConnected {
		n, err = c.fd.write(b)
		if err != nil {
			err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		}
		return
	}
	a, ok := c.fd.raddr.(*TCPAddr)
	if !ok {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: syscall.EINVAL}
	}
	sa, err := a.sockaddr(c.fd.family)
	if err != nil {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	for {
		var m int
		m, err = syscall.SendmsgN(c.fd.sysfd, b, nil, sa, syscall.MSG_FASTOPEN)
		switch err {
		case nil, syscall.EISCONN:
			c.fd.isConnected = true
			if m > 0 {
				n += m
			}
			if n < len(b) {
				m, err = c.fd.write(b[n:])
				if m > 0 {
					n += m
				}
			}
		case syscall.EAGAIN, syscall.EALREADY, syscall.EINPROGRESS, syscall.EINTR:
			if err = c.fd.pd.WaitWrite(); err == nil {
				continue
			}
		}
		if m == 0 {
			err = io.ErrUnexpectedEOF
		}
		break
	}
	if err != nil {
		if _, ok := err.(syscall.Errno); ok {
			err = os.NewSyscallError("sendmsg", err)
		}
		err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return
}
