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
	x.Do() // ERROR "x\.Do undefined \(pointer type \*Imported has no field or method Do, but interface Imported has method Do\)"
	x.do() // ERROR "x\.do undefined \(pointer type \*Imported has no field or method do, but interface Imported has Do\)"
	(*x).Do()
	x.Dont()    // ERROR "x\.Dont undefined \(pointer type \*Imported has no field or method Dont, and interface Imported also has no method Dont\)"
	(*x).Dont() // ERROR "\(\*x\)\.Dont undefined \(type Imported has no field or method Dont\)"

	y.Do()
	y.do() // ERROR "y\.do undefined \(type \*HasAMethod has no field or method do, but does have Do\)"
	(*y).Do()
	(*y).do()   // ERROR "\(\*y\)\.do undefined \(type HasAMethod has no field or method do, but does have Do\)"
	y.Dont()    // ERROR "y\.Dont undefined \(type \*HasAMethod has no field or method Dont\)"
	(*y).Dont() // ERROR "\(\*y\)\.Dont undefined \(type HasAMethod has no field or method Dont\)"

	z.Do() // ERROR "z\.Do undefined \(pointer type \*other\.Exported has no field or method Do, but interface other\.Exported has method Do\)"
	z.do() // ERROR "z\.do undefined \(pointer type \*other\.Exported has no field or method do, but interface other.Exported has Do\)"
	(*z).Do()
	(*z).do()     // ERROR "\(\*z\)\.do undefined \(type other.Exported has no field or method do, but does have Do\)"
	z.Dont()      // ERROR "z\.Dont undefined \(pointer type \*other\.Exported has no field or method Dont, and interface other\.Exported also has no method Dont\)"
	(*z).Dont()   // ERROR "\(\*z\)\.Dont undefined \(type other\.Exported has no field or method Dont\)"
	z.secret()    // ERROR "z\.secret undefined \(pointer type \*other\.Exported has no field or method secret; interface other\.Exported has method secret, but it is unexported\)"
	(*z).secret() // ERROR "\(\*z\)\.secret undefined \(cannot refer to unexported field or method secret\)"

}

func main() {
}
