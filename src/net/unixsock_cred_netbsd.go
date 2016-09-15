// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

func (c *UnixConn) getPeerCredentials() (creds *UnixPeerCreds, err error) {
	cred, err := syscall.GetsockoptUnpcbid(c.fd.sysfd, 0, syscall.LOCAL_PEEREID)
	if err != nil {
		return nil, err
	}

	return &UnixPeerCreds{int(cred.Unp_euid), int(cred.Unp_egid)}, nil
}
