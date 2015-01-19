// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
)

// Called from C. Returns the Go type *m.
func gc_m_ptr(ret *interface{}) {
	*ret = (*_core.M)(nil)
}

// Called from C. Returns the Go type *g.
func gc_g_ptr(ret *interface{}) {
	*ret = (*_core.G)(nil)
}

// Called from C. Returns the Go type *itab.
func gc_itab_ptr(ret *interface{}) {
	*ret = (*_core.Itab)(nil)
}

//go:linkname runtime_debug_freeOSMemory runtime/debug.freeOSMemory
func runtime_debug_freeOSMemory() {
	_gc.Gogc(2) // force GC and do eager sweep
	_lock.Systemstack(scavenge_m)
}

//go:linkname sync_runtime_registerPoolCleanup sync.runtime_registerPoolCleanup
func sync_runtime_registerPoolCleanup(f func()) {
	_gc.Poolcleanup = f
}
