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

func StartAllocTrace(w io.Writer) error {
	atState.mu.Lock()
	defer atState.mu.Unlock()
	if atState.done != nil {
		return errors.New("allocation trace already started")
	}
	done := make(chan struct{})
	atState.done = done

	runtime_startAllocTrace()
	go func() {
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

func StopAllocTrace() {
	atState.mu.Lock()
	defer atState.mu.Unlock()

	if atState.done == nil {
		// No trace started, nothing to do.
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
