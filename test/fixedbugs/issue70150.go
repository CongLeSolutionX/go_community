// errorcheck

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "context"

func main() {
	values := []int{10, 12, 20}
	vari(context.Background(), values...)             // ERROR "have \(context\.Context, \.\.\.int\)"
	vari(context.Background(), "ab", "cd", values...) // ERROR "have \(context\.Context, string, string, \.\.\.int\)"
}

func vari(ctx context.Context, method string, values ...int) {
	_ = values
}
