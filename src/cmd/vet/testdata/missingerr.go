// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains tests for the missing error assignment checker.

package testdata

func f() {
	err := g()
	if err != nil {
	}
	if g(); err != nil { // ERROR "possibly missing error assignment"
	}
	if g(); err == nil { // ERROR "possibly missing error assignment"
	}
	if g(); err != g() {
	}
	if g(); g() != nil {
	}
	if err = g(); err != nil {
	}
	if h(); err != nil { // ERROR "possibly missing error assignment"
	}
	if err != nil {
	}
}

func f2() {
	// Avoids warnings from the shadow checker.

	if _, err := h(); err != nil {
	}
	if err := g(); err != nil {
	}
}

func g() error {
	return nil
}

func h() (int, error) {
	return 0, nil
}
