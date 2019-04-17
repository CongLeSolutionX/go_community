// errorcheck -0 -m -l

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis for functions with variadic arguments

package foo

func debugf(format string, args ...interface{}) { // ERROR "format does not escape" "args does not escape"
	// Dummy implementation for non-debug build.
	// A non-empty implementation would be enabled with a build tag.
}

func bar() {
	value := 10
	debugf("value is %d", value) // ERROR "argument does not escape" "value does not escape"
}
