// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin nacl netbsd openbsd plan9 solaris windows

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func Notetsleep(n *_core.Note, ns int64) bool {
	gp := _core.Getg()
	if gp != gp.M.G0 && gp.M.Preemptoff != "" {
		_lock.Throw("notetsleep not on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = _lock.Semacreate()
	}
	return _sched.Notetsleep_internal(n, ns, nil, 0)
}
