// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

type Iface struct {
	Tab  *Itab
	Data unsafe.Pointer
}

type Eface struct {
	Type *_base.Type
	Data unsafe.Pointer
}

// layout of Itab known to compilers
// allocated in non-garbage-collected memory
type Itab struct {
	inter  *Interfacetype
	Type   *_base.Type
	Link   *Itab
	bad    int32
	unused int32
	fun    [1]uintptr // variable sized
}
