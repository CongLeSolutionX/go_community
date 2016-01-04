// errorcheck -0 -m -l

// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test, using compiler diagnostic flags, that the escape analysis is working.
// Compiles but does not run.  Inlining is disabled.
// Registerization is disabled too (-N), which should
// have no effect on escape analysis.

package main

import "fmt"

func main() {
	// Just run test over and over again. This main func is just for
	// convenience; if test were the main func, we could also trigger
	// the panic just by running the program over and over again
	// (sometimes it takes 1 time, sometimes it takes ~4,000+).
	for iter := 0; ; iter++ {
		if iter%50 == 0 {
			fmt.Println(iter) // ERROR "iter escapes to heap$" "main ... argument does not escape$"
		}
		test(iter)
	}
}

func test(iter int) {

	const maxI = 500
	m := make(map[int][]int) // ERROR "make\(map\[int\]\[\]int\) escapes to heap$" "moved to heap: m$"

	// The panic seems to be triggered when m is modified inside a
	// closure that is both recursively called and reassigned to in a
	// loop.

	// Cause of bug -- escape of closure failed to escape (shared) data structures
	// of map.  Assign to fn declared outside of loop triggers escape of closure.
	// Heap -> stack pointer eventually causes badness when stack reallocation
	// occurs.

	var fn func()               // ERROR "moved to heap: fn$"
	for i := 0; i < maxI; i++ { // ERROR "moved to heap: i$"
		// var fn func() // this makes it work, because fn stays off heap
		j := 0        // ERROR "moved to heap: j$"
		fn = func() { // ERROR "func literal escapes to heap$"
			m[i] = append(m[i], 0) // ERROR "&i escapes to heap$" "&m escapes to heap$"
			if j < 25 {            // ERROR "&j escapes to heap$"
				j++
				fn() // ERROR "&fn escapes to heap$"
			}
		}
		fn()
	}

	if len(m) != maxI {
		panic(fmt.Sprintf("iter %d: maxI = %d, len(m) = %d", iter, maxI, len(m))) // ERROR "iter escapes to heap$" "len\(m\) escapes to heap$" "maxI escapes to heap$" "test ... argument does not escape$"
	}
}
