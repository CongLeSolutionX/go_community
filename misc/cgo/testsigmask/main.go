// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <signal.h>
#include <stdlib.h>
#include <stdio.h>

static void checkmask() {
	sigset_t mask;
	sigprocmask(SIG_BLOCK, NULL, &mask);
	if (sigismember(&mask, SIGPIPE) != 1) {
		fprintf(stderr, "ERROR: SIGPIPE was unblocked by the Go runtime\n");
		exit(2);
	}
}

static void __attribute__ ((constructor)) sigsetup(void) {
	sigset_t mask;
	sigemptyset(&mask);
	sigaddset(&mask, SIGPIPE);
	sigprocmask(SIG_BLOCK, &mask, NULL);
}
*/
import "C"
import "runtime"

func init() {
	runtime.LockOSThread()
}

func main() {
	C.checkmask()
}
