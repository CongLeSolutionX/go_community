// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import (
	"cmd/internal/src"
	"fmt"
	"sort"
)

func hoistloopiv(f *Func) {
	nest := f.loopnest()
	nest.findBlocks()
	loops := nest.loops

	if nest.hasIrreducible {
		// TODO need to get this right for individual loops
		return
	}
	if len(loops) == 0 {
		return
	}

	// Reuse working storage.
	ss := newSparseSet(f.NumValues())

	var byID sortByCostAndID

	for _, l := range loops {
		if !l.isInner {
			continue
		}
		iv := nest.loopinvariants(l, ss)
		filteredIV := make([]*Value, 0, len(iv))
		filteredIVset := make(map[*Value]uint8)
		for v, c := range iv {
			filteredIV = append(filteredIV, v)
			filteredIVset[v] = c
		}

		rootCount := len(filteredIV)
		if rootCount == 0 {
			continue
		}

		byID.a = filteredIV // Values must be sorted for repeatable order between compilations.
		byID.m = filteredIVset
		sort.Sort(byID)

		const arbitraryLimit = 4
		if len(filteredIV) > arbitraryLimit {
			for i := arbitraryLimit; i < len(filteredIV); i++ {
				filteredIVset[filteredIV[i]] = 0
			}
			filteredIV = filteredIV[:arbitraryLimit]
		}

		var parent *Block // Destination of invariant value hoisting
		for _, p := range l.header.Preds {
			if nest.b2l[p.Block().ID] != l {
				parent = p.Block()
				break
			}
		}
		if parent == nil { // Can only happen if entry block contains nothing; unlikely, but cannot prove it won't happen.
			continue // Since it is so unlikely (not observed in testing), don't sweat the lost optimization.
		}

		if f.pass.debug == 1 {
			for _, v := range filteredIV {
				f.Warnl(v.Pos, "Inner loop root invariant cost %d, header %s, val %s, func %s", iv[v], l.header, v.LongString(), f.Name)
			}
		}

		// TODO flags are not hoisted because they cannot be separated from their uses.
		// Could speculatively hoist, if all uses are also hoisted then retain the hoist,
		// otherwise remove flags and all dependent "invariant" expressions.
		// This matters a lot more with 64-bit arithmetic on 32-bit machines.

		// Some values are not worth hoisting by themselves, but need to be hoisted
		// if other worthy values depend on them.
		change := true
		for change {
			change = false
			for _, v := range filteredIV {
				for _, a := range v.Args {
					if iv[a] > 0 && 0 == filteredIVset[a] {
						filteredIV = append(filteredIV, a)
						filteredIVset[a] = 1 // arbitrary non-zero "cost"
						change = true
					}
				}
			}
		}

		if f.pass.debug > 1 {
			for i, v := range filteredIV {
				what := "root"
				if i >= rootCount {
					what = "input"
				}
				f.Warnl(v.Pos, "Inner loop %s invariant cost %d, header %s, val %s, func %s", what, iv[v], l.header, v.LongString(), f.Name)
			}
		}

		// Move all invariants into parent
		for _, v := range filteredIV {
			b := v.Block
			v.Pos = src.NoXPos // Hoisting is guaranteed to degrade debugging experience.
			// TODO this is algorithmically ugly, fortunately problem size is small
			for i, vv := range b.Values {
				if v != vv {
					continue
				}
				parent.Values = append(parent.Values, v)
				v.Block = parent
				last := len(b.Values) - 1
				b.Values[i] = b.Values[last]
				b.Values[last] = nil
				b.Values = b.Values[:last]
				break
			}
		}
	}
}

// articulationBlocks returns, FOR AN INNER REDUCIBLE LOOP,
// those blocks within the loop that must execute if the loop iterates,
// in order of their execution.
func (nest *loopnest) articulationBlocks(l *loop) []*Block {
	// An articulation block must dominate all the backedges in the loop.
	head := l.header
	backedges := []*Block{}
	sdom := nest.f.Sdom()
	artBlocks := []*Block{}
	for _, p := range head.Preds {
		b := p.Block()
		if sdom.IsAncestorEq(head, b) {
			backedges = append(backedges, b)
		}
	}
blockloop:
	for i := len(l.blocks) - 1; i >= 0; i-- {
		b := l.blocks[i]
		for _, be := range backedges {
			if !sdom.IsAncestorEq(b, be) {
				continue blockloop
			}
		}
		artBlocks = append(artBlocks, b)
	}
	return artBlocks
}

