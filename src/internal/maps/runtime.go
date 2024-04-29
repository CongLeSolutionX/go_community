// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
	"unsafe"
)

//go:linkname fastrand64 runtime.fastrand64
func fastrand64() uint64

//go:linkname typedmemmove runtime.typedmemmove
func typedmemmove(typ *abi.Type, dst, src unsafe.Pointer)

//go:linkname typedmemclr runtime.typedmemclr
func typedmemclr(typ *abi.Type, ptr unsafe.Pointer)
