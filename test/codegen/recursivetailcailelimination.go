// asmcheck

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test defer.

package codegen

func recursiveTailCallElimination(x, y, z uint) uint {
	if x != 0 {
		// amd64:-"CALL"
		// arm64:-"CALL"
		return recursiveTailCallElimination(y, x, z)
	}
	return 42
}
