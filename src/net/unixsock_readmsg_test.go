// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"log"
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
	defer scmFile.Close()

	rights := syscall.UnixRights(int(scmFile.Fd()))
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
		t.Fatalf("unexpected FileConn type; expected UnixConn, got %T", cw)
	}
	ucr, ok := cr.(*UnixConn)
	if !ok {
		t.Fatalf("unexpected FileConn type; expected UnixConn, got %T", cr)
	}

	oob := make([]byte, syscall.CmsgSpace(4))
	err = ucw.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		t.Fatalf("Can't set unix connection timeout: %v", err)
	}
	_, _, err = ucw.WriteMsgUnix(nil, rights, nil)
	if err != nil {
		t.Fatalf("UnixConn readMsg: %v", err)
	}
	err = ucw.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		t.Fatalf("Can't set unix connection timeout: %v", err)
	}
	_, oobn, _, _, err := ucr.ReadMsgUnix(nil, oob)
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
	defer func() {
		if err := syscall.Close(int(gotFds[0])); err != nil {
			log.Fatalf("fail to close gotFds: %v", err)
		}
	}()

	oldFlags, _, err := syscall.Syscall(syscall.SYS_FCNTL, uintptr(gotFds[0]),
		uintptr(syscall.F_GETFL), 0)
	if err != nil {
		t.Fatalf("Can't get flags of fd:%#v, with err:%v", gotFds[0], err)
	}
	if oldFlags&syscall.FD_CLOEXEC == 0 {
		t.Fatalf("Fail to set close-on-exec flag, the oldFlags is %#v", oldFlags)
	}
}
