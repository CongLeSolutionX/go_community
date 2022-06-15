// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"unsafe"
)

func F[T any](v T) uintptr {
	return unsafe.Alignof(func() T {
		_ = reflect.DeepEqual(struct{ _ T }{}, nil)
		return v
	}())
}

func main() {
	F(0)
}
