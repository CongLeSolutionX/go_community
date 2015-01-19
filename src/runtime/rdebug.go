// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_sched "runtime/internal/sched"
)

func setMaxStack(in int) (out int) {
	out = int(maxstacksize)
	maxstacksize = uintptr(in)
	return out
}

func setPanicOnFault(new bool) (old bool) {
	mp := _sched.Acquirem()
	old = mp.Curg.Paniconfault
	mp.Curg.Paniconfault = new
	_sched.Releasem(mp)
	return old
}
