// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package testblock provides test helpers for blocking goroutines.
package testblock

import (
	"fmt"
	"regexp"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

// BlockMutex causes the calling goroutine to block on a sync.Mutex Lock call
// before returning.
func BlockMutex(t *testing.T) {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		AwaitBlockedGoroutine(t, "sync.Mutex.Lock", "internal/testblock.BlockMutex", 1)
		mu.Unlock()
	}()
	// Note: Unlock releases mu before recording the mutex event,
	// so it's theoretically possible for this to proceed and
	// capture the profile before the event is recorded. As long
	// as this is blocked before the unlock happens, it's okay.
	mu.Lock()

}

// AwaitBlockedGoroutine spins on runtime.Gosched until a runtime stack dump
// shows a goroutine in the given state with a stack frame in fName.
func AwaitBlockedGoroutine(t *testing.T, state, fName string, count int) {
	re := fmt.Sprintf(`(?m)^goroutine \d+ \[%s\]:\n(?:.+\n\t.+\n)*%s`, regexp.QuoteMeta(state), regexp.QuoteMeta(fName))
	fmt.Printf("re: %v\n", re)
	r := regexp.MustCompile(re)

	if deadline, ok := t.Deadline(); ok {
		if d := time.Until(deadline); d > 1*time.Second {
			timer := time.AfterFunc(d-1*time.Second, func() {
				debug.SetTraceback("all")
				panic(fmt.Sprintf("timed out waiting for %#q", re))
			})
			defer timer.Stop()
		}
	}

	buf := make([]byte, 64<<10)
	for {
		runtime.Gosched()
		n := runtime.Stack(buf, true)
		if n == len(buf) {
			// Buffer wasn't large enough for a full goroutine dump.
			// Resize it and try again.
			buf = make([]byte, 2*len(buf))
			continue
		}
		if len(r.FindAll(buf[:n], -1)) >= count {
			return
		}
	}
}
