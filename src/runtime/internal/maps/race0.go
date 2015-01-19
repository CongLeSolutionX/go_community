// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package maps

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func Racewritepc(addr unsafe.Pointer, callerpc, pc uintptr) { _lock.Throw("race") }
