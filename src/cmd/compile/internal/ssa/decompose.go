// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
)

type decomposedValueMap map[*Value][]*Value

// decompose converts phi ops on compound builtin types into phi
// ops on simple types, then invokes rewrite rules to decompose
// other ops on those types.
func decomposeBuiltIn(f *Func) {

	// When an aggregate value is decomposed, its new parts are recorded in this map,
	// so that field names can be directly bound to where they need to be.
	dvm := make(decomposedValueMap)

	// Decompose phis
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}
			decomposeBuiltInPhi(v, dvm)
		}
	}

	// Decompose other values
	applyRewrite(f, rewriteBlockdec, func(v *Value) bool {
		ret := rewriteValuedec(v)
		if ret {
			dvm.noteDecomposed(v)
		}
		return ret
	})
	if f.Config.RegSize == 4 {
		applyRewrite(f, rewriteBlockdec64, func(v *Value) bool {
			ret := rewriteValuedec64(v)
			if ret {
				dvm.noteDecomposed(v)
			}
			return ret
		})
	}

	// Split up named values into their components.
	var newNames []LocalSlot
	for _, name := range f.Names {
		t := name.Type
		switch {
		case t.IsInteger() && t.Size() > f.Config.RegSize:
			hiName, loName := f.fe.SplitInt64(name)
			newNames = append(newNames, hiName, loName)
			for _, v := range f.NamedValues[name] {
				if ev, ok := dvm[v]; ok {
					f.NamedValues[hiName] = append(f.NamedValues[hiName], ev[0])
					f.NamedValues[loName] = append(f.NamedValues[loName], ev[1])
				}
			}
			delete(f.NamedValues, name)
		case t.IsComplex():
			rName, iName := f.fe.SplitComplex(name)
			newNames = append(newNames, rName, iName)
			for _, v := range f.NamedValues[name] {
				if ev, ok := dvm[v]; ok {
					f.NamedValues[rName] = append(f.NamedValues[rName], ev[0])
					f.NamedValues[iName] = append(f.NamedValues[iName], ev[1])
				}
			}
			delete(f.NamedValues, name)
		case t.IsString():
			ptrName, lenName := f.fe.SplitString(name)
			newNames = append(newNames, ptrName, lenName)
			for _, v := range f.NamedValues[name] {
				if ev, ok := dvm[v]; ok {
					f.NamedValues[ptrName] = append(f.NamedValues[ptrName], ev[0])
					f.NamedValues[lenName] = append(f.NamedValues[lenName], ev[1])
				}
			}
			delete(f.NamedValues, name)
		case t.IsSlice():
			ptrName, lenName, capName := f.fe.SplitSlice(name)
			newNames = append(newNames, ptrName, lenName, capName)
			for _, v := range f.NamedValues[name] {
				if ev, ok := dvm[v]; ok {
					f.NamedValues[ptrName] = append(f.NamedValues[ptrName], ev[0])
					f.NamedValues[lenName] = append(f.NamedValues[lenName], ev[1])
					f.NamedValues[capName] = append(f.NamedValues[capName], ev[2])
				}
			}
			delete(f.NamedValues, name)
		case t.IsInterface():
			typeName, dataName := f.fe.SplitInterface(name)
			newNames = append(newNames, typeName, dataName)
			for _, v := range f.NamedValues[name] {
				if ev, ok := dvm[v]; ok {
					f.NamedValues[typeName] = append(f.NamedValues[typeName], ev[0])
					f.NamedValues[dataName] = append(f.NamedValues[dataName], ev[1])
				}
			}
			delete(f.NamedValues, name)
		case t.IsFloat():
			// floats are never decomposed, even ones bigger than RegSize
			newNames = append(newNames, name)
		case t.Size() > f.Config.RegSize:
			f.Fatalf("undecomposed named type %s %v", name, t)
		default:
			newNames = append(newNames, name)
		}
	}
	f.Names = newNames
}

func decomposeBuiltInPhi(v *Value, dvm decomposedValueMap) {
	switch {
	case v.Type.IsInteger() && v.Type.Size() > v.Block.Func.Config.RegSize:
		decomposeInt64Phi(v, dvm)
	case v.Type.IsComplex():
		decomposeComplexPhi(v, dvm)
	case v.Type.IsString():
		decomposeStringPhi(v, dvm)
	case v.Type.IsSlice():
		decomposeSlicePhi(v, dvm)
	case v.Type.IsInterface():
		decomposeInterfacePhi(v, dvm)
	case v.Type.IsFloat():
		// floats are never decomposed, even ones bigger than RegSize
	case v.Type.Size() > v.Block.Func.Config.RegSize:
		v.Fatalf("undecomposed type %s", v.Type)
	}
}

// noteDecomposed handles the case where a Foo-valued Phi or Load
// is converted to a FooMake of its parts; the name needs
// to be pushed to its args with appropriate field-name suffixing.
func (dvm decomposedValueMap) noteDecomposed(v *Value) {
	// TODO This appears to be over-general; so far all cases are handled
	// merely by recording the args, which don't need to be stored in a map
	// because they are attached to the value already.
	dvm[v] = v.Args
}