// Highest costs come first
type sortByCostAndID struct {
	a []*Value
	m map[*Value]uint8
}

func (sv sortByCostAndID) Len() int      { return len(sv.a) }
func (sv sortByCostAndID) Swap(i, j int) { sv.a[i], sv.a[j] = sv.a[j], sv.a[i] }
func (sv sortByCostAndID) Less(i, j int) bool {
	vi, vj := sv.a[i], sv.a[j]
	ci, cj := sv.m[vi], sv.m[vj]
	return ci > cj || ci == cj && vi.ID < vj.ID
}

// loopinvariants returns a map from *Value to greater-than-zero score for loop-invariant
// values.  l is the loop in question, vv is a scratch sparse set indexed by Value.ID.
func (nest *loopnest) loopinvariants(l *loop, vv *sparseSet) map[*Value]uint8 {
	// TODO loop invariant control flow?
	if !l.isInner {
		// innermost loops only.
		return nil
	}

	filteredblocks := nest.articulationBlocks(l)
	if nest.f.pass.debug > 2 {
		fmt.Printf("Articulation blocks for %s = %v\n", nest.f.Name, filteredblocks)
	}
	vv.clear()
	m := make(map[*Value]uint8)
	for l.loopinvariants1(vv, nest.b2l, filteredblocks, &m) {
	}
	return m
}

func possiblyFaultingLoadAddr(a *Value) bool {
	return a.Op != OpAddr &&
		!((a.Op == OpOffPtr || a.Op == OpPtrIndex || a.Op == OpAddPtr) && a.Args[0].Op == OpAddr)
}

// alwaysVariant returns true if v should be regarded as loop variant or
// otherwise unsuitable for hoisting (for example, if v can fault).
func alwaysVariant(v *Value) bool {
	if opcodeTable[v.Op].hasSideEffects {
		// Most of these also update memory, so within a loop they will be naturally variant.
		return true
	}

	if (opcodeTable[v.Op].faultOnNilArg0 || v.Op == OpLoad) && possiblyFaultingLoadAddr(v.Args[0]) {
		// TODO use inference to be less conservative about possibly faulting.
		return true
	}

	switch v.Op {
	case OpZero, OpMove, OpPhi, OpNilCheck,
		OpClosureCall, OpStaticCall, OpInterCall:
		return true
	}
	// Flags aren't really variant, but they're a pain to hoist
	// and need use checks and copying and aren't common case on 64-bit.
	// So don't hoist right now.
	if v.Type.IsFlags() || v.Type.IsTuple() && v.Type.FieldType(1).IsFlags() {
		return true
	}

	return false
}

