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

func sw[T any](a Ixg[T]) {
	switch t := a.(type) {
	case Iy:
		t.Y()
	default:
		println("other")
	}
	return
}

func main() {
	sw[int]((*XY)(nil))
}
