// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type (
	F   func() bool
	Fi  func(int) bool
	Fii func(int, int) bool
)

func _[P F](f func(P)) {
	for range f {
	}
	for k /* ERROR "range over f (variable of type func(P)) permits no iteration variables" */ := range f {
		_ = k
	}
}

func _[P F | Fi](f func(P)) {
	for range f {
	}
	for k /* ERROR "range over f (variable of type func(P)) permits no iteration variables" */ := range f {
		_ = k
	}
}

func _[P Fi | Fii](f func(P)) {
	for range f {
	}
	for k := range f {
		_ = k
	}
	for k, v /* ERROR "range over f (variable of type func(P)) permits only one iteration variable" */ := range f {
		_, _ = k, v
	}
}

func _[P Fi | func(string) bool](f func(P)) {
	for range f {
	}
	// keys exist but they have different types (int, string)
	for k /* ERROR "range over f (variable of type func(P)) permits no iteration variables" */ := range f {
		_ = k
	}
}

func _[P Fii | func(int, string) bool](f func(P)) {
	for range f {
	}
	for k := range f {
		_ = k
	}
	for k, v /* ERROR "range over f (variable of type func(P)) permits only one iteration variable" */ := range f {
		_, _ = k, v
	}
}
