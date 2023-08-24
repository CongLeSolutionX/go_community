// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This test uses reflect.(*rtype).Method() which depends
// on reflect.StructOf() and reflect.FuncOf(). Those must
// not be flagged as ReflectMethods.

package main

import "reflect"

type S struct{}

func (s *S) F() {}

func main() {
	var s S
	m := reflect.TypeOf(s).Method(0)
	println(m.Name)
}
