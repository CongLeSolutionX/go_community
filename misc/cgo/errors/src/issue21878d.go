// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

/*
static void return_void() { return; }
static void consume_void(void) { return; }
*/
import "C"

func F() {
	C.consume_void(C.return_void()) // ERROR HERE
}
