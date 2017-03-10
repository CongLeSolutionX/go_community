// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Compiler rejected initialization of structs to composite literals
// in a non-static setting (e.g. in a function)
// when the struct contained a field named _.

package p

type T struct {
	_ string
}

var (
	y = T{"stare"}
	w = T{_: "look"} // ERROR "cannot refer to blank field or method"
	_ = T{"page"}
	_ = T{_: "out"} // ERROR "cannot refer to blank field or method"
)

func main() {
	var x = T{"check"}
	_ = x
	_ = T{"et"}

	var z = T{_: "verse"} // ERROR "cannot refer to blank field or method"
	_ = z
	_ = T{_: "itinerary"} // ERROR "cannot refer to blank field or method"
}
