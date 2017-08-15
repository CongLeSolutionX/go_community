// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/obj/s390x"
	"cmd/internal/obj/x86"
	"cmd/internal/src"
	"fmt"
	"testing"
)

var CheckFunc = checkFunc
var Opt = opt
var Deadcode = deadcode
var Copyelim = copyelim

var testCtxts = map[string]*obj.Link{
	"amd64": obj.Linknew(&x86.Linkamd64),
	"s390x": obj.Linknew(&s390x.Links390x),
}

func testConfig(tb testing.TB) *Conf      { return testConfigArch(tb, "amd64") }
func testConfigS390X(tb testing.TB) *Conf { return testConfigArch(tb, "s390x") }

func testConfigArch(tb testing.TB, arch string) *Conf {
	ctxt, ok := testCtxts[arch]
	if !ok {
		tb.Fatalf("unknown arch %s", arch)
	}
	if ctxt.Arch.PtrSize != 8 {
		tb.Fatal("dummyTypes is 64-bit only")
	}
	c := &Conf{
		config: NewConfig(arch, dummyTypes, ctxt, true),
		tb:     tb,
	}
	return c
}

type Conf struct {
	config *Config
	tb     testing.TB
	fe     Frontend
}

func (c *Conf) Frontend() Frontend {
	if c.fe == nil {
		c.fe = DummyFrontend{t: c.tb, ctxt: c.config.ctxt}
	}
	return c.fe
}

// DummyFrontend is a test-only frontend.
// It assumes 64 bit integers and pointers.
type DummyFrontend struct {
	t    testing.TB
	ctxt *obj.Link
}

type DummyAuto struct {
	t *types.Type
	s string
}

func (d *DummyAuto) Typ() *types.Type {
	return d.t
}

func (d *DummyAuto) String() string {
	return d.s
}

func (d *DummyAuto) StorageClass() StorageClass {
	return ClassAuto
}

func (DummyFrontend) StringData(s string) interface{} {
	return nil
}
func (DummyFrontend) Auto(pos src.XPos, t *types.Type) GCNode {
	return &DummyAuto{t: t, s: "aDummyAuto"}
}
func (d DummyFrontend) SplitString(s LocalSlot) (LocalSlot, LocalSlot) {
	return LocalSlot{N: s.N, Type: dummyTypes.BytePtr, Off: s.Off}, LocalSlot{N: s.N, Type: dummyTypes.Int, Off: s.Off + 8}
}
func (d DummyFrontend) SplitInterface(s LocalSlot) (LocalSlot, LocalSlot) {
	return LocalSlot{N: s.N, Type: dummyTypes.BytePtr, Off: s.Off}, LocalSlot{N: s.N, Type: dummyTypes.BytePtr, Off: s.Off + 8}
}
func (d DummyFrontend) SplitSlice(s LocalSlot) (LocalSlot, LocalSlot, LocalSlot) {
	return LocalSlot{N: s.N, Type: s.Type.ElemType().PtrTo(), Off: s.Off},
		LocalSlot{N: s.N, Type: dummyTypes.Int, Off: s.Off + 8},
		LocalSlot{N: s.N, Type: dummyTypes.Int, Off: s.Off + 16}
}
func (d DummyFrontend) SplitComplex(s LocalSlot) (LocalSlot, LocalSlot) {
	if s.Type.Size() == 16 {
		return LocalSlot{N: s.N, Type: dummyTypes.Float64, Off: s.Off}, LocalSlot{N: s.N, Type: dummyTypes.Float64, Off: s.Off + 8}
	}
	return LocalSlot{N: s.N, Type: dummyTypes.Float32, Off: s.Off}, LocalSlot{N: s.N, Type: dummyTypes.Float32, Off: s.Off + 4}
}
func (d DummyFrontend) SplitInt64(s LocalSlot) (LocalSlot, LocalSlot) {
	if s.Type.IsSigned() {
		return LocalSlot{N: s.N, Type: dummyTypes.Int32, Off: s.Off + 4}, LocalSlot{N: s.N, Type: dummyTypes.UInt32, Off: s.Off}
	}
	return LocalSlot{N: s.N, Type: dummyTypes.UInt32, Off: s.Off + 4}, LocalSlot{N: s.N, Type: dummyTypes.UInt32, Off: s.Off}
}
func (d DummyFrontend) SplitStruct(s LocalSlot, i int) LocalSlot {
	return LocalSlot{N: s.N, Type: s.Type.FieldType(i), Off: s.Off + s.Type.FieldOff(i)}
}
func (d DummyFrontend) SplitArray(s LocalSlot) LocalSlot {
	return LocalSlot{N: s.N, Type: s.Type.ElemType(), Off: s.Off}
}
func (DummyFrontend) Line(_ src.XPos) string {
	return "unknown.go:0"
}
func (DummyFrontend) AllocFrame(f *Func) {
}
func (d DummyFrontend) Syslook(s string) *obj.LSym {
	return d.ctxt.Lookup(s)
}
func (DummyFrontend) UseWriteBarrier() bool {
	return true // only writebarrier_test cares
}

