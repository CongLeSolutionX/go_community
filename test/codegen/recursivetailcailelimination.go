// asmcheck

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import "runtime"

func recursiveTailCallElimination(x, y, z uint) uint {
	if x != 0 {
		// amd64:-"CALL"
		// arm64:-"CALL"
		return recursiveTailCallElimination(y, x, z)
	}
	return 42
}

func recursiveTailCallEliminationDoNotRunIfObservable(x, y, z uint) uint {
	if x != 0 {
		// amd64:"CALL"
		// arm64:"CALL"
		return recursiveTailCallElimination(y, x, z)
	}
	return observeStackTrace()
}

func observeStackTrace() uint {
	var b [1024]byte
	n := runtime.Stack(b[:], false)
	println(string(b[:n]))

	return 42
}
