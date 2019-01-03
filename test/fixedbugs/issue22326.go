// run

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

var (
	a = c + b
	b = f()
	c = f()
	d = 3
)

func f() int {
	d++
	return d
}

func expect(varname string, expected int, actual int) {
	if actual != expected {
		panic(fmt.Sprintf("%s is %v not %v", varname, actual, expected))
	}
}

func main() {
	expect("b", 4, b)
	expect("c", 5, c)
}
