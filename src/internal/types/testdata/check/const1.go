// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// constant conversions

package const1

import "math"

const(
	mi = ^int(0)
	mu = ^uint(0)
	mp = ^uintptr(0)

	logSizeofInt     = uint(mi>>8&1 + mi>>16&1 + mi>>32&1)
	logSizeofUint    = uint(mu>>8&1 + mu>>16&1 + mu>>32&1)
	logSizeofUintptr = uint(mp>>8&1 + mp>>16&1 + mp>>32&1)
)

const (
	minInt8 = -1<<(8<<iota - 1)
	minInt16
	minInt32
	minInt64
	minInt = -1<<(8<<logSizeofInt - 1)
)

const (
	maxInt8 = 1<<(8<<iota - 1) - 1
	maxInt16
	maxInt32
	maxInt64
	maxInt = 1<<(8<<logSizeofInt - 1) - 1
)

const (
	maxUint8 = 1<<(8<<iota) - 1
	maxUint16
	maxUint32
	maxUint64
	maxUint    = 1<<(8<<logSizeofUint) - 1
	maxUintptr = 1<<(8<<logSizeofUintptr) - 1
)

const (
	smallestFloat32 = 1.0 / (1<<(127 - 1 + 23))
	// TODO(gri) The compiler limits integers to 512 bit and thus
	//           we cannot compute the value (1<<(1023 - 1 + 52))
	//           without overflow. For now we match the compiler.
	//           See also issue #44057.
	// smallestFloat64 = 1.0 / (1<<(1023 - 1 + 52))
	smallestFloat64 = math.SmallestNonzeroFloat64
)

const (
	_ = assert(smallestFloat32 > 0)
	_ = assert(smallestFloat64 > 0)
)

const (
	maxFloat32 = 1<<127 * (1<<24 - 1) / (1.0<<23)
	// TODO(gri) The compiler limits integers to 512 bit and thus
	//           we cannot compute the value 1<<1023
	//           without overflow. For now we match the compiler.
	//           See also issue #44057.
	// maxFloat64 = 1<<1023 * (1<<53 - 1) / (1.0<<52)
	maxFloat64 = math.MaxFloat64
)

