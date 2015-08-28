// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
)

func setMaxStack(in int) (out int) {
	out = int(maxstacksize)
	maxstacksize = uintptr(in)
	return out
}

func setPanicOnFault(new bool) (old bool) {
	mp := _base.Acquirem()
	old = mp.Curg.Paniconfault
	mp.Curg.Paniconfault = new
	_base.Releasem(mp)
	return old
}
