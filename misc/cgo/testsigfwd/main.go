// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <signal.h>
#include <stdlib.h>
#include <stdio.h>
#ifndef _WIN32
#include <pthread.h>
#endif

int *p;
static void sigsegv() {
	*p = 1;
	fprintf(stderr, "ERROR: C SIGSEGV not thrown on caught?\n");
	exit(2);
}

static void sighandler(int signum) {
	if (signum == SIGSEGV) {
		exit(0);  // success
	}
}

static void __attribute__ ((constructor)) sigsetup(void) {
	struct sigaction act;
	act.sa_handler = &sighandler;
	sigaction(SIGSEGV, &act, 0);
}

void *crash(void *p) {
	raise(SIGABRT);
	return NULL;
}

#ifdef _WIN32
int start_crashing_thread(void) {
	exit(1);
}
#else
int start_crashing_thread(void) {
	pthread_t tid;
	return pthread_create(&tid, NULL, crash, NULL);
}
#endif
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

var p *byte

func f() (ret bool) {
	defer func() {
		if recover() == nil {
			fmt.Fprintf(os.Stderr, "ERROR: couldn't raise SIGSEGV in Go\n")
			C.exit(2)
		}
		ret = true
	}()
	*p = 1
	return false
}

func crashInCThread() {
	// Calling this will cause a C thread to raise SIGABRT, which
	// should cause the program to crash.
	C.start_crashing_thread()

	// It should crash immediately, but give it plenty of time in
	// case we are running on a slow system.
	time.Sleep(5 * time.Second)
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "crashInCThread" {
		crashInCThread()
		return
	}

	// Test for issue 10139: a segmentation violation in a C
	// thread should cause the program to crash.
	if exec.Command(os.Args[0], "crashInCThread").Run() == nil {
		fmt.Fprintf(os.Stderr, "C signal did not crash as expected\n")
		C.exit(2)
	}

	// Test that the signal originating in Go is handled (and recovered) by Go.
	if !f() {
		fmt.Fprintf(os.Stderr, "couldn't recover from SIGSEGV in Go\n")
		C.exit(2)
	}

	// Test that the signal originating in C is handled by C.
	C.sigsegv()
}
