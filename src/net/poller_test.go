// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux,!386 netbsd openbsd solaris

package net

import (
	"bytes"
	"syscall"
	"testing"
	"unsafe"
)

func TestPoller(t *testing.T) {
	c, err := newLocalPacketListener("udp")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	data := []byte("HELLO-R-U-THERE")
	if _, err := c.WriteTo(data, c.LocalAddr()); err != nil {
		t.Fatal(err)
	}

	p := c.(*UDPConn).Poller()
	b := make([]byte, 64)
	sa := make([]byte, 128)
	l := uint32(len(sa))
	err = p.Run(PolledRead, func(s uintptr) (bool, error) {
		n, _, errno := syscall.Syscall6(syscall.SYS_RECVFROM, s, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), syscall.MSG_PEEK, uintptr(unsafe.Pointer(&sa[0])), uintptr(unsafe.Pointer(&l)))
		switch errno {
		case 0:
			b = b[:n]
			return true, nil
		case syscall.EAGAIN:
			return false, errno
		default:
			return true, errno
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(b, data) != 0 {
		t.Fatalf("got %#v; want %#v", b, data)
	}
}
