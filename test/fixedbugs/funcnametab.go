// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"runtime"
	_ "unsafe"
)

//go:linkname firstFunc type:.eq.internal/abi.RegArgs
func firstFunc()

//go:linkname processOptions internal/cpu.processOptions
func processOptions()

func main() {
	name := runtime.FuncForPC(funcPC(processOptions)).Name()
	if name != "internal/cpu.processOptions" {
		panic("FAIL: wrong name of processOptions")
	}
	name = runtime.FuncForPC(funcPC(firstFunc)).Name()
	if name != "type:.eq.internal/abi.RegArgs" {
		panic("FAIL: wrong name of firstFunc: " + name)
	}
}

// funcPC returns the PC for the func value f.
func funcPC(f interface{}) uintptr {
	return reflect.ValueOf(f).Pointer()
}
