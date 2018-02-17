// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"errors"
	"syscall"
	"testing"
	"unsafe"
)

func TestRawConn(t *testing.T) {
	c, err := newLocalPacketListener("udp")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	cc, err := c.(*UDPConn).SyscallConn()
	if err != nil {
		t.Fatal(err)
	}

	var operr error
	fn := func(s uintptr) {
		operr = syscall.SetsockoptInt(syscall.Handle(s), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	}
	err = cc.Control(fn)
	if err != nil || operr != nil {
		t.Fatal(err, operr)
	}
	c.Close()
	err = cc.Control(fn)
	if err == nil {
		t.Fatal("should fail")
	}
}

func TestRawConnListener(t *testing.T) {
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cc, err := ln.(*TCPListener).SyscallConn()
	if err != nil {
		t.Fatal(err)
	}

	called := false
	op := func(uintptr) bool {
		called = true
		return true
	}

	err = cc.Write(op)
	if err == nil {
		t.Error("Write should return an error")
	}
	if called {
		t.Error("Write shouldn't call op")
	}

	called = false
	err = cc.Read(op)
	if err == nil {
		t.Error("Read should return an error")
	}
	if called {
		t.Error("Read shouldn't call op")
	}

	var operr error
	fn := func(s uintptr) {
		var v, l int32
		l = int32(unsafe.Sizeof(v))
		operr = syscall.Getsockopt(syscall.Handle(s), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, (*byte)(unsafe.Pointer(&v)), &l)
	}
	err = cc.Control(fn)
	if err != nil || operr != nil {
		t.Fatal(err, operr)
	}
	ln.Close()
	err = cc.Control(fn)
	if err == nil {
		t.Fatal("Control after Close should fail")
	}
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
