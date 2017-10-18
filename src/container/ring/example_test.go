// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ring_test

import (
	"container/ring"
	"fmt"
)

func Example() {
	r := ring.New(5)
	head := r
	for i := 0; i < 10; i++ {
		head.Value = i
		head = head.Next()
	}

	for i := 0; i < r.Len(); i++ {
		fmt.Printf("#%d: %v\n", i, r.Value.(int))
		r = r.Next()
	}

	fmt.Println("\nIterating in reverse")
	for i := 0; i < 7; i++ {
		fmt.Printf("%d: %v\n", i, r.Value)
		r = r.Prev()
	}
	// Output:
	// #0: 5
	// #1: 6
	// #2: 7
	// #3: 8
	// #4: 9
	//
	// Iterating in reverse
	// 0: 5
	// 1: 9
	// 2: 8
	// 3: 7
	// 4: 6
	// 5: 5
	// 6: 9
}

func ExampleRing_Do() {
	r := ring.New(4)
	head := r
	for i := 0; i < 6; i++ {
		head.Value = fmt.Sprintf("foo::%d", i)
		head = head.Next()
	}
	r.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
	// Output:
	// foo::4
	// foo::5
	// foo::2
	// foo::3
}

func ExampleRing_Link() {
	r1 := ring.New(3)
	for i := 0; i < r1.Len(); i++ {
		r1.Value = i
		r1 = r1.Next()
	}
	fmt.Println("r1:")
	r1.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})

	fmt.Println("\nr2:")
	r2 := ring.New(2)
	for i := 1; i <= r2.Len(); i++ {
		r2.Value = 10 * i
		r2 = r2.Next()
	}
	r2.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})

	r := r1.Link(r2)
	fmt.Printf("\nAfter link. Element count: %d\n", r.Len())
	r.Do(func(v interface{}) {
		fmt.Printf("%v\n", v)
	})
	// Output:
	// r1:
	// 0
	// 1
	// 2
	//
	// r2:
	// 10
	// 20
	//
	// After link. Element count: 5
	// 1
	// 2
	// 0
	// 10
	// 20
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
	fmt.Printf("%d elements now in the unlinked ring\n", tri.Len())
	tri.Do(func(v interface{}) {
		fmt.Printf("Unlinked: %v\n", v)
	})
	fmt.Printf("%d elements left in the original ring\n", r.Len())
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
	// 3 elements now in the unlinked ring
	// Unlinked: 4
	// Unlinked: 3
	// Unlinked: 2
	// 2 elements left in the original ring
	// 5
	// 1
}
