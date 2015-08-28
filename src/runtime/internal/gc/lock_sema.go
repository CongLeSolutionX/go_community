// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin nacl netbsd openbsd plan9 solaris windows

package gc

import (
	_base "runtime/internal/base"
)

func Notetsleep(n *_base.Note, ns int64) bool {
	gp := _base.Getg()
	if gp != gp.M.G0 && gp.M.Preemptoff != "" {
		_base.Throw("notetsleep not on g0")
	}
	if gp.M.Waitsema == 0 {
		gp.M.Waitsema = _base.Semacreate()
	}
	return _base.Notetsleep_internal(n, ns, nil, 0)
}
