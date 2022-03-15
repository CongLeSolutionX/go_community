// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var ok = false

func f() func(int, int) int {
	ok = true
	return func(int, int) int { return 0 }
}

func g() (int, int) {
	if !ok {
		panic("FAIL")
	}
	return 0, 0
}

var _ = f()(g())

func main() {
	f1()
	f2()
}

func f1() {
	ok := false

	f := func() func(int, int) {
		ok = true
		return func(int, int) {}
	}
	g := func() (int, int) {
		if !ok {
			panic("FAIL")
		}
		return 0, 0
	}

	f()(g())
}

type S struct{}

func (S) f(int, int) {}

func f2() {
	ok := false

	f := func() S {
		ok = true
		return S{}
	}
	g := func() (int, int) {
		if !ok {
			panic("FAIL")
		}
		return 0, 0
	}

	f().f(g())
}
