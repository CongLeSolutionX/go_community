// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	"unsafe"
)

//go:linkname reflect_unsafe_New reflect.unsafe_New
func reflect_unsafe_New(typ *_core.Type) unsafe.Pointer {
	return _maps.Newobject(typ)
}

//go:linkname reflect_unsafe_NewArray reflect.unsafe_NewArray
func reflect_unsafe_NewArray(typ *_core.Type, n uintptr) unsafe.Pointer {
	return _maps.Newarray(typ, n)
}

func GCcheckmarkenable() {
	_lock.Systemstack(gccheckmarkenable_m)
}

func GCcheckmarkdisable() {
	_lock.Systemstack(gccheckmarkdisable_m)
}

// GCstarttimes initializes the gc timess. All previous timess are lost.
func GCstarttimes(verbose int64) {
	_gc.Gctimer = _gc.Gcchronograph{Verbose: verbose}
}

// GCendtimes stops the gc timers.
func GCendtimes() {
	_gc.Gctimer.Verbose = 0
}
