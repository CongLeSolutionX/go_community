// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package race

import (
	_base "runtime/internal/base"
	"unsafe"
)

// Because raceenabled is false, none of these functions should be called.

func RaceReadObjectPC(t *_base.Type, addr unsafe.Pointer, callerpc, pc uintptr)  { _base.Throw("race") }
func RaceWriteObjectPC(t *_base.Type, addr unsafe.Pointer, callerpc, pc uintptr) { _base.Throw("race") }
func Raceinit() uintptr                                                          { _base.Throw("race"); return 0 }
func Racefini()                                                                  { _base.Throw("race") }
func Racewritepc(addr unsafe.Pointer, callerpc, pc uintptr)                      { _base.Throw("race") }
func Racereadpc(addr unsafe.Pointer, callerpc, pc uintptr)                       { _base.Throw("race") }
func Racereadrangepc(addr unsafe.Pointer, sz, callerpc, pc uintptr)              { _base.Throw("race") }
func Racewriterangepc(addr unsafe.Pointer, sz, callerpc, pc uintptr)             { _base.Throw("race") }
func raceacquireg(gp *_base.G, addr unsafe.Pointer)                              { _base.Throw("race") }
func Racerelease(addr unsafe.Pointer)                                            { _base.Throw("race") }
func racereleaseg(gp *_base.G, addr unsafe.Pointer)                              { _base.Throw("race") }
func Racereleasemerge(addr unsafe.Pointer)                                       { _base.Throw("race") }
func racereleasemergeg(gp *_base.G, addr unsafe.Pointer)                         { _base.Throw("race") }
func Racefingo()                                                                 { _base.Throw("race") }
func Racegoend()                                                                 { _base.Throw("race") }
