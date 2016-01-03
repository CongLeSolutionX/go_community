// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"syscall"
	"unsafe"
)

// getentropy(2)'s syscall number, from /usr/include/sys/syscall.h
const entropyTrap uintptr = 7

// GetEntropy calls the OpenBSD getentropy system call.
func GetEntropy(p []byte) (err error) {
	// XXX: do we want to feign an errno if plen == 0?
	_, _, errno := syscall.Syscall(entropyTrap,
		uintptr(unsafe.Pointer(&p[0])),
		uintptr(len(p)),
		0)
	if errno != 0 {
		return errno
	}
	return nil
}
