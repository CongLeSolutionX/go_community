// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort_test

import (
	"fmt"
	"sort"
)

func ExampleSearch_lowerBound() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 5

	// Lower Bound: the position of the smallest integer greater than or equal to an element
	lb := sort.Search(len(a), func(i int) bool {
		return a[i] >= x
	})

	if lb < len(a) {
		fmt.Printf("lower bound for %d at index: %d, value: %d\n", x, lb, a[lb])
	} else {
		fmt.Printf("lower bound for value %d doesn't exist\n", x)
	}

	// for int slices SearchInts by default calculates the lower bound
	lb = sort.SearchInts(a, x)

	if lb < len(a) {
		fmt.Printf("lower bound for %d at index: %d, value: %d\n", x, lb, a[lb])
	} else {
		fmt.Printf("lower bound for value %d doesn't exist\n", x)
	}

	// if the array is sorted in descending order
	// Search can still be used to find lower bound

	a = []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x = 6

	// Lower Bound for descending order array: the first number before a number definitely less than an element
	lb = sort.Search(len(a), func(i int) bool {
		return a[i] < x
	}) - 1

	if lb >= 0 {
		fmt.Printf("lower bound for %d at index: %d, value: %d\n", x, lb, a[lb])
	} else {
		fmt.Printf("lower bound for value %d doesn't exist\n", x)
	}
	// Output:
	// lower bound for value 5 at index: 2, value: 6
	// lower bound for value 5 at index: 2, value: 6
	// lower bound for value 6 at index: 7, value: 6
}

func ExampleSearch_upperBound() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	// Upper Bound: the position of the smallest integer greater than an element
	ub := sort.Search(len(a), func(i int) bool {
		return a[i] > x
	})

	if ub < len(a) {
		fmt.Printf("upper bound for value %d at index: %d, value: %d\n", x, ub, a[ub])
	} else {
		fmt.Printf("upper bound for value %d doesn't exist\n", x)
	}

	// if the array is sorted in descending order
	// Search can still be used to find upper bound

	a = []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}

	// Upper Bound for descending order array: the first number before a number less than or equal to an element
	ub = sort.Search(len(a), func(i int) bool {
		return a[i] <= x
	}) - 1

	if ub >= 0 {
		fmt.Printf("upper bound for value %d at index: %d, value: %d\n", x, ub, a[ub])
	} else {
		fmt.Printf("upper bound for value %d doesn't exist\n", 6)
	}
	// Output:
	// upper bound for value 6 at index: 3, value: 10
	// upper bound for value 6 at index: 6, value: 10
}

func ExampleSearch() {
	a := []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	x := 6

	// Binary Search
	// this is different compared to other languages,
	// Search returns the index of the lower bound,
	// so an explicit check for equality is needed
	a = []int{1, 3, 6, 10, 15, 21, 28, 36, 45, 55}
	if in := sort.Search(len(a), func(i int) bool { return a[i] >= x }); in < len(a) && a[in] == x {
		fmt.Printf("found %d at index %d in %v\n", x, in, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}

	x = 5
	// SearchInts is a shorthand for Search for int slices
	if in := sort.SearchInts(a, x); in < len(a) && a[in] == x {
		fmt.Printf("found %d at index %d in %v\n", x, in, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}

	// Search can also be used with arrays sorted in descending order

	a = []int{55, 45, 36, 28, 21, 15, 10, 6, 3, 1}
	x = 6
	if in := sort.Search(len(a), func(i int) bool { return a[i] < x }) - 1; in >= 0 && a[in] == x {
		fmt.Printf("found %d at index %d in %v\n", x, in, a)
	} else {
		fmt.Printf("%d not found in %v\n", x, a)
	}

	// Output:
	// found 6 at index 2 in [1 3 6 10 15 21 28 36 45 55]
	// 5 not found in [1 3 6 10 15 21 28 36 45 55]
	// found 6 at index 7 in [55 45 36 28 21 15 10 6 3 1]
}
