// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Per-P malloc cache for small objects.
//
// See malloc.h for an overview.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func freemcache(c *_core.Mcache) {
	_lock.Systemstack(func() {
		mCache_ReleaseAll(c)
		stackcache_clear(c)

		// NOTE(rsc,rlh): If gcworkbuffree comes back, we need to coordinate
		// with the stealing of gcworkbufs during garbage collection to avoid
		// a race where the workbuf is double-freed.
		// gcworkbuffree(c.gcworkbuf)

		_lock.Lock(&_lock.Mheap_.Lock)
		Purgecachedstats(c)
		_sched.FixAlloc_Free(&_lock.Mheap_.Cachealloc, unsafe.Pointer(c))
		_lock.Unlock(&_lock.Mheap_.Lock)
	})
}

func mCache_ReleaseAll(c *_core.Mcache) {
	for i := 0; i < _core.NumSizeClasses; i++ {
		s := c.Alloc[i]
		if s != &_lock.Emptymspan {
			mCentral_UncacheSpan(&_lock.Mheap_.Central[i].Mcentral, s)
			c.Alloc[i] = &_lock.Emptymspan
		}
	}
}
