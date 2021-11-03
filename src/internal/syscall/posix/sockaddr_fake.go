// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build js

package posix

import "syscall"

type RawSockaddrAny struct{}

func AsSockaddrInet4(rsa *RawSockaddrAny) syscall.SockaddrInet4 {
	return syscall.SockaddrInet4{}
}

func AsSockaddrInet6(rsa *RawSockaddrAny) syscall.SockaddrInet6 {
	return syscall.SockaddrInet6{}
}

func AsSockaddrUnix(rsa *RawSockaddrAny) (syscall.SockaddrUnix, error) {
	return syscall.SockaddrUnix{}, syscall.ENOSYS
}

func AsSockaddr(rsa *RawSockaddrAny) (syscall.Sockaddr, error) {
	return nil, syscall.ENOSYS
}
