// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build !nacl,!plan9,!solaris,!windows

package socket

/*
#include <sys/types.h>
#include <sys/socket.h>

#include <netinet/in.h>
#include <netdb.h>

#include <stdlib.h>
*/
import "C"

import (
	"syscall"
	"unsafe"
)

// GetaddrinfoPort returns a service port record.
// Network must be "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6".
func GetaddrinfoPort(network, service string) (int, error) {
	var hints C.struct_addrinfo
	switch network {
	case "": // no hints
	case "tcp", "tcp4", "tcp6":
		hints.ai_socktype = C.SOCK_STREAM
		hints.ai_protocol = C.IPPROTO_TCP
	case "udp", "udp4", "udp6":
		hints.ai_socktype = C.SOCK_DGRAM
		hints.ai_protocol = C.IPPROTO_UDP
	default:
		return 0, syscall.EINVAL
	}
	if len(network) >= 4 {
		switch network[3] {
		case '4':
			hints.ai_family = C.AF_INET
		case '6':
			hints.ai_family = C.AF_INET6
		}
	}
	s := C.CString(service)
	defer C.free(unsafe.Pointer(s))
	var res *C.struct_addrinfo
	errno, _ := C.getaddrinfo(nil, s, &hints, &res)
	if errno != 0 {
		return 0, AddrinfoErrno(errno)
	}
	defer C.freeaddrinfo(res)
	for r := res; r != nil; r = r.ai_next {
		switch r.ai_family {
		case C.AF_INET:
			sa := (*syscall.RawSockaddrInet4)(unsafe.Pointer(r.ai_addr))
			p := (*[2]byte)(unsafe.Pointer(&sa.Port))
			return int(p[0])<<8 | int(p[1]), nil
		case C.AF_INET6:
			sa := (*syscall.RawSockaddrInet6)(unsafe.Pointer(r.ai_addr))
			p := (*[2]byte)(unsafe.Pointer(&sa.Port))
			return int(p[0])<<8 | int(p[1]), nil
		}
	}
	return 0, nil
}

// GetaddrinfoAddr returns a list of address records.
func GetaddrinfoAddr(name string) ([]syscall.Sockaddr, error) {
	sas, _, err := getaddrinfoAddrCNAME(name)
	return sas, err
}

// GetaddrinfoCNAME returns a canonical name.
func GetaddrinfoCNAME(name string) (string, error) {
	_, cname, err := getaddrinfoAddrCNAME(name)
	return cname, err
}

func getaddrinfoAddrCNAME(name string) ([]syscall.Sockaddr, string, error) {
	var hints C.struct_addrinfo
	hints.ai_flags = addrinfoFlags
	hints.ai_socktype = C.SOCK_STREAM
	h := C.CString(name)
	defer C.free(unsafe.Pointer(h))
	var res *C.struct_addrinfo
	errno, err := C.getaddrinfo(h, nil, &hints, &res)
	if errno != 0 {
		if errno == C.EAI_SYSTEM && err == nil {
			// The err should not be nil, but sometimes
			// getaddrinfo returns errno == C.EAI_SYSTEM
			// with err == nil on Linux.
			// The report claims that it happens when we
			// have too many open files, so use EMFILE
			// (too many open files in system). Most
			// system calls would return ENFILE, so at the
			// least EMFILE should be easy to recognize if
			// this comes up again.
			// See golang.org/issue/6232.
			err = syscall.EMFILE
		} else {
			err = AddrinfoErrno(errno)
		}
		return nil, "", err
	}
	defer C.freeaddrinfo(res)
	var cname string
	if res != nil {
		cname = C.GoString(res.ai_canonname)
		if cname == "" {
			cname = name
		}
		if len(cname) > 0 && cname[len(cname)-1] != '.' {
			cname += "."
		}
	}
	var sas []syscall.Sockaddr
	for r := res; r != nil; r = r.ai_next {
		// We only asked for SOCK_STREAM, but check anyhow.
		if r.ai_socktype != C.SOCK_STREAM {
			continue
		}
		switch r.ai_family {
		case C.AF_INET, C.AF_INET6:
			b := (*[unsafe.Sizeof(*r.ai_addr)]byte)(unsafe.Pointer(r.ai_addr))
			sa := rawToAddr(b[:])
			if sa != nil {
				sas = append(sas, sa)
			}
		}
	}
	return sas, cname, nil
}
