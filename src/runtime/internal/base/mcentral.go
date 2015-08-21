// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Central free lists.
//
// See malloc.go for an overview.
//
// The MCentral doesn't actually contain the list of free objects; the MSpan does.
// Each MCentral is two lists of MSpans: those with free objects (c->nonempty)
// and those that are completely allocated (c->empty).

package base

// Central list of free objects of a given size.
type Mcentral struct {
	Lock      Mutex
	Sizeclass int32
	Nonempty  Mspan // list of spans with a free object
	Empty     Mspan // list of spans with no free objects (or cached in an mcache)
}
