// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi_test

import (
	"internal/abi"
	"testing"
)

func TestFuncPC(t *testing.T) {
	pcFromAsm := abi.FuncPCTestFnAddr

	// Test FuncPC for locally defined function
	pcFromGo := abi.FuncPCTest()
	if pcFromGo != pcFromAsm {
		t.Errorf("FuncPC returns wrong PC, want %x, got %x", pcFromAsm, pcFromGo)
	}

	// Test FuncPC for imported function
	pcFromGo = abi.FuncPC(abi.FuncPCTestFn)
	if pcFromGo != pcFromAsm {
		t.Errorf("FuncPC returns wrong PC, want %x, got %x", pcFromAsm, pcFromGo)
	}
}
