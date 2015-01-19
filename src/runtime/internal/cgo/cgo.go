// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgo

import (
	"unsafe"
)

//go:cgo_export_static main

// Filled in by runtime/cgo when linked into binary.

//go:linkname _cgo_malloc runtime/internal/cgo.Cgo_malloc
//go:linkname _cgo_free runtime/internal/cgo.Cgo_free

var (
	Cgo_malloc unsafe.Pointer
	Cgo_free   unsafe.Pointer
)
