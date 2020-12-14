// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestUnixConnReadMsgUnixSCMRightsCloseOnExec(t *testing.T) {
	if !testableNetwork("unix") {
		t.Skip("not unix system")
	}

	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Socketpair: %v", err)
	}
	writeFile := os.NewFile(uintptr(fds[0]), "parent-reads")
	readFile := os.NewFile(uintptr(fds[1]), "parent-reads")
	defer readFile.Close()

	c, err := FileConn(readFile)
	if err != nil {
		t.Fatalf("FileConn: %v", err)
	}
	defer c.Close()

	uc, ok := c.(*UnixConn)
	if !ok {
		t.Fatalf("unexpected FileConn type; expected UnixConn, got %T", c)
	}

	buf := make([]byte, 32) // expect 1 byte
	oob := make([]byte, 32) // expect 24 bytes
	err = uc.SetReadDeadline(5 * time.Second)
	if err != nil {
		t.Fatalf("Can't set unix connection timeout: %v", err)
	}
	_, oobn, _, _, err := uc.ReadMsgUnix(buf, oob)
	if err != nil {
		t.Fatalf("UnixConn readMsg: %v", err)
	}

	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		t.Fatalf("ParseSocketControlMessage: %v", err)
	}
	if len(scms) != 1 {
		t.Fatalf("expected 1 SocketControlMessage; got scms = %#v", scms)
	}
	scm := scms[0]
	gotFds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		t.Fatalf("syscall.ParseUnixRights: %v", err)
	}
	if len(gotFds) != 1 {
		t.Fatalf("wanted 1 fd; got %#v", gotFds)
	}

	oldFlags, _, err := syscall.Syscall(syscall.SYS_FCNTL, uintptr(gotFds[0]),
		uintptr(syscall.F_GETFL), 0)
	if err != nil {
		t.Fatalf("Can't get flags of fd:%#v", gotFds[0])
	}
	if oldFlags&syscall.FD_CLOEXEC == 0 {
		t.Fatalf("Fail to set close-on-exec flag, the oldFlags is %#v", oldFlags)
	}
}
