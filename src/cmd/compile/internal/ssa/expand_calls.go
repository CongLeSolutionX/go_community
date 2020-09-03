// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"sort"
)

type selKey struct {
	from   *Value
	offset int64
	size   int64
	typ    types.EType
}

type offsetKey struct {
	from   *Value
	offset int64
	pt     *types.Type
}

// expandCalls converts LE (Late Expansion) calls that act like they receive value args into a lower-level form
// that is more oriented to a platform's ABI.  The SelectN operations that extract results are rewritten into
// more appropriate forms, and any StructMake or ArrayMake inputs are decomposed until non-struct values are
// reached.
func expandCalls(f *Func) {
	// Calls that need lowering have some number of inputs, including a memory input,
	// and produce a tuple of (value1, value2, ..., mem) where valueK may or may not be SSA-able.

	// With the current ABI those inputs need to be converted into stores to memory,
	// rethreading the call's memory input to the first, and the new call now receiving the last.

	// With the current ABI, the outputs need to be converted to loads, which will all use the call's
	// memory output as their input.
	if !LateCallExpansionEnabledWithin(f) {
		return
	}
	debug := f.pass.debug > 0

	canSSAType := f.fe.CanSSA
	regSize := f.Config.RegSize
	sp, _ := f.spSb()
	typ := &f.Config.Types
	ptrSize := f.Config.PtrSize

	// For 32-bit, need to deal with decomposition of 64-bit integers, which depends on endianness.
	var hiOffset, lowOffset int64
	if f.Config.BigEndian {
		lowOffset = 4
	} else {
		hiOffset = 4
	}

	namedSelects := make(map[*Value][]namedVal)

	// intPairTypes returns the pair of 32-bit int types needed to encode a 64-bit integer type on a target
	// that has no 64-bit integer registers.
	intPairTypes := func(et types.EType) (tHi, tLo *types.Type) {
		tHi = typ.UInt32
		if et == types.TINT64 {
			tHi = typ.Int32
		}
		tLo = typ.UInt32
		return
	}

	// isAlreadyExpandedAggregateType returns whether a type is an SSA-able "aggregate" (multiple register) type
	// that was expanded in an earlier phase (small user-defined arrays and structs, lowered in decomposeUser).
	// Other aggregate types are expanded in decomposeBuiltin, which comes later.
	isAlreadyExpandedAggregateType := func(t *types.Type) bool {
		if !canSSAType(t) {
			return false
		}
		et := t.Etype
		switch et {
		case types.TINT64, types.TUINT64:
			return regSize == 4
		case types.TINTER, types.TSTRING, types.TSLICE, types.TCOMPLEX64, types.TCOMPLEX128:
			return true
		case types.TARRAY, types.TSTRUCT:
			return true
		}
		return false
	}

	offsets := make(map[offsetKey]*Value)

	// offsetFrom creates an offset from a pointer, simplifying chained offsets and offsets from SP
	// TODO should also optimize offsets from SB?
	offsetFrom := func(from *Value, offset int64, pt *types.Type) *Value {
		if offset == 0 && from.Type == pt { // this is not actually likely
			return from
		}
		// Simplify, canonicalize
		for from.Op == OpOffPtr {
			offset += from.AuxInt
			from = from.Args[0]
		}
		if from == sp {
			return f.ConstOffPtrSP(pt, offset, sp)
		}
		key := offsetKey{from, offset, pt}
		v := offsets[key]
		if v != nil {
			return v
		}
		v = from.Block.NewValue1I(from.Pos.WithNotStmt(), OpOffPtr, pt, offset, from)
		offsets[key] = v
		return v
	}

	splitSlots := func(ls []LocalSlot, sfx string, offset int64, ty *types.Type) []LocalSlot {
		var locs []LocalSlot
		for i := range ls {
			locs = append(locs, f.fe.SplitSlot(&ls[i], sfx, offset, ty))
		}
		return locs
	}

	// rewriteSelect recursively walks through a chain of Struct/Array/Builtin Select
	// operations until a select from call results is reached.  It emits the
	// code necessary to implement the leaf select operation that leads to the call.
	var rewriteSelect func(v *Value, sel *Value, offset int64, t *types.Type) []LocalSlot
	rewriteSelect = func(v *Value, sel *Value, offset int64, t *types.Type) []LocalSlot {
		var locs []LocalSlot
		switch sel.Op {
		case OpSelectN:
			for _, s := range namedSelects[sel] {
				locs = append(locs, f.Names[s.locIndex])
			}
			call := sel.Args[0]
			aux := call.Aux.(*AuxCall)
			which := sel.AuxInt
			if which == aux.ResultsLen() { // mem is after the results.
				// rewrite v as a Copy of call -- the replacement call will produce a mem.
				v.copyOf(call)
			} else {
				if canSSAType(t) {
					for t.Etype == types.TSTRUCT && t.NumFields() == 1 {
						// This may not be adequately general -- consider [1]etc but this is caused by immediate IDATA
						t = t.Field(0).Type
					}
					pt := types.NewPtr(t)
					off := offsetFrom(sp, offset+aux.OffsetOfResult(which), pt)
					// Any selection right out of the arg area/registers has to be same Block as call, use call as mem input.
					if v.Block == call.Block {
						v.reset(OpLoad)
						v.SetArgs2(off, call)
					} else {
						w := call.Block.NewValue2(v.Pos, OpLoad, t, off, call)
						v.copyOf(w)
					}
				} else {
					panic("Should not have non-SSA-able OpSelectN")
				}
			}
			v.Type = t // not right for the mem operand yet, but will be when call is rewritten.

		case OpStructSelect:
			w := sel.Args[0]
			var ls []LocalSlot
			if w.Type.Etype != types.TSTRUCT {
				fmt.Printf("Bad type for w: v=%v; sel=%v; w=%v; ,f=%s\n", v.LongString(), sel.LongString(), w.LongString(), f.Name)
				// Artifact of immediate interface idata
				ls = rewriteSelect(v, w, offset, t)
			} else {
				ls = rewriteSelect(v, w, offset+w.Type.FieldOff(int(sel.AuxInt)), t)
				for _, l := range ls {
					locs = append(locs, f.fe.SplitStruct(l, int(sel.AuxInt)))
				}
			}

		case OpArraySelect:
			w := sel.Args[0]
			rewriteSelect(v, w, offset+sel.Type.Size()*sel.AuxInt, t)

		case OpInt64Hi:
			w := sel.Args[0]
			ls := rewriteSelect(v, w, offset+hiOffset, t)
			locs = splitSlots(ls, ".ptr", hiOffset, t)

		case OpInt64Lo:
			w := sel.Args[0]
			ls := rewriteSelect(v, w, offset+lowOffset, t)
			locs = splitSlots(ls, ".ptr", lowOffset, t)

		case OpStringPtr:
			ls := rewriteSelect(v, sel.Args[0], offset, t)
			locs = splitSlots(ls, ".ptr", 0, typ.BytePtr)
			//for i := range ls {
			//	locs = append(locs, f.fe.SplitSlot(&ls[i], ".ptr", 0, typ.BytePtr))
			//}
		case OpSlicePtr:
			w := sel.Args[0]
			ls := rewriteSelect(v, w, offset, t)
			locs = splitSlots(ls, ".ptr", 0, types.NewPtr(w.Type.Elem()))

		case OpITab:
			w := sel.Args[0]
			ls := rewriteSelect(v, w, offset, t)
			sfx := ".itab"
			if w.Type.IsEmptyInterface() {
				sfx = ".type"
			}
			locs = splitSlots(ls, sfx, 0, typ.Uintptr)

		case OpComplexReal:
			ls := rewriteSelect(v, sel.Args[0], offset, t)
			locs = splitSlots(ls, ".real", 0, t)

		case OpStringLen, OpSliceLen:
			ls := rewriteSelect(v, sel.Args[0], offset+ptrSize, t)
			locs = splitSlots(ls, ".len", ptrSize, t)

		case OpIData:
			ls := rewriteSelect(v, sel.Args[0], offset+ptrSize, t)
			locs = splitSlots(ls, ".data", ptrSize, t)

		case OpSliceCap:
			ls := rewriteSelect(v, sel.Args[0], offset+2*ptrSize, t)
			locs = splitSlots(ls, ".cap", 2*ptrSize, t)

		case OpComplexImag:
			ls := rewriteSelect(v, sel.Args[0], offset+t.Width, t) // result is FloatNN, width of result is offset of imaginary part.
			locs = splitSlots(ls, ".image", t.Width, t)

		case OpCopy: // If it's an intermediate result, recurse
			locs = rewriteSelect(v, sel.Args[0], offset, t)
			for _, s := range namedSelects[sel] {
				// this copy may have had its own name, preserve that, too.
				locs = append(locs, f.Names[s.locIndex])
			}

		default:
			// panic(fmt.Sprintf("Did not expect to arrive here, sel = %v", sel))
		}

		return locs
	}

	// storeArg converts stores of SSA-able aggregate arguments (passed to a call) into a series of stores of
	// smaller types into individual parameter slots.
	var storeArg func(pos src.XPos, b *Block, a *Value, t *types.Type, offset int64, mem *Value) *Value
	storeArg = func(pos src.XPos, b *Block, a *Value, t *types.Type, offset int64, mem *Value) *Value {
		if debug {
			fmt.Printf("\tstoreArg(%s;  %s;  %v;  %d;  %s)\n", b, a.LongString(), t, offset, mem.String())
		}

		switch a.Op {
		case OpArrayMake0, OpStructMake0:
			return mem

		case OpStructMake1, OpStructMake2, OpStructMake3, OpStructMake4:
			for i := 0; i < t.NumFields(); i++ {
				fld := t.Field(i)
				mem = storeArg(pos, b, a.Args[i], fld.Type, offset+fld.Offset, mem)
			}
			return mem

		case OpArrayMake1:
			return storeArg(pos, b, a.Args[0], t.Elem(), offset, mem)

		case OpInt64Make:
			tHi, tLo := intPairTypes(t.Etype)
			mem = storeArg(pos, b, a.Args[0], tHi, offset+hiOffset, mem)
			return storeArg(pos, b, a.Args[1], tLo, offset+lowOffset, mem)

		case OpComplexMake:
			tPart := typ.Float32
			wPart := t.Width / 2
			if wPart == 8 {
				tPart = typ.Float64
			}
			mem = storeArg(pos, b, a.Args[0], tPart, offset, mem)
			return storeArg(pos, b, a.Args[1], tPart, offset+wPart, mem)

		case OpIMake:
			mem = storeArg(pos, b, a.Args[0], typ.Uintptr, offset, mem)
			return storeArg(pos, b, a.Args[1], typ.BytePtr, offset+ptrSize, mem)

		case OpStringMake:
			mem = storeArg(pos, b, a.Args[0], typ.BytePtr, offset, mem)
			return storeArg(pos, b, a.Args[1], typ.Int, offset+ptrSize, mem)

		case OpSliceMake:
			mem = storeArg(pos, b, a.Args[0], typ.BytePtr, offset, mem)
			mem = storeArg(pos, b, a.Args[1], typ.Int, offset+ptrSize, mem)
			return storeArg(pos, b, a.Args[2], typ.Int, offset+2*ptrSize, mem)
		}

		dst := offsetFrom(sp, offset, types.NewPtr(t))
		x := b.NewValue3A(pos, OpStore, types.TypeMem, t, dst, a, mem)
		if debug {
			fmt.Printf("\t\tstoreArg returns %s\n", x.LongString())
		}
		return x
	}

	// splitStore converts a store of an SSA-able aggregate into a series of smaller stores, emitting
	// appropriate Struct/Array Select operations (which will soon go dead) to obtain the parts.
	// This has to handle aggregate types that have already been lowered by an earlier phase.
	var splitStore func(dest, source, mem, v *Value, t *types.Type, offset int64, firstStorePos src.XPos) *Value
	splitStore = func(dest, source, mem, v *Value, t *types.Type, offset int64, firstStorePos src.XPos) *Value {
		if debug {
			fmt.Printf("\tsplitStore(%s;  %s;  %s;  %s;  %v;  %d;  %v)\n", dest.LongString(), source.LongString(), mem.String(), v.LongString(), t, offset, firstStorePos)
		}
		pos := v.Pos.WithNotStmt()
		switch t.Etype {
		case types.TARRAY:
			elt := t.Elem()
			// TODO is the commented code necessary? Test by stuffing struct {[1]uintptr} into an interface.
			//if elt.Width == t.Width && t.Width <= regSize {
			//	// handle array version of (StructSelect [0] (IData x)) => (IData x)
			//	return splitStore(dest, source, mem, v, elt, offset, firstStorePos)
			//}
			for i := int64(0); i < t.NumElem(); i++ {
				sel := source.Block.NewValue1I(pos, OpArraySelect, elt, i, source)
				mem = splitStore(dest, sel, mem, v, elt, offset+i*elt.Width, firstStorePos)
				firstStorePos = firstStorePos.WithNotStmt()
			}
			return mem

		case types.TSTRUCT:
			if t.NumFields() == 1 && t.Field(0).Type.Width == t.Width && t.Width <= regSize {
				// handle (StructSelect [0] (IData x)) => (IData x)
				return splitStore(dest, source, mem, v, t.Field(0).Type, offset, firstStorePos)
			}
			for i := 0; i < t.NumFields(); i++ {
				fld := t.Field(i)
				sel := source.Block.NewValue1I(pos, OpStructSelect, fld.Type, int64(i), source)
				mem = splitStore(dest, sel, mem, v, fld.Type, offset+fld.Offset, firstStorePos)
				firstStorePos = firstStorePos.WithNotStmt()
			}
			return mem

		case types.TINT64, types.TUINT64:
			if t.Width == regSize {
				break
			}
			tHi, tLo := intPairTypes(t.Etype)
			sel := source.Block.NewValue1(pos, OpInt64Hi, tHi, source)
			mem = splitStore(dest, sel, mem, v, tHi, offset+hiOffset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpInt64Lo, tLo, source)
			return splitStore(dest, sel, mem, v, tLo, offset+lowOffset, firstStorePos)

		case types.TINTER:
			sel := source.Block.NewValue1(pos, OpITab, typ.BytePtr, source)
			mem = splitStore(dest, sel, mem, v, typ.BytePtr, offset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpIData, typ.BytePtr, source)
			return splitStore(dest, sel, mem, v, typ.BytePtr, offset+ptrSize, firstStorePos)

		case types.TSTRING:
			sel := source.Block.NewValue1(pos, OpStringPtr, typ.BytePtr, source)
			mem = splitStore(dest, sel, mem, v, typ.BytePtr, offset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpStringLen, typ.Int, source)
			return splitStore(dest, sel, mem, v, typ.Int, offset+ptrSize, firstStorePos)

		case types.TSLICE:
			et := types.NewPtr(t.Elem())
			sel := source.Block.NewValue1(pos, OpSlicePtr, et, source)
			mem = splitStore(dest, sel, mem, v, et, offset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpSliceLen, typ.Int, source)
			mem = splitStore(dest, sel, mem, v, typ.Int, offset+ptrSize, firstStorePos)
			sel = source.Block.NewValue1(pos, OpSliceCap, typ.Int, source)
			return splitStore(dest, sel, mem, v, typ.Int, offset+2*ptrSize, firstStorePos)

		case types.TCOMPLEX64:
			sel := source.Block.NewValue1(pos, OpComplexReal, typ.Float32, source)
			mem = splitStore(dest, sel, mem, v, typ.Float32, offset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpComplexImag, typ.Float32, source)
			return splitStore(dest, sel, mem, v, typ.Float32, offset+4, firstStorePos)

		case types.TCOMPLEX128:
			sel := source.Block.NewValue1(pos, OpComplexReal, typ.Float64, source)
			mem = splitStore(dest, sel, mem, v, typ.Float64, offset, firstStorePos)
			firstStorePos = firstStorePos.WithNotStmt()
			sel = source.Block.NewValue1(pos, OpComplexImag, typ.Float64, source)
			return splitStore(dest, sel, mem, v, typ.Float64, offset+8, firstStorePos)
		}
		// Default, including for aggregates whose single element exactly fills their container
		// TODO this will be a problem for cast interfaces containing floats when we move to registers.
		x := v.Block.NewValue3A(firstStorePos, OpStore, types.TypeMem, t, offsetFrom(dest, offset, types.NewPtr(t)), source, mem)
		if debug {
			fmt.Printf("\t\tsplitStore returns %s\n", x.LongString())
		}

		return x
	}

	// rewriteArgs removes all the Args from a call and converts the call args into appropriate
	// stores (or later, register movement).  Extra args for interface and closure calls are ignored,
	// but removed.
	rewriteArgs := func(v *Value, firstArg int) *Value {
		// Thread the stores on the memory arg
		aux := v.Aux.(*AuxCall)
		pos := v.Pos.WithNotStmt()
		m0 := v.Args[len(v.Args)-1]
		mem := m0
		for i, a := range v.Args {
			if i < firstArg {
				continue
			}
			if a == m0 { // mem is last.
				break
			}
			auxI := int64(i - firstArg)
			if a.Op == OpDereference {
				// "Dereference" of addressed (probably not-SSA-eligible) value becomes Move
				// TODO this will be more complicated with registers in the picture.
				source := a.Args[0]
				dst := f.ConstOffPtrSP(source.Type, aux.OffsetOfArg(auxI), sp)
				if a.Uses == 1 {
					a.reset(OpMove)
					a.Pos = pos
					a.Type = types.TypeMem
					a.Aux = aux.TypeOfArg(auxI)
					a.AuxInt = aux.SizeOfArg(auxI)
					a.SetArgs3(dst, source, mem)
					mem = a
				} else {
					mem = a.Block.NewValue3A(pos, OpMove, types.TypeMem, aux.TypeOfArg(auxI), dst, source, mem)
					mem.AuxInt = aux.SizeOfArg(auxI)
				}
			} else {
				if debug {
					fmt.Printf("storeArg %s, %v, %d\n", a.LongString(), aux.TypeOfArg(auxI), aux.OffsetOfArg(auxI))
				}
				mem = storeArg(pos, v.Block, a, aux.TypeOfArg(auxI), aux.OffsetOfArg(auxI), mem)
			}
		}
		v.resetArgs()
		return mem
	}

	// TODO if too slow, whole program iteration can be replaced w/ slices of appropriate values, accumulated in first loop here.

	// Step 0: rewrite the calls to convert incoming args to stores.
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			switch v.Op {
			case OpStaticLECall:
				mem := rewriteArgs(v, 0)
				v.SetArgs1(mem)
			case OpClosureLECall:
				code := v.Args[0]
				context := v.Args[1]
				mem := rewriteArgs(v, 2)
				v.SetArgs3(code, context, mem)
			case OpInterLECall:
				code := v.Args[0]
				mem := rewriteArgs(v, 1)
				v.SetArgs2(code, mem)
			}
		}
	}

	for i, name := range f.Names {
		t := name.Type
		if isAlreadyExpandedAggregateType(t) {
			for j, v := range f.NamedValues[name] {
				if v.Op == OpSelectN {
					ns := namedSelects[v]
					namedSelects[v] = append(ns, namedVal{locIndex: i, valIndex: j})
				}
			}
		}
	}

	// Step 1: any stores of aggregates remaining are believed to be sourced from call results.
	// Decompose those stores into a series of smaller stores, adding selection ops as necessary.
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op == OpStore {
				t := v.Aux.(*types.Type)
				iAEATt := isAlreadyExpandedAggregateType(t)
				if !iAEATt {
					// guarding against store immediate struct into interface data field -- store type is *uint8
					// TODO can this happen recursively?
					tSrc := v.Args[1].Type
					iAEATt = isAlreadyExpandedAggregateType(tSrc)
					if iAEATt {
						t = tSrc
					}
				}
				if iAEATt {
					if debug {
						fmt.Printf("Splitting store %s\n", v.LongString())
					}
					dst, source, mem := v.Args[0], v.Args[1], v.Args[2]
					mem = splitStore(dst, source, mem, v, t, 0, v.Pos)
					v.copyOf(mem)
				}
			}
		}
	}

	val2Preds := make(map[*Value]int32) // Used to accumulate dependency graph of selection operations for topological oprdering.

	// Step 2: transform or accumulate selection operations for rewrite in topological order.
	// Aggregate types that have already (in earlier phases) been transformed must be lowered
	// comprehensively to finish the transformation (this would be user-defined structs and arrays),
	// those that have yet to be transformed (strings, slices, interface, complex, and 64-bit integers in 32-bit registers)
	// can be transformed locally.
	//
	// Any select-for-addressing applied to call results can be transformed directly.
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			// Accumulate chains of selectors for processing in topological order
			switch v.Op {
			case OpStructSelect, OpArraySelect,
				OpIData, OpITab,
				OpStringPtr, OpStringLen,
				OpSlicePtr, OpSliceLen, OpSliceCap,
				OpComplexReal, OpComplexImag,
				OpInt64Hi, OpInt64Lo:
				w := v.Args[0]
				switch w.Op {
				case OpStructSelect, OpArraySelect, OpSelectN:
					val2Preds[w] += 1
					if debug {
						fmt.Printf("v2p[%s] = %d\n", w.LongString(), val2Preds[w])
					}
				}
				if _, ok := val2Preds[v]; !ok {
					val2Preds[v] = 0
					if debug {
						fmt.Printf("v2p[%s] = %d\n", v.LongString(), val2Preds[v])
					}
				}

			case OpSelectN:
				if _, ok := val2Preds[v]; !ok {
					val2Preds[v] = 0
					if debug {
						fmt.Printf("v2p[%s] = %d\n", v.LongString(), val2Preds[v])
					}
				}

			case OpSelectNAddr:
				// Do these directly, there are no chains of selectors.
				call := v.Args[0]
				which := v.AuxInt
				aux := call.Aux.(*AuxCall)
				pt := v.Type
				off := offsetFrom(sp, aux.OffsetOfResult(which), pt)
				v.copyOf(off)
			}
		}
	}

	// Step 3: Compute topological order of selectors,
	// then process it in reverse to eliminate duplicates,
	// then forwards to rewrite selectors.
	//
	// All chains of selectors end up in same block as the call.
	sdom := f.Sdom()

	// Compilation must be deterministic, so sort after extracting first zeroes from map.
	// Sorting allows dominators-last order within each batch,
	// so that the backwards scan for duplicates will most often find copies from dominating blocks (it is best-effort).
	var toProcess []*Value
	less := func(i, j int) bool {
		vi, vj := toProcess[i], toProcess[j]
		bi, bj := vi.Block, vj.Block
		if bi == bj {
			return vi.ID < vj.ID
		}
		return sdom.domorder(bi) > sdom.domorder(bj) // reverse the order to put dominators last.
	}

	// Accumulate order in allOrdered
	var allOrdered []*Value
	for v, n := range val2Preds {
		if n == 0 {
			allOrdered = append(allOrdered, v)
		}
	}
	last := 0 // allOrdered[0:last] has been top-sorted and processed
	for len(val2Preds) > 0 {
		toProcess = allOrdered[last:]
		last = len(allOrdered)
		sort.Slice(toProcess, less)
		for _, v := range toProcess {
			w := v.Args[0]
			delete(val2Preds, v)
			n, ok := val2Preds[w]
			if !ok {
				continue
			}
			if n == 1 {
				allOrdered = append(allOrdered, w)
				delete(val2Preds, w)
				continue
			}
			val2Preds[w] = n - 1
		}
	}

	common := make(map[selKey]*Value)
	// Rewrite duplicate selectors as copies where possible.
	for i := len(allOrdered) - 1; i >= 0; i-- {
		v := allOrdered[i]
		w := v.Args[0]
		for w.Op == OpCopy {
			w = w.Args[0]
		}
		typ := v.Type
		if typ.IsMemory() {
			continue // handled elsewhere, not an indexable result
		}
		size := typ.Width
		offset := int64(0)
		switch v.Op {
		case OpStructSelect:
			if w.Type.Etype == types.TSTRUCT {
				offset = w.Type.FieldOff(int(v.AuxInt))
			} else { // Immediate interface data artifact, offset is zero.
				fmt.Printf("Func %s, v=%s, w=%s\n", f.Name, v.LongString(), w.LongString())
			}
		case OpArraySelect:
			offset = size * v.AuxInt
		case OpSelectN:
			offset = w.Aux.(*AuxCall).OffsetOfResult(v.AuxInt)
		case OpInt64Hi:
			offset = hiOffset
		case OpInt64Lo:
			offset = lowOffset
		case OpStringLen, OpSliceLen, OpIData:
			offset = ptrSize
		case OpSliceCap:
			offset = 2 * ptrSize
		case OpComplexImag:
			offset = size
		}
		sk := selKey{from: w, size: size, offset: offset, typ: typ.Etype}
		dupe := common[sk]
		if dupe == nil {
			common[sk] = v
		} else {
			if sdom.IsAncestorEq(dupe.Block, v.Block) {
				v.copyOf(dupe)
			} else {
				if f.pass.debug > 0 {
					fmt.Printf("%s is non-dominated copy of %s\n", v.LongString(), dupe.LongString())
				}
			}
		}
	}

	// Indices of entries in f.Names that need to be deleted.
	var toDelete []namedVal

	// Rewrite selectors.
	for i, v := range allOrdered {
		if debug {
			b := v.Block
			fmt.Printf("allOrdered[%d] = b%d, %s, uses=%d\n", i, b.ID, v.LongString(), v.Uses)
		}
		if v.Uses == 0 {
			v.reset(OpInvalid)
			continue
		}
		if v.Op == OpCopy {
			continue
		}
		locs := rewriteSelect(v, v, 0, v.Type)
		// Install new names.
		// TODO will any of the intermediate selectors have names?
		if v.Type.IsMemory() {
			continue
		}
		// Leaf types may have debug locations
		if !isAlreadyExpandedAggregateType(v.Type) {
			for _, l := range locs {
				f.NamedValues[l] = append(f.NamedValues[l], v)
			}
			f.Names = append(f.Names, locs...)
			continue
		}
		// Not-leaf types that had debug locations need to lose them.
		if ns, ok := namedSelects[v]; ok {
			toDelete = append(toDelete, ns...)
		}
	}

	deleteNamedVals(f, toDelete)

	// Step 4: rewrite the calls themselves, correcting the type
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			switch v.Op {
			case OpStaticLECall:
				v.Op = OpStaticCall
				v.Type = types.TypeMem
			case OpClosureLECall:
				v.Op = OpClosureCall
				v.Type = types.TypeMem
			case OpInterLECall:
				v.Op = OpInterCall
				v.Type = types.TypeMem
			}
		}
	}

	// Step 5: elide any copies introduced.
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			for i, a := range v.Args {
				if a.Op != OpCopy {
					continue
				}
				aa := copySource(a)
				v.SetArg(i, aa)
				for a.Uses == 0 {
					b := a.Args[0]
					a.reset(OpInvalid)
					a = b
				}
			}
		}
	}
}