func (d DummyFrontend) Logf(msg string, args ...interface{}) { d.t.Logf(msg, args...) }
func (d DummyFrontend) Log() bool                            { return true }

func (d DummyFrontend) Fatalf(_ src.XPos, msg string, args ...interface{}) { d.t.Fatalf(msg, args...) }
func (d DummyFrontend) Warnl(_ src.XPos, msg string, args ...interface{})  { d.t.Logf(msg, args...) }
func (d DummyFrontend) Debug_checknil() bool                               { return false }
func (d DummyFrontend) Debug_wb() bool                                     { return false }

var dummyTypes Types

func init() {
	// Initialize just enough of the universe and the types package to make our tests function.
	// TODO(josharian): move universe initialization to the types package,
	// so this test setup can share it.

	types.Tconv = func(t *types.Type, flag, mode, depth int) string {
		return t.Etype.String()
	}
	types.Sconv = func(s *types.Sym, flag, mode int) string {
		return "sym"
	}
	types.FormatSym = func(sym *types.Sym, s fmt.State, verb rune, mode int) {
		fmt.Fprintf(s, "sym")
	}
	types.FormatType = func(t *types.Type, s fmt.State, verb rune, mode int) {
		fmt.Fprintf(s, "%v", t.Etype)
	}
	types.Dowidth = func(t *types.Type) {}

	types.Tptr = types.TPTR64
	for _, typ := range [...]struct {
		width int64
		et    types.EType
	}{
		{1, types.TINT8},
		{1, types.TUINT8},
		{1, types.TBOOL},
		{2, types.TINT16},
		{2, types.TUINT16},
		{4, types.TINT32},
		{4, types.TUINT32},
		{4, types.TFLOAT32},
		{4, types.TFLOAT64},
		{8, types.TUINT64},
		{8, types.TINT64},
		{8, types.TINT},
		{8, types.TUINTPTR},
	} {
		t := types.New(typ.et)
		t.Width = typ.width
		t.Align = uint8(typ.width)
		types.Types[typ.et] = t
	}

	dummyTypes = Types{
		Bool:       types.Types[types.TBOOL],
		Int8:       types.Types[types.TINT8],
		Int16:      types.Types[types.TINT16],
		Int32:      types.Types[types.TINT32],
		Int64:      types.Types[types.TINT64],
		UInt8:      types.Types[types.TUINT8],
		UInt16:     types.Types[types.TUINT16],
		UInt32:     types.Types[types.TUINT32],
		UInt64:     types.Types[types.TUINT64],
		Float32:    types.Types[types.TFLOAT32],
		Float64:    types.Types[types.TFLOAT64],
		Int:        types.Types[types.TINT],
		Uintptr:    types.Types[types.TUINTPTR],
		String:     types.Types[types.TSTRING],
		BytePtr:    types.NewPtr(types.Types[types.TUINT8]),
		Int32Ptr:   types.NewPtr(types.Types[types.TINT32]),
		UInt32Ptr:  types.NewPtr(types.Types[types.TUINT32]),
		IntPtr:     types.NewPtr(types.Types[types.TINT]),
		UintptrPtr: types.NewPtr(types.Types[types.TUINTPTR]),
		Float32Ptr: types.NewPtr(types.Types[types.TFLOAT32]),
		Float64Ptr: types.NewPtr(types.Types[types.TFLOAT64]),
		BytePtrPtr: types.NewPtr(types.NewPtr(types.Types[types.TUINT8])),
		CardMarks:  types.New(types.TSTRUCT), // types.New(types.TSTRUCT),
	}
}

