// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"syscall"
	"unsafe"
)

const (
	offsetofRawSockaddrInet4 = unsafe.Offsetof(syscall.SockaddrInet4{}.Addr) + unsafe.Sizeof(syscall.SockaddrInet4{}.Addr)
	offsetofRawSockaddrInet6 = unsafe.Offsetof(syscall.SockaddrInet6{}.Addr) + unsafe.Sizeof(syscall.SockaddrInet6{}.Addr)
)

func addrToRaw(sa syscall.Sockaddr) []byte {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&b[offsetofRawSockaddrInet4]))
		rsa.Family = syscall.AF_INET
		b[offsetofRawSockaddrInet4+2] = byte(sa.Port >> 8)
		b[offsetofRawSockaddrInet4+3] = byte(sa.Port)
		copy(rsa.Addr[:], sa.Addr[:])
		return b[offsetofRawSockaddrInet4:]
	case *syscall.SockaddrInet6:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&b[offsetofRawSockaddrInet6]))
		rsa.Family = syscall.AF_INET6
		b[offsetofRawSockaddrInet6+2] = byte(sa.Port >> 8)
		b[offsetofRawSockaddrInet6+3] = byte(sa.Port)
		copy(rsa.Addr[:], sa.Addr[:])
		rsa.Scope_id = uint32(sa.ZoneId)
		return b[offsetofRawSockaddrInet6:]
	default:
		return nil
	}
}

func rawToAddr(rsa []byte) syscall.Sockaddr {
	switch rsa[0] {
	case syscall.AF_INET:
		rsa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&rsa[0]))
		sa := &syscall.SockaddrInet4{Port: int(rsa.Port)}
		copy(sa.Addr[:], rsa.Addr[:])
		return sa
	case syscall.AF_INET6:
		rsa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&rsa[0]))
		sa := &syscall.SockaddrInet6{Port: int(rsa.Port)}
		copy(sa.Addr[:], rsa.Addr[:])
		sa.ZoneId = uint32(rsa.Scope_id)
		return sa
	default:
		return nil
	}
}
