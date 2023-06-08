// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import (
	"unsafe"
)

func f[T int64](a T) T {
	return 1 << ((unsafe.Sizeof(a) * 8) - 1)
}

func g() {
	f(int64(1))
}
