// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "math"

func main() {
	if math.Trunc(18446744073709549568.0) != 18446744073709549568.0 {
		panic("¿")
	}
}
