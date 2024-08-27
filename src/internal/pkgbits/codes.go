// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkgbits

// A Code is an enum value that can be encoded into bitstreams.
//
// Code types are preferable for enum types, because they allow
// Decoder to detect desyncs.
type Code interface {
	// Marker returns the SyncMarker for the Code's dynamic type.
	Marker() SyncMarker

	// Value returns the Code's ordinal value.
	Value() int
}

// A MutableCode is a Code whose value can be set.
type MutableCode interface {
	Code

	Set(int)
}

// A CodeVal distinguishes among go/constant.Value encodings.
type CodeVal int

func (c CodeVal) Marker() SyncMarker { return SyncVal }
func (c CodeVal) Value() int         { return int(c) }

func (c *CodeVal) Set(v int) { *c = CodeVal(v) }

// Note: These values are public and cannot be changed without
// updating the go/types importers.

const (
	ValBool CodeVal = iota
	ValString
	ValInt64
	ValBigInt
	ValBigRat
	ValBigFloat
)

// A CodeType distinguishes among go/types.Type encodings.
type CodeType int

func (c CodeType) Marker() SyncMarker { return SyncType }
func (c CodeType) Value() int         { return int(c) }

func (c *CodeType) Set(v int) { *c = CodeType(v) }

// Note: These values are public and cannot be changed without
// updating the go/types importers.

const (
	TypeBasic CodeType = iota
	TypeNamed
	TypePointer
	TypeSlice
	TypeArray
	TypeChan
	TypeMap
	TypeSignature
	TypeStruct
	TypeInterface
	TypeUnion
	TypeTypeParam
)

// A CodeObj distinguishes among go/types.Object encodings.
type CodeObj int

func (c CodeObj) Marker() SyncMarker { return SyncCodeObj }
func (c CodeObj) Value() int         { return int(c) }

func (c *CodeObj) Set(v int) { *c = CodeObj(v) }

// Note: These values are public and cannot be changed without
// updating the go/types importers.

const (
	ObjAlias CodeObj = iota
	ObjConst
	ObjType
	ObjFunc
	ObjVar
	ObjStub
)
