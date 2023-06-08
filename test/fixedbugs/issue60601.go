// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import (
	"unsafe"
)

func f[T any]() int64 {
	return 1 << unsafe.Sizeof(*new(T))
}

func g() {
	f[[62]byte]()
	f[[63]byte]()
	f[[64]byte]()
	f[[100]byte]()
	f[[1e6]byte]()
}
