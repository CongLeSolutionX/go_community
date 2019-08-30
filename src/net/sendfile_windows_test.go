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
	"path/filepath"
	"sync"
	"testing"
)

// SendFile should be able to send files even >= 2GiB.
// See issue #33193.
func TestSendFile_biggerThan2GiB(t *testing.T) {
	if testing.Short() {
		t.Skip("This is a long running test")
	}

	tmpDir, err := ioutil.TempDir("", "sendfile-large")
	if err != nil {
		t.Fatalf("Failed to create tmp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Populate the file with 3GB.
	bigFilepath := filepath.Join(tmpDir, "sendfile")
	f, err := os.Create(bigFilepath)
	if err != nil {
		t.Fatalf("Failed to create the test file path: %v", err)
	}
	// We'll writing a 3GiB file to disk.

	chunk := int64(500 << 20)
	_3GiB := chunk * 6
	for i := chunk; i <= _3GiB; i += chunk {
		f.Write(bytes.Repeat([]byte("A"), int(chunk)))
	}
	// Now close the file and ensure it is written to disk.
	f.Sync()
	if err := f.Close(); err != nil {
		t.Fatalf("Failed to close written file: %v", err)
	}

	rf, err := os.Open(bigFilepath)
	if err != nil {
		t.Fatalf("Failed to open the file: %v", err)
	}
	defer rf.Close()

	fi, err := rf.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if gotSize, wantSize := fi.Size(), _3GiB; gotSize != wantSize {
		t.Fatalf("Size mismatch of written file: got %d want %d", gotSize, wantSize)
	}

	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
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

	tcpConn := conn.(*TCPConn)

	src := io.LimitReader(rf, _3GiB)
	written, err, handled := sendFile(tcpConn.conn.fd, src)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !handled {
		t.Fatal("Did not invoke sendfile!")
	}
	if written != _3GiB {
		t.Fatalf("Bytes written mismatch\n\tgot:  %d\n\twant: %d", written, _3GiB)
	}
}
