// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !go1.21

package sort

func ints_impl(x []int)         { Sort(IntSlice(x)) }
func float64s_impl(x []float64) { Sort(Float64Slice(x)) }
func strings_impl(x []string)   { Sort(StringSlice(x)) }

func intsAreSorted_impl(x []int) bool         { return IsSorted(IntSlice(x)) }
func float64sAreSorted_impl(x []float64) bool { return IsSorted(Float64Slice(x)) }
func stringsAreSorted_impl(x []string) bool   { return IsSorted(StringSlice(x)) }
