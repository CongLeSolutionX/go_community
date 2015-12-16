// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that a signal handler that uses up stack space does not crash
// if the signal is delivered to a thread running a goroutine.

#include <setjmp.h>
#include <signal.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <semaphore.h>
#include <dlfcn.h>

static void die(const char* msg) {
	perror(msg);
	exit(EXIT_FAILURE);
}

static sem_t sigioSem;

// Use up some stack space.
static recur(int i, char *p) {
	char a[1024];

	*p = '\0';
	if (i > 0) {
		recur(i - 1, a);
	}
}

// Signal handler that uses up more stack space than a goroutine will have.
static void ioHandler(int signo, siginfo_t* info, void* ctxt) {
	char a[1024];

	recur(4, a);
	if (sem_post(&sigioSem) < 0) {
		die("sem_post");
	}
}

static jmp_buf jmp;
static char* nullPointer;

// Signal handler for SIGSEGV on a C thread.
static void segvHandler(int signo, siginfo_t* info, void* ctxt) {
	// Don't try this at home.
	longjmp(jmp, signo);

	// We should never get here.
	abort();
}

int main(int argc, char** argv) {
	struct sigaction sa;
	void* handle;
	void (*fn)(void);
	sigset_t mask;
	struct timespec ts;
	int i;

	memset(&sa, 0, sizeof sa);
	sa.sa_sigaction = ioHandler;
	if (sigemptyset(&sa.sa_mask) < 0) {
		die("sigemptyset");
	}
	sa.sa_flags = SA_SIGINFO;
	if (sigaction(SIGIO, &sa, NULL) < 0) {
		die("sigaction");
	}

	sa.sa_sigaction = segvHandler;
	if (sigaction(SIGSEGV, &sa, NULL) < 0 || sigaction(SIGBUS, &sa, NULL) < 0) {
		die("sigaction");
	}

	handle = dlopen(argv[1], RTLD_NOW | RTLD_GLOBAL);
	if (handle == NULL) {
		fprintf(stderr, "%s\n", dlerror());
		exit(EXIT_FAILURE);
	}

	// Start some goroutines.
	fn = (void(*)(void))dlsym(handle, "RunGoroutines");
	if (fn == NULL) {
		fprintf(stderr, "%s\n", dlerror());
		exit(EXIT_FAILURE);
	}
	fn();

	if (sem_init(&sigioSem, 0, 0) < 0) {
		die("sem_init");
	}

	// Block SIGIO in this thread to make it more likely that it
	// will be delivered to a goroutine.
	if (sigemptyset(&mask) < 0) {
		die("sigemptyset");
	}
	if (sigaddset(&mask, SIGIO) < 0) {
		die("sigaddset");
	}
	i = pthread_sigmask(SIG_BLOCK, &mask, NULL);
	if (i != 0) {
		fprintf(stderr, "pthread_sigmask: %s\n", strerror(i));
		exit(EXIT_FAILURE);
	}

	if (kill(getpid(), SIGIO) < 0) {
		die("kill");
	}

	// Wait until the signal has been delivered.
	ts.tv_sec = time(NULL) + 2;
	ts.tv_nsec = 0;
	if (sem_timedwait(&sigioSem, &ts) < 0) {
		die("sem_timedwait");
	}

	// Test that a SIGSEGV on this thread is delivered to us.
	if (setjmp(jmp) == 0) {
		*nullPointer = '\0';

		fprintf(stderr, "continued after address error\n");
		exit(EXIT_FAILURE);
	}

	// Make sure that a SIGSEGV in Go causes a run-time panic.
	fn = (void (*)(void))dlsym(handle, "TestSEGV");
	if (fn == NULL) {
		fprintf(stderr, "%s\n", dlerror());
		exit(EXIT_FAILURE);
	}
	fn();

	printf("PASS\n");
	return 0;
}
