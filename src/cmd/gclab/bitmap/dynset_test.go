// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitmap

import "testing"

func TestDynSetGrow(t *testing.T) {
	for i := range uint64(256) {
		t.Logf("set bit %d", i)
		set := DynSet[uint64]{}
		set.Add(i)
		if !set.Has(i) {
			t.Log(set.bits)
			t.Fatalf("bit %d not set", i)
		}
	}
}
