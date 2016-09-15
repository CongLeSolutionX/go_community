// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Systems that get peer credentials via GetsockoptXucred.
// +build darwin dragonfly freebsd

package net

import "syscall"

func (c *UnixConn) getPeerCredentials() (creds *UnixPeerCreds, err error) {
	xucred, err := syscall.GetsockoptXucred(c.fd.sysfd, 0, syscall.LOCAL_PEERCRED)
	if err != nil {
		return nil, err
	}

	return &UnixPeerCreds{int(xucred.Cr_uid), int(xucred.Cr_gid)}, nil
}
