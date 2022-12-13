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
	// “When opening a FIFO with O_RDONLY or O_WRONLY set: If O_NONBLOCK is set,
	// an open() for reading-only shall return without delay. An open() for
	// writing-only shall return an error if no process currently has the file
	// open for reading.”
	//
	// So we open the read side first, then the write side.

	r, err := os.OpenFile(fifoName, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		t.Fatal(err)
	}

	w, err := os.OpenFile(fifoName, os.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}

	testPipeEOF(t, r, w)
}
