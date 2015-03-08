// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"context"
	"internal/nettrace"
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
	subCtx, cancel, relayCancel := c.fastOpen.Dialer.extractDialContext(c.fastOpen.Context)
	if cancel != nil {
		defer cancel()
	}
	if relayCancel != nil {
		go relayCancel()
	}
	if subCtx != nil {
		c.fastOpen.Context = subCtx
	}
	c.fastOpen.addrs, err = resolveAddrList(c.fastOpen.Context, "dial", c.fastOpen.network, c.fastOpen.address, c.fastOpen.Dialer.LocalAddr)
	if err != nil {
		return 0, &OpError{Op: "write", Net: c.fastOpen.network, Source: nil, Addr: nil, Err: err}
	}
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
		case <-c.fastOpen.Context.Done():
			return 0, &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: mapErr(c.fastOpen.Context.Err())}
		default:
		}
		deadline, _ := c.fastOpen.Context.Deadline()
		partialDeadline, err := partialDeadline(time.Now(), deadline, len(c.fastOpen.addrs)-i)
		if err != nil {
			if firstErr == nil {
				firstErr = &OpError{Op: "write", Net: c.fastOpen.network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: err}
			}
			break
		}
		dialCtx := c.fastOpen.Context
		if partialDeadline.Before(deadline) {
			var cancel context.CancelFunc
			dialCtx, cancel = context.WithDeadline(c.fastOpen.Context, partialDeadline)
			defer cancel()
		}
		if laddr.IP == nil || laddr.IP.IsUnspecified() {
			if raddr.IP.To4() != nil {
				laddr.IP = IPv4zero
			}
			if raddr.IP.To16() != nil && raddr.IP.To4() == nil {
				laddr.IP = IPv6unspecified
			}
		}
		nfd, n, err := dialTCPFastOpen(dialCtx, c.fastOpen.dialParam, laddr, raddr, b)
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

func dialTCPFastOpen(ctx context.Context, dp *dialParam, laddr, raddr *TCPAddr, b []byte) (fd *netFD, n int, err error) {
	trace, _ := ctx.Value(nettrace.TraceKey{}).(*nettrace.Trace)
	if trace != nil {
		raddrStr := raddr.String()
		if trace.ConnectStart != nil {
			trace.ConnectStart(dp.network, raddrStr)
		}
		if trace.ConnectDone != nil {
			defer func() { trace.ConnectDone(dp.network, raddrStr, err) }()
		}
	}
	fd, err = internetSocket(ctx, dp.network, laddr, nil, syscall.SOCK_STREAM, 0, "dial", true)
	if err == nil {
		fd.raddr = raddr
	}
	for i := 0; i < 2 && laddr.Port == 0 && tcpSelfConnect(fd, err); i++ {
		if err == nil {
			fd.Close()
			fd = nil
		}
		fd, err = internetSocket(ctx, dp.network, laddr, nil, syscall.SOCK_STREAM, 0, "dial", true)
		if err == nil {
			fd.raddr = raddr
		}
	}
	if err != nil {
		return
	}
	var sa syscall.Sockaddr
	sa, err = raddr.sockaddr(fd.family)
	if err != nil {
		fd.Close()
		fd = nil
		return
	}
	if deadline, _ := ctx.Deadline(); !deadline.IsZero() {
		fd.setWriteDeadline(deadline)
		defer fd.setWriteDeadline(noDeadline)
	}
	done := make(chan bool)
	defer func() { done <- true }()
	go func() {
		select {
		case <-ctx.Done():
			// Force the runtime's poller to immediately
			// give up waiting for writability.
			fd.setWriteDeadline(aLongTimeAgo)
			<-done
		case <-done:
		}
	}()
	n, err = fd.writeMsgFastOpen(ctx, sa, raddr, b)
	if err != nil {
		fd.Close()
	}
	return
}

func (fd *netFD) writeMsgFastOpen(ctx context.Context, sa syscall.Sockaddr, addr sockaddr, b []byte) (nn int, err error) {
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
			case <-ctx.Done():
				err = mapErr(ctx.Err())
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
