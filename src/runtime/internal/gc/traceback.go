// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

var (
	Systemstack_switchPC uintptr
)

// Traceback over the deferred function calls.
// Report them like calls that have been invoked but not started executing yet.
func tracebackdefers(gp *_base.G, callback func(*_base.Stkframe, unsafe.Pointer) bool, v unsafe.Pointer) {
	var frame _base.Stkframe
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
			f := _base.Findfunc(frame.Pc)
			if f == nil {
				print("runtime: unknown pc in defer ", _base.Hex(frame.Pc), "\n")
				_base.Throw("unknown pc")
			}
			frame.Fn = f
			frame.Argp = uintptr(DeferArgs(d))
			_base.SetArgInfo(&frame, f, true)
		}
		frame.Continpc = frame.Pc
		if !callback((*_base.Stkframe)(_base.Noescape(unsafe.Pointer(&frame))), v) {
			return
		}
	}
}
