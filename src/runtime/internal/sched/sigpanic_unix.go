// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

func Sigpanic() {
	g := _core.Getg()
	if !canpanic(g) {
		_lock.Throw("unexpected signal during runtime execution")
	}

	switch g.Sig {
	case _lock.SIGBUS:
		if g.Sigcode0 == _lock.BUS_ADRERR && g.Sigcode1 < 0x1000 || g.Paniconfault {
			panicmem()
		}
		print("unexpected fault address ", _core.Hex(g.Sigcode1), "\n")
		_lock.Throw("fault")
	case _lock.SIGSEGV:
		if (g.Sigcode0 == 0 || g.Sigcode0 == _lock.SEGV_MAPERR || g.Sigcode0 == _lock.SEGV_ACCERR) && g.Sigcode1 < 0x1000 || g.Paniconfault {
			panicmem()
		}
		print("unexpected fault address ", _core.Hex(g.Sigcode1), "\n")
		_lock.Throw("fault")
	case _lock.SIGFPE:
		switch g.Sigcode0 {
		case _lock.FPE_INTDIV:
			panicdivide()
		case _lock.FPE_INTOVF:
			panicoverflow()
		}
		panicfloat()
	}

	if g.Sig >= uint32(len(Sigtable)) {
		// can't happen: we looked up g.sig in sigtable to decide to call sigpanic
		_lock.Throw("unexpected signal value")
	}
	panic(ErrorString(Sigtable[g.Sig].name))
}
