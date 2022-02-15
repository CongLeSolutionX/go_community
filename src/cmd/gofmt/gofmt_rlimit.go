// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux || darwin || freebsd || openbsd || netbsd || solaris || dragonfly || aix

package main

import (
	"syscall"
)

func openFilelimit() int {
	rl := &syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, rl)
	if err != nil {
		return 200
	}

	if rl.Cur > maxFileOpen {
		return maxFileOpen
	}

	// limit 16 may be approximate.
	if rl.Cur < 16 {
		panic("insuffcient open file for gofmt")
	}

	return int(rl.Cur) - 16
}
