// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "a"

type O[T any] struct {
	A int
	B a.I1[T]
	C a.I2[T]
}

func main() {
	_ = O[int]{
		C: a.I2[int]{},
	}
}
