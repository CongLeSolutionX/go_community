// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package strings

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

func Racereadrangepc(addr unsafe.Pointer, sz, callerpc, pc uintptr) { _lock.Throw("race") }
