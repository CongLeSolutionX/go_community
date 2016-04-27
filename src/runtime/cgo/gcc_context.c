// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

#include <pthread.h>
#include <errno.h>
#include <string.h>

#include "libcgo.h"

// The context function, used when tracing back C calls into Go.
void (*x_cgo_context_function)(struct context_arg*);

// The pthread key to hold the context value until Go has figured out
// the goroutine in which to store it. This could be a __thread
// variable but pthread keys are sufficient.
static pthread_key_t key;

// Whether key has been initialized.
static int key_initialized;

// Sets the context function to call to record the traceback context
// when calling a Go function from C code. Called from runtime.SetCgoTraceback.
void x_cgo_set_context_function(void (*context)(struct context_arg*)) {
	int err;

	x_cgo_context_function = context;

	// Allocate the pthread_key to use. We don't worry about races
	// here--the program isn't permitted to race on calls to
	// SetCgoTraceback and isn't likely to race on it anyhow.
	if (context != nil && !key_initialized) {
		key_initialized = 1;
		err = pthread_key_create(&key, NULL);
		if (err != 0) {
			fatalf("pthread_key_create failed: %s", strerror(errno));
		}
	}
}

// Releases the cgo traceback context.
void _cgo_release_context(uintptr_t ctxt) {
	if (ctxt != 0 && x_cgo_context_function != nil) {
		struct context_arg arg;

		arg.Context = ctxt;
		(*x_cgo_context_function)(&arg);
	}
}
