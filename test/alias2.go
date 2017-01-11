// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test basic restrictions on type aliases.

package p

import (
	"reflect"
	. "reflect"
)

type T0 struct{}

// Valid type alias declarations.

type _ = int
type _ = struct{}
type _ = reflect.Value
type _ = Value
type _ = T0

type (
	A1 = int
	A2 = struct{}
	A3 = reflect.Value
	A4 = Value
	A5 = T0
)

func (T0) m1() {}
func (A5) m1() {} // TODO(gri) this should be an error
func (A5) m2() {}

var _ interface {
	m1()
	m2()
} = T0{}

var _ interface {
	m1()
	m2()
} = A5{}

func _() {
	type _ = int
	type _ = struct{}
	type _ = reflect.Value
	type _ = Value
	type _ = T0

	type (
		A1 = int
		A2 = struct{}
		A3 = reflect.Value
		A4 = Value
		A5 = T0
	)
}

// Invalid type alias declarations.

type _ = reflect.ValueOf // ERROR "reflect.ValueOf is not a type"

func (A1) m() {} // ERROR "cannot define new methods on non-local type int"

type B1 = struct{}

func (B1) m() {} // ERROR "invalid receiver type"

// TODO(gri) expand
