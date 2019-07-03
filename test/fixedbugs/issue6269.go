// run

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 6269: name collision on method names for function local types.

package main

type foo struct{}

func (foo) Error() string {
	return "ok"
}

type bar struct{}

func (bar) Error() string {
	return "fail"
}

func (bar) Unwrap() wrapper { return nil }

func unused() {
	type collision struct {
		bar
	}
	_ = collision{}
}

type collision struct {
	foo
}

func (collision) Unwrap() wrapper { return nil }

func main() {
	s := error(collision{})
	if str := s.Error(); str != "ok" {
		println("s.Error() ==", str)
		panic(`s.Error() != "ok"`)
	}
}
