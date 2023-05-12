// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WARNING: Please avoid updating this file. If this file needs to be updated,
// then a new shape.pprof file should be generated:
//
//	$ cd $GOROOT/src/cmd/compile/internal/test/testdata/pgo/devirtualize/
//	$ go mod init example.com/pgo/devirtualize
//	$ go test -bench=. -cpuprofile ./shape.pprof

package main

type Shape interface {
	Perimeter() int
	Area() int
}

type Square struct {
	x int
	y int
}

// Perimeter calculate the square perimeter.
func (a Square) Perimeter() int {
	sumP := 0
	iter := 100
	for i := 0; i < iter; i++ {
		sumP = i + (a.x+a.y)<<2
	}
	return sumP
}

// Area calculates area.
func (a Square) Area() int {
	sumA := 0
	iter := 100
	for i := 0; i < iter; i++ {
		sumA = i + (a.x * a.y)
	}
	return sumA
}

type Circle struct {
	diameter int
	Radius   int
}

// Perimeter calculates the Circle perimeter (circumference).
func (c Circle) Perimeter() int {
	sumP := 0
	iter := 100
	for i := 0; i < iter; i++ {
		sumP = i + int(2*3*c.Radius)
	}
	return sumP
}

// Area calculates the area of the Circle.
func (c Circle) Area() int {
	sumA := 0
	iter := 100
	for i := 0; i < iter; i++ {
		sumA = i + int(3*c.Radius*c.Radius)
	}
	return sumA
}

// Slow calls both Set and Perimeter (this function is for test).
//
//go:noinline
func Slow(i1 Shape, i2 Shape, iter int) (int, int) {
	sumA := 0
	sum := 0
	for i := 0; i < iter; i++ {
		if i < iter-2 {
			sumA += i2.Area()
			sum += i1.Perimeter() + i2.Area()
		} else {
			sumA += i1.Area()
			sum += i2.Perimeter() + i1.Area()
		}
	}
	return sumA, sum
}
