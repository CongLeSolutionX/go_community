// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(a, b, c, d ...int)       {} // ERROR "cannot use ... with non-final parameter a"
func g(a ...int, b ...int)      {} // ERROR "cannot use ... with non-final parameter a"
func h(...int, ...int, float32) {} // ERROR "cannot use ... with non-final parameter"

type a func(...float32, ...interface{}) // ERROR "cannot use ... with non-final parameter"
type b interface {
	f(...int, ...int)                // ERROR "cannot use ... with non-final parameter"
	g(a ...int, b ...int, c float32) // ERROR "cannot use ... with non-final parameter a"
	valid(...int)
}

func main() {}
