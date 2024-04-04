// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"errors"
	"sync"
	"syscall"
	"unsafe"
)

//go:linkname procUname libc_uname

var procUname uintptr

// Utsname represents the fields of a struct utsname.
type Utsname struct {
	Sysname  [257]byte
	Nodename [257]byte
	Release  [257]byte
	Version  [257]byte
	Machine  [257]byte
}

// KernelVersion returns major and minor kernel version numbers
// parsed from the syscall.Uname's Version field, or (0, 0) if the
// version can't be obtained or parsed.
func KernelVersion() (major int, minor int) {
	var un Utsname
	_, _, errno := syscall6(uintptr(unsafe.Pointer(&procUname)), 1, uintptr(unsafe.Pointer(&un)), 0, 0, 0, 0, 0)
	if errno != 0 {
		return 0, 0
	}

	ver := un.Version[:]
	// Parse the version string of "<version>.<update>.<sru>.<build>.<reserved>"
	// based off of https://blogs.oracle.com/solaris/post/whats-in-a-uname-
	parseNext := func() (n int) {
		for i, c := range ver {
			if c == '.' {
				ver = ver[i+1:]
				return
			}
			if '0' <= c && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		ver = nil
		return
	}

	major = parseNext()
	minor = parseNext()

	return
}

// SupportSockNonblockCloexec tests if SOCK_NONBLOCK and SOCK_CLOEXEC are supported
// for socket(3c) and accept4(3c), returns true if affirmative.
var SupportSockNonblockCloexec = sync.OnceValue(func() bool {
	// First test if the socket(3c) supports SOCK_NONBLOCK and SOCK_CLOEXEC directly.
	// Note that both accept4(3c) system call and support for SOCK_* flags as part of the type
	// parameter made their first appearance in Solaris 11.4. As a result of which, checking
	// socket(3c) with SOCK_NONBLOCK and SOCK_CLOEXEC should cover them both.
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, 0)
	if err == nil {
		syscall.Close(s)
		return true
	}
	if !errors.Is(err, syscall.EPROTONOSUPPORT) && !errors.Is(err, syscall.EINVAL) {
		// Something wrong with socket(3c), fall back to checking the kernel version.
		major, minor := KernelVersion()
		return major > 11 || (major == 11 && minor >= 4)
	}
	return false
})
