// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/ir"
	"fmt"
	"os"
)

// paramsAnalyzer holds state information for the phase
// that computes flags for a Go functions parameters, for
// use in inline heuristics.
type paramsAnalyzer struct {
	fname  string
	values []ParamPropBits
	params []*ir.Name
}

// getParams returns an *ir.Name slice containing all params for the
// function (plus rcvr as well if applicable).
func getParams(fn *ir.Func) []*ir.Name {
	params := []*ir.Name{}
	for _, n := range fn.Dcl {
		if n.Op() != ir.ONAME {
			continue
		}
		if n.Class != ir.PPARAM {
			continue
		}
		params = append(params, n)
	}
	return params
}

func makeParamsAnalyzer(fn *ir.Func) *paramsAnalyzer {
	params := getParams(fn) // includes receiver if applicable
	vals := make([]ParamPropBits, len(params))
	for i, pn := range params {
		// blank nodes should already have been stripped out
		if ir.IsBlank(pn) {
			panic("something went wrong")
		}
		pt := pn.Type()
		if !pt.IsScalar() && !pt.HasNil() {
			// existing properties not applicable here (for things
			// like structs, arrays, slices, etc).
			vals[i] = ParamNoInfo
			continue
		}
		// If param is reassigned, skip it.
		if ir.Reassigned(pn) {
			vals[i] = ParamNoInfo
			continue
		}
		vals[i] = ParamTop
	}

	if debugTrace&debugTraceParams != 0 {
		fmt.Fprintf(os.Stderr, "=-= param analysis of func %v:\n",
			fn.Sym().Name)
		for i := range vals {
			fmt.Fprintf(os.Stderr, "=-=   %d: %q %s\n",
				i, params[i].Sym().String(), vals[i].String())
		}
	}

	return &paramsAnalyzer{
		fname:  fn.Sym().Name,
		values: vals,
		params: params,
	}
}

func (pa *paramsAnalyzer) results() []ParamPropBits {
	// Map dataflow "top" element to bottom if is survives this long
	// (typically means params are not used in any interesting way).
	for i := range pa.values {
		pa.values[i] &^= ParamTop
	}
	return pa.values
}

// paramsAnalyzer invokes function 'testf' on the specified expression
// 'x' for each parameter, and if the result is TRUE, or's 'flag' into
// the flags for that param.
func (pa *paramsAnalyzer) checkParams(x ir.Node, flag ParamPropBits, testf func(x ir.Node, param *ir.Name) bool) {
	for idx, p := range pa.params {
		if pa.values[idx] == ParamNoInfo {
			continue
		}
		result := testf(x, p)
		if debugTrace&debugTraceParams != 0 {
			fmt.Fprintf(os.Stderr, "=-= test expr %v param %s result=%v flag=%s\n", x, p.Sym().Name, result, flag.String())
		}
		if result {
			pa.values[idx] |= flag
		}
	}
}

// foldCheckParams checks expression 'x' (an if condition or switch
// expr) to see if the expr would fold away if a specific parameter
// had a constant value.
func (pa *paramsAnalyzer) foldCheckParams(x ir.Node) {
	pa.checkParams(x, ParamFeedsIfOrSwitch,
		func(x ir.Node, p *ir.Name) bool {
			return ShouldFoldIfNameConstant(x, p)
		})
}

// callCheckParams examines the target of call expression 'ce' to see
// if it is making a call to the value passed in for some parameter.
func (pa *paramsAnalyzer) callCheckParams(ce *ir.CallExpr) {
	switch ce.Op() {
	case ir.OCALLINTER:
		if ce.Op() != ir.OCALLINTER {
			return
		}
		sel := ce.X.(*ir.SelectorExpr)
		r := ir.StaticValue(sel.X)
		if r.Op() != ir.ONAME {
			return
		}
		name := r.(*ir.Name)
		if name.Class != ir.PPARAM {
			return
		}
		pa.checkParams(r, ParamFeedsInterfaceMethodCall,
			func(x ir.Node, p *ir.Name) bool {
				name := x.(*ir.Name)
				return name == p
			})
	case ir.OCALLFUNC:
		if ce.X.Op() != ir.ONAME {
			return
		}
		called := ir.StaticValue(ce.X)
		if called.Op() != ir.ONAME {
			return
		}
		name := called.(*ir.Name)
		if name.Class != ir.PPARAM {
			return
		}
		pa.checkParams(called, ParamFeedsIndirectCall,
			func(x ir.Node, p *ir.Name) bool {
				name := x.(*ir.Name)
				return name == p
			})
	}
}

func (pa *paramsAnalyzer) nodeVisit(n ir.Node, aux interface{}) {
	if len(pa.values) == 0 {
		return
	}
	switch n.Op() {
	case ir.OCALLFUNC:
		ce := n.(*ir.CallExpr)
		pa.callCheckParams(ce)
	case ir.OCALLINTER:
		ce := n.(*ir.CallExpr)
		pa.callCheckParams(ce)
	case ir.OIF:
		ifst := n.(*ir.IfStmt)
		pa.foldCheckParams(ifst.Cond)
	case ir.OSWITCH:
		swst := n.(*ir.SwitchStmt)
		if swst.Tag != nil {
			pa.foldCheckParams(swst.Tag)
		}
	}
}
