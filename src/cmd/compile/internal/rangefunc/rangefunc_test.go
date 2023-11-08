// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.rangefunc

package rangefunc_test

import (
	"slices"
	"testing"
)

type Seq2[T1, T2 any] func(yield func(T1, T2) bool)

// OfSliceIndex returns a Seq over the elements of s. It is equivalent
// to range s.
func OfSliceIndex[T any, S ~[]T](s S) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
		return
	}
}

func BadOfSliceIndex[T any, S ~[]T](s S) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			yield(i, v)
		}
		return
	}
}

func VeryBadOfSliceIndex[T any, S ~[]T](s S) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			func() {
				defer func() {
					recover()
				}()
				yield(i, v)
			}()
		}
		return
	}
}

var i int

func TestCheck(t *testing.T) {
	i := 0
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
		} else {
			t.Error("Wanted to see a failure")
		}
	}()
	for _, x := range Check(BadOfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})) {
		i += x
		if i > 4*9 {
			break
		}
	}
}

func TestBreak1(t *testing.T) {
	var result []int
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3}
	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4}) {
		if x == -4 {
			break
		}
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			result = append(result, y)
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak2(t *testing.T) {
	var result []int
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3}
outer:
	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4}) {
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestContinue(t *testing.T) {
	var result []int
	var expect = []int{-1, 1, 2, -2, 1, 2, -3, 1, 2, -4}
outer:
	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4}) {
		result = append(result, x)
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				continue outer
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x-10)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak3(t *testing.T) {
	var result []int
	var expect = []int{100, 10, 2, 4, 200, 10, 2, 4, 20, 2, 4, 300, 10, 2, 4, 20, 2, 4, 30}
X:
	for _, x := range OfSliceIndex([]int{100, 200, 300, 400}) {
	Y:
		for _, y := range OfSliceIndex([]int{10, 20, 30, 40}) {
			if 10*y >= x {
				break
			}
			result = append(result, y)
			if y == 30 {
				continue X
			}
		Z:
			for _, z := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
				if z&1 == 1 {
					continue Z
				}
				result = append(result, z)
				if z >= 4 {
					continue Y
				}
			}
			result = append(result, -y) // should never be executed
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

//

func TestBreak1BadA(t *testing.T) {
	var result []int
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Error("Wanted to see a failure")
		}
	}()

	for _, x := range BadOfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		if x == -4 {
			break
		}
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			result = append(result, y)
		}
		result = append(result, x)
	}
}

func TestBreak1BadB(t *testing.T) {
	var result []int
	var expect = []int{1, 2} // inner breaks, panics, after before outer appends

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Error("Wanted to see a failure")
		}
	}()

	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		if x == -4 {
			break
		}
		for _, y := range BadOfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			result = append(result, y)
		}
		result = append(result, x)
	}
}

func TestBreak2BadA(t *testing.T) {
	var result []int
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)

		}
	}()

outer:
	for _, x := range BadOfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak2BadB(t *testing.T) {
	var result []int
	var expect = []int{1, 2} // inner breaks, panics, after before outer appends

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)
		}
	}()

outer:
	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		for _, y := range BadOfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestContinueBadA(t *testing.T) {
	var result []int
	var expect = []int{-1, 1, 2, -2, 1, 2, -3, 1, 2, -4}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)
		}
	}()

outer:
	for _, x := range BadOfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		result = append(result, x)
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				continue outer
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x-10)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestContinueBadB(t *testing.T) {
	var result []int
	var expect = []int{-1, 1, 2}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)
		}
	}()

outer:
	for _, x := range OfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		result = append(result, x)
		for _, y := range BadOfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				continue outer
			}
			if x == -4 {
				break outer
			}

			result = append(result, y)
		}
		result = append(result, x-10)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak3BadA(t *testing.T) {
	var result []int
	var expect = []int{100, 10, 2, 4, 200, 10, 2, 4, 20, 2, 4, 300, 10, 2, 4, 20, 2, 4, 30}

	// This one doesn't panic, it doesn't early exit the bad loop.

X:
	for _, x := range BadOfSliceIndex([]int{100, 200, 300, 400}) {
	Y:
		for _, y := range OfSliceIndex([]int{10, 20, 30, 40}) {
			if 10*y >= x {
				break
			}
			result = append(result, y)
			if y == 30 {
				continue X
			}
		Z:
			for _, z := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
				if z&1 == 1 {
					continue Z
				}
				result = append(result, z)
				if z >= 4 {
					continue Y
				}
			}
			result = append(result, -y) // should never be executed
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak3BadB(t *testing.T) {
	var result []int
	var expect = []int{} // fails at first execution of Y loop
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)
		}
	}()

