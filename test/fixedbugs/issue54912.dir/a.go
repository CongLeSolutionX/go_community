// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that inlining a function literal that captures both a type
// switch case variable and another local variable works correctly.

package a

var Sink any

func F(p *int, x any) {
	switch x := x.(type) {
	case int:
		Sink = func() {
			*p += x
		}
	}
}
