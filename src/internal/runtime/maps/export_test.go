// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
	sabi "internal/runtime/maps/internal/abi"
	"unsafe"
)

const DebugLog = debugLog

var AlignUpPow2 = alignUpPow2

type instantiatedGroup[K comparable, V any] struct {
	ctrls ctrlGroup
	slots [sabi.SwissMapGroupSlots]struct {
		key  K
		elem V
	}
}

func NewTestTable[K comparable, V any](length uint64) *table {
	var m map[K]V
	mTyp := abi.TypeOf(m)
	omt := (*abi.OldMapType)(unsafe.Pointer(mTyp))

	var grp instantiatedGroup[K, V]

	mt := &sabi.SwissMapType{
		Key:    omt.Key,
		Elem:   omt.Elem,
		Group:  abi.TypeOf(grp),
		Hasher: omt.Hasher,
	}
	if omt.NeedKeyUpdate() {
		mt.Flags |= 8
	}
	if omt.HashMightPanic() {
		mt.Flags |= 16
	}
	return newTable(mt, length)
}

func (t *table) Type() *sabi.SwissMapType {
	return t.typ
}
