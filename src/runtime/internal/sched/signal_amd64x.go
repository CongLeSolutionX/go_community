// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

func dumpregs(c *sigctxt) {
	print("rax    ", _core.Hex(c.rax()), "\n")
	print("rbx    ", _core.Hex(c.rbx()), "\n")
	print("rcx    ", _core.Hex(c.rcx()), "\n")
	print("rdx    ", _core.Hex(c.rdx()), "\n")
	print("rdi    ", _core.Hex(c.rdi()), "\n")
	print("rsi    ", _core.Hex(c.rsi()), "\n")
	print("rbp    ", _core.Hex(c.rbp()), "\n")
	print("rsp    ", _core.Hex(c.rsp()), "\n")
	print("r8     ", _core.Hex(c.r8()), "\n")
	print("r9     ", _core.Hex(c.r9()), "\n")
	print("r10    ", _core.Hex(c.r10()), "\n")
	print("r11    ", _core.Hex(c.r11()), "\n")
	print("r12    ", _core.Hex(c.r12()), "\n")
	print("r13    ", _core.Hex(c.r13()), "\n")
	print("r14    ", _core.Hex(c.r14()), "\n")
	print("r15    ", _core.Hex(c.r15()), "\n")
	print("rip    ", _core.Hex(c.rip()), "\n")
	print("rflags ", _core.Hex(c.rflags()), "\n")
	print("cs     ", _core.Hex(c.cs()), "\n")
	print("fs     ", _core.Hex(c.fs()), "\n")
	print("gs     ", _core.Hex(c.gs()), "\n")
}

func Sighandler(sig uint32, info *siginfo, ctxt unsafe.Pointer, gp *_core.G) {
	_g_ := _core.Getg()
	c := &sigctxt{info, ctxt}

	if sig == _lock.SIGPROF {
		sigprof((*byte)(unsafe.Pointer(uintptr(c.rip()))), (*byte)(unsafe.Pointer(uintptr(c.rsp()))), nil, gp, _g_.M)
		return
	}

	if _lock.GOOS == "darwin" {
		// x86-64 has 48-bit virtual addresses. The top 16 bits must echo bit 47.
		// The hardware delivers a different kind of fault for a malformed address
		// than it does for an attempt to access a valid but unmapped address.
		// OS X 10.9.2 mishandles the malformed address case, making it look like
		// a user-generated signal (like someone ran kill -SEGV ourpid).
		// We pass user-generated signals to os/signal, or else ignore them.
		// Doing that here - and returning to the faulting code - results in an
		// infinite loop. It appears the best we can do is rewrite what the kernel
		// delivers into something more like the truth. The address used below
		// has very little chance of being the one that caused the fault, but it is
		// malformed, it is clearly not a real pointer, and if it does get printed
		// in real life, people will probably search for it and find this code.
		// There are no Google hits for b01dfacedebac1e or 0xb01dfacedebac1e
		// as I type this comment.
		if sig == _lock.SIGSEGV && c.sigcode() == _core.SI_USER {
			c.set_sigcode(_core.SI_USER + 1)
			c.set_sigaddr(0xb01dfacedebac1e)
		}
	}

	flags := int32(SigThrow)
	if sig < uint32(len(Sigtable)) {
		flags = Sigtable[sig].Flags
	}
	if c.sigcode() != _core.SI_USER && flags&SigPanic != 0 {
		// Make it look like a call to the signal func.
		// Have to pass arguments out of band since
		// augmenting the stack frame would break
		// the unwinding code.
		gp.Sig = sig
		gp.Sigcode0 = uintptr(c.sigcode())
		gp.Sigcode1 = uintptr(c.sigaddr())
		gp.Sigpc = uintptr(c.rip())

		if _lock.GOOS == "darwin" {
			// Work around Leopard bug that doesn't set FPE_INTDIV.
			// Look at instruction to see if it is a divide.
			// Not necessary in Snow Leopard (si_code will be != 0).
			if sig == _lock.SIGFPE && gp.Sigcode0 == 0 {
				pc := (*[4]byte)(unsafe.Pointer(gp.Sigpc))
				i := 0
				if pc[i]&0xF0 == 0x40 { // 64-bit REX prefix
					i++
				} else if pc[i] == 0x66 { // 16-bit instruction prefix
					i++
				}
				if pc[i] == 0xF6 || pc[i] == 0xF7 {
					gp.Sigcode0 = _lock.FPE_INTDIV
				}
			}
		}

		// Only push runtime.sigpanic if rip != 0.
		// If rip == 0, probably panicked because of a
		// call to a nil func.  Not pushing that onto sp will
		// make the trace look like a call to runtime.sigpanic instead.
		// (Otherwise the trace will end at runtime.sigpanic and we
		// won't get to see who faulted.)
		if c.rip() != 0 {
			sp := c.rsp()
			if _lock.RegSize > _core.PtrSize {
				sp -= _core.PtrSize
				*(*uintptr)(unsafe.Pointer(uintptr(sp))) = 0
			}
			sp -= _core.PtrSize
			*(*uintptr)(unsafe.Pointer(uintptr(sp))) = uintptr(c.rip())
			c.set_rsp(sp)
		}
		c.set_rip(uint64(_lock.FuncPC(Sigpanic)))
		return
	}

	if c.sigcode() == _core.SI_USER || flags&SigNotify != 0 {
		if Sigsend(sig) {
			return
		}
	}

	if flags&SigKill != 0 {
		_core.Exit(2)
	}

	if flags&SigThrow == 0 {
		return
	}

	_g_.M.Throwing = 1
	_g_.M.Caughtsig = gp
	_lock.Startpanic()

	if sig < uint32(len(Sigtable)) {
		print(Sigtable[sig].name, "\n")
	} else {
		print("Signal ", sig, "\n")
	}

	print("PC=", _core.Hex(c.rip()), "\n")
	if _g_.M.Lockedg != nil && _g_.M.Ncgo > 0 && gp == _g_.M.G0 {
		print("signal arrived during cgo execution\n")
		gp = _g_.M.Lockedg
	}
	print("\n")

	var docrash bool
	if _lock.Gotraceback(&docrash) > 0 {
		_lock.Goroutineheader(gp)
		tracebacktrap(uintptr(c.rip()), uintptr(c.rsp()), 0, gp)
		_lock.Tracebackothers(gp)
		print("\n")
		dumpregs(c)
	}

	if docrash {
		_lock.Crash()
	}

	_core.Exit(2)
}