//func getCM() *types.Type {
//	result := types.New(types.TSTRUCT)
//	return result
//}

/*** RLH just for now Simply delete before submitting.

// It seems that none of the fields of CardMarks have to be provided.

func (t *Type) Fields() *Fields {
	switch t.Etype {
	case TSTRUCT:
		return &t.Extra.(*Struct).fields
	case TINTER:
		Dowidth(t)
		return &t.Extra.(*Interface).Fields
	}
	Fatalf("Fields: type %v does not have fields", t)
	return nil
}

// Field returns the i'th field/method of struct/interface type t.
func (t *Type) Field(i int) *Field {
	return t.Fields().Slice()[i]
}

// FieldSlice returns a slice containing all fields/methods of
// struct/interface type t.
func (t *Type) FieldSlice() []*Field {
	return t.Fields().Slice()
}

// SetFields sets struct/interface type t's fields/methods to fields.
func (t *Type) SetFields(fields []*Field) {
	// If we've calculated the width of t before,
	// then some other type such as a function signature
	// might now have the wrong type.
	// Rather than try to track and invalidate those,
	// enforce that SetFields cannot be called once
	// t's width has been calculated.
	if t.WidthCalculated() {
		Fatalf("SetFields of %v: width previously calculated", t)
	}
	t.wantEtype(TSTRUCT)
	for _, f := range fields {
		// If type T contains a field F with a go:notinheap
		// type, then T must also be go:notinheap. Otherwise,
		// you could heap allocate T and then get a pointer F,
		// which would be a heap pointer to a go:notinheap
		// type.
		if f.Type != nil && f.Type.NotInHeap() {
			t.SetNotInHeap(true)
			break
		}
	}
	t.Fields().Set(fields)
}
type Fields struct {
	s *[]*Field
}
// A Field represents a field in a struct or a method in an interface or
// associated with a named type.
type Field struct {
	flags bitset8

	Embedded uint8 // embedded field
	Funarg   Funarg

	Sym   *Sym
	Nname *Node

	Type *Type // field type

	// Offset in bytes of this field or method within its enclosing struct
	// or interface Type.
	Offset int64

	Note string // literal string annotation
}

func rlhgenCMStruct() *types.Type {
	newStruct := types.New(types.TSTRUCT)
	_ = newStruct
	//	newStruct.fields = new(types.Fields)
	// Create the 3 fields.
	f1 := types.NewField()
	//	f1.flags = 0
	f1.Embedded = 0
	f1.Sym = nil
	f1.Nname = nil
	f1.Type = nil
	f1.Offset = 0
	f1.Note = ""
	f2 := types.NewField()
	f2.Embedded = 0
	f2.Sym = nil
	f2.Nname = nil
	f2.Type = nil
	f2.Offset = 0
	f2.Note = ""
	f3 := types.NewField()
	f3.Embedded = 0
	f3.Sym = nil
	f3.Nname = nil
	f3.Type = nil
	f3.Offset = 0
	f3.Note = ""
	//newStruct.fields[0] = f1
	_ = f1
	_ = f2
	_ = f3
	return newStruct
	//newStruct.fields[1] = f2
	//newStruct.fields[2] = f3
}
************/

func (d DummyFrontend) DerefItab(sym *obj.LSym, off int64) *obj.LSym { return nil }

func (d DummyFrontend) CanSSA(t *types.Type) bool {
	// There are no un-SSAable types in dummy land.
	return true
}
