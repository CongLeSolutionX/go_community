// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "cmd/compile/internal/ir"

// recursiveTailCallElimination turns tail recursive calls to self into loops.
func recursiveTailCallElimination(f *Func) {
	var loopEntry *Block
	// Values may be nil if the argument is unused, then just don't assign it.
	var argumentIndexesToPhi []*Value
retLoop:
	for _, b := range f.Blocks {
		if b.Kind != BlockRet {
			continue
		}

		makeResult := b.Controls[0]
		if makeResult.Op != OpMakeResult {
			continue
		}

		var call *Value
		for i, selectN := range makeResult.Args {
			if selectN.Op != OpSelectN {
				continue retLoop
			}

			if selectN.AuxInt != int64(i) {
				if f.pass.debug > 1 {
					f.Warnl(b.Pos, "failed because results are out of order")
				}
				continue retLoop
			}

			if call == nil {
				call = selectN.Args[0]
				if call.Op != OpStaticLECall {
					continue retLoop
				}

				caux := call.Aux.(*AuxCall)
				faux := f.OwnAux
				if caux.Fn.Name != faux.Fn.Name {
					if f.pass.debug > 1 {
						f.Warnl(b.Pos, "failed because last call does not match current function; got %q; expected %q", caux.Fn.Name, faux.Fn.Name)
					}
					continue retLoop
				}
				if caux.reg == nil || faux.reg == nil ||
					caux.reg.clobbers != faux.reg.clobbers ||
					!slicesEqualInputInfo(caux.reg.inputs, faux.reg.inputs) ||
					!slicesEqualOutputInfo(caux.reg.outputs, faux.reg.outputs) {
					if f.pass.debug > 1 {
						f.Warnl(b.Pos, "failed because registers do not match")
					}
					continue retLoop
				}
				if caux.abiInfo.InRegistersUsed() != faux.abiInfo.InRegistersUsed() ||
					caux.abiInfo.OutRegistersUsed() != faux.abiInfo.OutRegistersUsed() ||
					caux.abiInfo.SpillAreaOffset() != faux.abiInfo.SpillAreaOffset() ||
					caux.abiInfo.SpillAreaSize() != faux.abiInfo.SpillAreaSize() {
					if f.pass.debug > 1 {
						f.Warnl(b.Pos, "failed because abis do not match")
					}
					continue retLoop
				}
			} else if selectN.Args[0] != call {
				if f.pass.debug > 1 {
					f.Warnl(b.Pos, "failed because last actions is not all the same call")
				}
				continue retLoop
			}
		}

		// we need to filter arguments that may have pointers to stack due to edge
		// cases like this:
		/*
			type Frame struct{
				next *Frame
				v uint
			}

			func StackAllocatedLinkedList(v *Frame, i uint) {
				if i != 0 {
				StackAllocatedLinkedList(&Frame{v, i}, i - 1) // do stack allocation
					return
				}
				DoStuff(v) // does not escape v
			}
		*/
		for _, arg := range call.Args {
			if mayHavePointerToStack(arg) {
				if f.pass.debug > 1 {
					f.Warnl(call.Pos, "failed because arguments may point to stack")
				}
				continue retLoop
			}
		}

		// At this point we have a block that is returning the result of the
		// function we are currently compiling. Let's turn this into a loop.
		if loopEntry == nil {
			// lazy build the loopEntry
			argumentIndexesToPhi = make([]*Value, len(call.Args))
			loopEntry = f.Entry
			newEntry := f.NewBlock(BlockPlain)
			f.Entry = newEntry
			newEntry.AddEdgeTo(loopEntry)

			// move and resetups init values
			toScan := loopEntry.Values
			loopEntry.Values = loopEntry.Values[:0] // reuse capacity in-place
			for _, v := range toScan {
				switch v.Op {
				case OpSP, OpSB,
					OpConst8, OpConst16, OpConst32, OpConst64:
					// Move thoses to the newEntry
					v.Block = newEntry
					newEntry.Values = append(newEntry.Values, v)

				case OpInitMem:
					// For InitMem and Args create a new equivalent Op in newEntry
					// and reset to be a phi node (thoses will be phying all the different tces)
					initMem := f.newValueCopyingValueInto(v, newEntry)
					argumentIndexesToPhi[len(argumentIndexesToPhi)-1] = v // last argument is always mem
					v.reset(OpPhi)
					v.AddArg(initMem)
					loopEntry.Values = append(loopEntry.Values, v)

				case OpArg:
					newArg := f.newValueCopyingValueInto(v, newEntry)

					name := v.Aux.(*ir.Name)
					var idx int
					for i, arg := range f.OwnAux.abiInfo.InParams() {
						if arg.Name != name {
							continue
						}
						idx = i
						goto found
					}
					panic("did not found argument within self abi")

				found:
					argumentIndexesToPhi[idx] = v
					v.reset(OpPhi)
					v.AddArg(newArg)
					fallthrough
				default:
					// for anything compuation related, leave inside the loop as it may depend on arguments
					loopEntry.Values = append(loopEntry.Values, v)
				}
			}

			// zero unused space in loopEntry.Values to avoid keeping alive values uselessly
			toScan = toScan[len(loopEntry.Values):]
			for i := range toScan {
				toScan[i] = nil
			}
		}

		if f.pass.debug > 0 {
			f.Warnl(call.Pos, "loopified recursive tail call")
		}

		b.Kind = BlockPlain
		b.Controls[0].Uses--
		b.Controls[0] = nil
		b.AddEdgeTo(loopEntry)
		for i, arg := range call.Args {
			phi := argumentIndexesToPhi[i]
			if phi == nil {
				// this is an unused argument, skip
				continue
			}

			phi.AddArg(arg)
		}

		// Don't keep alive two stores. This is safe to do because whatever may be
		// using results from this call is dead anyway since return directly map to
		// the call else we would have skipped this ret block.
		for _, selectN := range makeResult.Args {
			selectN.reset(OpInvalid)
		}
		makeResult.reset(OpInvalid)
		call.reset(OpInvalid)
	}
}

func mayHavePointerToStack(v *Value) bool {
	// TODO: we could maybe replace this with a proper analysis that allows to transform heap values.
	typ := v.Type
	return typ.IsPtrShaped() ||
		typ.IsSlice() ||
		typ.IsArray() ||
		typ.IsString() ||
		typ.IsStruct()
}

func slicesEqualInputInfo(s1, s2 []inputInfo) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func slicesEqualOutputInfo(s1, s2 []outputInfo) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
