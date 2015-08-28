// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

import (
	_base "runtime/internal/base"
)

const (
	KindBool          = _KindBool
	KindInt           = _KindInt
	KindInt8          = _KindInt8
	KindInt16         = _KindInt16
	KindInt32         = _KindInt32
	KindInt64         = _KindInt64
	KindUint          = _KindUint
	KindUint8         = _KindUint8
	KindUint16        = _KindUint16
	KindUint32        = _KindUint32
	KindUint64        = _KindUint64
	KindUintptr       = _KindUintptr
	KindFloat32       = _KindFloat32
	KindFloat64       = _KindFloat64
	KindComplex64     = _KindComplex64
	KindComplex128    = _KindComplex128
	KindArray         = _KindArray
	KindChan          = _KindChan
	KindFunc          = _KindFunc
	KindInterface     = _KindInterface
	KindMap           = _KindMap
	KindPtr           = _KindPtr
	KindSlice         = _KindSlice
	KindString        = _KindString
	KindStruct        = _KindStruct
	KindUnsafePointer = _KindUnsafePointer

	KindDirectIface = _KindDirectIface
	KindGCProg      = _KindGCProg
	KindNoPointers  = _KindNoPointers
	KindMask        = _KindMask
)

// isDirectIface reports whether t is stored directly in an interface value.
func IsDirectIface(t *_base.Type) bool {
	return t.Kind&KindDirectIface != 0
}
