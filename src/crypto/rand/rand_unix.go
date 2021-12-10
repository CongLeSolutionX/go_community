// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

// Unix cryptographically secure pseudorandom number
// generator.

package rand

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const urandomDevice = "/dev/urandom"

func init() {
	Reader = &reader{}
}

// A reader satisfies reads by reading from urandomDevice
type reader struct {
	f    io.Reader
	mu   sync.Mutex
	used int32 // atomic; whether this reader has been used
}

// altGetRandom if non-nil specifies an OS-specific function to get
// urandom-style randomness.
var altGetRandom func([]byte) (ok bool)

// batched returns a function that calls f to populate a []byte by chunking it
// into subslices of, at most, readMax bytes, buffering min(readMax, 4096)
// bytes at a time.
func batched(f func([]byte) bool, readMax int) func([]byte) bool {
	bufferSize := 4096
	if bufferSize > readMax {
		bufferSize = readMax
	}
	fullBuffer := make([]byte, bufferSize)
	var buf []byte
	return func(out []byte) bool {
		// First we copy any amount remaining in the buffer.
		n := copy(out, buf)
		out, buf = out[n:], buf[n:]

		// Then, if we're requesting more than the buffer size,
		// generate directly into the buffer, chunked by readMax.
		for len(out) >= len(fullBuffer) {
			read := len(out)
			if read > readMax {
				read = readMax
			}
			if !f(out[:read]) {
				return false
			}
			out = out[read:]
		}

		// If there's a partial block left over, fill the buffer,
		// and copy in the remainder.
		if len(out) > 0 {
			buf = fullBuffer[:]
			if !f(buf) {
				return false
			}
			n = copy(out, buf)
			out, buf = out[n:], buf[n:]
		}

		return true
	}
}

func warnBlocked() {
	println("crypto/rand: blocked for 60 seconds waiting to read random data from the kernel")
}

func (r *reader) Read(b []byte) (n int, err error) {
	if atomic.CompareAndSwapInt32(&r.used, 0, 1) {
		// First use of randomness. Start timer to warn about
		// being blocked on entropy not being available.
		t := time.AfterFunc(time.Minute, warnBlocked)
		defer t.Stop()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if altGetRandom != nil && altGetRandom(b) {
		return len(b), nil
	}
	if r.f == nil {
		f, err := os.Open(urandomDevice)
		if err != nil {
			return 0, err
		}
		r.f = bufio.NewReader(hideAgainReader{f})
	}
	return r.f.Read(b)
}

// hideAgainReader masks EAGAIN reads from /dev/urandom.
// See golang.org/issue/9205
type hideAgainReader struct {
	r io.Reader
}

func (hr hideAgainReader) Read(p []byte) (n int, err error) {
	n, err = hr.r.Read(p)
	if errors.Is(err, syscall.EAGAIN) {
		err = nil
	}
	return
}
