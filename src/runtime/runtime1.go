// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_core "runtime/internal/core"
	"unsafe"
)

var (
	// TODO: Retire in favor of GOOS== checks.
	isplan9   int32
	issolaris int32
	iswindows int32
)

var typelink, etypelink [0]byte

//go:linkname reflect_typelinks reflect.typelinks
//go:nosplit
func reflect_typelinks() []*_core.Type {
	var ret []*_core.Type
	sp := (*_core.Slice)(unsafe.Pointer(&ret))
	sp.Array = (*byte)(unsafe.Pointer(&typelink))
	sp.Len = uint((uintptr(unsafe.Pointer(&etypelink)) - uintptr(unsafe.Pointer(&typelink))) / unsafe.Sizeof(ret[0]))
	sp.Cap = sp.Len
	return ret
}
