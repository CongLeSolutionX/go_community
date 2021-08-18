// Copyright 2021 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package main

import (
	"a"
	"b"
	"reflect"
)

func main() {
	x := []a.T{}
	y := []b.T{}

	if reflect.ValueOf(x).Type().ConvertibleTo(reflect.ValueOf(y).Type()) {
		panic("shouldn't be convertible")
	}
}
