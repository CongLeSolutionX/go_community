// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ensure that typed non-integer len and cap make arguments do not cause a compiler panic.

package main

var sink []byte

type (
	TINT8    int8
	TUINT8   uint8
	TINT16   int16
	TUINT16  uint16
	TINT32   int32
	TUINT32  uint32
	TINT64   int64
	TUINT64  uint64
	TINT     int
	TUINT    uint
	TUINTPTR uintptr

	TCOMPLEX64  complex64
	TCOMPLEX128 complex128

	TFLOAT32 float32
	TFLOAT64 float64

	TRUNE rune
)

func main() {
	// len
	sink = make([]byte, TINT8(1))
	sink = make([]byte, TUINT8(1))
	sink = make([]byte, TINT16(1))
	sink = make([]byte, TUINT16(1))
	sink = make([]byte, TINT32(1))
	sink = make([]byte, TUINT32(1))
	sink = make([]byte, TINT64(1))
	sink = make([]byte, TUINT64(1))
	sink = make([]byte, TINT(1))
	sink = make([]byte, TUINT(1))
	sink = make([]byte, TUINTPTR(1))

	sink = make([]byte, TCOMPLEX64(1+0i))
	sink = make([]byte, TCOMPLEX128(1+0i))

	sink = make([]byte, TFLOAT32(1.0))
	sink = make([]byte, TFLOAT64(1.0))

	sink = make([]byte, TRUNE(1))

	// cap
	sink = make([]byte, 0, TINT8(1))
	sink = make([]byte, 0, TUINT8(1))
	sink = make([]byte, 0, TINT16(1))
	sink = make([]byte, 0, TUINT16(1))
	sink = make([]byte, 0, TINT32(1))
	sink = make([]byte, 0, TUINT32(1))
	sink = make([]byte, 0, TINT64(1))
	sink = make([]byte, 0, TUINT64(1))
	sink = make([]byte, 0, TINT(1))
	sink = make([]byte, 0, TUINT(1))
	sink = make([]byte, 0, TUINTPTR(1))

	sink = make([]byte, 0, TCOMPLEX64(1+0i))
	sink = make([]byte, 0, TCOMPLEX128(1+0i))

	sink = make([]byte, 0, TFLOAT32(1.0))
	sink = make([]byte, 0, TFLOAT64(1.0))

	sink = make([]byte, 0, TRUNE(1))
}
