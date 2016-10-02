// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd

package net

import (
	"os"
	"syscall"
	"unsafe"
)

func (c *conn) readBuffers(v *Buffers) (int64, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	n, err := c.fd.readBuffers(v)
	if err != nil {
		return n, &OpError{Op: "readv", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return n, nil
}

func (fd *netFD) readBuffers(v *Buffers) (n int64, err error) {
	if err := fd.readLock(); err != nil {
		return 0, err
	}
	defer fd.readUnlock()
	if err := fd.pd.prepareRead(); err != nil {
		return 0, err
	}

	var iovecs []syscall.Iovec
	if fd.iovecs != nil {
		iovecs = *fd.iovecs
	}
	// TODO: read from sysconf(_SC_IOV_MAX)? The Linux default is
	// 1024 and this seems conservative enough for now. Darwin's
	// UIO_MAXIOV also seems to be 1024.
	maxVec := 1024

	vc := *v
	pos := 0

	for len(vc) > 0 {
		iovecs = iovecs[:0]
		for i, chunk := range vc {
			if len(chunk) == 0 {
				continue
			}
			if i == 0 { // subtract pos bytes for first chunk
				iovecs = append(iovecs, syscall.Iovec{Base: &chunk[pos]})
				iovecs[len(iovecs)-1].SetLen(len(chunk) - pos)
			} else {
				iovecs = append(iovecs, syscall.Iovec{Base: &chunk[0]})
				iovecs[len(iovecs)-1].SetLen(len(chunk))
			}
			if len(iovecs) == maxVec {
				break
			}
		}
		if len(iovecs) == 0 {
			break
		}
		fd.iovecs = &iovecs // cache

		read, _, e0 := syscall.Syscall(syscall.SYS_READV,
			uintptr(fd.sysfd),
			uintptr(unsafe.Pointer(&iovecs[0])),
			uintptr(len(iovecs)))
		if read < 0 || read == ^uintptr(0) {
			read = 0
		}
		testHookDidReadv(int(read))
		n += int64(read)
		for read > 0 {
			if len(vc[0])-pos <= int(read) {
				pos, read = 0, read-uintptr(len(vc[0]))
				vc = vc[1:]
			} else {
				pos, read = int(read), 0
			}
		}
		if e0 == syscall.EAGAIN {
			if err = fd.pd.waitRead(); err == nil {
				continue
			}
		} else if e0 != 0 {
			err = syscall.Errno(e0)
		}
		if err != nil {
			break
		}
	}
	if _, ok := err.(syscall.Errno); ok {
		err = os.NewSyscallError("readv", err)
	}
	return n, err
}
