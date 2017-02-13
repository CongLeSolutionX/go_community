// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris windows

package socket_test

import (
	"encoding/binary"
	"internal/socket"
	"runtime"
	"syscall"
	"testing"
	"unsafe"
)

var nativeEndian binary.ByteOrder

func init() {
	i := uint32(1)
	b := (*[4]byte)(unsafe.Pointer(&i))
	if b[0] == 1 {
		nativeEndian = binary.LittleEndian
	} else {
		nativeEndian = binary.BigEndian
	}
}

func TestSocket(t *testing.T) {
	// On Windows, an unamed socket has really no name.
	// The protocol stack inside the kernel returns an error for
	// the query on unnamed sockets.
	if runtime.GOOS != "windows" {
		t.Run("Sockname", func(t *testing.T) {
			testSockname(t, socket.Getsockname)
		})
	}
	t.Run("Sockopt", func(t *testing.T) {
		testSockopt(t, socket.Setsockopt, socket.Getsockopt)
	})
}

func testSockname(t *testing.T, fn func(uintptr) (int, []byte, error)) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Skipf("not supported on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
	}
	defer closeFunc(uintptr(s))
	if _, _, err := fn(uintptr(s)); err != nil {
		t.Fatal(err)
	}
}

func testSockopt(t *testing.T, setFn func(uintptr, int, int, []byte) error, getFn func(uintptr, int, int, []byte) (int, error)) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Skipf("not supported on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
	}
	defer closeFunc(uintptr(s))
	var b [4]byte
	nativeEndian.PutUint32(b[:], uint32(1))
	if err := setFn(uintptr(s), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, b[:]); err != nil {
		t.Fatal(err)
	}
	n, err := getFn(uintptr(s), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, b[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != len(b) {
		t.Fatalf("got %d; want %d", n, len(b))
	}
}
