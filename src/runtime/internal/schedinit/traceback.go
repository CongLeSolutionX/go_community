// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func tracebackinit() {
	// Go variable initialization happens late during runtime startup.
	// Instead of initializing the variables above in the declarations,
	// schedinit calls this function so that the variables are
	// initialized and available earlier in the startup sequence.
	_lock.GoexitPC = _lock.FuncPC(Goexit)
	_lock.JmpdeferPC = _lock.FuncPC(Jmpdefer)
	_lock.McallPC = _lock.FuncPC(_sched.Mcall)
	_lock.MorestackPC = _lock.FuncPC(morestack)
	_lock.MstartPC = _lock.FuncPC(_sched.Mstart)
	_lock.Rt0_goPC = _lock.FuncPC(rt0_go)
	_lock.SigpanicPC = _lock.FuncPC(_sched.Sigpanic)
	_gc.Systemstack_switchPC = _lock.FuncPC(systemstack_switch)
}
