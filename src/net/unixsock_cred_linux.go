// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

func (c *UnixConn) getPeerCredentials() (creds *UnixPeerCreds, err error) {
	ucred, err := syscall.GetsockoptUcred(c.fd.sysfd, syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, err
	}

	return &UnixPeerCreds{int(ucred.Uid), int(ucred.Gid), int(ucred.Pid)}, nil
}
