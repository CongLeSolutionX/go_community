// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package base

const (
	SIG_DFL uintptr = 0
	SIG_IGN uintptr = 1
)

// Stores the signal handlers registered before Go installed its own.
// These signal handlers will be invoked in cases where Go doesn't want to
// handle a particular signal (e.g., signal occurred on a non-Go thread).
// See sigfwdgo() for more information on when the signals are forwarded.
//
// Signal forwarding is currently available only on Linux.
var FwdSig [NSIG]uintptr

// sigmask represents a general signal mask compatible with the GOOS
// specific sigset types: the signal numbered x is represented by bit x-1
// to match the representation expected by sigprocmask.
type Sigmask [(NSIG + 31) / 32]uint32

func initsig() {
	// _NSIG is the number of signals on this operating system.
	// sigtable should describe what to do for all the possible signals.
	if len(Sigtable) != NSIG {
		print("runtime: len(sigtable)=", len(Sigtable), " _NSIG=", NSIG, "\n")
		Throw("initsig")
	}

	// First call: basic setup.
	for i := int32(0); i < NSIG; i++ {
		t := &Sigtable[i]
		if t.Flags == 0 || t.Flags&SigDefault != 0 {
			continue
		}
		FwdSig[i] = Getsig(i)
		// For some signals, we respect an inherited SIG_IGN handler
		// rather than insist on installing our own default handler.
		// Even these signals can be fetched using the os/signal package.
		switch i {
		case SIGHUP, SIGINT:
			if Getsig(i) == SIG_IGN {
				t.Flags = SigNotify | SigIgnored
				continue
			}
		}

		if t.Flags&SigSetStack != 0 {
			setsigstack(i)
			continue
		}

		t.Flags |= SigHandling
		Setsig(i, FuncPC(Sighandler), true)
	}
}

func Resetcpuprofiler(hz int32) {
	var it itimerval
	if hz == 0 {
		setitimer(ITIMER_PROF, &it, nil)
	} else {
		it.it_interval.tv_sec = 0
		it.it_interval.set_usec(1000000 / hz)
		it.it_value = it.it_interval
		setitimer(ITIMER_PROF, &it, nil)
	}
	_g_ := Getg()
	_g_.M.Profilehz = hz
}

func crash() {
	if GOOS == "darwin" {
		// OS X core dumps are linear dumps of the mapped memory,
		// from the first virtual byte to the last, with zeros in the gaps.
		// Because of the way we arrange the address space on 64-bit systems,
		// this means the OS X core file will be >128 GB and even on a zippy
		// workstation can take OS X well over an hour to write (uninterruptible).
		// Save users from making that mistake.
		if PtrSize == 8 {
			return
		}
	}

	Updatesigmask(Sigmask{})
	Setsig(SIGABRT, SIG_DFL, false)
	Raise(SIGABRT)
}
