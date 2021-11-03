// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package posix

import (
	"syscall"
	"unsafe"
)

func AsSockaddrInet4(rsa *RawSockaddrAny) syscall.SockaddrInet4 {
	pp := (*syscall.RawSockaddrInet4)(unsafe.Pointer(rsa))
	var sa syscall.SockaddrInet4
	p := (*[2]byte)(unsafe.Pointer(&pp.Port))
	sa.Port = int(p[0])<<8 + int(p[1])
	sa.Addr = pp.Addr
	return sa
}

func AsSockaddrInet6(rsa *RawSockaddrAny) syscall.SockaddrInet6 {
	pp := (*syscall.RawSockaddrInet6)(unsafe.Pointer(rsa))
	var sa syscall.SockaddrInet6
	p := (*[2]byte)(unsafe.Pointer(&pp.Port))
	sa.Port = int(p[0])<<8 + int(p[1])
	sa.ZoneId = pp.Scope_id
	sa.Addr = pp.Addr
	return sa
}

func AsSockaddrUnix(rsa *RawSockaddrAny) (syscall.SockaddrUnix, error) {
	pp := (*syscall.RawSockaddrUnix)(unsafe.Pointer(rsa))
	var sa syscall.SockaddrUnix
	// Assume path ends at NUL.
	// This is not technically the Solaris semantics for
	// abstract Unix domain sockets -- they are supposed
	// to be uninterpreted fixed-size binary blobs -- but
	// everyone uses this convention.
	n := 0
	for n < len(pp.Path) && pp.Path[n] != 0 {
		n++
	}
	bytes := (*[len(pp.Path)]byte)(unsafe.Pointer(&pp.Path[0]))[0:n]
	sa.Name = string(bytes)
	return sa, nil
}

//go:linkname AsSockaddr syscall.anyToSockaddr
func AsSockaddr(rsa *RawSockaddrAny) (syscall.Sockaddr, error) {
	switch rsa.Addr.Family {
	case syscall.AF_UNIX:
		sa, _ := AsSockaddrUnix(rsa)
		return &sa, nil
	case syscall.AF_INET:
		sa := AsSockaddrInet4(rsa)
		return &sa, nil
	case syscall.AF_INET6:
		sa := AsSockaddrInet6(rsa)
		return &sa, nil
	}
	return nil, syscall.EAFNOSUPPORT
}
