// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

//go:nosplit
func canpanic(gp *_core.G) bool {
	// Note that g is m->gsignal, different from gp.
	// Note also that g->m can change at preemption, so m can go stale
	// if this function ever makes a function call.
	_g_ := _core.Getg()
	_m_ := _g_.M

	// Is it okay for gp to panic instead of crashing the program?
	// Yes, as long as it is running Go code, not runtime code,
	// and not stuck in a system call.
	if gp == nil || gp != _m_.Curg {
		return false
	}
	if _m_.Locks-_m_.Softfloat != 0 || _m_.Mallocing != 0 || _m_.Throwing != 0 || _m_.Gcing != 0 || _m_.Dying != 0 {
		return false
	}
	status := _lock.Readgstatus(gp)
	if status&^_lock.Gscan != _lock.Grunning || gp.Syscallsp != 0 {
		return false
	}
	if _lock.GOOS == "windows" && _m_.Libcallsp != 0 {
		return false
	}
	return true
}
