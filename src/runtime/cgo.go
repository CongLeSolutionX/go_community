// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

//go:cgo_export_static main

// Filled in by runtime/cgo when linked into binary.

//go:linkname _cgo_init _cgo_init

var (
	_cgo_init unsafe.Pointer
)
