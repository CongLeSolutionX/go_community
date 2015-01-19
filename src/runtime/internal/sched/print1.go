// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

var debuglock _core.Mutex

// The compiler emits calls to printlock and printunlock around
// the multiple calls that implement a single Go print or println
// statement. Some of the print helpers (printsp, for example)
// call print recursively. There is also the problem of a crash
// happening during the print routines and needing to acquire
// the print lock to print information about the crash.
// For both these reasons, let a thread acquire the printlock 'recursively'.

func Printlock() {
	mp := _core.Getg().M
	mp.Printlock++
	if mp.Printlock == 1 {
		_lock.Lock(&debuglock)
	}
}

func Printunlock() {
	mp := _core.Getg().M
	mp.Printlock--
	if mp.Printlock == 0 {
		_lock.Unlock(&debuglock)
	}
}
