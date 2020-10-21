// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo
// +build linux
// +build mips64 mips64le

#include <limits.h>
#include <pthread.h>
#include <string.h>
#include <signal.h>
#include <stdlib.h>
#include "libcgo.h"
#include "libcgo_unix.h"

#define magic (0xc476c475c47957UL)

static void *threadentry(void*);
static void (*setg_gcc)(void*);

// inittls allocates a thread-local storage slot for g.
//
// It finds the first available slot using pthread_key_create and uses
// it as the offset value for runtime.tlsg.
static void
inittls(void **tlsg, void **tlsbase)
{
        pthread_key_t k;
        int i, err;

        err = pthread_key_create(&k, nil);
        if(err != 0) {
                fprintf(stderr, "runtime/cgo: pthread_key_create failed: %d\n", err);
                abort();
        }
        //fprintf(stderr, "runtime/cgo: k = %d, tlsbase = %p\n", (int)k, tlsbase); // debug
        pthread_setspecific(k, (void*)magic);
        // The first key should be at 257.
        for (i=0; i<PTHREAD_KEYS_MAX; i++) {
                if (*(tlsbase+i) == (void*)magic) {
                        *tlsg = (void*)(i*sizeof(void *));
                        pthread_setspecific(k, 0);
                        return;
                }
        }
        fprintf(stderr, "runtime/cgo: could not find pthread key.\n");
        abort();
}

void
_cgo_sys_thread_start(ThreadStart *ts)
{
	pthread_attr_t attr;
	sigset_t ign, oset;
	pthread_t p;
	size_t size;
	int err;

	sigfillset(&ign);
	pthread_sigmask(SIG_SETMASK, &ign, &oset);

	// Not sure why the memset is necessary here,
	// but without it, we get a bogus stack size
	// out of pthread_attr_getstacksize.  C'est la Linux.
	memset(&attr, 0, sizeof attr);
	pthread_attr_init(&attr);
	size = 0;
	pthread_attr_getstacksize(&attr, &size);
	// Leave stacklo=0 and set stackhi=size; mstart will do the rest.
	ts->g->stackhi = size;
	err = _cgo_try_pthread_create(&p, &attr, threadentry, ts);

	pthread_sigmask(SIG_SETMASK, &oset, nil);

	if (err != 0) {
		fatalf("pthread_create failed: %s", strerror(err));
	}
}

extern void crosscall1(void (*fn)(void), void (*setg_gcc)(void*), void *g);
static void*
threadentry(void *v)
{
	ThreadStart ts;

	ts = *(ThreadStart*)v;
	free(v);

	crosscall1(ts.fn, setg_gcc, (void*)ts.g);
	return nil;
}

void
x_cgo_init(G *g, void (*setg)(void*), void **tlsg, void **tlsbase)
{
	pthread_attr_t attr;
	size_t size;

	setg_gcc = setg;
	pthread_attr_init(&attr);
	pthread_attr_getstacksize(&attr, &size);
	g->stacklo = (uintptr)&attr - size + 4096;
	pthread_attr_destroy(&attr);

	inittls(tlsg, tlsbase);
}

