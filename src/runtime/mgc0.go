// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
)

//go:linkname runtime_debug_freeOSMemory runtime/debug.freeOSMemory
func runtime_debug_freeOSMemory() {
	_gc.Gogc(2) // force GC and do eager sweep
	_lock.Systemstack(scavenge_m)
}

//go:linkname sync_runtime_registerPoolCleanup sync.runtime_registerPoolCleanup
func sync_runtime_registerPoolCleanup(f func()) {
	_gc.Poolcleanup = f
}
