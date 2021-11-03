// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

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
	if pp.Len < 2 || pp.Len > syscall.SizeofSockaddrUnix {
		return syscall.SockaddrUnix{}, syscall.EINVAL
	}

	// Some BSDs include the trailing NUL in the length, whereas
	// others do not. Work around this by subtracting the leading
	// family and len. The path is then scanned to see if a NUL
	// terminator still exists within the length.
	n := int(pp.Len) - 2 // subtract leading Family, Len
	for i := 0; i < n; i++ {
		if pp.Path[i] == 0 {
			// found early NUL; assume Len included the NUL
			// or was overestimating.
			n = i
			break
		}
	}
	bytes := (*[len(pp.Path)]byte)(unsafe.Pointer(&pp.Path[0]))[0:n]
	var sa syscall.SockaddrUnix
	sa.Name = string(bytes)
	return sa, nil
}

//go:linkname AsSockaddr syscall.anyToSockaddr
func AsSockaddr(rsa *RawSockaddrAny) (syscall.Sockaddr, error) {
	switch rsa.Addr.Family {
	case syscall.AF_LINK:
		pp := (*syscall.RawSockaddrDatalink)(unsafe.Pointer(rsa))
		sa := new(syscall.SockaddrDatalink)
		sa.Len = pp.Len
		sa.Family = pp.Family
		sa.Index = pp.Index
		sa.Type = pp.Type
		sa.Nlen = pp.Nlen
		sa.Alen = pp.Alen
		sa.Slen = pp.Slen
		sa.Data = pp.Data
		return sa, nil
	case syscall.AF_UNIX:
		sa, err := AsSockaddrUnix(rsa)
		if err != nil {
			return nil, err
		}
		return &sa, err
	case syscall.AF_INET:
		sa := AsSockaddrInet4(rsa)
		return &sa, nil
	case syscall.AF_INET6:
		sa := AsSockaddrInet6(rsa)
		return &sa, nil
	}
	return nil, syscall.EAFNOSUPPORT
}
