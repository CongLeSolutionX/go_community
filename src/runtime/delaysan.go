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

// delayN delays around 1ns per spin on my 2Ghz Intel laptop
//go:noinline
//go:nosplit
func delayN(spin uint64) {
	sum := 0
	if !neverTrue {
		sum += neverAccessed
	}
	for i := uint64(0); i < spin; i++ {
		if sum&4096 != 0 {
			sum = (sum + sum) ^ sum
		}
		sum++
	}
	if neverTrue {
		neverAccessed = sum
	}
}