const (
	_ int8 = minInt8 /* ERR overflows */ - 1
	_ int8 = minInt8
	_ int8 = maxInt8
	_ int8 = maxInt8 /* ERR overflows */ + 1
	_ int8 = smallestFloat64 /* ERR truncated */

	_ = int8(minInt8 /* ERR cannot convert */ - 1)
	_ = int8(minInt8)
	_ = int8(maxInt8)
	_ = int8(maxInt8 /* ERR cannot convert */ + 1)
	_ = int8(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ int16 = minInt16 /* ERR overflows */ - 1
	_ int16 = minInt16
	_ int16 = maxInt16
	_ int16 = maxInt16 /* ERR overflows */ + 1
	_ int16 = smallestFloat64 /* ERR truncated */

	_ = int16(minInt16 /* ERR cannot convert */ - 1)
	_ = int16(minInt16)
	_ = int16(maxInt16)
	_ = int16(maxInt16 /* ERR cannot convert */ + 1)
	_ = int16(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ int32 = minInt32 /* ERR overflows */ - 1
	_ int32 = minInt32
	_ int32 = maxInt32
	_ int32 = maxInt32 /* ERR overflows */ + 1
	_ int32 = smallestFloat64 /* ERR truncated */

	_ = int32(minInt32 /* ERR cannot convert */ - 1)
	_ = int32(minInt32)
	_ = int32(maxInt32)
	_ = int32(maxInt32 /* ERR cannot convert */ + 1)
	_ = int32(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ int64 = minInt64 /* ERR overflows */ - 1
	_ int64 = minInt64
	_ int64 = maxInt64
	_ int64 = maxInt64 /* ERR overflows */ + 1
	_ int64 = smallestFloat64 /* ERR truncated */

	_ = int64(minInt64 /* ERR cannot convert */ - 1)
	_ = int64(minInt64)
	_ = int64(maxInt64)
	_ = int64(maxInt64 /* ERR cannot convert */ + 1)
	_ = int64(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ int = minInt /* ERR overflows */ - 1
	_ int = minInt
	_ int = maxInt
	_ int = maxInt /* ERR overflows */ + 1
	_ int = smallestFloat64 /* ERR truncated */

	_ = int(minInt /* ERR cannot convert */ - 1)
	_ = int(minInt)
	_ = int(maxInt)
	_ = int(maxInt /* ERR cannot convert */ + 1)
	_ = int(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uint8 = 0 /* ERR overflows */ - 1
	_ uint8 = 0
	_ uint8 = maxUint8
	_ uint8 = maxUint8 /* ERR overflows */ + 1
	_ uint8 = smallestFloat64 /* ERR truncated */

	_ = uint8(0 /* ERR cannot convert */ - 1)
	_ = uint8(0)
	_ = uint8(maxUint8)
	_ = uint8(maxUint8 /* ERR cannot convert */ + 1)
	_ = uint8(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uint16 = 0 /* ERR overflows */ - 1
	_ uint16 = 0
	_ uint16 = maxUint16
	_ uint16 = maxUint16 /* ERR overflows */ + 1
	_ uint16 = smallestFloat64 /* ERR truncated */

	_ = uint16(0 /* ERR cannot convert */ - 1)
	_ = uint16(0)
	_ = uint16(maxUint16)
	_ = uint16(maxUint16 /* ERR cannot convert */ + 1)
	_ = uint16(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uint32 = 0 /* ERR overflows */ - 1
	_ uint32 = 0
	_ uint32 = maxUint32
	_ uint32 = maxUint32 /* ERR overflows */ + 1
	_ uint32 = smallestFloat64 /* ERR truncated */

	_ = uint32(0 /* ERR cannot convert */ - 1)
	_ = uint32(0)
	_ = uint32(maxUint32)
	_ = uint32(maxUint32 /* ERR cannot convert */ + 1)
	_ = uint32(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uint64 = 0 /* ERR overflows */ - 1
	_ uint64 = 0
	_ uint64 = maxUint64
	_ uint64 = maxUint64 /* ERR overflows */ + 1
	_ uint64 = smallestFloat64 /* ERR truncated */

	_ = uint64(0 /* ERR cannot convert */ - 1)
	_ = uint64(0)
	_ = uint64(maxUint64)
	_ = uint64(maxUint64 /* ERR cannot convert */ + 1)
	_ = uint64(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uint = 0 /* ERR overflows */ - 1
	_ uint = 0
	_ uint = maxUint
	_ uint = maxUint /* ERR overflows */ + 1
	_ uint = smallestFloat64 /* ERR truncated */

	_ = uint(0 /* ERR cannot convert */ - 1)
	_ = uint(0)
	_ = uint(maxUint)
	_ = uint(maxUint /* ERR cannot convert */ + 1)
	_ = uint(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ uintptr = 0 /* ERR overflows */ - 1
	_ uintptr = 0
	_ uintptr = maxUintptr
	_ uintptr = maxUintptr /* ERR overflows */ + 1
	_ uintptr = smallestFloat64 /* ERR truncated */

	_ = uintptr(0 /* ERR cannot convert */ - 1)
	_ = uintptr(0)
	_ = uintptr(maxUintptr)
	_ = uintptr(maxUintptr /* ERR cannot convert */ + 1)
	_ = uintptr(smallestFloat64 /* ERR cannot convert */)
)

const (
	_ float32 = minInt64
	_ float64 = minInt64
	_ complex64 = minInt64
	_ complex128 = minInt64

	_ = float32(minInt64)
	_ = float64(minInt64)
	_ = complex64(minInt64)
	_ = complex128(minInt64)
)

const (
	_ float32 = maxUint64
	_ float64 = maxUint64
	_ complex64 = maxUint64
	_ complex128 = maxUint64

	_ = float32(maxUint64)
	_ = float64(maxUint64)
	_ = complex64(maxUint64)
	_ = complex128(maxUint64)
)

// TODO(gri) find smaller deltas below

const delta32 = maxFloat32/(1 << 23)

const (
	_ float32 = - /* ERR overflow */ (maxFloat32 + delta32)
	_ float32 = -maxFloat32
	_ float32 = maxFloat32
	_ float32 = maxFloat32 /* ERR overflow */ + delta32

	_ = float32(- /* ERR cannot convert */ (maxFloat32 + delta32))
	_ = float32(-maxFloat32)
	_ = float32(maxFloat32)
	_ = float32(maxFloat32 /* ERR cannot convert */ + delta32)

	_ = assert(float32(smallestFloat32) == smallestFloat32)
	_ = assert(float32(smallestFloat32/2) == 0)
	_ = assert(float32(smallestFloat64) == 0)
	_ = assert(float32(smallestFloat64/2) == 0)
)

const delta64 = maxFloat64/(1 << 52)

const (
	_ float64 = - /* ERR overflow */ (maxFloat64 + delta64)
	_ float64 = -maxFloat64
	_ float64 = maxFloat64
	_ float64 = maxFloat64 /* ERR overflow */ + delta64

	_ = float64(- /* ERR cannot convert */ (maxFloat64 + delta64))
	_ = float64(-maxFloat64)
	_ = float64(maxFloat64)
	_ = float64(maxFloat64 /* ERR cannot convert */ + delta64)

	_ = assert(float64(smallestFloat32) == smallestFloat32)
	_ = assert(float64(smallestFloat32/2) == smallestFloat32/2)
	_ = assert(float64(smallestFloat64) == smallestFloat64)
	_ = assert(float64(smallestFloat64/2) == 0)
)

const (
	_ complex64 = - /* ERR overflow */ (maxFloat32 + delta32)
	_ complex64 = -maxFloat32
	_ complex64 = maxFloat32
	_ complex64 = maxFloat32 /* ERR overflow */ + delta32

	_ = complex64(- /* ERR cannot convert */ (maxFloat32 + delta32))
	_ = complex64(-maxFloat32)
	_ = complex64(maxFloat32)
	_ = complex64(maxFloat32 /* ERR cannot convert */ + delta32)
)

const (
	_ complex128 = - /* ERR overflow */ (maxFloat64 + delta64)
	_ complex128 = -maxFloat64
	_ complex128 = maxFloat64
	_ complex128 = maxFloat64 /* ERR overflow */ + delta64

	_ = complex128(- /* ERR cannot convert */ (maxFloat64 + delta64))
	_ = complex128(-maxFloat64)
	_ = complex128(maxFloat64)
	_ = complex128(maxFloat64 /* ERR cannot convert */ + delta64)
)

// Initialization of typed constant and conversion are the same:
const (
	f32 = 1 + smallestFloat32
	x32 float32 = f32
	y32 = float32(f32)
	_ = assert(x32 - y32 == 0)
)

const (
	f64 = 1 + smallestFloat64
	x64 float64 = f64
	y64 = float64(f64)
	_ = assert(x64 - y64 == 0)
)

const (
	_ = int8(-1) << 7
	_ = int8 /* ERR overflows */ (-1) << 8

	_ = uint32(1) << 31
	_ = uint32 /* ERR overflows */ (1) << 32
)
