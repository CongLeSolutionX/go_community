// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris windows

package net

import "syscall"

func setNoDelay(fd *netFD, noDelay bool) error {
	if err := fd.incref(); err != nil {
		return err
	}
	defer fd.decref()
	return syscall.SetsockoptInt(fd.sysfd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, boolint(noDelay))
}
