// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include <pthread.h>
#include "libcgo.h"

void
x_cgo_getstackbound(G *g)
{
	size_t size;
	size = pthread_get_stacksize_np(pthread_self());
	g->stacklo = (uintptr)__builtin_frame_address(0) - size + 1024;
}
