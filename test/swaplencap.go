// run -goexperiment swaplencap

//go:build !wasm && !arm && !386

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"unicode/utf8"
)

//go:noinline
func G(x []byte) []byte {
	if cap(x) > len(x) {
		x = x[:len(x)+1]
	}
	return x
}

//go:noinline
func PG(px *[]byte) []byte {
	x := *px
	if cap(x) > len(x) {
		x = x[:len(x)+1]
	}
	return x
}

type arena struct {
	aa []byte
}

//go:noinline
func AA(ar *arena) {
	ar.aa = ar.aa[:len(ar.aa)+1]
}

var a [10]byte = [10]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
var b [10]byte = [10]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
var c [10]byte
var h = []byte{'H', 'e', 'l', 'l', 'o'}
var w = []byte{'W', 'o', 'r', 'l', 'd'}

var s = "some string"

//go:noinline
func H(x []byte) []byte {
	y := c[:]
	copy(y, x)
	return y
}

//go:noinline
func F(x, y []byte) []byte {
	if len(x) > len(y) {
		x, y = y, x
	}
	var a []byte = c[:0]
	for i := range x {
		a = append(append(a, x[i]), y[i])
	}
	for i := len(x); i < len(y); i++ {
		a = append(a, y[i])
	}
	return a
}

//go:noinline
func FF(x, y []byte) []byte {
	if len(x) > len(y) {
		x, y = y, x
	}
	var a []byte
	for i := range x {
		a = append(append(a, x[i]), y[i])
	}
	for i := len(x); i < len(y); i++ {
		a = append(a, y[i])
	}
	return a
}

//go:noinline
func id[T any](x T) T {
	return x
}

//go:noinline
func ok(x []byte) []byte {
	if id(len(x)) > id(cap(x)) {
		os.Exit(86)
	}
	return id(x)
}

//go:noinline
func s2b(s string) []byte {
	return []byte(s)
}

//go:noinline
func slc(s []byte, l, c int) []byte {
	if len(s) != -l {
		os.Exit(91)
	}
	if cap(s) != -c {
		os.Exit(92)
	}
	return lsc(l, s, c)
}

//go:noinline
func lsc(l int, s []byte, c int) []byte {
	if len(s) != -l {
		os.Exit(93)
	}
	if cap(s) != -c {
		os.Exit(94)
	}
	_, _, _, _, _, _, r := lcs(l, c, l, c, l, c, s)
	return ok(r)
}

// Goal here is to force the parameters into memory, just in case that matters.
//
//go:noinline
func lcs(l, c, _, _, _, _ int, s []byte) (int, int, int, int, int, int, []byte) {
	if len(s) != -l {
		os.Exit(95)
	}
	if cap(s) != -c {
		os.Exit(96)
	}
	return 100, 200, 300, 400, 500, 600, append(s[0:-l], 11, 12, 13)
}

//go:noinline
func AP(p *[]byte, s []byte) {
	*p = append(*p, s...)
}

func main() {
	if len(h) != 5 {
		os.Exit(1)
	}
	if len(G(h)) != 5 {
		os.Exit(2)
	}
	cs := c[:4]
	if len(G(cs)) != 5 {
		os.Exit(3)
	}
	if len(PG(&cs)) != 5 {
		os.Exit(4)
	}

	cs = F(h, w)
	if len(cs) != 10 {
		os.Exit(5)
	}
	for i, x := range cs {
		if i&1 == 0 {
			if h[i/2] != x {
				os.Exit(6)
			}
		} else {
			if w[i/2] != x {
				os.Exit(7)
			}
		}
	}
	cs = H(h)
	if cap(cs) != 10 {
		os.Exit(8)
	}
	if len(cs) != 10 {
		os.Exit(9)
	}
	for i, x := range cs[:5] {
		if h[i] != x {
			os.Exit(10)
		}
	}
	ar := arena{aa: cs[:5]}
	AA(&ar)
	if len(ar.aa) != 6 {
		os.Exit(11)
	}
	if cap(ar.aa) != 10 {
		os.Exit(12)
	}

	bytes := s2b(s)
	if len(bytes) != id(len(s)) {
		os.Exit(13)
	}

	if cap(bytes) < id(len(s)) {
		os.Exit(14)
	}

	bytes = bytes[0 : len(bytes)-1]
	if len(bytes) != len(s)-1 {
		os.Exit(15)
	}

	{
		s := id("Hello, 世界")
		b := []byte("Hello, 世界")
		if len(b) != len(s) {
			os.Exit(16)
		}
		b2 := slc(b, -len(s), -cap(b))
		if len(b2)-3 != len(b) {
			os.Exit(17)
		}
	}

	cs = c[:0]
	AP(&cs, h)

	if len(cs) != len(h) {
		os.Exit(18)
	}

	if cap(cs) != len(c) {
		os.Exit(19)
	}

	// more daring tests that print.

	s := fmt.Sprintf("%s", string(FF([]byte("HloWrd"), []byte("el ol!"))))
	if s != "Hello World!" {
		os.Exit(99)
	}

	u()

}

//go:noinline
func u() {
	b := []byte("XYZ")

	for len(b) > 0 {
		r, size := utf8.DecodeLastRune(b)
		fmt.Printf("%c %d\n", r, size)
		b = b[:len(b)-size]
	}

}
