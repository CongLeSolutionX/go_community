// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func goenvs() {
	goenvs_unix()

	// Register our thread-creation callback (see sys_darwin_{amd64,386}.s)
	// but only if we're not using cgo.  If we are using cgo we need
	// to let the C pthread library install its own thread-creation callback.
	if !_sched.Iscgo {
		if bsdthread_register() != 0 {
			if Gogetenv("DYLD_INSERT_LIBRARIES") != "" {
				_lock.Gothrow("runtime: bsdthread_register error (unset DYLD_INSERT_LIBRARIES)")
			}
			_lock.Gothrow("runtime: bsdthread_register error")
		}
	}
}
