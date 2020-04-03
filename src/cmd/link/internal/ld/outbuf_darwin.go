// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package ld

import (
	"syscall"
	"unsafe"
)

// Implemented in the syscall package.
//go:linkname fcntl syscall.fcntl
func fcntl(fd uintptr, cmd int, arg uintptr) (int, error)

func (out *OutBuf) Fallocate(length uint64) error {
	store := syscall.Fstore_t{
		Flags:      syscall.F_ALLOCATEALL,
		Posmode:    syscall.F_PEOFPOSMODE,
		Offset:     0,
		Length:     int64(length),
		Bytesalloc: 0,
	}
	_, err := fcntl(out.f.Fd(), syscall.F_PREALLOCATE, uintptr(unsafe.Pointer(&store)))
	return err
}
