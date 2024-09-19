// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swissmap

package runtime

import (
	"internal/abi"
	"internal/runtime/maps"
	"unsafe"
)

func mapaccess1_fast32(t *abi.SwissMapType, m *maps.Map, key uint32) unsafe.Pointer

func mapaccess2_fast32(t *abi.SwissMapType, m *maps.Map, key uint32) (unsafe.Pointer, bool)

func mapassign_fast32(t *abi.SwissMapType, m *maps.Map, key uint32) unsafe.Pointer

func mapassign_fast32ptr(t *abi.SwissMapType, m *maps.Map, key unsafe.Pointer) unsafe.Pointer {
	throw("mapassign_fast32ptr unimplemented")
	panic("unreachable")
}

func mapdelete_fast32(t *abi.SwissMapType, m *maps.Map, key uint32)
