// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package runtime

import (
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func sigenable(sig uint32) {
	if sig >= uint32(len(_sched.Sigtable)) {
		return
	}

	t := &_sched.Sigtable[sig]
	if t.Flags&_sched.SigNotify != 0 && t.Flags&_sched.SigHandling == 0 {
		t.Flags |= _sched.SigHandling
		if _sched.Getsig(int32(sig)) == _lock.SIG_IGN {
			t.Flags |= _sched.SigIgnored
		}
		_lock.Setsig(int32(sig), _lock.FuncPC(_sched.Sighandler), true)
	}
}

func sigdisable(sig uint32) {
	if sig >= uint32(len(_sched.Sigtable)) {
		return
	}

	t := &_sched.Sigtable[sig]
	if t.Flags&_sched.SigNotify != 0 && t.Flags&_sched.SigHandling != 0 {
		t.Flags &^= _sched.SigHandling
		if t.Flags&_sched.SigIgnored != 0 {
			_lock.Setsig(int32(sig), _lock.SIG_IGN, true)
		} else {
			_lock.Setsig(int32(sig), _lock.SIG_DFL, true)
		}
	}
}

func sigpipe() {
	_lock.Setsig(_lock.SIGPIPE, _lock.SIG_DFL, false)
	_lock.Raise(_lock.SIGPIPE)
}
