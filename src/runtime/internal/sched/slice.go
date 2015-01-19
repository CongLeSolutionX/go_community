// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sched

import (
	"unsafe"
)

type SliceStruct struct {
	Array unsafe.Pointer
	Len   int
	Cap   int
}
