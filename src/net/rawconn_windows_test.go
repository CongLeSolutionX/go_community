// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"errors"
	"syscall"
	"unsafe"
)

func readRawConn(c syscall.RawConn, b []byte) (int, error) {
	return 0, syscall.EWINDOWS
}

func writeRawConn(c syscall.RawConn, b []byte) error {
	return syscall.EWINDOWS
}

func controlRawConn(c syscall.RawConn, addr Addr) error {
	var operr error
	fn := func(s uintptr) {
		var v, l int32
		l = int32(unsafe.Sizeof(v))
		operr = syscall.Getsockopt(syscall.Handle(s), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, (*byte)(unsafe.Pointer(&v)), &l)
		if operr != nil {
			return
		}
		switch addr := addr.(type) {
		case *TCPAddr:
			// There's no guarantee that IP-level socket
			// options work well with dual stack sockets.
			if addr.IP.To16() != nil && addr.IP.To4() == nil {
				operr = syscall.SetsockoptInt(syscall.Handle(s), syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, 1)
			} else if addr.IP.To16() == nil && addr.IP.To4() != nil {
				operr = syscall.SetsockoptInt(syscall.Handle(s), syscall.IPPROTO_IP, syscall.IP_TTL, 1)
			}
		}
	}
	if err := c.Control(fn); err != nil {
		return err
	}
	if operr != nil {
		return operr
	}
	return nil
}

func controlOnConnSetup(network string, address string, c syscall.RawConn) error {
	var operr error
	var fn func(uintptr)
	switch network {
	case "tcp", "udp", "ip":
		return errors.New("ambiguous network: " + network)
	default:
		switch network[len(network)-1] {
		case '4':
			fn = func(s uintptr) {
				operr = syscall.SetsockoptInt(syscall.Handle(s), syscall.IPPROTO_IP, syscall.IP_TTL, 1)
			}
		case '6':
			fn = func(s uintptr) {
				operr = syscall.SetsockoptInt(syscall.Handle(s), syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, 1)
			}
		default:
			return errors.New("unknown network: " + network)
		}
	}
	if err := c.Control(fn); err != nil {
		return err
	}
	if operr != nil {
		return operr
	}
	return nil
}
