// run -gcflags=-G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "reflect"

type I interface { foo() }

type myint int

func (myint) foo() {}

type myfloat float64
func (myfloat) foo() {}

type myint32 int32
func (myint32) foo() {}

func f[T I](i I) {
	switch x := i.(type) {
	case T:
		println("T", x)
	case myint:
		println("myint", x)
	default:
		println("other", reflect.ValueOf(x).Int())
	}
}
func main() {
	f[myfloat](myint(6))
	f[myfloat](myfloat(7))
	f[myfloat](myint32(8))
}
