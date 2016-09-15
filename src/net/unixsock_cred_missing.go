// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Systems that have no way of getting peer credentials.
// +build plan9 windows

package net

import "syscall"

func (c *UnixConn) getPeerCredentials() (creds *UnixPeerCreds, err error) {
	return nil, syscall.EINVAL
}
