// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The windows C definitions for trace.go. That file uses //export so
// it can't put function definitions in the "C" import comment.

#include <pthread.h>
#include <assert.h>

extern void goCalledFromC(void);
extern void goCalledFromCThread(void);

__stdcall
static unsigned int cCalledFromCThread(void *p) {
	goCalledFromCThread();
	return NULL;
}

void cCalledFromGo(void) {
	goCalledFromC();

	uintptr_t thread;
  thread = _beginthreadex(NULL, 0, cCalledFromCThread, NULL, 0, NULL);
  WaitForSingleObject((HANDLE)thread, INFINITE);
  CloseHandle((HANDLE)thread);
}
