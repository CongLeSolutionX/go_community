// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	"fmt"
	"internal/abi"
	"testing"
	"unsafe"
)

func newTestTable[K comparable, V any](length uint32) *table {
	var m map[K]V
	mTyp := abi.TypeOf(m)
	omt := (*abi.OldMapType)(unsafe.Pointer(mTyp))

	mt := &abi.SwissMapType{
		Key:    omt.Key,
		Elem:   omt.Elem,
		Hasher: omt.Hasher,
	}
	return newTable(mt, length)
}

func TestTable(t *testing.T) {
	tab := newTestTable[uint32, uint64](32)

	key := uint32(0)
	elem := uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		tab.Put(unsafe.Pointer(&key), unsafe.Pointer(&elem))

		if debugLog {
			fmt.Printf("After put %d: %v\n", key, tab)
		}
	}

	key = uint32(0)
	elem = uint64(256+0)

	for i := 0; i < 31; i++ {
		key += 1
		elem += 1
		got, ok := tab.Get(unsafe.Pointer(&key))
		if !ok {
			t.Errorf("Get(%d) got ok false want true", key)
		}
		gotElem := *(*uint64)(got)
		if gotElem != elem {
			t.Errorf("Get(%d) got elem %d want %d", key, gotElem, elem)
		}
	}
}
