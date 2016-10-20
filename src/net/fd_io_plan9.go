// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
)

// asyncIO implements asynchronous cancelable I/O.
type asyncIO struct {
	res chan result

	// mu guards the pid field.
	mu sync.Mutex

	// pid holds the process id of
	// the process running the IO operation.
	pid int
}

type result struct {
	n   int
	err error
}

// NewAsyncIO returns a new asyncIO that performs I/O
// operation by calling fn, which must do one and only one
// interruptible system call.
func newAsyncIO(fn func([]byte) (int, error), b []byte) *asyncIO {
	var aio asyncIO
	aio.res = make(chan result)
	aio.mu.Lock()
	go func() {
		// Lock the current goroutine to its process
		// and store the pid in io so that Cancel can
		// interrupt it. We register a signal channel so
		// that the signal does not take down the entire
		// Go runtime, but we don't need to receive on it.
		runtime.LockOSThread()
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.Note("hangup"))
		aio.pid = os.Getpid()
		aio.mu.Unlock()

		n, err := fn(b)

		aio.mu.Lock()
		aio.pid = -1
		signal.Stop(sig)
		aio.mu.Unlock()

		aio.res <- result{n, err}
	}()
	return &aio
}

// Cancel interrupts the I/O operation, causing
// the Wait function to return.
func (aio *asyncIO) Cancel() {
	aio.mu.Lock()
	defer aio.mu.Unlock()
	if aio.pid == -1 {
		return
	}
	proc, err := os.FindProcess(aio.pid)
	if err != nil {
		return
	}
	proc.Signal(syscall.Note("hangup"))
}

// Wait for the I/O operation to complete.
func (aio *asyncIO) Wait() (int, error) {
	res := <-aio.res
	return res.n, res.err
}
