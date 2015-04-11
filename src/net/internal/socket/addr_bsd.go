// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package socket

import (
	"syscall"
	"unsafe"
)

const (
	offsetofRawSockaddrDatalink = unsafe.Offsetof(syscall.SockaddrDatalink{}.Data) + unsafe.Sizeof(syscall.SockaddrDatalink{}.Data)
	offsetofRawSockaddrInet4    = unsafe.Offsetof(syscall.SockaddrInet4{}.Addr) + unsafe.Sizeof(syscall.SockaddrInet4{}.Addr)
	offsetofRawSockaddrInet6    = unsafe.Offsetof(syscall.SockaddrInet6{}.Addr) + unsafe.Sizeof(syscall.SockaddrInet6{}.Addr)
	offsetofRawSockaddrUnix     = unsafe.Sizeof(syscall.SockaddrUnix{}.Name)
)

func addrToRaw(sa syscall.Sockaddr) []byte {
	switch sa := sa.(type) {
	case *syscall.SockaddrDatalink:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrDatalink)(unsafe.Pointer(&b[offsetofRawSockaddrDatalink]))
		rsa.Len = syscall.SizeofSockaddrDatalink
		rsa.Family = syscall.AF_LINK
		rsa.Index = sa.Index
		rsa.Type = sa.Type
		rsa.Nlen = sa.Nlen
		rsa.Alen = sa.Alen
		rsa.Slen = sa.Slen
		copy(rsa.Data[:], sa.Data[:])
		return b[offsetofRawSockaddrDatalink:]
	case *syscall.SockaddrInet4:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(&b[offsetofRawSockaddrInet4]))
		rsa.Len = syscall.SizeofSockaddrInet4
		rsa.Family = syscall.AF_INET
		b[offsetofRawSockaddrInet4+2] = byte(sa.Port >> 8)
		b[offsetofRawSockaddrInet4+3] = byte(sa.Port)
		copy(rsa.Addr[:], sa.Addr[:])
		return b[offsetofRawSockaddrInet4:]
	case *syscall.SockaddrInet6:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(&b[offsetofRawSockaddrInet6]))
		rsa.Len = syscall.SizeofSockaddrInet4
		rsa.Family = syscall.AF_INET6
		b[offsetofRawSockaddrInet6+2] = byte(sa.Port >> 8)
		b[offsetofRawSockaddrInet6+3] = byte(sa.Port)
		copy(rsa.Addr[:], sa.Addr[:])
		rsa.Scope_id = uint32(sa.ZoneId)
		return b[offsetofRawSockaddrInet6:]
	case *syscall.SockaddrUnix:
		b := (*[unsafe.Sizeof(*sa)]byte)(unsafe.Pointer(sa))
		rsa := (*syscall.RawSockaddrUnix)(unsafe.Pointer(&b[offsetofRawSockaddrUnix]))
		rsa.Len = syscall.SizeofSockaddrUnix
		rsa.Family = syscall.AF_UNIX
		for i := 0; len(rsa.Path) > i && i < len(sa.Name); i++ {
			rsa.Path[i] = int8(sa.Name[i])
		}
		return b[offsetofRawSockaddrUnix:]
	default:
		return nil
	}
}

func rawToAddr(rsa []byte) syscall.Sockaddr {
	switch rsa[1] {
	case syscall.AF_LINK:
		rsa := (*syscall.RawSockaddrDatalink)(unsafe.Pointer(&rsa[0]))
		sa := &syscall.SockaddrDatalink{
			Len:    rsa.Len,
			Family: rsa.Family,
			Index:  rsa.Index,
			Type:   rsa.Type,
			Nlen:   rsa.Nlen,
			Alen:   rsa.Alen,
			Slen:   rsa.Slen,
		}
		copy(sa.Data[:], rsa.Data[:])
		return sa
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
	case syscall.AF_UNIX:
		rsa := (*syscall.RawSockaddrUnix)(unsafe.Pointer(&rsa[0]))
		sa := &syscall.SockaddrUnix{}
		n := 0
		for n < len(rsa.Path) && rsa.Path[n] != 0 {
			n++
		}
		name := (*[10000]byte)(unsafe.Pointer(&rsa.Path[0]))[:n]
		sa.Name = string(name)
		return sa
	default:
		return nil
	}
}
