// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package timerpool

import (
	"testing"
	"time"
)

func TestTimerPool(t *testing.T) {
	var tp TimerPool

	for i := 0; i < 100; i++ {
		timer := tp.Get(20 * time.Millisecond)

		select {
		case <-timer.C:
			t.Errorf("timer expired too early")
			continue
		default:
		}

		select {
		case <-time.After(100 * time.Millisecond):
			t.Errorf("timer didn't expire on time")
		case <-timer.C:
		}

		tp.Put(timer)
	}
}
