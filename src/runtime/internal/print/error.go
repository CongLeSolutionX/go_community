// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package print

import (
	_iface "runtime/internal/iface"
	"unsafe"
)

type stringer interface {
	String() string
}

func typestring(x interface{}) string {
	e := (*_iface.Eface)(unsafe.Pointer(&x))
	return *e.Type.String
}

// For calling from C.
// Prints an argument passed to panic.
// There's room for arbitrary complexity here, but we keep it
// simple and handle just a few important cases: int, string, and Stringer.
func Printany(i interface{}) {
	switch v := i.(type) {
	case nil:
		print("nil")
	case stringer:
		print(v.String())
	case error:
		print(v.Error())
	case int:
		print(v)
	case string:
		print(v)
	default:
		print("(", typestring(i), ") ", i)
	}
}
