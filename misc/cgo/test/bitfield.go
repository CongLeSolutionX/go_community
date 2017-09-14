// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

/*
struct with_bitfield {
	int before;
	int tiny:2;
	short after;
	int tiny_after:3;
};
static int sum(struct with_bitfield wb) {
	return wb.before + wb.tiny + wb.after + wb.tiny_after;
}
static struct with_bitfield sum15() {
	struct with_bitfield wb = { .before = 8, .tiny = 1, .after = 4, .tiny_after = 2 };
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

	wb = C.sum15()
	if s := C.sum(wb); s != 15 {
		t.Errorf("C.sum(<struct with bitfields summing to 15>) = %v; want 15", s)
	}
}
