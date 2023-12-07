// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package debug contains facilities for programs to debug themselves while
// they are running.
package debug

import (
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	_ "unsafe" // for linkname
)

// PrintStack prints to standard error the stack trace returned by runtime.Stack.
func PrintStack() {
	os.Stderr.Write(Stack())
}

// Stack returns a formatted stack trace of the goroutine that calls it.
// It calls [runtime.Stack] with a large enough buffer to capture the entire trace.
func Stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

// SetCrashOutput sets the optional file to which unhandled panics and
// other fatal errors should be written in addition to the standard
// error. Call SetCrashOutput(nil) to disable the feature.
//
// SetCrashOutput does not close the file.
func SetCrashOutput(f *os.File) {
	crashFileMu.Lock()
	defer crashFileMu.Unlock()
	if f != nil {
		runtime_crashFD = f.Fd()
	} else {
		runtime_crashFD = ^uintptr(0)
	}

	// Keep f alive, across the critical section above,
	// and until the end of the next call to SetCrashOutput.
	crashFile.Store(f)
}

var (
	crashFileMu sync.Mutex
	crashFile   atomic.Pointer[os.File] // just to ensure liveness
	//go:linkname runtime_crashFD runtime.crashFD
	runtime_crashFD uintptr
)
