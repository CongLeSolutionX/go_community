// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris windows

package socket

import (
	"runtime"
	"syscall"
	"testing"
)

func TestSocket(t *testing.T) {
	t.Run("Sockname", func(t *testing.T) {
		testSockname(t, Getsockname)
	})
	t.Run("Sockopt", func(t *testing.T) {
		testSockopt(t, Setsockopt, Getsockopt)
	})
}

func testSockname(t *testing.T, fn func(uintptr) (int, []byte, error)) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Logf("not supported on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
		return
	}
	defer closeFunc(uintptr(s))
	if _, _, err := fn(uintptr(s)); err != nil {
		t.Fatal(err)
	}
}

func testSockopt(t *testing.T, setFn func(uintptr, int, int, []byte) error, getFn func(uintptr, int, int, []byte) (int, error)) {
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		t.Logf("not supported on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
		return
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
		t.Fatalf("god %d; want %d", n, len(b))
	}
}