func (l *loop) loopinvariants1(ss *sparseSet, b2l []*loop, filteredblocks []*Block, m *map[*Value]uint8) bool {
	change := false
	for _, b := range filteredblocks {
	valloop:
		for _, v := range b.Values {
			if alwaysVariant(v) {
				continue
			}
			cost := 2
			switch v.Op {
			default:
				if opcodeTable[v.Op].rematerializeable {
					cost = 1
				}
			case // TODO check for never-zero divisor (and MININT/-1) and allow some DIV/MOD to be invariant.
				Op386DIVL,
				Op386DIVW,
				Op386DIVLU,
				Op386DIVWU,
				Op386MODL,
				Op386MODW,
				Op386MODLU,
				Op386MODWU,

				OpAMD64DIVQ,
				OpAMD64DIVL,
				OpAMD64DIVW,
				OpAMD64DIVQU,
				OpAMD64DIVLU,
				OpAMD64DIVWU,
				OpAMD64DIVQU2,

				OpARMCALLudiv,

				OpARM64DIV,
				OpARM64UDIV,
				OpARM64DIVW,
				OpARM64UDIVW,
				OpARM64MOD,
				OpARM64UMOD,
				OpARM64MODW,
				OpARM64UMODW,

				OpMIPSDIV,
				OpMIPSDIVU,

				OpMIPS64DIVV,
				OpMIPS64DIVVU,

				OpS390XDIVD,
				OpS390XDIVW,
				OpS390XDIVDU,
				OpS390XDIVWU,
				OpS390XMODD,
				OpS390XMODW,
				OpS390XMODDU,
				OpS390XMODWU,

				OpDiv128u, // Also works for generic ops
				OpDiv8,
				OpDiv8u,
				OpDiv16,
				OpDiv16u,
				OpDiv32,
				OpDiv32u,
				OpDiv64,
				OpDiv64u,
				OpMod8,
				OpMod8u,
				OpMod16,
				OpMod16u,
				OpMod32,
				OpMod32u,
				OpMod64,
				OpMod64u:
				continue // Can't hoist from loop if divisor might be zero or if MININT/-1, because it could fault.

			// Believe that these ops are slightly more expensive and worth hoisting by themselves.
			case Op386MULSS,
				Op386MULSD,
				Op386DIVSS,
				Op386DIVSD,
				Op386MULL,
				Op386HMULL,
				Op386HMULLU,
				Op386MULLQU,

				OpAMD64MULSS,
				OpAMD64MULSD,
				OpAMD64DIVSS,
				OpAMD64DIVSD,
				OpAMD64MULQ,
				OpAMD64MULL,
				OpAMD64HMULQ,
				OpAMD64HMULL,
				OpAMD64HMULQU,
				OpAMD64HMULLU,
				OpAMD64MULQU2,
				OpAMD64SQRTSD,
				OpAMD64ROUNDSD,
				OpAMD64POPCNTQ,
				OpAMD64POPCNTL,

				OpARMMUL,
				OpARMHMUL,
				OpARMHMULU,
				OpARMMULLU,
				OpARMMULA,
				OpARMMULS,
				OpARMADDF,
				OpARMADDD,
				OpARMSUBF,
				OpARMSUBD,
				OpARMMULF,
				OpARMMULD,
				OpARMNMULF,
				OpARMNMULD,
				OpARMDIVF,
				OpARMDIVD,
				OpARMMULAF,
				OpARMMULAD,
				OpARMMULSF,
				OpARMMULSD,

				OpARM64FSQRTD,
				OpARM64FDIVD,
				OpARM64FDIVS,
				OpARM64FMULD,
				OpARM64FMULS,
				OpARM64FSUBD,
				OpARM64FSUBS,
				OpARM64FADDD,
				OpARM64FADDS,
				OpARM64MULL,
				OpARM64UMULH,
				OpARM64UMULL,
				OpARM64MUL,
				OpARM64MULW,
				OpARM64MULH,

				OpMIPSMUL,
				OpMIPSMULT,
				OpMIPSMULTU,
				OpMIPSADDF,
				OpMIPSADDD,
				OpMIPSSUBF,
				OpMIPSSUBD,
				OpMIPSMULF,
				OpMIPSMULD,
				OpMIPSDIVF,
				OpMIPSDIVD,

				OpMIPS64MULV,
				OpMIPS64MULVU,
				OpMIPS64ADDF,
				OpMIPS64ADDD,
				OpMIPS64SUBF,
				OpMIPS64SUBD,
				OpMIPS64MULF,
				OpMIPS64MULD,
				OpMIPS64DIVF,
				OpMIPS64DIVD,

				OpPPC64MULLD,
				OpPPC64MULLW,
				OpPPC64MULHD,
				OpPPC64MULHW,
				OpPPC64MULHDU,
				OpPPC64MULHWU,
				OpPPC64FMUL,
				OpPPC64FMULS,
				OpPPC64FMADD,
				OpPPC64FMADDS,
				OpPPC64FMSUB,
				OpPPC64FMSUBS,

				OpS390XFMULS,
				OpS390XFMUL,
				OpS390XFDIVS,
				OpS390XFDIV,
				OpS390XFNEGS,
				OpS390XFNEG,
				OpS390XFMADDS,
				OpS390XFMADD,
				OpS390XFMSUBS,
				OpS390XFMSUB,
				OpS390XMULLD,
				OpS390XMULLW,
				OpS390XMULHD,
				OpS390XMULHDU:
				cost = 4
			}
			if len(v.Args) > 0 && v.Args[len(v.Args)-1].Type.IsMemory() { // loads are semi-expensive
				cost = 4
			}
			for _, a := range v.Args {
				bb := a.Block
				if b2l[bb.ID] == l && !ss.contains(a.ID) {
					continue valloop
				}
				s := cost + int((*m)[a])
				if s > 255 {
					s = 255
				}
				cost = s
			}
			// all args for v originate elsewhere or are loop invariant
			if !ss.contains(v.ID) {
				change = true
				ss.add(v.ID)
				(*m)[v] = uint8(cost)
			}
		}
	}
	return change
}
