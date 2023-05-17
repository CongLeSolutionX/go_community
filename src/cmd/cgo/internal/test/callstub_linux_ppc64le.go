// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

// extern void notoc_func(void);
// void TestPPC64Stubs(void) {
//	notoc_func();
// }
import "C"
import "testing"

func testPPC64CallStubs(t *testing.T) {
	// Verify the trampolines run on the testing machine.
	C.TestPPC64Stubs()
}
