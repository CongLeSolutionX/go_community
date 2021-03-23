// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64
// +build amd64

package abi_test

import (
	"internal/abi"
	"testing"
)

func TestFuncPC(t *testing.T) {
	pcFromGo, pcFromAsm := abi.FuncPCTest()
	if pcFromGo != pcFromAsm {
		t.Errorf("FuncPC returns wrong PC, want %x, got %x", pcFromAsm, pcFromGo)
	}
}
