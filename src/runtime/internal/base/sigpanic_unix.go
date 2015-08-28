// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package base

func Sigpanic() {
	g := Getg()
	if !canpanic(g) {
		Throw("unexpected signal during runtime execution")
	}

	switch g.Sig {
	case SIGBUS:
		if g.Sigcode0 == BUS_ADRERR && g.Sigcode1 < 0x1000 || g.Paniconfault {
			panicmem()
		}
		print("unexpected fault address ", Hex(g.Sigcode1), "\n")
		Throw("fault")
	case SIGSEGV:
		if (g.Sigcode0 == 0 || g.Sigcode0 == SEGV_MAPERR || g.Sigcode0 == SEGV_ACCERR) && g.Sigcode1 < 0x1000 || g.Paniconfault {
			panicmem()
		}
		print("unexpected fault address ", Hex(g.Sigcode1), "\n")
		Throw("fault")
	case SIGFPE:
		switch g.Sigcode0 {
		case FPE_INTDIV:
			panicdivide()
		case FPE_INTOVF:
			panicoverflow()
		}
		panicfloat()
	}

	if g.Sig >= uint32(len(Sigtable)) {
		// can't happen: we looked up g.sig in sigtable to decide to call sigpanic
		Throw("unexpected signal value")
	}
	panic(ErrorString(Sigtable[g.Sig].name))
}
