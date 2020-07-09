// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sizeof defines the sizes of basic Go types.
package sizeof

// The sizes of basic Go types in bytes.
const (
	size = 4 << (^uint(0) >> 32 & 1) // 4 or 8

	Bool       = 1
	Byte       = Uint8
	Complex64  = 2 * Float32
	Complex128 = 2 * Float64
	Float32    = 4
	Float64    = 8
	Int        = size
	Int8       = 1
	Int16      = 2
	Int32      = 4
	Int64      = 8
	Rune       = Int32
	Uint       = size
	Uint8      = 1
	Uint16     = 2
	Uint32     = 4
	Uint64     = 8
	Uintptr    = size
)
