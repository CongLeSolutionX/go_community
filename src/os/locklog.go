// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build locklog

// When runtime lock logging is enabled, this connects to the lock
// logging server and initializes runtime lock logging.

package os

import "syscall"

func runtime_lockLogInit(fd int, exePath string)

func init() {
	sockPath := Getenv("GOLOCKLOG")
	if sockPath == "" {
		return
	}

	exe, err := Executable()
	if err != nil {
		panic(err)
	}

	// Connect to the lock log service.
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}
	syscall.CloseOnExec(fd)
	err = syscall.Connect(fd, &syscall.SockaddrUnix{Name: sockPath})
	if err != nil {
		panic(err)
	}

	runtime_lockLogInit(fd, exe)
}
