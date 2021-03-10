// run -gcflags=-G=3

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I interface{}

type _S[T any] struct {
	*T
}

// Testing that F is exported correctly, with its instantiated type _S[I].
func F() {
	v := _S[I]{}
	if v.T != nil {
		panic(v)
	}
}

// Testing the various combinations of method expressions.
type S1 struct{}
func (*S1) M() {}

type S2 struct{}
func (S2) M() {}

func _F1[T interface{ M() }](t T) {
	_ = T.M
}

func F2() {
        _F1(&S1{})
        _F1(S2{})
        _F1(&S2{})
}

func main() {
	F()
	F2()
}
