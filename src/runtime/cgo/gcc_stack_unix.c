// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build unix && !darwin

#include <pthread.h>
#include "libcgo.h"

void
x_cgo_getstackbound(G *g)
{
	pthread_attr_t attr;
	size_t size;

	pthread_attr_init(&attr);
	pthread_attr_getstacksize(&attr, &size);

	g->stacklo = (uintptr)__builtin_frame_address(0) - size + 1024;
}
