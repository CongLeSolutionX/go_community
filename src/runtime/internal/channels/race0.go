// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package channels

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

// Because raceenabled is false, none of these functions should be called.

func RaceReadObjectPC(t *_core.Type, addr unsafe.Pointer, callerpc, pc uintptr) { _lock.Gothrow("race") }
func raceWriteObjectPC(t *_core.Type, addr unsafe.Pointer, callerpc, pc uintptr) {
	_lock.Gothrow("race")
}
func Racereadpc(addr unsafe.Pointer, callerpc, pc uintptr) { _lock.Gothrow("race") }
func raceacquireg(gp *_core.G, addr unsafe.Pointer)        { _lock.Gothrow("race") }
func Racerelease(addr unsafe.Pointer)                      { _lock.Gothrow("race") }
func racereleaseg(gp *_core.G, addr unsafe.Pointer)        { _lock.Gothrow("race") }
