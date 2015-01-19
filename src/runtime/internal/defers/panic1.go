// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package defers

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// Unwind the stack after a deferred function calls recover
// after a panic.  Then arrange to continue running as though
// the caller of the deferred function returned normally.
func recovery(gp *_core.G) {
	// Info about defer passed in G struct.
	sp := gp.Sigcode0
	pc := gp.Sigcode1

	// d's arguments need to be in the stack.
	if sp != 0 && (sp < gp.Stack.Lo || gp.Stack.Hi < sp) {
		print("recover: ", _core.Hex(sp), " not in [", _core.Hex(gp.Stack.Lo), ", ", _core.Hex(gp.Stack.Hi), "]\n")
		_lock.Gothrow("bad recovery")
	}

	// Make the deferproc for this d return again,
	// this time returning 1.  The calling function will
	// jump to the standard return epilogue.
	gp.Sched.Sp = sp
	gp.Sched.Pc = pc
	gp.Sched.Lr = 0
	gp.Sched.Ret = 1
	_sched.Gogo(&gp.Sched)
}
