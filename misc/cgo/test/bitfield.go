// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

/*
struct with_bitfield {
	int before;
	int tiny:2;
	int after;
};
static int sum(struct with_bitfield wb) {
	return wb.before + wb.tiny + wb.after;
}
static struct with_bitfield sum7() {
	struct with_bitfield wb = { .before = 2, .tiny = 1, .after = 4 };
	return wb;
}
*/
import "C"

import "testing"

func testBitfield(t *testing.T) {
	wb := C.struct_with_bitfield{before: 1, after: 2}
	if s := C.sum(wb); s != 3 {
		t.Errorf("C.sum(%#v) = %v; want 3", s)
	}

	wb = C.sum7()
	if s := C.sum(wb); s != 7 {
		t.Errorf("C.sum(<struct with bitfield set>) = %v; want 7", s)
	}
}
