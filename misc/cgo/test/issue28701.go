// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that the scheduler can preempt goroutines that make repeated blocking
// cgo calls.

package cgotest

// #include <unistd.h>
import "C"

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func test28701(t *testing.T) {
	// This test sleeps for 2ms 50 times. Setting a timeout of 1s is thus
	// extremely conservative.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	maxProcs := runtime.GOMAXPROCS(-1)
	for i := 0; i < maxProcs; i++ {
		go func() {
			for ctx.Err() == nil {
				// Sleep for long enough that sysmon is practically guaranteed
				// to see this P blocked in cgo, but short enough that sysmon
				// will never observe this P blocked in the same cgo call twice.
				C.usleep(1000 /* 1ms */)
			}
		}()
	}

	// sysmon occasionally gets lucky and retakes a P. Sleep 50 times to ensure
	// that sysmon is actually preempting Ps stuck in syscalls. Sysmon is
	// unlikely to get lucky 50 times in a row.
	for j := 0; j < 50; j++ {
		time.Sleep(2 * time.Millisecond)
	}

	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}
}
