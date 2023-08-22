// run

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "maps"

func main() {
	m := map[string]struct{}{}

	ss := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
		"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V",
		"hmm",
	}
	for _, s := range ss {
		m[s] = struct{}{}
	}
	_ = maps.Clone(m)

	delete(m, "hmm")

	_ = maps.Clone(m)
}
