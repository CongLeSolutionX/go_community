// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_ "unsafe" // for linkname.
)

// These functions are called from C code via cgo/callbacks.go.

// Panic.

func _cgo_panic_internal(p *byte) {
	panic(gostringnocopy(p))
}

//go:linkname _cgo_panic _cgo_panic
func _cgo_panic(a *struct{ cstr *byte }) {
	_cgo_panic_internal(a.cstr)
}
