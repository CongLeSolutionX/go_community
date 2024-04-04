// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unique

import (
	"internal/abi"
	"unsafe"
)

// clone shallow-clones value. For most types this is a no-op, but we explicitly
// clone directly-referenced strings to avoid accidentally giving a large string
// a long lifetime. This includes strings that are fields of structs, but not
// those stored in interfaces.
func clone[T comparable](value T, seq *cloneSeq) T {
	for _, offset := range seq.strOffsets {
		ps := (*string)(unsafe.Pointer(uintptr(unsafe.Pointer(&value)) + offset))
		*ps = cloneString(*ps)
	}
	return value
}

// singleStringClone describes how to clone a single string.
var singleStringClone = cloneSeq{strOffsets: []uintptr{0}}

// cloneSeq describes how to clone a value of a particular type.
type cloneSeq struct {
	strOffsets []uintptr
}

// makeCloneSeq creates a cloneSeq for a type.
func makeCloneSeq(typ *abi.Type) cloneSeq {
	if typ == nil {
		return cloneSeq{}
	}
	if typ.Kind() == abi.String {
		return singleStringClone
	}
	var seq cloneSeq
	if typ.Kind() == abi.Struct {
		buildStructCloneSeq(typ, &seq, 0)
	}
	return seq
}

// buildStructCloneSeq populates a cloneSeq for an abi.Type that has Kind abi.Struct.
// baseOffset should always be zero when calling this function.
func buildStructCloneSeq(typ *abi.Type, seq *cloneSeq, baseOffset uintptr) {
	styp := typ.StructType()
	for i := range styp.Fields {
		f := &styp.Fields[i]
		switch f.Typ.Kind() {
		case abi.String:
			seq.strOffsets = append(seq.strOffsets, baseOffset+f.Offset)
		case abi.Struct:
			buildStructCloneSeq(f.Typ, seq, baseOffset+f.Offset)
		}
	}
}

// cloneString is a copy of strings.Clone, because we can't depend on the strings
// package here. Several packages that might make use of unique, like net, explicitly
// forbid depending on unicode, which strings depends on.
func cloneString(s string) string {
	if len(s) == 0 {
		return ""
	}
	b := make([]byte, len(s))
	copy(b, s)
	return unsafe.String(&b[0], len(b))
}
