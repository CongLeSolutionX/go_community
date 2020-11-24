// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
	"sync"
)

//......................................................................
//
// Public/exported bits of the ABI utilities.
//

// ABIParamResultInfo stores the results of processing a given
// function type to compute stack layout and register assignments.
// For each param/result we capture whether the param was register-assigned
// (and to which register(s)) or the stack offset for the param if
// is not going to be passed in registers according to the rules
// in the Go internal ABI specification (1.17).
type ABIParamResultInfo struct {
	params            []ABIParamAssignment // Includes receiver for method calls.  Does NOT include hidden closure pointer.
	results           []ABIParamAssignment
	intSpillSlots     int
	floatSpillSlots   int
	offsetToSpillArea int64
}

// RegIndex stores the index into the set of machine registers used by
// the ABI on a specific architecture for parameter passing.  RegIndex
// values 0 through N-1 (where N is the number of integer registers
// used for param passing according to the ABI rules) describe integer
// registers; values N through M (where M is the number of floating
// point registers used).  Thus if the ABI says there are 5 integer
// registers and 7 floating point registers, then RegIndex value of 4
// indicates the 5th integer register, and a RegIndex value of 11
// indicates the 7th floating point register.
type RegIndex uint8

// ABIParamAssignment holds information about how a specific param or
// result will be passed: in registers (in which case 'Registers' is
// populated) or on the stack (in which case 'Offset' is set to a
// non-negative stack offset. The values in 'Registers' are indices (as
// described above), not architected registers.
type ABIParamAssignment struct {
	Type      *types.Type
	Registers []RegIndex
	Offset    int32
}

// RegAmounts holds a specified number of integer/float registers.
type RegAmounts struct {
	intRegs   int
	floatRegs int
}

// ABIConfig captures the number of registers made available
// by the ABI rules for parameter passing and result returning.
type ABIConfig struct {
	// Do we need anything more than this?
	regAmounts RegAmounts
}

// ABIAnalyze takes a function type 't' and an ABI rules description
// 'config' and analyzes the function to determine how its parameters
// and results will be passed (in registers or on the stack), returning
// a ABIParamResultInfo object that holds the results of the analysis.
func ABIAnalyze(t *types.Type, config ABIConfig) ABIParamResultInfo {
	setup()
	state := assignState{
		rTotal: config.regAmounts,
		which:  doingParams,
	}

	// Receiver
	ft := t.FuncType()
	if t.NumRecvs() != 0 {
		rfsl := ft.Receiver.FieldSlice()
		state.doParmResult(rfsl[0].Type)
	}

	// Inputs
	ifsl := ft.Params.FieldSlice()
	for _, f := range ifsl {
		state.doParmResult(f.Type)
	}
	state.stackOffset = Rnd(state.stackOffset, int64(Widthreg))

	// Outputs
	state.rUsed = RegAmounts{}
	state.pUsed = RegAmounts{}
	state.which = doingResults
	ofsl := ft.Results.FieldSlice()
	for _, f := range ofsl {
		state.doParmResult(f.Type)
	}
	state.result.offsetToSpillArea = state.stackOffset

	return state.result
}

//......................................................................
//
// Non-public portions.

// whichAssign records a given stage of the register assignment
// processing process (either params or results).
type whichAssign int

const (
	doingParams  whichAssign = 0
	doingResults whichAssign = 1
)

// regString produces a human-readable version of a RegIndex.
func (c *RegAmounts) regString(r RegIndex) string {
	if int(r) < c.intRegs {
		return fmt.Sprintf("I%d", int(r))
	} else if int(r) < c.intRegs+c.floatRegs {
		return fmt.Sprintf("F%d", int(r)-c.intRegs)
	}
	return fmt.Sprintf("<?>%d", r)
}

// toString method renders an ABIParamAssignment in human-readable
// form, suitable for debugging or unit testing.
func (ri *ABIParamAssignment) toString(config ABIConfig) string {
	regs := "R{"
	for _, r := range ri.Registers {
		regs += " " + config.regAmounts.regString(r)
	}
	return fmt.Sprintf("%s } offset: %d typ: %v", regs, ri.Offset, ri.Type)
}

