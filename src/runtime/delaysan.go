// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

var neverTrue bool
var neverAccessed int

const _SPIN = 100

//go:noinline
//go:nosplit
func delay() {
	sum := 0
	if !neverTrue {
		sum += neverAccessed
	}
	for i := 0; i < _SPIN; i++ {
		if sum&4096 != 0 {
			sum = (sum + sum) ^ sum
		}
		sum++
	}
	if neverTrue {
		neverAccessed = sum
	}
}
