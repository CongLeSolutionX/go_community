// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

func funcPCTestFn()

var funcPCTestFnAddr uintptr // address of funcPCTestFn, directly retrieved from assembly

//go:noinline
func FuncPCTest() (uintptr, uintptr) {
	return FuncPC(funcPCTestFn), funcPCTestFnAddr
}
