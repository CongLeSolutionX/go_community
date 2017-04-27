// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

/*
#define HELLO "hello"
#define WORLD "world"
#define HELLO_WORLD HELLO "\000" WORLD

#define VAR1 VAR
#define VAR var
int var = 5;
*/
import "C"
import "testing"

func test18720(t *testing.T) {
	if C.HELLO_WORLD != "hello\000world" {
		t.Errorf(`expected "hello\000world", but got %q`, C.HELLO_WORLD)
	}

	if C.VAR1 != 5 {
		t.Errorf(`expected 5, but got %v`, C.VAR1)
	}
}
