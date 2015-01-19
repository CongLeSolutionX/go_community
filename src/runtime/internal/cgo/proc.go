// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgo

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
)

func Allgadd(gp *_core.G) {
	if _lock.Readgstatus(gp) == _lock.Gidle {
		_lock.Gothrow("allgadd: bad status Gidle")
	}

	_lock.Lock(&_lock.Allglock)
	_lock.Allgs = append(_lock.Allgs, gp)
	allg = &_lock.Allgs[0]
	_gc.Allglen = uintptr(len(_lock.Allgs))
	_lock.Unlock(&_lock.Allglock)
}
