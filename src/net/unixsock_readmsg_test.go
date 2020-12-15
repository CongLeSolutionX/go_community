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

	scmFile, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("file open: %v", err)
	}
	scmFd := scmFile.Fd()

	fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("Socketpair: %v", err)
	}
	writeFile := os.NewFile(uintptr(fds[0]), "write-socket")
	readFile := os.NewFile(uintptr(fds[1]), "read-socket")
	defer writeFile.Close()
	defer readFile.Close()

	cw, err := FileConn(writeFile)
	if err != nil {
		t.Fatalf("FileConn: %v", err)
	}
	cr, err := FileConn(readFile)
	if err != nil {
		t.Fatalf("FileConn: %v", err)
	}
	defer cr.Close()
	defer cw.Close()

	ucw, ok := cw.(*UnixConn)
	if !ok {
		t.Fatalf("unexpected FileConn type; expected UnixConn, got %T", c)
	}
	ucr, ok := cr.(*UnixConn)
	if !ok {
		t.Fatalf("unexpected FileConn type; expected UnixConn, got %T", c)
	}

	buf := make([]byte, 32) // expect 1 byte
	oob := make([]byte, 32) // expect 24 bytes
	// err = ucw.SetWriteDeadline(5 * time.Second)
	// if err != nil {
	// 	t.Fatalf("Can't set unix connection timeout: %v", err)
	// }
	// _, oobn, _, _, err := ucr.WriteMsgUnix(buf, oob)
	// if err != nil {
	// 	t.Fatalf("UnixConn readMsg: %v", err)
	// }
	err = ucr.SetReadDeadline(5 * time.Second)
	if err != nil {
		t.Fatalf("Can't set unix connection timeout: %v", err)
	}
	_, oobn, _, _, err := ucr.ReadMsgUnix(buf, oob)
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
