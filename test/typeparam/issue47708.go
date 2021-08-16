// run -gcflags=-G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

type Foo[t any] interface {
	Foo(Bar[t])string
}
type Bar[t any] interface {
	Bar(Foo[t])string
}

type Baz[t any] t
func (l Baz[t]) Foo(v Bar[t]) string {
	if v,ok := v.(Bob[t]);ok{
		return fmt.Sprintf("%v%v",l,v)
	}
	return ""
}
type Bob[t any] t
func (l Bob[t]) Bar(v Foo[t]) string {
	if v,ok := v.(Baz[t]);ok{
		return fmt.Sprintf("%v%v",l,v)
	}
	return ""
}


func main() {
	var baz Baz[int] = 123
	var bob Bob[int] = 456

	if got, want := baz.Foo(bob), "123456"; got != want {
		panic(fmt.Sprintf("got %d want %d", got, want))
	}
	if got, want := bob.Bar(baz), "456123"; got != want {
		panic(fmt.Sprintf("got %d want %d", got, want))
	}
}