// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// test case 1

func f0[T any](T, T) {}

func _() {
	var a chan string
	var b <-chan string
	f0(a, b)
	f0(b, a)
}

// test case 2

type F[T any] func(T) bool

func g[T any](T) F[<-chan T] { return nil }

func f1[T any](T, F[T]) {}
func f2[T any](F[T], T) {}

func _() {
	var ch chan string
	f1(ch, g(""))
	f2(g(""), ch)
}

// (simplified) test case from issue

type Matcher[T any] func(T) bool

func Produces[T any](T) Matcher[<-chan T] { return nil }

func Assert1[T any](Matcher[T], T) {}
func Assert2[T any](T, Matcher[T]) {}

func _() {
	var ch chan string
	Assert1(Produces(""), ch)
	Assert2(ch, Produces(""))
}
