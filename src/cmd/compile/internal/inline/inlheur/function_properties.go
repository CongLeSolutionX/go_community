// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

// This file defines a set of Go function "properties" intended to
// guide inlining heuristics; these properties may apply to the
// function as a whole, or to one or more function return values or
// parameters.
//
// IMPORTANT: function properties are produced on a "best effort"
// basis, meaning that the code that computes them doesn't verify that
// the properties are guaranteed to be true in 100% of cases. For this
// reason, properties should only be used to drive always-safe
// optimization decisions (e.g. "should I inline this call", or
// "should I unroll this loop") as opposed to potentially unsafe IR
// alterations that could change program semantics (e.g. "can I delete
// this variable" or "can I move this statement to a new location").
//
//----------------------------------------------------------------

// FuncProps describes a set of function or method properties that
// may be useful for inlining heuristics. Here 'Flags' are properties
// that we think apply to the entire function; 'RecvrParamFlags'
// are properties of specific function params (or the receiver), and
// 'ReturnFlags' are things properties we think will apply to values
// of specific returns. Note that 'RecvrParamFlags' contains
// only non-blank names; for a function such as "func foo(_ int,
// b byte, _ float32)" the length of RecvrParamFlags will be 1.
type FuncProps struct {
	Flags           FuncPropBits
	RecvrParamFlags []ParamPropBits // slot 0 receiver if applicable
	ReturnFlags     []ReturnPropBits
}

type FuncPropBits uint32

const (
	// Function always panics or invokes os.Exit() or a func that does
	// likewise.
	FuncPropUnconditionalPanicExit FuncPropBits = 1 << iota
)

type ParamPropBits uint32

const (
	// No info about this param
	ParamNoInfo ParamPropBits = 0
	// Parameter value feeds unmodified into one or more interface
	// calls (this assumes the parameter is of interface type).
	ParamFeedsInterfaceMethodCall ParamPropBits = 1 << iota
	// Parameter value feeds unmodified into one or more indirect
	// function calls (assumes parameter is of function type).
	ParamFeedsIndirectCall
	// Parameter value feeds unmodified into one or more "switch"
	// statements or "if" statement simple expressions (see more
	// on "simple" expression classification below).
	ParamFeedsIfOrSwitch
	// Top element in data flow lattice
	ParamTop
)

type ReturnPropBits uint32

const (
	// No info about this return
	ReturnNoInfo ReturnPropBits = 0
	// This return always contains allocated memory.
	ReturnIsAllocatedMem ReturnPropBits = 1 << iota
	// This return is always a single concrete type that is
	// implicitly converted to interface.
	ReturnIsConcreteTypeConvertedToInterface
	// Return is always the same non-composite compile time constant.
	ReturnAlwaysSameConstant
	// Top element in data flow lattice
	ReturnTop
)
