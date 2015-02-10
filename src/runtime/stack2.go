// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_lock "runtime/internal/lock"
)

// Goroutine preemption request.
// Stored into g->stackguard0 to cause split stack check failure.
// Must be greater than any real sp.
// 0xfffffade in hex.
const (
	_StackPreempt = _lock.UintptrMask & -1314
	_StackFork    = _lock.UintptrMask & -1234
)
