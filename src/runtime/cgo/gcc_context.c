// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

#include <pthread.h>
#include "libcgo.h"

static pthread_mutex_t cgo_context_mu = PTHREAD_MUTEX_INITIALIZER;

// The context function, used when tracing back C calls into Go.
static void (*cgo_context_function)(struct context_arg*);

// Sets the context function to call to record the traceback context
// when calling a Go function from C code. Called from runtime.SetCgoTraceback.
void x_cgo_set_context_function(void (*context)(struct context_arg*)) {
	pthread_mutex_lock(&cgo_context_mu);
	cgo_context_function = context;
	pthread_mutex_unlock(&cgo_context_mu);
}

// Gets the context function.
void (*(_cgo_get_context_function(void)))(struct context_arg*) {
	void (*ret)(struct context_arg*);

	pthread_mutex_lock(&cgo_context_mu);
	ret = cgo_context_function;
	pthread_mutex_unlock(&cgo_context_mu);
	return ret;
}

// Releases the cgo traceback context.
void _cgo_release_context(uintptr_t ctxt) {
	void (*pfn)(struct context_arg*);

	pfn = _cgo_get_context_function();
	if (ctxt != 0 && pfn != nil) {
		struct context_arg arg;

		arg.Context = ctxt;
		(*pfn)(&arg);
	}
}
