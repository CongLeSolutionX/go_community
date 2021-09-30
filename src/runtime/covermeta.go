// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

type covmetablob struct {
	p    unsafe.Pointer
	len  uint32
	hash [16]byte
}

var covmetalist []covmetablob

func addcovmeta(p unsafe.Pointer, dlen uint32, hash [16]byte) uint32 {
	slot := len(covmetalist)
	covmetalist = append(covmetalist,
		covmetablob{
			p:    p,
			len:  dlen,
			hash: hash,
		})
	return uint32(slot)
}
