// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 22904: Make sure the compiler emits a proper error message about
// invalid recursive types rather than crashing.

package p

type a struct{ b }
type b struct{ a } // ERROR "(?s)invalid recursive type: b.\tb refers to.\ta refers to.\tb$"

var x interface{}

func f() {
	x = a{}
}
