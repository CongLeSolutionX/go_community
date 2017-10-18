// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"container/ring"
	"fmt"
)

func Example() {
	head := ring.New(5)
	for i := 0; i < 10; i++ {
		// Set the value of each element in the ring and move on to the next one.
		head.Value = i
		head = head.Next()
	}
	// Iterate and apply the function on each element.
	head.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})

	fmt.Println("\nIterating in reverse and circularly")
	for i := 0; i < 7; i++ {
		fmt.Printf("%v\n", head.Value)
		head = head.Prev()
	}
	// Output:
	// 5
	// 6
	// 7
	// 8
	// 9
	//
	// Iterating in reverse and circularly
	// 5
	// 9
	// 8
	// 7
	// 6
	// 5
	// 9
}

func ExampleRing_Do() {
	r := ring.New(4)
	for i := 0; i < 6; i++ {
		// Set the value of each element in the ring and move on to the next one
		r.Value = fmt.Sprintf("gopher::%d", i)
		r = r.Next()
	}
	// Iterate and apply the function on each element.
	r.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
	// Output:
	// gopher::2
	// gopher::3
	// gopher::4
	// gopher::5
}

func ExampleRing_Link() {
	r1 := ring.New(3)
	for i := 0; i < r1.Len(); i++ {
		r1.Value = 'a' + i
		r1 = r1.Next()
	}
	fmt.Println("r1:")
	r1.Do(func(v interface{}) {
		fmt.Printf("%c\n", v)
	})

	fmt.Println("\nr2:")
	r2 := ring.New(2)
	for i := 0; i < r2.Len(); i++ {
		r2.Value = 'A' + i
		r2 = r2.Next()
	}
	r2.Do(func(v interface{}) {
		fmt.Printf("%c\n", v)
	})

	r := r1.Link(r2)
	fmt.Printf("\nAfter link. Element count: %d\n", r.Len())
	r.Do(func(v interface{}) {
		fmt.Printf("%c\n", v)
	})
	// Output:
	// r1:
	// a
	// b
	// c
	//
	// r2:
	// A
	// B
	//
	// After link. Element count: 5
	// b
	// c
	// a
	// A
	// B
}

func ExampleRing_Unlink() {
	r := ring.New(5)
	head := r
	for i := 0; i < 6; i++ {
		head.Value = i
		head = head.Prev()
	}
	fmt.Println("Before Unlink")
	r.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
	tri := r.Unlink(3)
	fmt.Printf("\n%d elements now in the unlinked ring\n", tri.Len())
	tri.Do(func(v interface{}) {
		fmt.Printf("Unlinked: %v\n", v)
	})
	fmt.Printf("\n%d elements left in the original ring\n", r.Len())
	r.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
	// Output:
	// Before Unlink
	// 5
	// 4
	// 3
	// 2
	// 1
	//
	// 3 elements now in the unlinked ring
	// Unlinked: 4
	// Unlinked: 3
	// Unlinked: 2
	//
	// 2 elements left in the original ring
	// 5
	// 1
}
