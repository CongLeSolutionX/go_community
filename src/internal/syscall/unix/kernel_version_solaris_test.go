// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build solaris

package unix_test

import (
	"errors"
	"internal/syscall/unix"
	"runtime"
	"syscall"
	"testing"
)

func TestSupportSockNonblockCloexec(t *testing.T) {
	// Test that SupportSockNonblockCloexec returns true if socket succeeds with SOCK_NONBLOCK and SOCK_CLOEXEC.
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, 0)
	if err == nil {
		syscall.Close(s)
	}
	want := !errors.Is(err, syscall.EPROTONOSUPPORT) && !errors.Is(err, syscall.EINVAL)
	got := unix.SupportSockNonblockCloexec()
	if want != got {
		t.Fatalf("SupportSockNonblockCloexec, got %t; want %t", got, want)
	}

	// Test that the version returned by KernelVersion matches expectations.
	major, minor := unix.KernelVersion()
	t.Logf("Kernel version: %d.%d", major, minor)
	if runtime.GOOS == "illumos" {
		if got && (major < 5 || (major == 5 && minor < 11)) {
			t.Fatalf("SupportSockNonblockCloexec is true, but kernel version is older than 5.11, SunOS version: %d.%d", major, minor)
		}
		if !got && (major > 5 || (major == 5 && minor >= 11)) {
			t.Fatalf("SupportSockNonblockCloexec is false, but kernel version is 5.11 or newer, SunOS version: %d.%d", major, minor)
		}
	} else { // Solaris
		if got && (major < 11 || (major == 11 && minor < 4)) {
			t.Fatalf("SupportSockNonblockCloexec is true, but kernel version is older than 11.4, Solaris version: %d.%d", major, minor)
		}
		if !got && (major > 11 || (major == 11 && minor >= 4)) {
			t.Errorf("SupportSockNonblockCloexec is false, but kernel version is 11.4 or newer, Solaris version: %d.%d", major, minor)
		}
	}
}
