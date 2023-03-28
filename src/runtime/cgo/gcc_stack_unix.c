// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !darwin

#ifndef _GNU_SOURCE // pthread_getattr_np
#define _GNU_SOURCE
#endif

#include <pthread.h>
#include "libcgo.h"

void
x_cgo_getstackbound(G *g)
{
	pthread_attr_t attr;
	void *addr;
	size_t size;

	pthread_attr_init(&attr);
	pthread_getattr_np(pthread_self(), &attr);  // GNU extension
	pthread_attr_getstack(&attr, &addr, &size); // low address
	g->stacklo = (uintptr)addr;
	g->stackhi = (uintptr)addr + size;
}
