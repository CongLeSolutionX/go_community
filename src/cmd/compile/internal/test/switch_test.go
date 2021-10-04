// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import "testing"

func BenchmarkSwitch(b *testing.B) {
	x := 0
	for i := 0; i < b.N; i++ {
		switch x {
		case 0:
			x = 2
		case 2:
			x = 7
		case 7:
			x = 3
		case 3:
			x = 9
		case 9:
			x = 1
		case 1:
			x = 4
		case 4:
			x = 0
		}
	}
	sink = x
}
