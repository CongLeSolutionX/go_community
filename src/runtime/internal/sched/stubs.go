// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	_core "runtime/internal/core"
	"unsafe"
)

// mcall switches from the g to the g0 stack and invokes fn(g),
// where g is the goroutine that made the call.
// mcall saves g's current PC/SP in g->sched so that it can be restored later.
// It is up to fn to arrange for that later execution, typically by recording
// g in a data structure, causing something to call ready(g) later.
// mcall returns to the original goroutine g later, when g has been rescheduled.
// fn must not return at all; typically it ends by calling schedule, to let the m
// run other goroutines.
//
// mcall can only be called from g stacks (not g0, not gsignal).
//go:noescape
func Mcall(fn func(*_core.G))

// memmove copies n bytes from "from" to "to".
// in memmove_*.s
//go:noescape
func Memmove(to, from unsafe.Pointer, n uintptr)
func Gogo(buf *_core.Gobuf)
func gosave(buf *_core.Gobuf)

//go:noescape
func Cas(ptr *uint32, old, new uint32) bool

//go:noescape
func Asmcgocall(fn, arg unsafe.Pointer)
