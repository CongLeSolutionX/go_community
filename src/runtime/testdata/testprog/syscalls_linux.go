// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"internal/testenv"
	"os"
	"syscall"
)

func gettid() int {
	return syscall.Gettid()
}

func tidExists(tid int) (exists, supported bool) {
	statFile := fmt.Sprintf("/proc/self/task/%d/stat", tid)
	stat, err := os.ReadFile(statFile)
	if os.IsNotExist(err) {
		return false, true
	}
	if err != nil {
		return false, false
	}
	fields := bytes.Fields(stat)
	if len(fields) < 3 {
		// This has been observed to fail on the builders.
		return false, false
	}
	// Check if it's a zombie thread.
	state := fields[2]
	return !(len(state) == 1 && state[0] == 'Z'), true
}

func getcwd() (string, error) {
	if !syscall.ImplementsGetwd {
		return "", nil
	}
	// Use the syscall to get the current working directory.
	// This is imperative for checking for OS thread state
	// after an unshare since os.Getwd might just check the
	// environment, or use some other mechanism.
	var buf [4096]byte
	n, err := syscall.Getcwd(buf[:])
	if err != nil {
		return "", err
	}
	// Subtract one for null terminator.
	return string(buf[:n-1]), nil
}

func unshareFs() error {
	err := syscall.Unshare(syscall.CLONE_FS)
	if testenv.SyscallIsNotSupported(err) {
		return errNotPermitted
	}
	return err
}

func chdir(path string) error {
	return syscall.Chdir(path)
}
