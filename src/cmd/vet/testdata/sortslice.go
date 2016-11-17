// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testdata

import "sort"

func goodSliceSort() {
	a := []int{1, 2, 3, 4, 5}
	sort.Slice(a, func(i, j int) { return a[i] < a[j] })
}

func badSliceSort() {
	a := map[string]int{}
	sort.Slice(a, func(i, j int) { return a[i] < a[j] })
}
