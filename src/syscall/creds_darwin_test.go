// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin

package syscall_test

import (
	"net"
	"os"
	"syscall"
	"testing"
)

func TestGetsockoptPID(t *testing.T) {
	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Socketpair: %v", err)
	}
	defer syscall.Close(fds[0])
	defer syscall.Close(fds[1])

	srvFile := os.NewFile(uintptr(fds[0]), "server")
	defer srvFile.Close()
	srv, err := net.FileConn(srvFile)
	if err != nil {
		t.Errorf("FileConn: %v", err)
		return
	}
	defer srv.Close()

	cliFile := os.NewFile(uintptr(fds[1]), "client")
	defer cliFile.Close()
	cli, err := net.FileConn(cliFile)
	if err != nil {
		t.Errorf("FileConn: %v", err)
		return
	}
	defer cli.Close()

	pid, err := syscall.GetsockoptPID(fds[1], syscall.SOL_LOCAL, syscall.LOCAL_PEERPID)
	if err != nil {
		t.Errorf("GetsockoptXUcred: %v", err)
		return
	}
	if pid != os.Getpid() {
		t.Errorf("pid = %d, want %d", pid, os.Getpid())
	}
}
