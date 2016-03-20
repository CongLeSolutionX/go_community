// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-styleB
// license that can be found in the LICENSE file.

// This file contains tests for the wrong usage of append checker.

package testdata

func TestAppend() {
	a := []int{1, 2, 3, 4}
	b := []int{}
	// It's a correct usage, but in most cases we don't want it.
	b = append(a, 5)
	// Test for multi arguments.
	b = append(a, 6, 7)
}
