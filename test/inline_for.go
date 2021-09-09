// errorcheck -0 -m=2

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test, using compiler diagnostic flags, that inlining is working.
// Compiles but does not run.

package foo

import "runtime"

func func_with() int { // ERROR "can inline func_with .*"
	return 10
}

func func_with_cost_88() { // ERROR "can inline only into small FORs .*"
	x := 200
	for i := 0; i < x; i++ { // ERROR "add for to stack \[\{25 25 0\}\]"
		if i%2 == 0 {
			runtime.GC()
		} else {
			i += 2
			x += 1
		}
	}
}

func func_with_fors() { // ERROR "cannot inline .*"
	for { // ERROR "add for to stack \[\{6 6 2\}\]"
		for { // ERROR "add for to stack \[\{5 6 2\} \{2 2 1\}\]"
			func_with_cost_88() // ERROR "inlining call to func_with_cost_88" "fixup inline \[\{36 6 2\} \{33 2 1\}\]" "add for to stack \[\{29 6 2\} \{26 2 1\} \{25 25 0\}\]"
		}
		for { // ERROR "add for to stack"
			func_with() // ERROR "inlining call to func_with" "fixup inline \[\{10 6 2\} \{10 2 1\}\]"
		}
	}
}
