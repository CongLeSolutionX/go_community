// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package netpoll

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func Netpollcheckerr(pd *_sched.PollDesc, mode int32) int {
	if pd.Closing {
		return 1 // errClosing
	}
	if (mode == 'r' && pd.Rd < 0) || (mode == 'w' && pd.Wd < 0) {
		return 2 // errTimeout
	}
	return 0
}

func netpollblockcommit(gp *_core.G, gpp unsafe.Pointer) bool {
	return _core.Casuintptr((*uintptr)(gpp), _sched.PdWait, uintptr(unsafe.Pointer(gp)))
}

// returns true if IO is ready, or false if timedout or closed
// waitio - wait only for completed IO, ignore errors
func Netpollblock(pd *_sched.PollDesc, mode int32, waitio bool) bool {
	gpp := &pd.Rg
	if mode == 'w' {
		gpp = &pd.Wg
	}

	// set the gpp semaphore to WAIT
	for {
		old := *gpp
		if old == _sched.PdReady {
			*gpp = 0
			return true
		}
		if old != 0 {
			_lock.Throw("netpollblock: double wait")
		}
		if _core.Casuintptr(gpp, 0, _sched.PdWait) {
			break
		}
	}

	// need to recheck error states after setting gpp to WAIT
	// this is necessary because runtime_pollUnblock/runtime_pollSetDeadline/deadlineimpl
	// do the opposite: store to closing/rd/wd, membarrier, load of rg/wg
	if waitio || Netpollcheckerr(pd, mode) == 0 {
		_sched.Gopark(netpollblockcommit, unsafe.Pointer(gpp), "IO wait")
	}
	// be careful to not lose concurrent READY notification
	old := xchguintptr(gpp, 0)
	if old > _sched.PdWait {
		_lock.Throw("netpollblock: corrupted state")
	}
	return old == _sched.PdReady
}

func netpolldeadlineimpl(pd *_sched.PollDesc, seq uintptr, read, write bool) {
	_lock.Lock(&pd.Lock)
	// Seq arg is seq when the timer was set.
	// If it's stale, ignore the timer event.
	if seq != pd.Seq {
		// The descriptor was reused or timers were reset.
		_lock.Unlock(&pd.Lock)
		return
	}
	var rg *_core.G
	if read {
		if pd.Rd <= 0 || pd.Rt.F == nil {
			_lock.Throw("netpolldeadlineimpl: inconsistent read deadline")
		}
		pd.Rd = -1
		_sched.Atomicstorep(unsafe.Pointer(&pd.Rt.F), nil) // full memory barrier between store to rd and load of rg in netpollunblock
		rg = _sched.Netpollunblock(pd, 'r', false)
	}
	var wg *_core.G
	if write {
		if pd.Wd <= 0 || pd.Wt.F == nil && !read {
			_lock.Throw("netpolldeadlineimpl: inconsistent write deadline")
		}
		pd.Wd = -1
		_sched.Atomicstorep(unsafe.Pointer(&pd.Wt.F), nil) // full memory barrier between store to wd and load of wg in netpollunblock
		wg = _sched.Netpollunblock(pd, 'w', false)
	}
	_lock.Unlock(&pd.Lock)
	if rg != nil {
		_sched.Goready(rg)
	}
	if wg != nil {
		_sched.Goready(wg)
	}
}

func NetpollDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_sched.PollDesc), seq, true, true)
}

func NetpollReadDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_sched.PollDesc), seq, true, false)
}

func NetpollWriteDeadline(arg interface{}, seq uintptr) {
	netpolldeadlineimpl(arg.(*_sched.PollDesc), seq, false, true)
}