// toString method renders an ABIParamResultInfo in human-readable
// form, suitable for debugging or unit testing.
func (ri *ABIParamResultInfo) toString(config ABIConfig) string {
	res := ""
	for k, p := range ri.params {
		res += fmt.Sprintf("P%d: %s\n", k, p.toString(config))
	}
	for k, r := range ri.results {
		res += fmt.Sprintf("R%d: %s\n", k, r.toString(config))
	}
	res += fmt.Sprintf("intspill: %d floatspill: %d offsetToSpillArea: %d",
		ri.intSpillSlots, ri.floatSpillSlots, ri.offsetToSpillArea)
	return res
}

var noRegisters = RegAmounts{0, 0}

// assignState holds intermediate state during the register assigning process
// for a given function signature.
type assignState struct {
	result      ABIParamResultInfo // result we are constructing
	rTotal      RegAmounts         // total reg amounts from ABI rules
	rUsed       RegAmounts         // regs used by params completely assigned so far
	pUsed       RegAmounts         // regs used by the current param (or pieces therein)
	stackOffset int64              // current stack offset
	which       whichAssign        // stage of assignment we're in
}

// stackSlot returns a stack offset for a param or result of the
// specified type.
func (state *assignState) stackSlot(t *types.Type) int64 {
	if t.Align > 0 {
		state.stackOffset = Rnd(state.stackOffset, int64(t.Align))
	}
	rv := state.stackOffset
	state.stackOffset += t.Width
	return rv
}

// allocateRegs returns a set of register indicates for a parameter or result
// that we've just determined to be register-assignable. The number of registers
// needed is assumed to be stored in state.pUsed.
func (state *assignState) allocateRegs() []RegIndex {
	regs := []RegIndex{}

	// integer
	for r := state.rUsed.intRegs; r < state.rUsed.intRegs+state.pUsed.intRegs; r++ {
		regs = append(regs, RegIndex(r))
	}
	state.rUsed.intRegs += state.pUsed.intRegs

	// floating
	for r := state.rUsed.floatRegs; r < state.rUsed.floatRegs+state.pUsed.floatRegs; r++ {
		regs = append(regs, RegIndex(r+state.rTotal.intRegs))
	}
	state.rUsed.floatRegs += state.pUsed.floatRegs

	// record spill slots
	if state.which == doingParams {
		state.result.intSpillSlots += state.pUsed.intRegs
		state.result.floatSpillSlots += state.pUsed.floatRegs
	}

	return regs
}

// recordParamResult records an assignment for a given parameter or
// result, appending to the proper slice depending on which stage
// we're in.
func (state *assignState) recordParamResult(asgn ABIParamAssignment) {
	if state.which == doingParams {
		state.result.params = append(state.result.params, asgn)
	} else {
		state.result.results = append(state.result.results, asgn)
	}
}

// regAllocate creates a register ABIParamAssignment object for a param
// or result with the specified type, as a final step (this assumes
// that all of the safety/suitability analysis is complete).
func (state *assignState) regAllocate(t *types.Type) {
	state.recordParamResult(ABIParamAssignment{
		Type:      t,
		Registers: state.allocateRegs(),
		Offset:    -1,
	})
}

// stackAllocate creates a stack memory ABIParamAssignment object for
// a param or result with the specified type, as a final step (this
// assumes that all of the safety/suitability analysis is complete).
func (state *assignState) stackAllocate(t *types.Type) {
	state.recordParamResult(ABIParamAssignment{
		Type:   t,
		Offset: int32(state.stackSlot(t)),
	})
}

// intUsed returns the number of integer registers consumed
// at a given point within an assignment stage.
func (state *assignState) intUsed() int {
	return state.rUsed.intRegs + state.pUsed.intRegs
}

// floatUsed returns the number of floating point registers consumed at
// a given point within an assignment stage.
func (state *assignState) floatUsed() int {
	return state.rUsed.floatRegs + state.pUsed.floatRegs
}

// visitIntegral examines a param/result of integral type 't' to
// determines whether it can be register-assigned. Returns TRUE if we
// can register allocate, FALSE otherwise (and updates state
// accordingly).
func (state *assignState) visitIntegral(t *types.Type) bool {
	w := Rnd(t.Width, int64(Widthptr))

	// Floating point and complex.
	if t.IsFloat() || t.IsComplex() {
		floatRegsNeeded := int(w / int64(Widthptr))
		if floatRegsNeeded+state.floatUsed() > state.rTotal.floatRegs {
			// not enough regs
			return false
		}
		state.pUsed.floatRegs += floatRegsNeeded
		return true
	}

	// Non-floating point
	intRegsNeeded := int(w / int64(Widthptr))
	if intRegsNeeded+state.intUsed() > state.rTotal.intRegs {
		// not enough regs
		return false
	}
	state.pUsed.intRegs += intRegsNeeded
	return true
}

