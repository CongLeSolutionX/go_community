// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prof

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

// Arrange to call fn with a traceback hz times a second.
func setcpuprofilerate_m(hz int32) {
	// Force sane arguments.
	if hz < 0 {
		hz = 0
	}

	// Disable preemption, otherwise we can be rescheduled to another thread
	// that has profiling enabled.
	_g_ := _core.Getg()
	_g_.M.Locks++

	// Stop profiler on this thread so that it is safe to lock prof.
	// if a profiling signal came in while we had prof locked,
	// it would deadlock.
	_sched.Resetcpuprofiler(0)

	for !_sched.Cas(&_sched.Prof.Lock, 0, 1) {
		_core.Osyield()
	}
	_sched.Prof.Hz = hz
	_lock.Atomicstore(&_sched.Prof.Lock, 0)

	_lock.Lock(&_core.Sched.Lock)
	_core.Sched.Profilehz = hz
	_lock.Unlock(&_core.Sched.Lock)

	if hz != 0 {
		_sched.Resetcpuprofiler(hz)
	}

	_g_.M.Locks--
}
