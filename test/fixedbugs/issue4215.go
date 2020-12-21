// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo() (int, int) {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return 2.3 // ERROR "not enough arguments to return\n\thave \(number\)\n\twant \(int, int\)|wrong number of return values"
=======
	return 2.3 // ERROR "not enough arguments to return\n\thave \(number\)\n\twant \(int, int\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func foo2() {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return int(2), 2 // ERROR "too many arguments to return\n\thave \(int, number\)\n\twant \(\)|no result values expected"
=======
	return int(2), 2 // ERROR "too many arguments to return\n\thave \(int, number\)\n\twant \(\)|return with value in function with no return type"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func foo3(v int) (a, b, c, d int) {
	if v >= 0 {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return 1 // ERROR "not enough arguments to return\n\thave \(number\)\n\twant \(int, int, int, int\)|wrong number of return values"
=======
		return 1 // ERROR "not enough arguments to return\n\thave \(number\)\n\twant \(int, int, int, int\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return 2, 3 // ERROR "not enough arguments to return\n\thave \(number, number\)\n\twant \(int, int, int, int\)|wrong number of return values"
=======
	return 2, 3 // ERROR "not enough arguments to return\n\thave \(number, number\)\n\twant \(int, int, int, int\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func foo4(name string) (string, int) {
	switch name {
	case "cow":
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return "moo" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(string, int\)|wrong number of return values"
=======
		return "moo" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(string, int\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	case "dog":
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return "dog", 10, true // ERROR "too many arguments to return\n\thave \(string, number, bool\)\n\twant \(string, int\)|wrong number of return values"
=======
		return "dog", 10, true // ERROR "too many arguments to return\n\thave \(string, number, bool\)\n\twant \(string, int\)|too many values in return statement"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	case "fish":
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return "" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(string, int\)|wrong number of return values"
=======
		return "" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(string, int\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	default:
		return "lizard", 10
	}
}

type S int
type T string
type U float64

func foo5() (S, T, U) {
	if false {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return "" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(S, T, U\)|wrong number of return values"
=======
		return "" // ERROR "not enough arguments to return\n\thave \(string\)\n\twant \(S, T, U\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	} else {
		ptr := new(T)
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
		return ptr // ERROR "not enough arguments to return\n\thave \(\*T\)\n\twant \(S, T, U\)|wrong number of return values"
=======
		return ptr // ERROR "not enough arguments to return\n\thave \(\*T\)\n\twant \(S, T, U\)|not enough arguments to return"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
	}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return new(S), 12.34, 1 + 0i, 'r', true // ERROR "too many arguments to return\n\thave \(\*S, number, number, number, bool\)\n\twant \(S, T, U\)|wrong number of return values"
=======
	return new(S), 12.34, 1 + 0i, 'r', true // ERROR "too many arguments to return\n\thave \(\*S, number, number, number, bool\)\n\twant \(S, T, U\)|too many values in return statement"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}

func foo6() (T, string) {
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
	return "T", true, true // ERROR "too many arguments to return\n\thave \(string, bool, bool\)\n\twant \(T, string\)|wrong number of return values"
=======
	return "T", true, true // ERROR "too many arguments to return\n\thave \(string, bool, bool\)\n\twant \(T, string\)|too many values in return statement"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
}
