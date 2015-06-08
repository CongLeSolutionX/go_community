// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 10303. Pointers passed to C were not marked as escaping (bug in cgo).

package cgotest

/*
typedef int *intptr;

void setintstar(int *x) {
	*x = 1;
}

void setintptr(intptr x) {
	*x = 1;
}
*/
import "C"

import (
	"testing"
	"unsafe"
)

func test10303(t *testing.T, n int) {
	// Run at a few different stack depths just to avoid an unlucky pass
	// due to variables ending up on different pages.
	if n > 0 {
		test10303(t, n-1)
	}
	if t.Failed() {
		return
	}
	var x, y, z C.int
	C.setintstar(&x)
	C.setintptr(&y)
	if uintptr(unsafe.Pointer(&x))&^0xfff == uintptr(unsafe.Pointer(&z))&^0xfff {
		t.Error("C int* argument on stack")
	}
	if uintptr(unsafe.Pointer(&y))&^0xfff == uintptr(unsafe.Pointer(&z))&^0xfff {
		t.Error("C intptr argument on stack")
	}
}
