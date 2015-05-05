// errorcheck -0 -m -l

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "./other"

type Imported interface {
	Do()
}

type HasAMethod struct {
	x int
}

func (me *HasAMethod) Do() {
	println(me.x)
}

func InMyCode(x *Imported, y *HasAMethod, z *other.Exported) {
	x.Do() // ERROR "x\.Do undefined \(pointer type \*Imported has no field or method Do; interface Imported has method Do\)"
	(*x).Do()
	x.Dont()    // ERROR "x\.Dont undefined \(pointer type \*Imported has no field or method Dont; interface Imported also lacks method Dont\)"
	(*x).Dont() // ERROR "\(\*x\)\.Dont undefined \(type Imported has no field or method Dont\)"

	y.Do()
	(*y).Do()
	y.Dont()    // ERROR "y\.Dont undefined \(type \*HasAMethod has no field or method Dont\)"
	(*y).Dont() // ERROR "\(\*y\)\.Dont undefined \(type HasAMethod has no field or method Dont\)"

	z.Do() // ERROR "z\.Do undefined \(pointer type \*other\.Exported has no field or method Do; interface other.Exported has method Do\)"
	(*z).Do()
	z.Dont()      // ERROR "z\.Dont undefined \(pointer type \*other\.Exported has no field or method Dont; interface other\.Exported also lacks method Dont\)"
	(*z).Dont()   // ERROR "\(\*z\)\.Dont undefined \(type other\.Exported has no field or method Dont\)"
	z.secret()    // ERROR "z\.secret undefined \(pointer type \*other\.Exported has no field or method secret; interface other\.Exported has method secret, but it is unexported\)"
	(*z).secret() // ERROR "\(\*z\)\.secret undefined \(cannot refer to unexported field or method secret\)"

}

func main() {
}
