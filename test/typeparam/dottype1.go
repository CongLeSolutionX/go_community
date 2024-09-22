// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Ixg[T any] interface {
	X()
}

type Iy interface {
	Y()
}

type XY struct{}

func (xy *XY) X() {
	println("X")
}

func (xy *XY) Y() {
	println("Y")
}

func ig2i[T any](a Ixg[T]) (y Iy) {
	y = a.(Iy)
	return
}

func main() {
	y := ig2i[int]((*XY)(nil))
	y.Y()
}
