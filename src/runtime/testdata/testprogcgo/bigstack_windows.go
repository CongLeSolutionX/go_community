// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <windows.h>
#include <stdio.h>

extern void goBigStack1(char*);

// Allocate a stack that's much larger than the default.
static const int STACK_SIZE = 16<<20;

static void useStack(int bytes) {
	// Windows doesn't like huge frames, so we grow the stack 64k at a time.
	char x[64<<10];
	if (bytes < sizeof x) {
		goBigStack1(x);
	} else {
		useStack(bytes - sizeof x);
	}
}

static DWORD WINAPI threadEntry(LPVOID lpParam) {
	useStack(STACK_SIZE - (128<<10));
	return 0;
}

static void bigStack(void) {
	HANDLE hThread = CreateThread(NULL, STACK_SIZE, threadEntry, NULL, STACK_SIZE_PARAM_IS_A_RESERVATION, NULL);
	if (hThread == NULL) {
		fprintf(stderr, "CreateThread failed\n");
		exit(1);
	}
	WaitForSingleObject(hThread, INFINITE);
}
*/
import "C"

func init() {
	register("BigStack", BigStack)
}

func BigStack() {
	// Create a large thread stack and call back into Go to test
	// if Go correctly determines the stack bounds.
	C.bigStack()
}

//export goBigStack1
func goBigStack1(x *C.char) {
	println("OK")
}
