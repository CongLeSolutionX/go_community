// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

type SigTabT struct {
	Flags int32
	name  string
}

var Sigtable = [...]SigTabT{
	/* 0 */ {0, "SIGNONE: no trap"},
	/* 1 */ {SigNotify + SigKill, "SIGHUP: terminal line hangup"},
	/* 2 */ {SigNotify + SigKill, "SIGINT: interrupt"},
	/* 3 */ {SigNotify + SigThrow, "SIGQUIT: quit"},
	/* 4 */ {SigThrow + SigUnblock, "SIGILL: illegal instruction"},
	/* 5 */ {SigThrow + SigUnblock, "SIGTRAP: trace trap"},
	/* 6 */ {SigNotify + SigThrow, "SIGABRT: abort"},
	/* 7 */ {SigThrow, "SIGEMT: emulate instruction executed"},
	/* 8 */ {SigPanic + SigUnblock, "SIGFPE: floating-point exception"},
	/* 9 */ {0, "SIGKILL: kill"},
	/* 10 */ {SigPanic + SigUnblock, "SIGBUS: bus error"},
	/* 11 */ {SigPanic + SigUnblock, "SIGSEGV: segmentation violation"},
	/* 12 */ {SigThrow, "SIGSYS: bad system call"},
	/* 13 */ {SigNotify, "SIGPIPE: write to broken pipe"},
	/* 14 */ {SigNotify, "SIGALRM: alarm clock"},
	/* 15 */ {SigNotify + SigKill, "SIGTERM: termination"},
	/* 16 */ {SigNotify, "SIGURG: urgent condition on socket"},
	/* 17 */ {0, "SIGSTOP: stop"},
	/* 18 */ {SigNotify + SigDefault, "SIGTSTP: keyboard stop"},
	/* 19 */ {0, "SIGCONT: continue after stop"},
	/* 20 */ {SigNotify + SigUnblock, "SIGCHLD: child status has changed"},
	/* 21 */ {SigNotify + SigDefault, "SIGTTIN: background read from tty"},
	/* 22 */ {SigNotify + SigDefault, "SIGTTOU: background write to tty"},
	/* 23 */ {SigNotify, "SIGIO: i/o now possible"},
	/* 24 */ {SigNotify, "SIGXCPU: cpu limit exceeded"},
	/* 25 */ {SigNotify, "SIGXFSZ: file size limit exceeded"},
	/* 26 */ {SigNotify, "SIGVTALRM: virtual alarm clock"},
	/* 27 */ {SigNotify + SigUnblock, "SIGPROF: profiling alarm clock"},
	/* 28 */ {SigNotify, "SIGWINCH: window size change"},
	/* 29 */ {SigNotify, "SIGINFO: status request from keyboard"},
	/* 30 */ {SigNotify, "SIGUSR1: user-defined signal 1"},
	/* 31 */ {SigNotify, "SIGUSR2: user-defined signal 2"},
}
