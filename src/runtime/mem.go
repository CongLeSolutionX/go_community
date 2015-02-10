// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sem "runtime/internal/sem"
)

//go:linkname runtime_debug_WriteHeapDump runtime/debug.WriteHeapDump
func runtime_debug_WriteHeapDump(fd uintptr) {
	_sem.Semacquire(&_gc.Worldsema, false)
	gp := _core.Getg()
	gp.M.Preemptoff = "write heap dump"
	_lock.Systemstack(_gc.Stoptheworld)

	_lock.Systemstack(func() {
		writeheapdump_m(fd)
	})

	gp.M.Preemptoff = ""
	gp.M.Locks++
	_sem.Semrelease(&_gc.Worldsema)
	_lock.Systemstack(_gc.Starttheworld)
	gp.M.Locks--
}
