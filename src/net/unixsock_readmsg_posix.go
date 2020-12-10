// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd netbsd openbsd solaris

package net

import (
	"syscall"
)

func (c *UnixConn) readMsg(b, oob []byte) (n, oobn, flags int, addr *UnixAddr, err error) {
	var sa syscall.Sockaddr
	n, oobn, flags, sa, err = c.fd.readMsg(b, oob)
	if oobn > 0 {
		scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil {
			return 0, 0, 0, nil, err
		}

		for _, scm := range scms {
			fds, err := syscall.ParseUnixRights(&scm)
			if err != nil {
				return 0, 0, 0, nil, err
			}
			for _, fd := range fds {
				syscall.CloseOnExec(fd)
				old_flags, _, err := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd),
					uintptr(syscall.F_GETFL), 0)
				if err != nil {
					return 0, 0, 0, nil, err
				}

				if old_flags&syscall.FD_CLOEXEC == 0 {
					// Fail to set close-on-exec flag
					return 0, 0, 0, nil, EINVAL
				}

			}
		}
	}
	switch sa := sa.(type) {
	case *syscall.SockaddrUnix:
		if sa.Name != "" {
			addr = &UnixAddr{Name: sa.Name, Net: sotypeToNet(c.fd.sotype)}
		}
	}
	return
}
