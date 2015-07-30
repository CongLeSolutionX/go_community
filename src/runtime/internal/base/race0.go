// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package base

import (
	"unsafe"
)

const Raceenabled = false

func racemapshadow(addr unsafe.Pointer, size uintptr) { Throw("race") }
func Raceacquire(addr unsafe.Pointer)                 { Throw("race") }
func Racemalloc(p unsafe.Pointer, sz uintptr)         { Throw("race") }
func Racegostart(pc uintptr) uintptr                  { Throw("race"); return 0 }
