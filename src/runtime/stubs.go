// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	"unsafe"
)

func badsystemstack() {
	_lock.Throw("systemstack called from unexpected goroutine")
}

//go:linkname reflect_memmove reflect.memmove
func reflect_memmove(to, from unsafe.Pointer, n uintptr) {
	_sched.Memmove(to, from, n)
}

// exported value for testing
var hashLoad = _maps.LoadFactor

func cgocallback(fn, frame unsafe.Pointer, framesize uintptr)
func mincore(addr unsafe.Pointer, n uintptr, dst *byte) int32
func exit1(code int32)
func breakpoint()
func cgocallback_gofunc(fv *_core.Funcval, frame unsafe.Pointer, framesize uintptr)

// return0 is a stub used to return 0 from deferproc.
// It is called at the very end of deferproc to signal
// the calling Go function that it should not jump
// to deferreturn.
// in asm_*.s
func return0()
