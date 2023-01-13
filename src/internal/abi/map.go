// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package abi

// Map constants common to several packages
// runtime/runtime-gdb.py:MapTypePrinter contains its own copy
// test/codegen/maps.go:MapLiteralSizing is only correct for MapBucketCountBits={3,4,5}
// runtime:TestMapIterOrder becomes flaky for maps of size 3 when MapBucketCountBits=5
const (
	MapBucketCountBits = 4 // log2 of number of elements in a bucket.
	MapBucketCount     = 1 << MapBucketCountBits
	MapMaxKeyBytes     = 128 // Must fit in a uint8.
	MapMaxElemBytes    = 128 // Must fit in a uint8.
)
