// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "./a"

func main() {
	// Test that inlined type switches without short variable
	// declarations work correctly.
	check(0, a.F(nil))
	check(1, a.F(0))
	check(2, a.F(0.0))
	check(3, a.F(""))

	// Test that inlined type switches with shart variable
	// declarations work correctly.
	_ = a.G(nil).(*interface{})
	_ = a.G(1).(*int)
	_ = a.G(2.0).(*float64)
	_ = (*a.G("").(*interface{})).(string)
	_ = (*a.G(([]byte)(nil)).(*interface{})).([]byte)
	_ = (*a.G(true).(*interface{})).(bool)
}

//go:noinline
func check(want, got int) {
	if want != got {
		println("want", want, "but got", got)
	}
}
