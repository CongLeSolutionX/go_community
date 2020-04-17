// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"errors"
	"io"
	"runtime"
	"sync"
	_ "unsafe" // for go:linkname
)

var empty [runtime.AllocTraceBatchSize]byte

//go:linkname runtime_readAllocTrace runtime.readAllocTrace
func runtime_readAllocTrace() []byte

//go:linkname runtime_stopAllocTrace runtime.stopAllocTrace
func runtime_stopAllocTrace()

// CopyAllocTrace starts streaming the allocation trace to the
// given Writer.
//
// This function only has any effect if the GOALLOCTRACE environment
// variable is set, which means tracing has started from program
// startup. If it is not set, this function will spin up a writer
// which writes nothing into w.
//
// This function may only be called once in the lifetime of an application.
//
// StopAllocTrace should always be called before or after CopyAllocTrace
// to ensure a consistent trace is produced.
//
// GOALLOCTRACE is only available on AIX, Darwin, Dragonfly BSD,
// FreeBSD, NetBSD, OpenBSD, Illumos, Solaris, and Linux, because
// environment variables need to be parsed early on in bootstrapping
// and only these platforms allow us to do so easily.
func CopyAllocTrace(w io.Writer) error {
	atState.mu.Lock()
	defer atState.mu.Unlock()
	if atState.done != nil {
		return errors.New("allocation trace writer already started")
	}
	done := make(chan struct{})
	atState.done = done

	go func() {
		header := []byte{0, 0, 1, 15}
		w.Write(header)

		for {
			bytes := runtime_readAllocTrace()
			if bytes == nil {
				break
			}
			w.Write(bytes)
			w.Write(empty[:runtime.AllocTraceBatchSize-len(bytes)])
		}
		close(done)
	}()
	return nil
}

// StopAllocTrace stops accumulating allocation trace events and
// blocks until a consistent trace is written out to the Writer
// passed to CopyAllocTrace if one exists.
func StopAllocTrace() {
	atState.mu.Lock()
	defer atState.mu.Unlock()

	// Stop tracing.
	runtime_stopAllocTrace()

	if atState.done == nil {
		// No trace streaming started, nothing to do.
		return
	}

	// Wait for the reader to finish.
	<-atState.done
}

var atState struct {
	mu   sync.Mutex
	done <-chan struct{}
}
