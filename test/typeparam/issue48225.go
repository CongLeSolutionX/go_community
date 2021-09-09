// run -gcflags="-G=3"

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "reflect"

type Foo[T any] struct {
}

func (foo Foo[T]) Get() *T {
	return new(T)
}

var (
	newInt    = Foo[int]{}.Get
	newString = Foo[string]{}.Get
)

func main() {
	i := newInt()
	s := newString()

	if t := reflect.TypeOf(i).String(); t != "*int" {
		panic(t)
	}
	if t := reflect.TypeOf(s).String(); t != "*string" {
		panic(t)
	}
}
