// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <stdio.h>
#include <stdlib.h>
#include "libcgo.h"

void
x_cgo_sys_thread_create(void* (*func)(void*), void* arg) {
	fprintf(stderr, "x_cgo_sys_thread_create not implemented");
	abort();
}

uintptr_t
_cgo_wait_runtime_init_done() {
	void (*pfn)(struct context_arg*);

	// TODO(spetrovic): implement this method.
	pfn = _cgo_get_context_function();
	if (pfn != nil) {
		struct context_arg arg;

		arg.Context = 0;
		(*pfn)(&arg);
		return arg.Context;
	}
	return 0;
}

void
x_cgo_notify_runtime_init_done(void* dummy) {
	// TODO(spetrovic): implement this method.
}
