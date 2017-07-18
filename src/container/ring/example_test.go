// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"container/ring"
	"fmt"
)

func Example() {
	ints := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	alphas := []string{"A", "B", "C", "D", "E"}

	// Create 2 new rings.
	intsring := ring.New(len(ints))
	alphasring := ring.New(len(alphas))

	// Fill intsring with numbers 1 to 10
	for _, i := range ints {
		intsring.Value = i
		intsring = intsring.Next()
	}

	// Fill alphasring with string numbers "A" to "E"
	for _, a := range alphas {
		alphasring.Value = a
		alphasring = alphasring.Next()
	}

	intsring = intsring.Move(2)     // intsring points to Value: 3
	alphasring = alphasring.Move(4) // alphasring points to Value: "E"

	// Link the 2 rings together. alphasring will get appended to intsring and the
	// new ring will point to the next position (Value: 4) of the append
	// point.
	linkedring := intsring.Link(alphasring) // linkedring points to Value: 4

	// Iterate over linkedring and print the value
	linkedring.Do(func(value interface{}) {
		fmt.Printf("%v, ", value)
	})

	// Output:
	// 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, E, A, B, C, D,
}
