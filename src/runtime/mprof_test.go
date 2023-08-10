// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	. "runtime"
	"testing"
)

func TestProfileMaxStackCompatibility(t *testing.T) {
	var (
		sr  StackRecord
		mpr MemProfileRecord
	)
	if n := len(sr.Stack0); n != MaxStack {
		t.Fatalf("Expected length of Stack0 for runtime.StackRecord to be %d, but got %d", MaxStack, n)
	}
	if n := len(mpr.Stack0); n != MaxStack {
		t.Fatalf("Expected length of Stack0 for runtime.MemProfileRecord to be %d, but got %d", MaxStack, n)
	}
}
