// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
	"testing"
)

func TestBuffers_write(t *testing.T) {
	var story = []byte("once upon a time in Gopherland ... ")
	var buffers Buffers
	for range story {
		buffers = append(buffers, make([]byte, 1))
	}
	n, err := buffers.Write(story)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(story) {
		t.Errorf("n = %d; want %d", n, len(story))
	}
	if got := bytes.Join([][]byte(buffers), nil); !bytes.Equal(got, story) {
		t.Errorf("read %q; want %q", string(got), string(story))
	}
	if len(buffers) != len(story) {
		t.Errorf("len(buffers) = %d; want 0%d", len(buffers), len(story))
	}
}

func TestBuffers_ReadFrom(t *testing.T) {
	for _, name := range []string{"ReadFrom", "Copy"} {
		for _, size := range []int{0, 10, 1023, 1024, 1025} {
			t.Run(fmt.Sprintf("%s/%d", name, size), func(t *testing.T) {
				testBuffer_readFrom(t, size, false, name == "Copy")
			})
			t.Run(fmt.Sprintf("%s/%d+1", name, size), func(t *testing.T) {
				testBuffer_readFrom(t, size, true, name == "Copy")
			})
		}
	}
}

func testBuffer_readFrom(t *testing.T, chunks int, extraByte, useCopy bool) {
	oldHook := testHookDidReadv
	defer func() { testHookDidReadv = oldHook }()
	var readLog struct {
		sync.Mutex
		log []int
	}
	testHookDidReadv = func(size int) {
		readLog.Lock()
		readLog.log = append(readLog.log, size)
		readLog.Unlock()
	}
	var want bytes.Buffer
	for i := 0; i < chunks; i++ {
		want.WriteByte(byte(i))
	}
	if extraByte { // write extra byte that doesn't fit into the Buffers
		want.WriteByte(byte(chunks))
	}

	readDone := make(chan struct{})

	withTCPConnPair(t, func(c *TCPConn) error {
		defer close(readDone)
		buffers := make(Buffers, chunks)
		for i := range buffers {
			buffers[i] = make([]byte, 1)
		}
		var n int64
		var err error
		if useCopy {
			n, err = io.Copy(&buffers, c)
		} else {
			n, err = buffers.ReadFrom(c)
		}
		if err != nil {
			return err
		}
		if n != int64(chunks) {
			return fmt.Errorf("Buffers.ReadFrom returned %d; want %d", n, chunks)
		}
		if extraByte {
			n, err := io.Copy(ioutil.Discard, c)
			if err != nil {
				return err
			}
			if n != 1 {
				return fmt.Errorf("found %d extra bytes; want 1", n)
			}
		}
		return nil
	}, func(c *TCPConn) error {
		var n int
		for i := 0; i < want.Len(); i++ {
			n0, err := c.Write(want.Bytes()[i : i+1])
			n += n0
			if err != nil {
				return err
			}
		}
		if err := c.CloseWrite(); err != nil {
			return err
		}
		if n != want.Len() {
			return fmt.Errorf("client wrote %d; want %d", n, want.Len())
		}
		<-readDone     // wait for the data to be fully read
		readLog.Lock() // no need to unlock (or even lock, but let's be safe)
		var gotSum int
		for _, v := range readLog.log {
			gotSum += v
		}

		var wantSum int
		var wantMinCalls int
		switch runtime.GOOS {
		case "darwin", "dragonfly", "freebsd", "linux", "netbsd", "openbsd":
			wantSum = want.Len()
			if extraByte {
				wantSum--
			}
			v := chunks
			for v > 0 {
				wantMinCalls++
				v -= 1024
			}
		}
		if len(readLog.log) < wantMinCalls {
			t.Errorf("write calls = %v < wanted min %v", len(readLog.log), wantMinCalls)
		}
		if gotSum != wantSum {
			t.Errorf("readv call sum = %v; want %v", gotSum, wantSum)
		}
		return nil
	})
}
