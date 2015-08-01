// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64 amd64p32
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package base

import (
	"unsafe"
)

func dumpregs(c *Sigctxt) {
	print("rax    ", Hex(c.rax()), "\n")
	print("rbx    ", Hex(c.rbx()), "\n")
	print("rcx    ", Hex(c.rcx()), "\n")
	print("rdx    ", Hex(c.rdx()), "\n")
	print("rdi    ", Hex(c.rdi()), "\n")
	print("rsi    ", Hex(c.rsi()), "\n")
	print("rbp    ", Hex(c.rbp()), "\n")
	print("rsp    ", Hex(c.rsp()), "\n")
	print("r8     ", Hex(c.r8()), "\n")
	print("r9     ", Hex(c.r9()), "\n")
	print("r10    ", Hex(c.r10()), "\n")
	print("r11    ", Hex(c.r11()), "\n")
	print("r12    ", Hex(c.r12()), "\n")
	print("r13    ", Hex(c.r13()), "\n")
	print("r14    ", Hex(c.r14()), "\n")
	print("r15    ", Hex(c.r15()), "\n")
	print("rip    ", Hex(c.rip()), "\n")
	print("rflags ", Hex(c.rflags()), "\n")
	print("cs     ", Hex(c.cs()), "\n")
	print("fs     ", Hex(c.fs()), "\n")
	print("gs     ", Hex(c.gs()), "\n")
}

var crashing int32

// May run during STW, so write barriers are not allowed.
//go:nowritebarrier
func Sighandler(sig uint32, info *Siginfo, ctxt unsafe.Pointer, gp *G) {
	_g_ := Getg()
	c := &Sigctxt{info, ctxt}

	if sig == SIGPROF {
		sigprof(uintptr(c.rip()), uintptr(c.rsp()), 0, gp, _g_.M)
		return
	}

	if GOOS == "darwin" {
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
		if sig == SIGSEGV && c.Sigcode() == SI_USER {
			c.set_sigcode(SI_USER + 1)
			c.set_sigaddr(0xb01dfacedebac1e)
		}
	}

	flags := int32(SigThrow)
	if sig < uint32(len(Sigtable)) {
		flags = Sigtable[sig].Flags
	}
	if c.Sigcode() != SI_USER && flags&SigPanic != 0 {
		// Make it look like a call to the signal func.
		// Have to pass arguments out of band since
		// augmenting the stack frame would break
		// the unwinding code.
		gp.Sig = sig
		gp.Sigcode0 = uintptr(c.Sigcode())
		gp.Sigcode1 = uintptr(c.sigaddr())
		gp.Sigpc = uintptr(c.rip())

		if GOOS == "darwin" {
			// Work around Leopard bug that doesn't set FPE_INTDIV.
			// Look at instruction to see if it is a divide.
			// Not necessary in Snow Leopard (si_code will be != 0).
			if sig == SIGFPE && gp.Sigcode0 == 0 {
				pc := (*[4]byte)(unsafe.Pointer(gp.Sigpc))
				i := 0
				if pc[i]&0xF0 == 0x40 { // 64-bit REX prefix
					i++
				} else if pc[i] == 0x66 { // 16-bit instruction prefix
					i++
				}
				if pc[i] == 0xF6 || pc[i] == 0xF7 {
					gp.Sigcode0 = FPE_INTDIV
				}
			}
		}

		pc := uintptr(c.rip())
		sp := uintptr(c.rsp())

		// If we don't recognize the PC as code
		// but we do recognize the top pointer on the stack as code,
		// then assume this was a call to non-code and treat like
		// pc == 0, to make unwinding show the context.
		if pc != 0 && Findfunc(pc) == nil && Findfunc(*(*uintptr)(unsafe.Pointer(sp))) != nil {
			pc = 0
		}

		// Only push runtime.sigpanic if pc != 0.
		// If pc == 0, probably panicked because of a
		// call to a nil func.  Not pushing that onto sp will
		// make the trace look like a call to runtime.sigpanic instead.
		// (Otherwise the trace will end at runtime.sigpanic and we
		// won't get to see who faulted.)
		if pc != 0 {
			if RegSize > PtrSize {
				sp -= PtrSize
				*(*uintptr)(unsafe.Pointer(sp)) = 0
			}
			sp -= PtrSize
			*(*uintptr)(unsafe.Pointer(sp)) = pc
			c.set_rsp(uint64(sp))
		}
		c.set_rip(uint64(FuncPC(Sigpanic)))
		return
	}

	if c.Sigcode() == SI_USER || flags&SigNotify != 0 {
		if Sigsend(sig) {
			return
		}
	}

	if flags&SigKill != 0 {
		Exit(2)
	}

	if flags&SigThrow == 0 {
		return
	}

	_g_.M.Throwing = 1
	_g_.M.caughtsig.Set(gp)

	if crashing == 0 {
		Startpanic()
	}

	if sig < uint32(len(Sigtable)) {
		print(Sigtable[sig].name, "\n")
	} else {
		print("Signal ", sig, "\n")
	}

	print("PC=", Hex(c.rip()), " m=", _g_.M.Id, "\n")
	if _g_.M.Lockedg != nil && _g_.M.Ncgo > 0 && gp == _g_.M.G0 {
		print("signal arrived during cgo execution\n")
		gp = _g_.M.Lockedg
	}
	print("\n")

	var docrash bool
	if gotraceback(&docrash) > 0 {
		Goroutineheader(gp)
		tracebacktrap(uintptr(c.rip()), uintptr(c.rsp()), 0, gp)
		if crashing > 0 && gp != _g_.M.Curg && _g_.M.Curg != nil && Readgstatus(_g_.M.Curg)&^Gscan == Grunning {
			// tracebackothers on original m skipped this one; trace it now.
			Goroutineheader(_g_.M.Curg)
			Traceback(^uintptr(0), ^uintptr(0), 0, gp)
		} else if crashing == 0 {
			Tracebackothers(gp)
			print("\n")
		}
		dumpregs(c)
	}

	if docrash {
		// TODO(rsc): Implement raiseproc on other systems
		// and then add to this if condition.
		if GOOS == "darwin" || GOOS == "linux" {
			crashing++
			if crashing < Sched.mcount {
				// There are other m's that need to dump their stacks.
				// Relay SIGQUIT to the next m by sending it to the current process.
				// All m's that have already received SIGQUIT have signal masks blocking
				// receipt of any signals, so the SIGQUIT will go to an m that hasn't seen it yet.
				// When the last m receives the SIGQUIT, it will fall through to the call to
				// crash below. Just in case the relaying gets botched, each m involved in
				// the relay sleeps for 5 seconds and then does the crash/exit itself.
				// In expected operation, the last m has received the SIGQUIT and run
				// crash/exit and the process is gone, all long before any of the
				// 5-second sleeps have finished.
				print("\n-----\n\n")
				raiseproc(SIGQUIT)
				Usleep(5 * 1000 * 1000)
			}
		}
		crash()
	}

	Exit(2)
}
