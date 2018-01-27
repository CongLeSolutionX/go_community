// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math_test

import (
	. "runtime/internal/math"
	"testing"
)

var SinkUintptr uintptr
var SinkBool bool

func BenchmarkMulUintptr(b *testing.B) {
	x, y := uintptr(1), uintptr(2)
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var overflow bool
			SinkUintptr, overflow = MulUintptr(x, y)
			if overflow {
				SinkUintptr = 0
			}
		}
	})
	x, y = MaxUintptr, MaxUintptr-1
	b.Run("large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var overflow bool
			SinkUintptr, overflow = MulUintptr(x, y)
			if overflow {
				SinkUintptr = 0
			}
		}
	})

}
