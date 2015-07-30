// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
)

const hasLinkRegister = _base.GOARCH == "arm" || _base.GOARCH == "arm64" || _base.GOARCH == "ppc64" || _base.GOARCH == "ppc64le"

// Unwind the stack after a deferred function calls recover
// after a panic.  Then arrange to continue running as though
// the caller of the deferred function returned normally.
func recovery(gp *_base.G) {
	// Info about defer passed in G struct.
	sp := gp.Sigcode0
	pc := gp.Sigcode1

	// d's arguments need to be in the stack.
	if sp != 0 && (sp < gp.Stack.Lo || gp.Stack.Hi < sp) {
		print("recover: ", _base.Hex(sp), " not in [", _base.Hex(gp.Stack.Lo), ", ", _base.Hex(gp.Stack.Hi), "]\n")
		_base.Throw("bad recovery")
	}

	// Make the deferproc for this d return again,
	// this time returning 1.  The calling function will
	// jump to the standard return epilogue.
	_iface.GcUnwindBarriers(gp, sp)
	gp.Sched.Sp = sp
	gp.Sched.Pc = pc
	gp.Sched.Lr = 0
	gp.Sched.Ret = 1
	_base.Gogo(&gp.Sched)
}
