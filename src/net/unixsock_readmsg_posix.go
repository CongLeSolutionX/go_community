// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd netbsd openbsd solaris

package net

import (
	"syscall"
)

func (c *UnixConn) readMsg(b, oob []byte) (n, oobn, flags int, addr *UnixAddr, err error) {
	var sa syscall.Sockaddr
	n, oobn, flags, sa, err = c.fd.readMsg(b, oob, 0)
	if oobn > 0 {
		setCloseOnExec(oob[:oobn])
	}

	switch sa := sa.(type) {
	case *syscall.SockaddrUnix:
		if sa.Name != "" {
			addr = &UnixAddr{Name: sa.Name, Net: sotypeToNet(c.fd.sotype)}
		}
	}
	return
}

func setCloseOnExec(oob []byte) {
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return
	}

	for _, scm := range scms {
		if scm.Header.Level == syscall.SOL_SOCKET &&
			scm.Header.Type == syscall.SCM_RIGHTS {
			fds, err := syscall.ParseUnixRights(&scm)
			if err != nil {
				continue
			}
			for _, fd := range fds {
				syscall.CloseOnExec(fd)
			}
		}
	}
}
