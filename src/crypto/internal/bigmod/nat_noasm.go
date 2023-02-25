// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build purego || !(386 || amd64 || arm || arm64 || ppc64 || ppc64le || s390x)

package bigmod

import "unsafe"

func addMulVVW1024(z, x *uint, y uint) (c uint) {
	return addMulVVW(unsafe.Slice(z, 16), unsafe.Slice(x, 16), y)
}

func addMulVVW1536(z, x *uint, y uint) (c uint) {
	return addMulVVW(unsafe.Slice(z, 24), unsafe.Slice(x, 24), y)
}

func addMulVVW2048(z, x *uint, y uint) (c uint) {
	return addMulVVW(unsafe.Slice(z, 32), unsafe.Slice(x, 32), y)
}
