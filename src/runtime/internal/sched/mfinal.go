// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
)

var Finlock _core.Mutex // protects the following variables
var Fingwait bool
var Fingwake bool

func wakefing() *_core.G {
	var res *_core.G
	_lock.Lock(&Finlock)
	if Fingwait && Fingwake {
		Fingwait = false
		Fingwake = false
		res = _core.Fing
	}
	_lock.Unlock(&Finlock)
	return res
}
