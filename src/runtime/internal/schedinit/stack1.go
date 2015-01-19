// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

func stackinit() {
	if _core.StackCacheSize&_core.PageMask != 0 {
		_lock.Gothrow("cache size must be a multiple of page size")
	}
	for i := range _sched.Stackpool {
		mSpanList_Init(&_sched.Stackpool[i])
	}
}
