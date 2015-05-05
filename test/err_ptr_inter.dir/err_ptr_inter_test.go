// errorcheck -0 -m -l

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gfpoly provides operations that treat big-nat quantities as if they were polynomials with coefficients {0,1}

package main

import "other"

type ImportedFromSomewhere interface {
	Do()
}

type HasAMethod struct {
	x int
}

func (me *HasAMethod) Do() {
	println(me.x)
}

func InMyCode(x *ImportedFromSomewhere, y *HasAMethod, z *other.ExportedInterface) {
	x.Do() // ERROR "x\.Do undefined \(pointer type \*ImportedFromSomewhere has no field or method Do; interface ImportedFromSomewhere however has method Do\)"
	(*x).Do()
	x.Dont()    // ERROR "x\.Dont undefined \(pointer type \*ImportedFromSomewhere has no field or method Dont; interface ImportedFromSomewhere also lacks method Dont\)"
	(*x).Dont() // ERROR "\(\*x\)\.Dont undefined \(type ImportedFromSomewhere has no field or method Dont\)"

	y.Do()
	(*y).Do()
	y.Dont()    // ERROR "y\.Dont undefined \(type \*HasAMethod has no field or method Dont\)"
	(*y).Dont() // ERROR "\(\*y\)\.Dont undefined \(type HasAMethod has no field or method Dont\)"

	z.Do() // ERROR "z\.Do undefined \(pointer type \*other\.ExportedInterface has no field or method Do; interface other.ExportedInterface however has method Do\)"
	(*z).Do()
	z.Dont()      // ERROR "z\.Dont undefined \(pointer type \*other\.ExportedInterface has no field or method Dont; interface other\.ExportedInterface also lacks method Dont\)"
	(*z).Dont()   // ERROR "\(\*z\)\.Dont undefined \(type other\.ExportedInterface has no field or method Dont\)"
	z.secret()    // ERROR "z\.secret undefined \(pointer type \*other\.ExportedInterface has no field or method secret; interface other\.ExportedInterface however has method secret, but it is inaccessible\)"
	(*z).secret() // ERROR "\(\*z\)\.secret undefined \(cannot refer to unexported field or method secret\)"

}

func main() {
}
