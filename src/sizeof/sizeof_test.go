// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sizeof_test

import (
	. "sizeof"
	"testing"
	"unsafe"
)

var SizeofTests = []struct {
	name string
	exp  uintptr
	act  uintptr
}{
	{"bool", Bool, unsafe.Sizeof(*new(bool))},
	{"byte", Byte, unsafe.Sizeof(*new(byte))},
	{"complex128", Complex128, unsafe.Sizeof(*new(complex128))},
	{"complex64", Complex64, unsafe.Sizeof(*new(complex64))},
	{"float32", Float32, unsafe.Sizeof(*new(float32))},
	{"float64", Float64, unsafe.Sizeof(*new(float64))},
	{"int", Int, unsafe.Sizeof(*new(int))},
	{"int16", Int16, unsafe.Sizeof(*new(int16))},
	{"int32", Int32, unsafe.Sizeof(*new(int32))},
	{"int64", Int64, unsafe.Sizeof(*new(int64))},
	{"int8", Int8, unsafe.Sizeof(*new(int8))},
	{"rune", Rune, unsafe.Sizeof(*new(rune))},
	{"uint", Uint, unsafe.Sizeof(*new(uint))},
	{"uint16", Uint16, unsafe.Sizeof(*new(uint16))},
	{"uint32", Uint32, unsafe.Sizeof(*new(uint32))},
	{"uint64", Uint64, unsafe.Sizeof(*new(uint64))},
	{"uint8", Uint8, unsafe.Sizeof(*new(uint8))},
	{"uintptr", Uintptr, unsafe.Sizeof(*new(uintptr))},
}

func TestSizeof(t *testing.T) {
	for _, tt := range SizeofTests {
		if tt.exp != tt.act {
			t.Errorf("%s should be %d bytes, not %d", tt.name, tt.exp, tt.act)
		}
	}
}
