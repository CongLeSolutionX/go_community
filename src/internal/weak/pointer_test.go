// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weak_test

import (
	"internal/weak"
	"runtime"
	"testing"
)

type T struct {
	// N.B. This must contain a pointer, otherwise the weak handle might get placed
	// in a tiny block making the tests in this package flaky.
	t *T
	a int
}

func TestPointer(t *testing.T) {
	bt := new(T)
	wt := weak.Make(bt)
	if st := wt.Strong(); st != bt {
		t.Fatalf("weak pointer is not the same as strong pointer: %p vs. %p", st, bt)
	}
	// bt is no longer referenced.

	// Two GC cycles are necessary to clean up a weak pointer that has been strong-ified.
	runtime.GC()
	runtime.GC()

	if st := wt.Strong(); st != bt {
		t.Fatalf("weak pointer is not the same as strong pointer after GC: %p vs. %p", st, bt)
	}

	// bt is no longer referenced.

	runtime.GC()
	runtime.GC()

	if st := wt.Strong(); st != nil {
		t.Fatalf("expected weak pointer to be nil, got %p", st)
	}
}
