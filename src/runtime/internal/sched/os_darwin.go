// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	"unsafe"
)

func bsdthread_create(stk unsafe.Pointer, mm *_core.M, gg *_core.G, fn uintptr) int32

//go:noescape
func setitimer(mode int32, new, old *itimerval)
func raiseproc(int32)
