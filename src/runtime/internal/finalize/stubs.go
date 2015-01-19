// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package finalize

import (
	"unsafe"
)

func Reflectcall(fn, arg unsafe.Pointer, n uint32, retoffset uint32)
