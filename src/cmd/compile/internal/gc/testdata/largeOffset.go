// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that we handle large constants correctly.
// On amd64, most instructions only accept a signed
// 32-bit constant or offset.

package main

const c = 0x1122334455

func f1(s []byte) byte {
	return s[c]
}

func f2(s []byte) *byte {
	return &s[c]
}

type T struct {
	pad [c]byte
	x   int
}

func f3(t *T) int {
	return t.x
}
func f4(t *T) *int {
	return &t.x
}

type U struct {
	pad [c]byte
	x   [8]byte
}

func f5(u *U, i int) byte {
	return u.x[i]
}
func f6(u *U, i int) *byte {
	return &u.x[i]
}

type W struct {
	pad [c]byte
	x   [8]int16
}

func f7(w *W, i int) int16 {
	return w.x[i]
}
func f8(w *W, i int) *int16 {
	return &w.x[i]
}

type X struct {
	pad [c]byte
	x   [8]int32
}

func f9(x *X, i int) int32 {
	return x.x[i]
}
func f10(x *X, i int) *int32 {
	return &x.x[i]
}

type Y struct {
	pad [c]byte
	x   [8]int64
}

func f11(y *Y, i int) int64 {
	return y.x[i]
}
func f12(y *Y, i int) *int64 {
	return &y.x[i]
}
