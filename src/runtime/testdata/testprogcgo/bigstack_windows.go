// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
extern void goBigStack1(char*);

static void bigStack(void) {
	char x[256<<10];
	goBigStack1(x);
}
*/
import "C"

func init() {
	register("BigStack", BigStack)
}

func BigStack() {
	// Use a lot of stack and call back into Go.
	C.bigStack()
}

var alwaysFalse bool

//export goBigStack1
func goBigStack1(x *C.char) {
	println("OK")
}
