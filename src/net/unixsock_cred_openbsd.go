// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

func (c *UnixConn) getPeerCredentials() (creds *UnixPeerCreds, err error) {
	cred, err := syscall.GetsockoptSockpeercred(c.fd.sysfd, syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	if err != nil {
		return nil, err
	}

	return &UnixPeerCreds{int(cred.Uid), int(cred.Gid), int(cred.Pid)}, nil
}
