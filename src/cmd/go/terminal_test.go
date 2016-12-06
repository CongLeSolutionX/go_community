// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux

package main_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"unsafe"
)

const ioctlReadTermios = syscall.TCGETS

// isTerminal reports whether fd is a terminal.
func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

// Issue 18153: streaming test output's stdout/stderr should be a terminal.
func TestStreamingTestOutputTerminal(t *testing.T) {
	if !isTerminal(1) {
		t.Skip("stdout is not a terminal; skipping test")
	}
	tg := testgo(t)
	_ = tg
	cmd := exec.Command("../../testgo", "test")
	cmd.Dir = "./testdata/testterminal18153"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	t.Logf("Got: %v", cmd.Run())
}
