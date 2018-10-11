// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"os"
	"syscall"
)

func newSyscallConn(network string, f *os.File) (*SyscallConn, error) {
	s := int(f.Fd())
	if err := syscall.SetNonblock(s, true); err != nil {
		return nil, err
	}
	var err error
	var c SyscallConn
	c.c.fd, err = newFD(s, 0, 0, network)
	if err != nil {
		return nil, err
	}
	if err := c.c.fd.init(); err != nil {
		return nil, err
	}
	return &c, nil
}
