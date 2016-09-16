// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort_test

import (
	"fmt"
	"sort"
)

// This example demonstrates searching a list sorted in ascending order.
func ExampleSearch() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	i := sort.Search(len(a), func(i int) bool { return a[i] >= x })
	if i < len(a) && a[i] == x {
		fmt.Printf("found %d at index %d in %v\n", x, i, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}
	// Output:
	// found 6 at index 2 in [1 3 6 10 15 21 28 36 45 55]
}

// This example demonstrates searching a list sorted in descending order.
// The approach is the same as searching a list in ascending order,
// but with the condition inverted.
func ExampleSearch_descendingOrder() {
	a := []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x := 6

	i := sort.Search(len(a), func(i int) bool { return a[i] <= x })
	if i < len(a) && a[i] == x {
		fmt.Printf("found %d at index %d in %v\n", x, i, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}
	// Output:
	// found 6 at index 7 in [55 45 36 28 21 15 10 6 3 1]
}

// Lower Bound: the position of the smallest integer greater than or equal to an element
func ExampleSearch_lowerBound() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 5

	lb := sort.Search(len(a), func(i int) bool { return a[i] >= x })
	if lb < len(a) {
		fmt.Printf("lower bound for %d at index: %d, value: %d\n", x, lb, a[lb])
	} else {
		fmt.Printf("lower bound for value %d doesn't exist\n", x)
	}
	// Output:
	// lower bound for 5 at index: 2, value: 6
}

// Lower Bound for descending order array: the first number before a number definitely less than an element
func ExampleSearch_lowerBound_descendingOrder() {
	a := []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x := 6

	lb := sort.Search(len(a), func(i int) bool { return a[i] < x }) - 1
	if lb >= 0 && lb < len(a) {
		fmt.Printf("lower bound for %d at index: %d, value: %d\n", x, lb, a[lb])
	} else {
		fmt.Printf("lower bound for value %d doesn't exist\n", x)
	}
	// Output:
	// lower bound for 6 at index: 7, value: 6
}

// Upper Bound: the position of the smallest integer greater than an element
func ExampleSearch_upperBound() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	ub := sort.Search(len(a), func(i int) bool { return a[i] > x })
	if ub < len(a) {
		fmt.Printf("upper bound for %d at index: %d, value: %d\n", x, ub, a[ub])
	} else {
		fmt.Printf("upper bound for value %d doesn't exist\n", x)
	}
	// Output:
	// upper bound for 6 at index: 3, value: 10
}

// Upper Bound for descending order array: the first number before a number less than or equal to an element
func ExampleSearch_upperBound_descendingOrder() {
	a := []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x := 6

	ub := sort.Search(len(a), func(i int) bool { return a[i] <= x }) - 1
	if ub >= 0 && ub < len(a) {
		fmt.Printf("upper bound for %d at index: %d, value: %d\n", x, ub, a[ub])
	} else {
		fmt.Printf("upper bound for value %d doesn't exist\n", 6)
	}
	// Output:
	// upper bound for 6 at index: 6, value: 10
}
