// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build !nacl,!plan9,!solaris,!windows

package socket

/*
#include <sys/types.h>
#include <sys/socket.h>

#include <netdb.h>
*/
import "C"

import "syscall"

// These are roughly enough for the following:
//
// Source		Encoding			Maximum length of single name entry
// Unicast DNS		ASCII or			<=253 + a NUL terminator
//			Unicode in RFC 5892		252 * total number of labels + delimiters + a NUL terminator
// Multicast DNS	UTF-8 in RFC 5198 or		<=253 + a NUL terminator
//			the same as unicast DNS ASCII	<=253 + a NUL terminator
// Local database	various				depends on implementation
const (
	nameinfoLen    = 64
	maxNameinfoLen = 4096
)

// GetnameinfoPTR returns a name of pointer record.
func GetnameinfoPTR(ip []byte) (string, error) {
	var sa *C.struct_sockaddr
	var salen C.socklen_t
	switch len(ip) {
	case 4:
		sa, salen = rawSockaddrInet4(ip), C.socklen_t(syscall.SizeofSockaddrInet4)
	case 16:
		sa, salen = rawSockaddrInet6(ip), C.socklen_t(syscall.SizeofSockaddrInet6)
	default:
		return "", syscall.EINVAL
	}
	var b []byte
	var err error
	for l := nameinfoLen; l <= maxNameinfoLen; l *= 2 {
		b = make([]byte, l)
		err = getnameinfoPTR(b, sa, salen)
		if err == nil || err != AddrinfoErrno(C.EAI_OVERFLOW) {
			break
		}
	}
	if err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		if b[i] != 0 {
			continue
		}
		b = b[:i]
		break
	}
	return string(b), nil
}
