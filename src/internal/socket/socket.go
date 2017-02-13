// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package socket provides a portable interface for socket system
// calls.
package socket

import "runtime"

// BUG(mikio): This package is not implemented on NaCl and Plan 9.

// Getsockname returns the local end point address of s.
//
// The returned af represents the address family of sa.
// It's not reasonable to assume that the address family of sa is
// equivalent to the communication domain of s.
// It depends on the implementation of protocol stack.
func Getsockname(s uintptr) (af int, sa []byte, err error) {
	var b []byte
	if b, err = getsockname(s); err != nil {
		return 0, nil, err
	}
	return addrFamily(b), b, nil
}

// Getpeername returns the remote end point address of s.
//
// The returned af represents the address family of sa.
// It's not reasonable to assume that the address family of sa is
// equivalent to the communication domain of s.
// It depends on the implementation of protocol stack.
func Getpeername(s uintptr) (af int, sa []byte, err error) {
	var b []byte
	if b, err = getpeername(s); err != nil {
		return 0, nil, err
	}
	return addrFamily(b), b, nil
}

func addrFamily(b []byte) int {
	switch runtime.GOOS {
	case "linux", "solaris", "windows":
		return int(nativeEndian.Uint16(b[:2]))
	default:
		return int(b[1])
	}
}

// Getsockopt reads a value for the option specified by level and
// name from the kernel.
// It returns the number of bytes written into b.
func Getsockopt(s uintptr, level, name int, b []byte) (int, error) {
	return getsockopt(s, level, name, b)
}

// Setsockopt writes the option and value to the kernel.
func Setsockopt(s uintptr, level, name int, b []byte) error {
	return setsockopt(s, level, name, b)
}
