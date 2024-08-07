// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"internal/abi"
)

type CtrlGroup = ctrlGroup

const DebugLog = debugLog

var AlignUpPow2 = alignUpPow2

const MaxTableCapacity = maxTableCapacity

func NewTestMap[K comparable, V any](length uint64) *Map {
	mt := newTestMapType[K, V]()
	return NewMap(mt, length)
}

//func NewTestTable[K comparable, V any](length uint64) *table {
//	mt := newTestMapType[K, V]()
//	return newTable(mt, length)
//}

func (t *table) Type() *abi.SwissMapType {
	return t.typ
}
