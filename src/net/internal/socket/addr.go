// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !nacl,!plan9

// Package socket provides utilities for the manipulation of socket
// facilities.
package socket

import "syscall"

// AddrToRaw takes a syscall.Sockaddr and returns a binary encoding of
// platform-specific socket address.
func AddrToRaw(sa syscall.Sockaddr) []byte {
	return addrToRaw(sa)
}

// RawToAddr takes a binary encoding of platform-specific socket
// address and returns a syscall.Sockaddr.
func RawToAddr(rsa []byte) syscall.Sockaddr {
	return rawToAddr(rsa)
}
