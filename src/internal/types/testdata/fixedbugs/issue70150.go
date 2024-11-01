// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	values := []int{10, 12, 20}
	vf(values...) /* ERROR "have (...int)" */

	vf("ab", "cd", values /* ERROR "have (string, string, ...int)" */ ...)
}

func vf(method string, values ...int) {
	_ = values
}
