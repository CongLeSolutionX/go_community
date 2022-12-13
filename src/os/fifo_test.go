// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || (linux && !android) || netbsd || openbsd

package os_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFifoEOF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	fifoName := filepath.Join(dir, "fifo")
	if err := syscall.Mkfifo(fifoName, 0600); err != nil {
		t.Fatal(err)
	}

	// Per https://pubs.opengroup.org/onlinepubs/9699919799/functions/open.html#tag_16_357_03:
	//
	// - “If O_NONBLOCK is set, an open() for reading-only shall return without
	//   delay. An open() for writing-only shall return an error if no process
	//   currently has the file open for reading.”
	//
	// - “If O_NONBLOCK is clear, an open() for reading-only shall block the
	//   calling thread until a thread opens the file for writing. An open() for
	//   writing-only shall block the calling thread until a thread opens the file
	//   for reading.”
	//
	// It appears that the O_NONBLOCK implementation on macOS 12 has a bug: in
	// https://storage.googleapis.com/go-build-log/e9552219/darwin-amd64-12_0_54304319.log,
	// a read on the FIFO was observed to fail with EAGAIN, even though the write
	// side was successfully opened.
	//
	// So instead, we leave O_NONBLOCK clear, and use the blocking behavior by
	// opening the two ends of the pipe in separate goroutines.

	rc := make(chan *os.File, 1)
	go func() {
		r, err := os.Open(fifoName)
		if err != nil {
			t.Error(err)
		}
		rc <- r
	}()

	w, err := os.OpenFile(fifoName, os.O_WRONLY, 0)
	if err != nil {
		t.Error(err)
	}

	r := <-rc
	if t.Failed() {
		if r != nil {
			r.Close()
		}
		if w != nil {
			w.Close()
		}
		return
	}

	testPipeEOF(t, r, w)
}