// visitArray processes an array type (or array component within some
// other enclosing type) to determine if if can be register assigned.
// Returns TRUE if we can register allocate, FALSE otherwise.
func (state *assignState) visitArray(t *types.Type) bool {

	nel := t.NumElem()
	if nel == 0 {
		return true
	}
	if nel > 1 {
		// Not an array of length 1: stack assign
		return false
	}
	// Visit element
	return state.visit(t.Elem())
}

// visitStruct processes a struct type (or struct component within some other
// enclosing type) to determine if if can be register assigned.  Returns TRUE if we
// can register allocate, FALSE otherwise.
func (state *assignState) visitStruct(t *types.Type) bool {
	for _, field := range t.FieldSlice() {
		if !state.visit(field.Type) {
			return false
		}
	}
	return true
}

// mkstruct is a helper routine to create a struct type with fields
// of the types specified in 'fieldtypes'.
func mkstruct(fieldtypes []*types.Type) *types.Type {
	fields := make([]*types.Field, len(fieldtypes))
	for k, t := range fieldtypes {
		if t == nil {
			panic("bad -- field has no type")
		}
		f := types.NewField(src.NoXPos, nil, t)
		fields[k] = f
	}
	s := types.NewStruct(ir.LocalPkg, fields)
	return s
}

// synthOnce ensures that we only create the synth* fake types once.
var synthOnce sync.Once

// synthSlice, synthString, and syncIface are synthesized struct types
// meant to capture the underlying implementations of string/slice/interface.
var synthSlice *types.Type
var synthString *types.Type
var synthIface *types.Type

// setup performs setup for the register assignment utilities, manufacturing
// a small set of synthesized types that we'll need along the way.
func setup() {
	fname := ir.BuiltinPkg.Lookup
	nxp := src.NoXPos
	unsp := types.Types[types.TUNSAFEPTR]
	ui := types.Types[types.TUINTPTR]
	synthOnce.Do(func() {
		synthSlice = types.NewStruct(types.NoPkg, []*types.Field{
			types.NewField(nxp, fname("ptr"), unsp),
			types.NewField(nxp, fname("len"), ui),
			types.NewField(nxp, fname("cap"), ui),
		})
		synthString = types.NewStruct(types.NoPkg, []*types.Field{
			types.NewField(nxp, fname("data"), unsp),
			types.NewField(nxp, fname("len"), ui),
		})
		synthIface = types.NewStruct(types.NoPkg, []*types.Field{
			types.NewField(nxp, fname("f1"), unsp),
			types.NewField(nxp, fname("f2"), unsp),
		})
	})
}

// visit examines a given param type (or component within some
// composite) to determine if it can be register assigned.  Returns
// TRUE if we can register allocate, FALSE otherwise.
func (state *assignState) visit(pt *types.Type) bool {
	typ := pt.Kind()
	if pt.IsScalar() {
		return state.visitIntegral(pt)
	}
	switch typ {
	case types.TARRAY:
		return state.visitArray(pt)
	case types.TSTRUCT:
		return state.visitStruct(pt)
	case types.TSLICE:
		return state.visitStruct(synthSlice)
	case types.TSTRING:
		return state.visitStruct(synthString)
	case types.TINTER:
		return state.visitStruct(synthIface)
	default:
		panic("not expected")
	}
	return false
}

// doParmResult processes a given receiver, param, or result
// of type 'pt' to determine whether it can be register assigned.
// The result of the analysis is recorded in the result
// ABIParamResultInfo held in 'state'.
func (state *assignState) doParmResult(pt *types.Type) {
	if pt.Width == types.BADWIDTH {
		panic("should never happen")
	} else if pt.Width == 0 {
		state.stackAllocate(pt)
	} else if state.visit(pt) {
		state.regAllocate(pt)
	} else {
		state.stackAllocate(pt)
	}
	state.pUsed = RegAmounts{}
}
