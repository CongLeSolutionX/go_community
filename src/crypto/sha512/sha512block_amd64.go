// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha512

import (
	"reflect"
	"unsafe"
)

//go:noescape
func blockPtr(dig *digest, p uintptr, n int)

func blockString(dig *digest, s string) {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	blockPtr(dig, hdr.Data, hdr.Len)
}

func block(dig *digest, p []byte) {
	blockPtr(dig, uintptr(unsafe.Pointer(&p[0])), len(p))
}
