// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
)

const (
	UintptrMask = 1<<(8*_core.PtrSize) - 1
	PoisonStack = UintptrMask & 0x6868686868686868

	// Goroutine preemption request.
	// Stored into g->stackguard0 to cause split stack check failure.
	// Must be greater than any real sp.
	// 0xfffffade in hex.
	StackPreempt = UintptrMask & -1314

	// Thread is forking.
	// Stored into g->stackguard0 to cause split stack check failure.
	// Must be greater than any real sp.
	StackFork = UintptrMask & -1234
)

// Cached value of haveexperiment("framepointer")
var Framepointer_enabled bool
