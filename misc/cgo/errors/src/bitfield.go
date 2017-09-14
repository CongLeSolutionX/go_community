// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

/*
struct with_bitfield {
	int before;
	int tiny:2;
	int corruptible:2;
	int after;
};
*/
import "C"

func F() {
	s := C.struct_with_bitfield{}
	s.tiny = 4 // ERROR HERE: no field or method tiny
}
