// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
extern void DestructorCallback(void);

static void (*destructorFn)(void);

static void callDestructorCallback() {
	GoDestructorCallback();
}

static void registerDestructor() {
	destructorFn = callDestructorCallback;
}

__attribute__((destructor))
static void destructor() {
	if (destructorFn) {
		destructorFn();
	}
}
*/
import "C"

import "fmt"

func init() {
	register("DestructorCallback", DestructorCallback)
}

//export GoDestructorCallback
func GoDestructorCallback() {
}

func DestructorCallback() {
	C.registerDestructor()
	fmt.Println("OK")
}