func decomposeStringPhi(v *Value, dvm decomposedValueMap) {
	types := &v.Block.Func.Config.Types
	ptrType := types.BytePtr
	lenType := types.Int

	ptr := v.Block.NewValue0(v.Pos, OpPhi, ptrType)
	len := v.Block.NewValue0(v.Pos, OpPhi, lenType)
	for _, a := range v.Args {
		ptr.AddArg(a.Block.NewValue1(v.Pos, OpStringPtr, ptrType, a))
		len.AddArg(a.Block.NewValue1(v.Pos, OpStringLen, lenType, a))
	}
	v.reset(OpStringMake)
	v.AddArg(ptr)
	v.AddArg(len)
	dvm.noteDecomposed(v)
}

func decomposeSlicePhi(v *Value, dvm decomposedValueMap) {
	types := &v.Block.Func.Config.Types
	ptrType := types.BytePtr
	lenType := types.Int

	ptr := v.Block.NewValue0(v.Pos, OpPhi, ptrType)
	len := v.Block.NewValue0(v.Pos, OpPhi, lenType)
	cap := v.Block.NewValue0(v.Pos, OpPhi, lenType)
	for _, a := range v.Args {
		ptr.AddArg(a.Block.NewValue1(v.Pos, OpSlicePtr, ptrType, a))
		len.AddArg(a.Block.NewValue1(v.Pos, OpSliceLen, lenType, a))
		cap.AddArg(a.Block.NewValue1(v.Pos, OpSliceCap, lenType, a))
	}
	v.reset(OpSliceMake)
	v.AddArg(ptr)
	v.AddArg(len)
	v.AddArg(cap)
	dvm.noteDecomposed(v)
}

func decomposeInt64Phi(v *Value, dvm decomposedValueMap) {
	cfgtypes := &v.Block.Func.Config.Types
	var partType *types.Type
	if v.Type.IsSigned() {
		partType = cfgtypes.Int32
	} else {
		partType = cfgtypes.UInt32
	}

	hi := v.Block.NewValue0(v.Pos, OpPhi, partType)
	lo := v.Block.NewValue0(v.Pos, OpPhi, cfgtypes.UInt32)
	for _, a := range v.Args {
		hi.AddArg(a.Block.NewValue1(v.Pos, OpInt64Hi, partType, a))
		lo.AddArg(a.Block.NewValue1(v.Pos, OpInt64Lo, cfgtypes.UInt32, a))
	}
	v.reset(OpInt64Make)
	v.AddArg(hi)
	v.AddArg(lo)
	dvm.noteDecomposed(v)
}

func decomposeComplexPhi(v *Value, dvm decomposedValueMap) {
	cfgtypes := &v.Block.Func.Config.Types
	var partType *types.Type
	switch z := v.Type.Size(); z {
	case 8:
		partType = cfgtypes.Float32
	case 16:
		partType = cfgtypes.Float64
	default:
		v.Fatalf("decomposeComplexPhi: bad complex size %d", z)
	}

	real := v.Block.NewValue0(v.Pos, OpPhi, partType)
	imag := v.Block.NewValue0(v.Pos, OpPhi, partType)
	for _, a := range v.Args {
		real.AddArg(a.Block.NewValue1(v.Pos, OpComplexReal, partType, a))
		imag.AddArg(a.Block.NewValue1(v.Pos, OpComplexImag, partType, a))
	}
	v.reset(OpComplexMake)
	v.AddArg(real)
	v.AddArg(imag)
	dvm.noteDecomposed(v)
}

func decomposeInterfacePhi(v *Value, dvm decomposedValueMap) {
	ptrType := v.Block.Func.Config.Types.BytePtr

	itab := v.Block.NewValue0(v.Pos, OpPhi, ptrType)
	data := v.Block.NewValue0(v.Pos, OpPhi, ptrType)
	for _, a := range v.Args {
		itab.AddArg(a.Block.NewValue1(v.Pos, OpITab, ptrType, a))
		data.AddArg(a.Block.NewValue1(v.Pos, OpIData, ptrType, a))
	}
	v.reset(OpIMake)
	v.AddArg(itab)
	v.AddArg(data)
	dvm.noteDecomposed(v)
}

func decomposeUser(f *Func) {
	dvm := make(decomposedValueMap)
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op != OpPhi {
				continue
			}
			decomposeUserPhi(v, dvm)
		}
	}
	// Split up named values into their components.
	i := 0
	var newNames []LocalSlot
	for _, name := range f.Names {
		t := name.Type
		switch {
		case t.IsStruct():
			newNames = append(newNames, decomposeUserStruct(f, name, dvm)...)
		case t.IsArray():
			newNames = append(newNames, decomposeUserArray(f, name, dvm)...)
		default:
			f.Names[i] = name
			i++
		}
	}
	f.Names = f.Names[:i]
	f.Names = append(f.Names, newNames...)
}

