// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package cgo

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func racereleasemerge(addr unsafe.Pointer) { _lock.Gothrow("race") }
func Racegostart(pc uintptr) uintptr       { _lock.Gothrow("race"); return 0 }
