// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

func probeTCPStack() (supportsTCPActiveFastOpen bool) {
	fd, err := open("/proc/sys/net/ipv4/tcp_fastopen")
	if err != nil {
		return false
	}
	defer fd.close()
	l, ok := fd.readLine()
	if !ok {
		return false
	}
	f := getFields(l)
	// See Documentation/networking/ip-sysctl.txt.
	n, _, ok := dtoi(f[0], 0)
	if !ok {
		return false
	}
	return n&0x1 != 0
}

func maxListenerBacklog() int {
	fd, err := open("/proc/sys/net/core/somaxconn")
	if err != nil {
		return syscall.SOMAXCONN
	}
	defer fd.close()
	l, ok := fd.readLine()
	if !ok {
		return syscall.SOMAXCONN
	}
	f := getFields(l)
	n, _, ok := dtoi(f[0], 0)
	if n == 0 || !ok {
		return syscall.SOMAXCONN
	}
	// Linux stores the backlog in a uint16.
	// Truncate number to avoid wrapping.
	// See issue 5030.
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return n
}
