// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

func initsig() {
	// _NSIG is the number of signals on this operating system.
	// sigtable should describe what to do for all the possible signals.
	if len(Sigtable) != _core.NSIG {
		print("runtime: len(sigtable)=", len(Sigtable), " _NSIG=", _core.NSIG, "\n")
		_lock.Gothrow("initsig")
	}

	// First call: basic setup.
	for i := int32(0); i < _core.NSIG; i++ {
		t := &Sigtable[i]
		if t.Flags == 0 || t.Flags&SigDefault != 0 {
			continue
		}

		// For some signals, we respect an inherited SIG_IGN handler
		// rather than insist on installing our own default handler.
		// Even these signals can be fetched using the os/signal package.
		switch i {
		case _lock.SIGHUP, _lock.SIGINT:
			if Getsig(i) == _lock.SIG_IGN {
				t.Flags = SigNotify | SigIgnored
				continue
			}
		}

		if t.Flags&SigSetStack != 0 {
			setsigstack(i)
			continue
		}

		t.Flags |= SigHandling
		_lock.Setsig(i, _lock.FuncPC(Sighandler), true)
	}
}

func Resetcpuprofiler(hz int32) {
	var it itimerval
	if hz == 0 {
		setitimer(_lock.ITIMER_PROF, &it, nil)
	} else {
		it.it_interval.tv_sec = 0
		it.it_interval.set_usec(1000000 / hz)
		it.it_value = it.it_interval
		setitimer(_lock.ITIMER_PROF, &it, nil)
	}
	_g_ := _core.Getg()
	_g_.M.Profilehz = hz
}
