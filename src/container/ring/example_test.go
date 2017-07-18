// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"container/ring"
	"fmt"
	"strconv"
)

func Example() {
	// Create 2 new rings.
	r10 := ring.New(10)
	r5 := ring.New(5)

	// Fill r10 with numbers 1 to 10
	for i := 1; i <= r10.Len(); i++ {
		r10.Value = i
		r10 = r10.Next()
	}

	// Fill r5 with string numbers "Num: 11" to "Num: 15"
	for i := 1; i <= r5.Len(); i++ {
		r5.Value = fmt.Sprintf("Num: %s", strconv.Itoa(i+10))
		r5 = r5.Next()
	}

	r10 = r10.Move(2) // r10 points to Value: 3
	r5 = r5.Move(4)   // r5 points to Value: "Num: 15"

	// Link the 2 rings together. r5 will get appended to r10 and the
	// new ring will point to the next position (Value: 4) of the append
	// point.
	r15 := r10.Link(r5) // r15 points to Value: 4

	for i := 1; i <= r15.Len(); i++ {
		fmt.Printf("%v, ", r15.Value)
		r15 = r15.Next()
	}

	// Output:
	// 4, 5, 6, 7, 8, 9, 10, 1, 2, 3, Num: 15, Num: 11, Num: 12, Num: 13, Num: 14,
}
