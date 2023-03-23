// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import "unsafe"

type E [1 << 31 - 1]int
var a [1 << 31]E
var _ = unsafe.Sizeof(a /* ERROR "too large" */ )

var s struct {
	_ [1 << 31]E
	x int
}
var _ = unsafe.Offsetof(s /* ERROR "too large" */ .x)