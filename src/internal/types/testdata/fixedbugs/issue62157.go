// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f[T any](...T) T { var x T; return x }

// Test case 1

func _() {
	var a chan string
	var b <-chan string
	f(a, b)
	f(b, a)
}

// Test case 2

type F[T any] func(T) bool

func g[T any](T) F[<-chan T] { return nil }

func f1[T any](T, F[T]) {}
func f2[T any](F[T], T) {}

func _() {
	var ch chan string
	f1(ch, g(""))
	f2(g(""), ch)
}

// Test case 3: named and directional types combined

func _() {
	type namedA chan int
	type namedB chan<- int

	var a chan int
	var A namedA
	var b chan<- int
	var B namedB

	// Ensure that all combinations of directional and
	// bidirectional channels with a named bidirectional
	// channel lead to the correct (unnamed) directional
	// channel.
	b = f(a, b)
	b = f(A, b)
	b = f(b, a)
	b = f(b, B)

	b = f(a, b, A)
	b = f(a, A, b)
	b = f(b, A, a)
	b = f(b, a, A)
	b = f(A, a, b)
	b = f(A, b, a)

	// verify type error
	a = f /* ERROR "cannot use f(A, b, a) (value of type chan<- int) as chan int value in assignment" */ (A, b, a)

	// Ensure that all combinations of directional and
	// bidirectional channels with a named directional
	// channel lead to the correct (named) directional
	// channel.
	B = f(a, b)
	B = f(a, B)
	B = f(b, a)
	B = f(B, a)

	B = f(a, b, B)
	B = f(a, B, b)
	B = f(b, B, a)
	B = f(b, a, B)
	B = f(B, a, b)
	B = f(B, b, a)

	// verify type error
	A = f /* ERROR "cannot use f(B, b, a) (value of type namedB) as namedA value in assignment" */ (B, b, a)
}

// Simplified test case from issue

type Matcher[T any] func(T) bool

func Produces[T any](T) Matcher[<-chan T] { return nil }

func Assert1[T any](Matcher[T], T) {}
func Assert2[T any](T, Matcher[T]) {}

func _() {
	var ch chan string
	Assert1(Produces(""), ch)
	Assert2(ch, Produces(""))
}
