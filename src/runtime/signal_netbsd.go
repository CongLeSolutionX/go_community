// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

type sigTabT struct {
	flags int32
	name  string
}

var sigtable = [...]sigTabT{
	/*  0 */ {0, "SIGNONE: no trap"},
	/*  1 */ {_SigNotify + _SigKill + _SigFatalDFL, "SIGHUP: terminal line hangup"},
	/*  2 */ {_SigNotify + _SigKill + _SigFatalDFL, "SIGINT: interrupt"},
	/*  3 */ {_SigNotify + _SigThrow + _SigFatalDFL, "SIGQUIT: quit"},
	/*  4 */ {_SigThrow + _SigUnblock + _SigFatalDFL, "SIGILL: illegal instruction"},
	/*  5 */ {_SigThrow + _SigUnblock + _SigFatalDFL, "SIGTRAP: trace trap"},
	/*  6 */ {_SigNotify + _SigThrow + _SigFatalDFL, "SIGABRT: abort"},
	/*  7 */ {_SigThrow + _SigFatalDFL, "SIGEMT: emulate instruction executed"},
	/*  8 */ {_SigPanic + _SigUnblock + _SigFatalDFL, "SIGFPE: floating-point exception"},
	/*  9 */ {0, "SIGKILL: kill"},
	/* 10 */ {_SigPanic + _SigUnblock + _SigFatalDFL, "SIGBUS: bus error"},
	/* 11 */ {_SigPanic + _SigUnblock + _SigFatalDFL, "SIGSEGV: segmentation violation"},
	/* 12 */ {_SigThrow + _SigFatalDFL, "SIGSYS: bad system call"},
	/* 13 */ {_SigNotify + _SigFatalDFL, "SIGPIPE: write to broken pipe"},
	/* 14 */ {_SigNotify + _SigFatalDFL, "SIGALRM: alarm clock"},
	/* 15 */ {_SigNotify + _SigKill + _SigFatalDFL, "SIGTERM: termination"},
	/* 16 */ {_SigNotify, "SIGURG: urgent condition on socket"},
	/* 17 */ {0, "SIGSTOP: stop"},
	/* 18 */ {_SigNotify + _SigDefault, "SIGTSTP: keyboard stop"},
	/* 19 */ {_SigNotify + _SigDefault, "SIGCONT: continue after stop"},
	/* 20 */ {_SigNotify + _SigUnblock, "SIGCHLD: child status has changed"},
	/* 21 */ {_SigNotify + _SigDefault, "SIGTTIN: background read from tty"},
	/* 22 */ {_SigNotify + _SigDefault, "SIGTTOU: background write to tty"},
	/* 23 */ {_SigNotify, "SIGIO: i/o now possible"},
	/* 24 */ {_SigNotify + _SigFatalDFL, "SIGXCPU: cpu limit exceeded"},
	/* 25 */ {_SigNotify + _SigFatalDFL, "SIGXFSZ: file size limit exceeded"},
	/* 26 */ {_SigNotify + _SigFatalDFL, "SIGVTALRM: virtual alarm clock"},
	/* 27 */ {_SigNotify + _SigUnblock + _SigFatalDFL, "SIGPROF: profiling alarm clock"},
	/* 28 */ {_SigNotify, "SIGWINCH: window size change"},
	/* 29 */ {_SigNotify, "SIGINFO: status request from keyboard"},
	/* 30 */ {_SigNotify + _SigFatalDFL, "SIGUSR1: user-defined signal 1"},
	/* 31 */ {_SigNotify + _SigFatalDFL, "SIGUSR2: user-defined signal 2"},
	/* 32 */ {_SigNotify, "SIGTHR: reserved"},
}
