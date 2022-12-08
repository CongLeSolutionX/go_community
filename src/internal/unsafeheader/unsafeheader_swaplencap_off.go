// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.swaplencap
// +build !goexperiment.swaplencap

package unsafeheader

import (
	"unsafe"
)

// Slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type Slice struct {
	_align [0][]byte // this ought to force alignment
	Data   unsafe.Pointer
	Len    int
	Cap    int
}
