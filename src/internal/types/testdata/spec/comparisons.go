// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package comparisons

type (
	B int // basic type representative
	A [10]func()
	L []byte
	S struct{ f []byte }
	P *S
	F func()
	I interface{}
	M map[string]int
	C chan int
)

var (
	b B
	a A
	l L
	s S
	p P
	f F
	i I
	m M
	c C
)

func _() {
	_ = nil == nil // ERROR operator == not defined on untyped nil
	_ = b == b
	_ = a /* ERR [10]func() cannot be compared */ == a
	_ = l /* ERR slice can only be compared to nil */ == l
	_ = s /* ERR struct containing []byte cannot be compared */ == s
	_ = p == p
	_ = f /* ERR func can only be compared to nil */ == f
	_ = i == i
	_ = m /* ERR map can only be compared to nil */ == m
	_ = c == c

	_ = b == nil /* ERR mismatched types */
	_ = a == nil /* ERR mismatched types */
	_ = l == nil
	_ = s == nil /* ERR mismatched types */
	_ = p == nil
	_ = f == nil
	_ = i == nil
	_ = m == nil
	_ = c == nil

	_ = nil /* ERR operator < not defined on untyped nil */ < nil
	_ = b < b
	_ = a /* ERR operator < not defined on array */ < a
	_ = l /* ERR operator < not defined on slice */ < l
	_ = s /* ERR operator < not defined on struct */ < s
	_ = p /* ERR operator < not defined on pointer */ < p
	_ = f /* ERR operator < not defined on func */ < f
	_ = i /* ERR operator < not defined on interface */ < i
	_ = m /* ERR operator < not defined on map */ < m
	_ = c /* ERR operator < not defined on chan */ < c
}

func _[
	B int,
	A [10]func(),
	L []byte,
	S struct{ f []byte },
	P *S,
	F func(),
	I interface{},
	J comparable,
	M map[string]int,
	C chan int,
](
	b B,
	a A,
	l L,
	s S,
	p P,
	f F,
	i I,
	j J,
	m M,
	c C,
) {
	_ = b == b
	_ = a /* ERR incomparable types in type set */ == a
	_ = l /* ERR incomparable types in type set */ == l
	_ = s /* ERR incomparable types in type set */ == s
	_ = p == p
	_ = f /* ERR incomparable types in type set */ == f
	_ = i /* ERR incomparable types in type set */ == i
	_ = j == j
	_ = m /* ERR incomparable types in type set */ == m
	_ = c == c

	_ = b == nil /* ERR mismatched types */
	_ = a == nil /* ERR mismatched types */
	_ = l == nil
	_ = s == nil /* ERR mismatched types */
	_ = p == nil
	_ = f == nil
	_ = i == nil /* ERR mismatched types */
	_ = j == nil /* ERR mismatched types */
	_ = m == nil
	_ = c == nil

	_ = b < b
	_ = a /* ERR type parameter A is not comparable with < */ < a
	_ = l /* ERR type parameter L is not comparable with < */ < l
	_ = s /* ERR type parameter S is not comparable with < */ < s
	_ = p /* ERR type parameter P is not comparable with < */ < p
	_ = f /* ERR type parameter F is not comparable with < */ < f
	_ = i /* ERR type parameter I is not comparable with < */ < i
	_ = j /* ERR type parameter J is not comparable with < */ < j
	_ = m /* ERR type parameter M is not comparable with < */ < m
	_ = c /* ERR type parameter C is not comparable with < */ < c
}
