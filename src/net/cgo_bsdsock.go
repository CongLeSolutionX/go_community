// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !netgo
// +build darwin dragonfly freebsd netbsd openbsd

package net

/*
#include <sys/types.h>
#include <sys/socket.h>

#include <netinet/in.h>
*/
import "C"

import (
	"syscall"
	"unsafe"
)

func cgoSockaddrInet4(ip IP) *C.struct_sockaddr {
	sa4 := syscall.RawSockaddrInet4{
		Len:    syscall.SizeofSockaddrInet4,
		Family: syscall.AF_INET,
	}
	copy(sa4.Addr[:], ip)
	return (*C.struct_sockaddr)(unsafe.Pointer(&sa4))
}

func cgoSockaddrInet6(ip IP) *C.struct_sockaddr {
	sa6 := syscall.RawSockaddrInet6{
		Len:    syscall.SizeofSockaddrInet6,
		Family: syscall.AF_INET6,
	}
	copy(sa6.Addr[:], ip)
	return (*C.struct_sockaddr)(unsafe.Pointer(&sa6))
}
