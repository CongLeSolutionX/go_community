// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package runtime

import (
	_base "runtime/internal/base"
)

type maptype struct {
	typ           _base.Type
	key           *_base.Type
	elem          *_base.Type
	bucket        *_base.Type // internal type representing a hash bucket
	hmap          *_base.Type // internal type representing a hmap
	keysize       uint8       // size of key slot
	indirectkey   bool        // store ptr to key instead of key itself
	valuesize     uint8       // size of value slot
	indirectvalue bool        // store ptr to value instead of value itself
	bucketsize    uint16      // size of bucket
	reflexivekey  bool        // true if k==k for all keys
}

type chantype struct {
	typ  _base.Type
	elem *_base.Type
	dir  uintptr
}

type slicetype struct {
	typ  _base.Type
	elem *_base.Type
}

type functype struct {
	typ       _base.Type
	dotdotdot bool
	in        _base.Slice
	out       _base.Slice
}
