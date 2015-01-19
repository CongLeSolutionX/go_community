// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_sched "runtime/internal/sched"
)

//go:nosplit

// Gosched yields the processor, allowing other goroutines to run.  It does not
// suspend the current goroutine, so execution resumes automatically.
func Gosched() {
	_sched.Mcall(_sched.Gosched_m)
}

func newP() *_core.P {
	return new(_core.P)
}
