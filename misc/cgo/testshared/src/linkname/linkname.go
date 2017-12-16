// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package linkname

import (
	"log"
	"reflect"
	_ "unsafe"

	_ "linknamed"
)

type life int64

//go:linkname life.Everything linknamed.universe.answer
func (l life) Everything() int64

//go:linkname newgalaxy linknamed.increment
func newgalaxy()

//go:linkname hitchhiker linknamed.singleton
var hitchhiker life

// Exercise the use of linknamed variables, functions and methods
// that are linked dynamically.
func Test() (int64, int64) {
	newgalaxy()
	return int64(hitchhiker), hitchhiker.Everything()
}

// Ensure that address offset to the tfn of a method is correctly
// set in the method table of a dynlinked type. reflect.Type.Method
// is the most obvious way to check this.
func TestReflect() int64 {
	m := reflect.TypeOf(hitchhiker).Method(0)
	v := m.Func.Call([]reflect.Value{reflect.ValueOf(hitchhiker)})
	if n, ok := v[0].Interface().(int64); ok {
		return n
	}
	log.Fatalf("reflected method call return value of incorrect type (%T)", v[0].Interface())
	return 0
}
