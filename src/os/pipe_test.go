// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9,!nacl,!js

package os_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func testClosedPipeRace(t *testing.T, read bool) {
	switch runtime.GOOS {
	case "freebsd":
		t.Skip("FreeBSD does not use the poller; issue 19093")
	}

	limit := 1
	if !read {
		// Get the amount we have to write to overload a pipe
		// with no reader.
		limit = 131073
		if b, err := ioutil.ReadFile("/proc/sys/fs/pipe-max-size"); err == nil {
			if i, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
				limit = i + 1
			}
		}
		t.Logf("using pipe write limit of %d", limit)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()

	// Close the read end of the pipe in a goroutine while we are
	// writing to the write end, or vice-versa.
	go func() {
		// Give the main goroutine a chance to enter the Read or
		// Write call. This is sloppy but the test will pass even
		// if we close before the read/write.
		time.Sleep(20 * time.Millisecond)

		var err error
		if read {
			err = r.Close()
		} else {
			err = w.Close()
		}
		if err != nil {
			t.Error(err)
		}
	}()

	b := make([]byte, limit)
	if read {
		_, err = r.Read(b[:])
	} else {
		_, err = w.Write(b[:])
	}
	if err == nil {
		t.Error("I/O on closed pipe unexpectedly succeeded")
	} else if pe, ok := err.(*os.PathError); !ok {
		t.Errorf("I/O on closed pipe returned unexpected error type %T; expected os.PathError", pe)
	} else if pe.Err != os.ErrClosed {
		t.Errorf("got error %q but expected %q", pe.Err, os.ErrClosed)
	} else {
		t.Logf("I/O returned expected error %q", err)
	}
}

func TestClosedPipeRaceRead(t *testing.T) {
	testClosedPipeRace(t, true)
}

func TestClosedPipeRaceWrite(t *testing.T) {
	testClosedPipeRace(t, false)
}

// Test that we don't let a blocking read prevent a close.
func testCloseWithBlockingRead(t *testing.T, r, w *os.File) {
	defer r.Close()
	defer w.Close()

	c1, c2 := make(chan bool), make(chan bool)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(c chan bool) {
		defer wg.Done()
		// Give the other goroutine a chance to enter the Read
		// or Write call. This is sloppy but the test will
		// pass even if we close before the read/write.
		time.Sleep(20 * time.Millisecond)

		if err := r.Close(); err != nil {
			t.Error(err)
		}
		close(c)
	}(c1)

	wg.Add(1)
	go func(c chan bool) {
		defer wg.Done()
		var b [1]byte
		_, err := r.Read(b[:])
		close(c)
		if err == nil {
			t.Error("I/O on closed pipe unexpectedly succeeded")
		}
	}(c2)

	for c1 != nil || c2 != nil {
		select {
		case <-c1:
			c1 = nil
			// r.Close has completed, but the blocking Read
			// is hanging. Close the writer to unblock it.
			w.Close()
		case <-c2:
			c2 = nil
		case <-time.After(1 * time.Second):
			switch {
			case c1 != nil && c2 != nil:
				t.Error("timed out waiting for Read and Close")
				w.Close()
			case c1 != nil:
				t.Error("timed out waiting for Close")
			case c2 != nil:
				t.Error("timed out waiting for Read")
			default:
				t.Error("impossible case")
			}
		}
	}

	wg.Wait()
}

func TestCloseWithBlockingRead(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	testCloseWithBlockingRead(t, r, w)
}

// Issue 24164, for pipes.
func TestPipeEOF(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			if err := w.Close(); err != nil {
				t.Errorf("error closing writer: %v", err)
			}
		}()

		for i := 0; i < 3; i++ {
			time.Sleep(10 * time.Millisecond)
			_, err := fmt.Fprintf(w, "line %d\n", i)
			if err != nil {
				t.Errorf("error writing to fifo: %v", err)
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}()

	defer wg.Wait()

	done := make(chan bool)
	go func() {
		defer close(done)

		defer func() {
			if err := r.Close(); err != nil {
				t.Errorf("error closing reader: %v", err)
			}
		}()

		rbuf := bufio.NewReader(r)
		for {
			b, err := rbuf.ReadBytes('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Error(err)
				return
			}
			t.Logf("%s\n", bytes.TrimSpace(b))
		}
	}()

	select {
	case <-done:
		// Test succeeded.
	case <-time.After(time.Second):
		t.Error("timed out waiting for read")
		// Close the reader to force the read to complete.
		r.Close()
	}
}
