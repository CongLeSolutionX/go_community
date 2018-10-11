// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package net

import (
	"os"
	"syscall"
)

func newSyscallConnFile() (string, *os.File, error) {
	s, err := syscall.Socket(syscall.AF_ROUTE, syscall.SOCK_RAW, syscall.AF_UNSPEC)
	if err != nil {
		return "", nil, err
	}
	return "route", os.NewFile(uintptr(s), ""), nil
}
