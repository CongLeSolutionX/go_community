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
	debug := f.pass.debug > 1

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

	// rewriteSelect recursively walks through a chain of Struct/Array Select
	// operations until a select from call results is reached.  It emits the
	// code necessary to implement the leaf select operation that leads to the call.
	var rewriteSelect func(v *Value, sel *Value, offset int64, t *types.Type)
	rewriteSelect = func(v *Value, sel *Value, offset int64, t *types.Type) {
		switch sel.Op {
		case OpSelectN:
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
			if w.Type.Etype != types.TSTRUCT {
				fmt.Printf("Bad type for w: v=%v; sel=%v; w=%v; ,f=%s\n", v.LongString(), sel.LongString(), w.LongString(), f.Name)
				// Artifact of immediate interface idata
				rewriteSelect(v, w, offset, t)
			} else {
				rewriteSelect(v, w, offset+w.Type.FieldOff(int(sel.AuxInt)), t)
			}

		case OpArraySelect:
			w := sel.Args[0]
			rewriteSelect(v, w, offset+sel.Type.Size()*sel.AuxInt, t)

		case OpCopy: // If it's an intermediate result, recurse
			rewriteSelect(v, sel.Args[0], offset, t)

		default:
			// panic(fmt.Sprintf("Did not expect to arrive here, sel = %v", sel))
		}
	}

	// storeArg converts stores of SSA-able aggregate arguments (passed to a call) into a series of stores of
	// smaller types into individual parameter slots.
	var storeArg func(pos src.XPos, b *Block, a *Value, t *types.Type, offset int64, mem *Value) *Value
	storeArg = func(pos src.XPos, b *Block, a *Value, t *types.Type, offset int64, mem *Value) *Value {
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
			fmt.Printf("storeArg(%v) returns %s\n", a, x.LongString())
		}
		return x
	}

	// splitStore converts a store of an SSA-able aggregate into a series of smaller stores, emitting
	// appropriate Struct/Array Select operations (which will soon go dead) to obtain the parts.
	// This has to handle aggregate types that have already been lowered by an earlier phase.
	var splitStore func(dest, source, mem, v *Value, t *types.Type, offset int64, firstStorePos src.XPos) *Value
	splitStore = func(dest, source, mem, v *Value, t *types.Type, offset int64, firstStorePos src.XPos) *Value {
		pos := v.Pos.WithNotStmt()
		switch t.Etype {
		case types.TARRAY:
			elt := t.Elem()
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
		}
		// Default, including for aggregates whose single element exactly fills their container
		// TODO this will be a problem for cast interfaces containing floats when we move to registers.
		x := v.Block.NewValue3A(firstStorePos, OpStore, types.TypeMem, t, offsetFrom(dest, offset, types.NewPtr(t)), source, mem)
		if debug {
			fmt.Printf("splitStore(%v, %v, %v, %v) returns %s\n", dest, source, mem, v, x.LongString())
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
				src := a.Args[0]
				dst := f.ConstOffPtrSP(src.Type, aux.OffsetOfArg(auxI), sp)
				if a.Uses == 1 {
					a.reset(OpMove)
					a.Pos = pos
					a.Type = types.TypeMem
					a.Aux = aux.TypeOfArg(auxI)
					a.AuxInt = aux.SizeOfArg(auxI)
					a.SetArgs3(dst, src, mem)
					mem = a
				} else {
					mem = a.Block.NewValue3A(pos, OpMove, types.TypeMem, aux.TypeOfArg(auxI), dst, src, mem)
					mem.AuxInt = aux.SizeOfArg(auxI)
				}
			} else {
				mem = storeArg(pos, v.Block, a, aux.TypeOfArg(auxI), aux.OffsetOfArg(auxI), mem)
			}
		}
		v.resetArgs()
		return mem
	}

	// splitSelectInPlace replaces a select of an aggregate type that has two parts,
	// with a FooMake of the loads of those two parts.
	splitSelectInPlace := func(v *Value, combine Op, t1, t2 *types.Type, off1, off2 int64) {
		call := v.Args[0] // will become a mem operand, later
		which := v.AuxInt
		aux := call.Aux.(*AuxCall)
		t := v.Type
		pos := v.Pos
		offset := aux.OffsetOfResult(which)
		first := offsetFrom(sp, offset+off1, types.NewPtr(t1))
		second := offsetFrom(sp, offset+off2, types.NewPtr(t2))
		v.reset(combine)
		v.Pos = pos
		v.SetArgs2(call.Block.NewValue2(v.Pos, OpLoad, t1, first, call),
			call.Block.NewValue2(v.Pos, OpLoad, t2, second, call))
		v.Type = t
	}

	splitSelectInPlace3 := func(v *Value, combine Op, t1, t2, t3 *types.Type, off1, off2, off3 int64) {
		call := v.Args[0] // will become a mem operand, later
		which := v.AuxInt
		aux := call.Aux.(*AuxCall)
		t := v.Type
		pos := v.Pos
		offset := aux.OffsetOfResult(which)
		first := offsetFrom(sp, offset+off1, types.NewPtr(t1))
		second := offsetFrom(sp, offset+off2, types.NewPtr(t2))
		third := offsetFrom(sp, offset+off3, types.NewPtr(t3))
		v.reset(combine)
		v.Pos = pos
		v.SetArgs3(call.Block.NewValue2(v.Pos, OpLoad, t1, first, call),
			call.Block.NewValue2(v.Pos, OpLoad, t2, second, call),
			call.Block.NewValue2(v.Pos, OpLoad, t3, third, call))
		v.Type = t
	}

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

	// Step 1: any stores of aggregates remaining are believed to be sourced from call results.
	// Decompose those stores into a series of smaller stores, adding selection ops as necessary.
	for _, b := range f.Blocks {
		for _, v := range b.Values {
			if v.Op == OpStore {
				t := v.Aux.(*types.Type)
				iAEATt := isAlreadyExpandedAggregateType(t)
				if !iAEATt {
					// guarding against store immediate struct into interface data field -- store type is *uint8
					tSrc := v.Args[1].Type
					iAEATt = isAlreadyExpandedAggregateType(tSrc)
					if iAEATt {
						t = tSrc
					}
				}
				if iAEATt {
					dst, src, mem := v.Args[0], v.Args[1], v.Args[2]
					mem = splitStore(dst, src, mem, v, t, 0, v.Pos)
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
			case OpStructSelect, OpArraySelect:
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
				switch v.Type.Etype {
				// Types that have not yet been decomposed, simply split and remake in place, from pieces.
				case types.TINTER:
					splitSelectInPlace(v, OpIMake, typ.Uintptr, typ.BytePtr, 0, ptrSize)
					continue
				case types.TCOMPLEX64:
					splitSelectInPlace(v, OpComplexMake, typ.Float32, typ.Float32, 0, 4)
					continue
				case types.TCOMPLEX128:
					splitSelectInPlace(v, OpComplexMake, typ.Float64, typ.Float64, 0, 8)
					continue

				case types.TSTRING: // TODO
					splitSelectInPlace(v, OpStringMake, typ.BytePtr, typ.Int, 0, ptrSize)
					continue

				case types.TINT64, types.TUINT64: // for 32-bit
					t := v.Type
					if t.Width != regSize { // split into a pair of loads.
						tHi, tLo := intPairTypes(t.Etype)
						splitSelectInPlace(v, OpInt64Make, tHi, tLo, hiOffset, lowOffset)
						continue
					}

				case types.TSLICE: // TODO
					splitSelectInPlace3(v, OpSliceMake, typ.BytePtr, typ.Int, typ.Int, 0, ptrSize, 2*ptrSize)
					continue

				}

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
		rewriteSelect(v, v, 0, v.Type)
	}

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