X:
	for _, x := range OfSliceIndex([]int{100, 200, 300, 400}) {
	Y:
		for _, y := range BadOfSliceIndex([]int{10, 20, 30, 40}) {
			if 10*y >= x {
				break
			}
			result = append(result, y)
			if y == 30 {
				continue X
			}
		Z:
			for _, z := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
				if z&1 == 1 {
					continue Z
				}
				result = append(result, z)
				if z >= 4 {
					continue Y
				}
			}
			result = append(result, -y) // should never be executed
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func TestBreak3BadC(t *testing.T) {
	var result []int
	var expect = []int{100, 10, 2, 4}
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("Expected %v, got %v", expect, result)
			}
		} else {
			t.Errorf("Wanted to see a failure, result was %v", result)
		}
	}()

X:
	for _, x := range OfSliceIndex([]int{100, 200, 300, 400}) {
	Y:
		for _, y := range OfSliceIndex([]int{10, 20, 30, 40}) {
			if 10*y >= x {
				break
			}
			result = append(result, y)
			if y == 30 {
				continue X
			}
		Z:
			for _, z := range BadOfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
				if z&1 == 1 {
					continue Z
				}
				result = append(result, z)
				if z >= 4 {
					continue Y
				}
			}
			result = append(result, -y) // should never be executed
		}
		result = append(result, x)
	}
	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func veryBad(s []int) []int {
	var result []int
X:
	for _, x := range OfSliceIndex([]int{1, 2, 3}) {

		result = append(result, x)

		for _, y := range VeryBadOfSliceIndex(s) {
			result = append(result, y)
			break X
		}
		for _, z := range OfSliceIndex([]int{100, 200, 300}) {
			result = append(result, z)
			if z == 100 {
				break
			}
		}
	}
	return result
}

func okay(s []int) []int {
	var result []int
X:
	for _, x := range OfSliceIndex([]int{1, 2, 3}) {

		result = append(result, x)

		for _, y := range OfSliceIndex(s) {
			result = append(result, y)
			break X
		}
		for _, z := range OfSliceIndex([]int{100, 200, 300}) {
			result = append(result, z)
			if z == 100 {
				break
			}
		}
	}
	return result
}

// TestVeryBad1 demonstrates the behavior of an extremely poorly behaved iterator.
func TestVeryBad1(t *testing.T) {
	result := veryBad([]int{10, 20, 30, 40, 50}) // ODD length
	expect := []int{1, 10}

	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

// TestVeryBad2 demonstrates the behavior of an extremely poorly behaved iterator.
func TestVeryBad2(t *testing.T) {
	result := veryBad([]int{10, 20, 30, 40}) // EVEN length
	expect := []int{1, 10}

	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

// TestOk is the nice version of the very bad iterator.
func TestOk(t *testing.T) {
	result := okay([]int{10, 20, 30, 40, 50}) // ODD length
	expect := []int{1, 10}

	if !slices.Equal(expect, result) {
		t.Errorf("Expected %v, got %v", expect, result)
	}
}

func testBreak1BadDefer(t *testing.T) (result []int) {
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3, -30, -20, -10}

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Saw panic %v", r)
			if !slices.Equal(expect, result) {
				t.Errorf("(Inner) Expected %v, got %v", expect, result)
			}
		} else {
			t.Error("Wanted to see a failure")
		}
	}()

	for _, x := range BadOfSliceIndex([]int{-1, -2, -3, -4, -5}) {
		if x == -4 {
			break
		}
		defer func() {
			result = append(result, x*10)
		}()
		for _, y := range OfSliceIndex([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) {
			if y == 3 {
				break
			}
			result = append(result, y)
		}
		result = append(result, x)
	}
	return
}

func TestBreak1BadDefer(t *testing.T) {
	var result []int
	var expect = []int{1, 2, -1, 1, 2, -2, 1, 2, -3, -30, -20, -10}
	result = testBreak1BadDefer(t)
	if !slices.Equal(expect, result) {
		t.Errorf("(Outer) Expected %v, got %v", expect, result)
	}
}

func Check[U, V any](forall Seq2[U, V]) Seq2[U, V] {
	return func(body func(U, V) bool) {
		ret := true
		forall(func(u U, v V) bool {
			if !ret {
				panic("Iterator access after exit")
			}
			ret = body(u, v)
			return ret
		})
	}
}
