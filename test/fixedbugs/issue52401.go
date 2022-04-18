// errorcheck

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	const x = int(10)
	x += 1 // ERROR "cannot assign to x"
	_ = x

}