func decomposeUserArray(f *Func, name LocalSlot, dvm decomposedValueMap) []LocalSlot {
	t := name.Type
	if t.NumElem() == 0 {
		// TODO(khr): Not sure what to do here.  Probably nothing.
		// Names for empty arrays aren't important.
		return []LocalSlot{}
	}
	if t.NumElem() != 1 {
		// shouldn't get here due to CanSSA
		f.Fatalf("array not of size 1")
	}
	elemName := f.fe.SplitArray(name)
	for _, v := range f.NamedValues[name] {
		if ev, ok := dvm[v]; ok {
			f.NamedValues[elemName] = append(f.NamedValues[elemName], ev[0])
		}
	}
	// delete the name for the array as a whole
	delete(f.NamedValues, name)

	if t.ElemType().IsArray() {
		ret := decomposeUserArray(f, elemName, dvm)
		delete(f.NamedValues, elemName)
		return ret
	} else if t.ElemType().IsStruct() {
		ret := decomposeUserStruct(f, elemName, dvm)
		delete(f.NamedValues, elemName)
		return ret
	}

	// no need to record the name, as it's being decomposed further
	return []LocalSlot{elemName}
}

func decomposeUserStruct(f *Func, name LocalSlot, dvm decomposedValueMap) []LocalSlot {
	fnames := []LocalSlot{} // slots for struct in name
	ret := []LocalSlot{}    // slots for struct in name plus nested struct slots
	t := name.Type
	n := t.NumFields()

	for i := 0; i < n; i++ {
		fs := f.fe.SplitStruct(name, i)
		fnames = append(fnames, fs)
		// arrays and structs will be decomposed further, so
		// there's no need to record a name
		if !fs.Type.IsArray() && !fs.Type.IsStruct() {
			ret = append(ret, fs)
		}
	}

	// create named values for each struct field
	for _, v := range f.NamedValues[name] {
		if ev, ok := dvm[v]; ok {
			for i := 0; i < len(fnames); i++ {
				f.NamedValues[fnames[i]] = append(f.NamedValues[fnames[i]], ev[i])
			}
		}
	}
	// remove the name of the struct as a whole
	delete(f.NamedValues, name)

	// now that this f.NamedValues contains values for the struct
	// fields, recurse into nested structs
	for i := 0; i < n; i++ {
		if name.Type.FieldType(i).IsStruct() {
			ret = append(ret, decomposeUserStruct(f, fnames[i], dvm)...)
			delete(f.NamedValues, fnames[i])
		} else if name.Type.FieldType(i).IsArray() {
			ret = append(ret, decomposeUserArray(f, fnames[i], dvm)...)
			delete(f.NamedValues, fnames[i])
		}
	}
	return ret
}
func decomposeUserPhi(v *Value, dvm decomposedValueMap) {
	switch {
	case v.Type.IsStruct():
		decomposeStructPhi(v, dvm)
	case v.Type.IsArray():
		decomposeArrayPhi(v, dvm)
	}
}

// decomposeStructPhi replaces phi-of-struct with structmake(phi-for-each-field),
// and then recursively decomposes the phis for each field.
func decomposeStructPhi(v *Value, dvm decomposedValueMap) {
	t := v.Type
	n := t.NumFields()
	var fields [MaxStruct]*Value
	for i := 0; i < n; i++ {
		fields[i] = v.Block.NewValue0(v.Pos, OpPhi, t.FieldType(i))
	}
	for _, a := range v.Args {
		for i := 0; i < n; i++ {
			fields[i].AddArg(a.Block.NewValue1I(v.Pos, OpStructSelect, t.FieldType(i), int64(i), a))
		}
	}
	v.reset(StructMakeOp(n))
	v.AddArgs(fields[:n]...)
	dvm.noteDecomposed(v)

	// Recursively decompose phis for each field.
	for _, f := range fields[:n] {
		decomposeUserPhi(f, dvm)
	}
}

// decomposeArrayPhi replaces phi-of-array with arraymake(phi-of-array-element),
// and then recursively decomposes the element phi.
func decomposeArrayPhi(v *Value, dvm decomposedValueMap) {
	t := v.Type
	if t.NumElem() == 0 {
		v.reset(OpArrayMake0)
		return
	}
	if t.NumElem() != 1 {
		v.Fatalf("SSAable array must have no more than 1 element")
	}
	elem := v.Block.NewValue0(v.Pos, OpPhi, t.ElemType())
	for _, a := range v.Args {
		elem.AddArg(a.Block.NewValue1I(v.Pos, OpArraySelect, t.ElemType(), 0, a))
	}
	v.reset(OpArrayMake1)
	v.AddArg(elem)
	dvm.noteDecomposed(v)

	// Recursively decompose elem phi.
	decomposeUserPhi(elem, dvm)
}

// MaxStruct is the maximum number of fields a struct
// can have and still be SSAable.
const MaxStruct = 4

// StructMakeOp returns the opcode to construct a struct with the
// given number of fields.
func StructMakeOp(nf int) Op {
	switch nf {
	case 0:
		return OpStructMake0
	case 1:
		return OpStructMake1
	case 2:
		return OpStructMake2
	case 3:
		return OpStructMake3
	case 4:
		return OpStructMake4
	}
	panic("too many fields in an SSAable struct")
}
