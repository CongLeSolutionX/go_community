// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that using the Go profiler in a C program does not crash.

#include <signal.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>

#include "libgo6.h"

static void die(const char* msg) {
	perror(msg);
	exit(EXIT_FAILURE);
}

static void dummyProfHandler(int signo, siginfo_t* info, void* ctxt) {}

int main(int argc, char **argv) {
	struct timeval tvstart, tvnow;
	int diff;
	struct sigaction sa;

	gettimeofday(&tvstart, NULL);

	go_start_profile();

	// Busy wait so we have something to profile.
	// If we just sleep the profiling signal will never fire.
	while (1) {
		gettimeofday(&tvnow, NULL);
		diff = (tvnow.tv_sec - tvstart.tv_sec) * 1000 * 1000 + (tvnow.tv_usec - tvstart.tv_usec);

		// Profile frequency is 100Hz so we should definitely
		// get a signal in 50 milliseconds.
		if (diff > 50 * 1000)
			break;
	}

	// On (at least) Darwin, SIGPROF can be delivered even after the profiling timer has stopped.
	// See issue #19320 and #22151. Register a dummy SIGPROF handler before stopping profiling
	// to avoid crashing.
	memset(&sa, 0, sizeof sa);
	sa.sa_sigaction = dummyProfHandler;
	if (sigemptyset(&sa.sa_mask) < 0) {
		die("sigemptyset");
	}
	sa.sa_flags = SA_SIGINFO;
	if (sigaction(SIGPROF, &sa, NULL) < 0) {
		die("sigaction");
	}

	go_stop_profile();
	return 0;
}
