// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"unsafe"

	"./b"
)

var s = []rune{0, 1, 2, 3}

func main() {
	var k = pathKey(s)
	var m = map[any]int{}
	m[k] = 1
	b.Init()
}

func pathKey(p []int32) interface{} {
	if p == nil {
		return [0]int32{}
	}
	hdr := unsafe.SliceData(p)
	array := reflect.NewAt(reflect.ArrayOf(len(p), reflect.TypeOf(int32(0))), unsafe.Pointer(hdr))
	return array.Elem().Interface()
}
