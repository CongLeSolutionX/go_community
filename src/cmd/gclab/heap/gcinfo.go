// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package heap

import (
	"cmd/gclab/bitmap"
	"log"
)

type GCInfo struct {
	Ps int

	Roots      []VAddr
	WBRoots    []VAddr // Pointers marked by write barriers
	AllocBlack []VAddr // Base address of objects allocated black

	Marks bitmap.Set[ObjectID]
}

func (i *GCInfo) CompareMarks(h *Heap, got bitmap.Set[ObjectID]) {
	for addr, objID := range h.Objects() {
		if i.Marks.Has(objID) && !got.Has(objID) {
			log.Printf("object %s: want marked, got not marked", addr)
		} else if !i.Marks.Has(objID) && got.Has(objID) {
			log.Printf("object %s: want not marked, got marked", addr)
		}
	}
}
