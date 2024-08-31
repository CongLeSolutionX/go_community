// asmcheck

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

func f1(i uint) uint {
	// This is not a complete fix, but is the best I could do naively.
	p1 := &i
	p2 := &p1
	// amd64:-"MOVQ\s\\("
	return **p2
}
