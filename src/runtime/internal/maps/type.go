// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package maps

import (
	_core "runtime/internal/core"
)

type Maptype struct {
	typ           _core.Type
	key           *_core.Type
	elem          *_core.Type
	bucket        *_core.Type // internal type representing a hash bucket
	hmap          *_core.Type // internal type representing a hmap
	keysize       uint8       // size of key slot
	indirectkey   bool        // store ptr to key instead of key itself
	valuesize     uint8       // size of value slot
	indirectvalue bool        // store ptr to value instead of value itself
	bucketsize    uint16      // size of bucket
	reflexivekey  bool        // true if k==k for all keys
}
