// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

func checkptrAlignment(p unsafe.Pointer, elem *_type) {
	// TODO(mdempsky): What about fieldAlign?
	if uintptr(p)&elem.align-1 != 0 {
		throw("bad pointer alignment")
	}
}

func checkptrArithmetic(p unsafe.Pointer, originals []unsafe.Pointer) {
	if 0 < uintptr(p) && uintptr(p) < minLegalPointer {
		throw("bad pointer arithmetic")
	}

	base := checkptrBase(p)
	if base == 0 {
		return
	}

	for _, original := range originals {
		if base == checkptrBase(original) {
			return
		}
	}

	throw("bad pointer arithmetic")
}

func checkptrBase(p unsafe.Pointer) uintptr {
	base, _, _ := findObject(uintptr(p), 0, 0)
	// TODO(mdempsky): If base == 0, then check if p points to the
	// stack or a global variable.
	return base
}
