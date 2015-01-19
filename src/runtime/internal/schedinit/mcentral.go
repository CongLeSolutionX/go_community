// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Central free lists.
//
// See malloc.h for an overview.
//
// The MCentral doesn't actually contain the list of free objects; the MSpan does.
// Each MCentral is two lists of MSpans: those with free objects (c->nonempty)
// and those that are completely allocated (c->empty).

package schedinit

import (
	_lock "runtime/internal/lock"
)

// Initialize a single central free list.
func mCentral_Init(c *_lock.Mcentral, sizeclass int32) {
	c.Sizeclass = sizeclass
	mSpanList_Init(&c.Nonempty)
	mSpanList_Init(&c.Empty)
}
