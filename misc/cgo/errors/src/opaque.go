// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

/*
#include <stdint.h>

union ptr_or_eight {
	void *p;
	uint32_t endian[2];
};
*/
import "C"

func F() {
	blob := [8]byte{42}
	var _ C.union_ptr_or_eight = blob // ERROR HERE
}
