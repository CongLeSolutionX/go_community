// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 8385: provide a more descriptive error when a method expression
// is called without a receiver.

package main

type Fooer interface {
	Foo(i, j int)
}

func f(x int) {
}

type I interface {
	M(int)
}
type T struct{}

func (t T) M(x int) {
}

func g() func(int)

func main() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	Fooer.Foo(5, 6) // ERROR "not enough arguments in call to method expression Fooer.Foo|not enough arguments in call"
=======
	Fooer.Foo(5, 6) // ERROR "not enough arguments in call to method expression Fooer.Foo|incompatible type|not enough arguments"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

	var i I
	var t *T

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	g()()    // ERROR "not enough arguments in call to g\(\)"
	f()      // ERROR "not enough arguments in call to f"
	i.M()    // ERROR "not enough arguments in call to i\.M"
	I.M()    // ERROR "not enough arguments in call to method expression I\.M|not enough arguments in call"
	t.M()    // ERROR "not enough arguments in call to t\.M"
	T.M()    // ERROR "not enough arguments in call to method expression T\.M|not enough arguments in call"
	(*T).M() // ERROR "not enough arguments in call to method expression \(\*T\)\.M|not enough arguments in call"
=======
	g()()    // ERROR "not enough arguments in call to g\(\)|not enough arguments"
	f()      // ERROR "not enough arguments in call to f|not enough arguments"
	i.M()    // ERROR "not enough arguments in call to i\.M|not enough arguments"
	I.M()    // ERROR "not enough arguments in call to method expression I\.M|not enough arguments"
	t.M()    // ERROR "not enough arguments in call to t\.M|not enough arguments"
	T.M()    // ERROR "not enough arguments in call to method expression T\.M|not enough arguments"
	(*T).M() // ERROR "not enough arguments in call to method expression \(\*T\)\.M|not enough arguments"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
