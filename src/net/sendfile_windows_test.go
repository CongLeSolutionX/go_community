// +build windows

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

// SendFile should be able to send files even >= 2GiB.
// See issue #33193.
func TestSendFile_fileSizeThan2GiB(t *testing.T) {
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	prc, pwc, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create os.Pipe(): %v", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	// Putting this here so that ln.Accept() in the goroutine right
	// below can exit when this listener is closed, and in test failure.
	defer ln.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			io.Copy(ioutil.Discard, conn)
		}
	}()

	conn, err := Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("Failed to Dial to the listener: %v", err)
	}
	defer conn.Close()

	// We'll be testing with a 5GiB file.
	_5GiB := int64(5 << 30)

	go func() {
		defer pwc.Close()
		chunk := int64(40 << 20)
		for i := chunk; i <= _5GiB; i += chunk {
			pwc.Write(bytes.Repeat([]byte("A"), int(chunk)))
		}
	}()

	tcpConn := conn.(*TCPConn)

	src := io.LimitReader(prc, _5GiB)
	written, err, handled := sendFile(tcpConn.conn.fd, src)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !handled {
		t.Fatal("Unexpected not handled")
	}
	if written != _5GiB {
		t.Fatalf("Bytes written mismatch\n\tgot:  %d\n\twant: %d", written, _5GiB)
	}
}
