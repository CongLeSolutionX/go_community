// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import (
	"unsafe"
)

func shift[T any]() int64 {
	return 1 << unsafe.Sizeof(*new(T))
}

func div[T any]() uintptr {
	return 1 / unsafe.Sizeof(*new(T))
}

func add[T any]() int64 {
	return 1<<63 - 1 + int64(unsafe.Sizeof(*new(T)))
}

func f() {
	shift[[62]byte]()
	shift[[63]byte]()
	shift[[64]byte]()
	shift[[100]byte]()
	shift[[1e6]byte]()

	div[[0]byte]()
	add[[1]byte]()
}
