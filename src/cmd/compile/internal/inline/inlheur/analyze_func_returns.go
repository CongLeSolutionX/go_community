// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/ir"
	"fmt"
	"go/constant"
	"go/token"
	"os"
)

// returnsAnalyzer stores state information for the process of
// computing flags/properties for the return values of a specific
// Go function, as part of inline heuristics synthesis.
type returnsAnalyzer struct {
	fname    string
	values   []ReturnPropBits
	literals []constant.Value
}

func makeReturnsAnalyzer(fn *ir.Func) *returnsAnalyzer {
	results := fn.Type().Results().FieldSlice()
	vals := make([]ReturnPropBits, len(results))
	literals := make([]constant.Value, len(results))
	for i := range results {
		rt := results[i].Type
		if !rt.IsScalar() && !rt.HasNil() {
			// existing properties not applicable here (for things
			// like structs, arrays, slices, etc).
			vals[i] = ReturnNoInfo
			continue
		}
		// initialize to "top" element of data flow lattice,
		// meaining "we have no info yet, but we might later on".
		vals[i] = ReturnTop
	}
	return &returnsAnalyzer{
		fname:    fn.Sym().Name,
		values:   vals,
		literals: literals,
	}
}

func (ra *returnsAnalyzer) results() []ReturnPropBits {
	return ra.values
}

func (ra *returnsAnalyzer) pessimize() {
	for i := range ra.values {
		ra.values[i] = ReturnNoInfo
	}
}

func (ra *returnsAnalyzer) nodeVisit(n ir.Node, aux interface{}) {
	if len(ra.values) == 0 {
		return
	}
	if n.Op() != ir.ORETURN {
		return
	}
	if debugTrace&debugTraceReturns != 0 {
		fmt.Fprintf(os.Stderr, "=+= returns nodevis %v %s\n",
			ir.Line(n), n.Op().String())
	}

	// No support currently for named returns, so if we see an empty
	// return, throw out any results.
	rs := n.(*ir.ReturnStmt)
	if len(rs.Results) == 0 || len(rs.Results) != len(ra.values) {
		ra.pessimize()
		return
	}
	for i, r := range rs.Results {
		ra.analyzeReturn(i, r)
	}
}

// analyzeReturn examines the expression 'n' being returned
// as the 'ii'th argument in some return statement to see whether
// has interesting characteristics (for example, returns a constant),
// then applies a dataflow "meet" operation to combine this result
// with any previous result (for the given return slot) that we've
// already processed.
func (ra *returnsAnalyzer) analyzeReturn(ii int, n ir.Node) {
	isAllocMem := isAllocatedMem(n)
	isConcConvItf := isConcreteConvIface(n)
	lit, isConst := isLiteral(n)
	curp := ra.values[ii]
	newp := ReturnNoInfo
	var newlit constant.Value

	if debugTrace&debugTraceReturns != 0 {
		fmt.Fprintf(os.Stderr, "=-= %v: analyzeReturn n=%s ismem=%v isconcconv=%v isconst=%v\n", ir.Line(n), n.Op().String(), isAllocMem, isConcConvItf, isConst)
	}

	switch curp {
	case ReturnTop:
		// top element: this is the first return we're seen
		switch {
		case isAllocMem:
			newp = ReturnIsAllocatedMem
		case isConcConvItf:
			newp = ReturnIsConcreteTypeConvertedToInterface
		case isConst:
			newp = ReturnAlwaysSameConstant
			newlit = lit
		}
	case ReturnIsAllocatedMem:
		if isAllocatedMem(n) {
			newp = ReturnIsAllocatedMem
		}
	case ReturnIsConcreteTypeConvertedToInterface:
		if isConcreteConvIface(n) {
			newp = ReturnIsConcreteTypeConvertedToInterface
		}
	case ReturnAlwaysSameConstant:
		if isConst && isSameLiteral(lit, ra.literals[ii]) {
			newp = ReturnAlwaysSameConstant
		}
	}
	ra.values[ii] = newp
	ra.literals[ii] = newlit
}

func isAllocatedMem(n ir.Node) bool {
	sv := ir.StaticValue(n)
	switch sv.Op() {
	case ir.OMAKESLICE, ir.ONEW, ir.OPTRLIT, ir.OSLICELIT:
		return true
	}
	return false
}

func isLiteral(n ir.Node) (constant.Value, bool) {
	sv := ir.StaticValue(n)
	if sv.Op() == ir.ONIL {
		return nil, true
	}
	if sv.Op() != ir.OLITERAL {
		return nil, false
	}
	ce := sv.(*ir.ConstExpr)
	return ce.Val(), true
}

// isSameLiteral checks to see if 'v1' and 'v2' correspond to the same
// literal value, or if they are both nil.
func isSameLiteral(v1, v2 constant.Value) bool {
	if v1 == nil && v2 == nil {
		return true
	}
	if v1 == nil || v2 == nil {
		return false
	}
	return constant.Compare(v1, token.EQL, v2)
}

func isConcreteConvIface(n ir.Node) bool {
	sv := ir.StaticValue(n)
	if sv.Op() != ir.OCONVIFACE {
		return false
	}
	cie := sv.(*ir.ConvExpr)
	return isAllocatedMem(cie.X)
}
