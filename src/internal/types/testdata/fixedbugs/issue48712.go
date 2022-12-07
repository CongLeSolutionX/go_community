// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _[P comparable](x, y P) {
	_ = x == x
	_ = x == y
	_ = y == x
	_ = y == y

	_ = x /* ERR type parameter P is not comparable with < */ < y
}

func _[P comparable](x P, y any) {
	_ = x == x
	_ = x == y
	_ = y == x
	_ = y == y

	_ = x /* ERR type parameter P is not comparable with < */ < y
}

func _[P any](x, y P) {
	_ = x /* ERR incomparable types in type set */ == x
	_ = x /* ERR incomparable types in type set */ == y
	_ = y /* ERR incomparable types in type set */ == x
	_ = y /* ERR incomparable types in type set */ == y

	_ = x /* ERR type parameter P is not comparable with < */ < y
}

func _[P any](x P, y any) {
	_ = x /* ERR incomparable types in type set */ == x
	_ = x /* ERR incomparable types in type set */ == y
	_ = y == x // ERR incomparable types in type set
	_ = y == y

	_ = x /* ERR type parameter P is not comparable with < */ < y
}
