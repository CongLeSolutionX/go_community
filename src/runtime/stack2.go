// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
)

// Goroutine preemption request.
// Stored into g->stackguard0 to cause split stack check failure.
// Must be greater than any real sp.
// 0xfffffade in hex.
const (
	_StackPreempt = _base.UintptrMask & -1314
	_StackFork    = _base.UintptrMask & -1234
)
