// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "a"

func main() {
	if a.F(4, 5, 1) != 5 {
		panic("bad")
	}
}
