// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 8507
// used to call algtype on invalid recursive type and get into infinite recursion

package p

type T struct{ T } // ERROR "(?s)invalid recursive type: T.*T refers to.*T$"

func f() {
	println(T{} == T{})
}
