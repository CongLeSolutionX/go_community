// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"io"
	"os"
	"syscall"
	"time"
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
func (c *TCPConn) Write(b []byte) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	err := c.fd.openLock()
	if c.fd.isConnected || c.fastOpen == nil {
		if err != nil {
			return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		}
		defer c.fd.writeUnlock()
		if err := c.fd.pd.prepareWrite(); err != nil {
			return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		}
		n, err := c.fd.write(b)
		if err != nil {
			err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		}
		return n, err
	}
	c.fastOpen.finalDeadline = c.fastOpen.Dialer.deadline(time.Now())
	n, err := dialSerialTCPFastOpen(c, b)
	if err != nil {
		c.fd.writeUnlock()
	} else {
		c.fd.openUnlock()
	}
	return n, err
}

// dialSerialTCPFastOpen connects to a list of addresses in sequence,
// returning either the first successful connection, or the first
// error.
func dialSerialTCPFastOpen(c *TCPConn, b []byte) (int, error) {
	laddr, _ := c.fastOpen.Dialer.LocalAddr.(*TCPAddr)
	if laddr == nil {
		laddr = &TCPAddr{}
	}
	var firstErr error
	for i, raddr := range c.fastOpen.addrs {
		raddr, ok := raddr.(*TCPAddr)
		if !ok {
			continue
		}
		select {
		case <-c.fastOpen.Dialer.Cancel:
			return 0, &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: errCanceled}
		default:
		}
		partialDeadline, err := partialDeadline(time.Now(), c.fastOpen.finalDeadline, len(c.fastOpen.addrs)-i)
		if err != nil {
			if firstErr == nil {
				firstErr = &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
			}
			break
		}
		nfd, n, err := dialTCPFastOpen(c.fastOpen.network, b, laddr, raddr, partialDeadline, c.fastOpen.Dialer.Cancel)
		if err != nil {
			if firstErr == nil {
				firstErr = &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
			}
			continue
		}
		c.fd.subst(nfd)
		if c.fastOpen.Dialer.KeepAlive > 0 {
			setKeepAlive(c.fd, true)
			setKeepAlivePeriod(c.fd, c.fastOpen.Dialer.KeepAlive)
			testHookSetKeepAlive()
		}
		c.fastOpen = nil
		return n, nil
	}
	if firstErr == nil {
		firstErr = &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: nil, Err: errMissingAddress}
	}
	return 0, firstErr
}

func dialTCPFastOpen(network string, b []byte, laddr, raddr *TCPAddr, deadline time.Time, cancel <-chan struct{}) (*netFD, int, error) {
	fd, err := internetSocket(network, laddr, nil, syscall.SOCK_STREAM, 0, "dial", true, noDeadline, noCancel)
	if err == nil {
		fd.raddr = raddr
	}
	for i := 0; i < 2 && laddr.Port == 0 && tcpSelfConnect(fd, err); i++ {
		if err == nil {
			fd.Close()
		}
		fd, err = internetSocket(network, laddr, nil, syscall.SOCK_STREAM, 0, "dial", true, noDeadline, noCancel)
		if err == nil {
			fd.raddr = raddr
		}
	}
	if err != nil {
		return nil, 0, err
	}
	sa, err := raddr.sockaddr(fd.family)
	if err != nil {
		fd.Close()
		return nil, 0, err
	}
	if !deadline.IsZero() {
		fd.setWriteDeadline(deadline)
		defer fd.setWriteDeadline(noDeadline)
	}
	if cancel != nil {
		done := make(chan struct{})
		defer func() {
			// This is unbuffered; wait for the goroutine
			// before returning.
			println("fire")
			done <- struct{}{}
		}()
		go func() {
			select {
			case <-cancel:
				// Force the runtime's poller to
				// immediately give up waiting for
				// writability.
				println("cancel")
				fd.setWriteDeadline(aLongTimeAgo)
				<-done
			case <-done:
				println("done")
			}
		}()
	}
	n, err := fd.writeMsgFastOpen(b, sa, raddr, cancel)
	if err != nil {
		fd.Close()
		return nil, n, err
	}
	return fd, n, nil
}

func (fd *netFD) writeMsgFastOpen(b []byte, sa syscall.Sockaddr, addr sockaddr, cancel <-chan struct{}) (nn int, err error) {
	if err := fd.writeLock(); err != nil {
		return 0, err
	}
	defer fd.writeUnlock()
	if err := fd.pd.prepareWrite(); err != nil {
		return 0, err
	}
	for {
		var n int
		n, err = syscall.SendmsgN(fd.sysfd, b, nil, sa, syscall.MSG_FASTOPEN)
		switch err {
		case nil, syscall.EISCONN:
			fd.isConnected = true
			if rsa, _ := syscall.Getpeername(fd.sysfd); rsa != nil {
				fd.raddr = fd.addrFunc()(rsa)
			} else {
				fd.raddr = addr
			}
			if n > 0 {
				nn += n
			}
			if nn < len(b) {
				n, err = fd.write(b[n:])
				if n > 0 {
					nn += n
				}
			}
		case syscall.EAGAIN, syscall.EALREADY, syscall.EINPROGRESS, syscall.EINTR:
			if err = fd.pd.waitWrite(); err == nil {
				continue
			}
			select {
			case <-cancel:
				err = errCanceled
			default:
			}
		}
		if n == 0 && err == nil {
			err = io.ErrUnexpectedEOF
		}
		break
	}
	if _, ok := err.(syscall.Errno); ok {
		err = os.NewSyscallError("sendmsg", err)
	}
	return
}
