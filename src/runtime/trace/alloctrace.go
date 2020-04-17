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

//go:linkname runtime_startAllocTrace runtime.startAllocTrace
func runtime_startAllocTrace()

//go:linkname runtime_readAllocTrace runtime.readAllocTrace
func runtime_readAllocTrace() []byte

//go:linkname runtime_stopAllocTrace runtime.stopAllocTrace
func runtime_stopAllocTrace()

// StartAllocTrace starts streaming the allocation trace to the
// given Writer, ending the trace at around maxBytes in size.
//
// It will also start the allocation trace if it hasn't yet been
// started.
//
// maxBytes is only fuzzy because an unknown (but bounded) amount
// of additional data may need to be written to produce a consistent
// trace after the limit is reached, so the value should be chosen
// conservatively.
//
// Even if maxBytes is passed, StopAllocTrace must still always be
// called in order to ensure that
//
// If maxBytes == -1, then the trace will continue indefinitely,
// until StopAllocTrace is called.
func StartAllocTrace(w io.Writer, maxBytes int64) error {
	atState.mu.Lock()
	defer atState.mu.Unlock()
	if atState.done != nil {
		return errors.New("allocation trace already started")
	}
	done := make(chan struct{})
	atState.done = done

	runtime_startAllocTrace()
	go func() {
		written := int64(0)
		sentStop := false
		for {
			if maxBytes != -1 && !sentStop && written > maxBytes {
				sentStop = true

				// Asynchronously stop the allocation trace.
				go func() {
					StopAllocTrace()
				}()
			}
			bytes := runtime_readAllocTrace()
			if bytes == nil {
				break
			}
			w.Write(bytes)
			w.Write(empty[:runtime.AllocTraceBatchSize-len(bytes)])
			written += runtime.AllocTraceBatchSize
		}
		close(done)
	}()
	return nil
}

// StopAllocTrace stops accumulating allocation trace events and
// blocks until a consistent trace is written out to the Writer
// passed to StartAllocTrace.
//
// If StartAllocTrace has never been called, or StopAllocTrace has
// already been called after a StartAllocTrace, then this function
// simply returns.
func StopAllocTrace() {
	atState.mu.Lock()
	defer atState.mu.Unlock()

	if atState.done == nil {
		// No trace streaming started, nothing to do.
		return
	}

	// Stop trace and wait for the reader to finish.
	runtime_stopAllocTrace()
	<-atState.done
	atState.done = nil
}

var atState struct {
	mu   sync.Mutex
	done <-chan struct{}
}
