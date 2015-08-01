// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

func freemcache(c *_base.Mcache) {
	_base.Systemstack(func() {
		mCache_ReleaseAll(c)
		stackcache_clear(c)

		// NOTE(rsc,rlh): If gcworkbuffree comes back, we need to coordinate
		// with the stealing of gcworkbufs during garbage collection to avoid
		// a race where the workbuf is double-freed.
		// gcworkbuffree(c.gcworkbuf)

		_base.Lock(&_base.Mheap_.Lock)
		Purgecachedstats(c)
		_base.FixAlloc_Free(&_base.Mheap_.Cachealloc, unsafe.Pointer(c))
		_base.Unlock(&_base.Mheap_.Lock)
	})
}

func mCache_ReleaseAll(c *_base.Mcache) {
	for i := 0; i < _base.NumSizeClasses; i++ {
		s := c.Alloc[i]
		if s != &_base.Emptymspan {
			mCentral_UncacheSpan(&_base.Mheap_.Central[i].Mcentral, s)
			c.Alloc[i] = &_base.Emptymspan
		}
	}
}
