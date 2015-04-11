// Copyright 2015 The Go Authors.  All rights reserved.
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

// NewAsyncIO returns a new asyncIO that performs IO
// by calling iofn, which must do one and only one
// interruptible system call.
func NewAsyncIO(iofn func([]byte) (int, error), b []byte) *asyncIO {
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

		n, err := iofn(b)

		aio.mu.Lock()
		aio.pid = -1
		signal.Stop(sig)
		aio.mu.Unlock()

		aio.res <- result{n, err}
	}()
	return &aio
}

// Cancel interrupts the IO operation, causing
// Wait to return.
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

func (aio *asyncIO) Wait() (int, error) {
	res := <-aio.res
	return res.n, res.err
}
