// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

var (
	Systemstack_switchPC uintptr
)

// Traceback over the deferred function calls.
// Report them like calls that have been invoked but not started executing yet.
func tracebackdefers(gp *_core.G, callback func(*_lock.Stkframe, unsafe.Pointer) bool, v unsafe.Pointer) {
	var frame _lock.Stkframe
	for d := gp.Defer; d != nil; d = d.Link {
		fn := d.Fn
		if fn == nil {
			// Defer of nil function. Args don't matter.
			frame.Pc = 0
			frame.Fn = nil
			frame.Argp = 0
			frame.Arglen = 0
			frame.Argmap = nil
		} else {
			frame.Pc = uintptr(fn.Fn)
			f := _lock.Findfunc(frame.Pc)
			if f == nil {
				print("runtime: unknown pc in defer ", _core.Hex(frame.Pc), "\n")
				_lock.Gothrow("unknown pc")
			}
			frame.Fn = f
			frame.Argp = uintptr(DeferArgs(d))
			_lock.SetArgInfo(&frame, f, true)
		}
		frame.Continpc = frame.Pc
		if !callback((*_lock.Stkframe)(_core.Noescape(unsafe.Pointer(&frame))), v) {
			return
		}
	}
}
