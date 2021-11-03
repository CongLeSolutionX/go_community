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
	sa.Addr = pp.Addr
	return sa
}

func getLen(sa *syscall.RawSockaddrUnix) (int, error) {
	// Some versions of AIX have a bug in getsockname (see IV78655).
	// We can't rely on sa.Len being set correctly.
	n := syscall.SizeofSockaddrUnix - 3 // subtract leading Family, Len, terminating NUL.
	for i := 0; i < n; i++ {
		if sa.Path[i] == 0 {
			n = i
			break
		}
	}
	return n, nil
}

func AsSockaddrUnix(rsa *RawSockaddrAny) (syscall.SockaddrUnix, error) {
	pp := (*syscall.RawSockaddrUnix)(unsafe.Pointer(rsa))
	n, err := getLen(pp)
	if err != nil {
		return syscall.SockaddrUnix{}, err
	}
	bytes := (*[len(pp.Path)]byte)(unsafe.Pointer(&pp.Path[0]))
	var sa syscall.SockaddrUnix
	sa.Name = string(bytes[0:n])
	return sa, nil
}

//go:linkname AsSockaddr syscall.anyToSockaddr
func AsSockaddr(rsa *RawSockaddrAny) (syscall.Sockaddr, error) {
	switch rsa.Addr.Family {
	case syscall.AF_UNIX:
		sa, err := AsSockaddrUnix(rsa)
		if err != nil {
			return nil, err
		}
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
