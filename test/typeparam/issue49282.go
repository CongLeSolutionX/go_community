// compile -G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f[G uint]() {
	var a, m []int
	var s struct {
		a, b, c, d, e int
	}

	func(G) {
		_ = a
		_ = m
		_ = s
		func() {
			for i := 0; i < 5; i++ {
				_ = a
				_ = m
				_, _ = s, s
			}
		}()
	}(G(1.0))

	defer func() uint {
		return 0
	}()
}

func g() {
	f[uint]()
}
