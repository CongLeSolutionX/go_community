// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
)

const (
	Thechar          = '6'
	BigEndian        = 0
	CacheLineSize    = 64
	RuntimeGogoBytes = 64 + (_core.Goos_plan9|goos_solaris|_core.Goos_windows)*16
	PhysPageSize     = 4096
	PCQuantum        = 1
	Int64Align       = 8
)
