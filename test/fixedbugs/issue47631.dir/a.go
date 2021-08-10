// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

func F(x, y, z int) int { // inlineable and exported
	type T [2]int
	t := T{x, y}
	return t[z]
}
