// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package net

import "syscall"

func (fd *netFD) openRead(b []byte, flags int) (int, error) {
	if err := fd.readLock(); err != nil {
		return 0, err
	}
	defer fd.readUnlock()
	if err := fd.pd.PrepareRead(); err != nil {
		return 0, err
	}
	for {
		switch n, err := syscall.Read(fd.sysfd, b); err {
		case nil:
			return n, fd.eofError(n, err)
		case syscall.ENOTCONN:
			if flags&syscall.MSG_FASTOPEN == 0 {
				return 0, err
			}
			fallthrough
		case syscall.EAGAIN:
			if err := fd.pd.WaitRead(); err == nil {
				continue
			}
			fallthrough
		default:
			return 0, err
		}
	}
}

func (fd *netFD) openWrite(b []byte, flags int) (int, error) {
	if fd.isConnected {
		return fd.write(b)
	}
	if err := fd.writeLock(); err != nil {
		return 0, err
	}
	defer fd.writeUnlock()
	if err := fd.pd.PrepareWrite(); err != nil {
		return 0, err
	}
	raddr, ok := fd.raddr.(*TCPAddr)
	if !ok {
		return 0, err
	}
	sa, err := raddr.sockaddr(fd.family)
	if err != nil {
		return 0, err
	}
	for {
		switch n, err := syscall.SendmsgN(fd.sysfd, b, nil, sa, flags); err {
		case nil:
			fd.isConnected = true
			return n, nil
		case syscall.EISCONN:
			fd.isConnected = true
			return fd.write(b)
		case syscall.EAGAIN, syscall.EALREADY, syscall.EINPROGRESS, syscall.EINTR:
			if err := fd.pd.WaitWrite(); err == nil {
				continue
			}
			fallthrough
		default:
			return 0, err
		}
	}
}
