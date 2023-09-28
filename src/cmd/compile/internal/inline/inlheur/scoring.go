// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/pgo"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// These constants enumerate the set of possible ways/scenarios
// in which we'll adjust the score of a given callsite.
type scoreAdjustTyp uint

const (
	panicPathAdj scoreAdjustTyp = (1 << iota)
	initFuncAdj
	inLoopAdj
	passConstToIfAdj
	passConstToNestedIfAdj
	passConcreteToItfCallAdj
	passConcreteToNestedItfCallAdj
	passFuncToIndCallAdj
	passFuncToNestedIndCallAdj
	passInlinableFuncToIndCallAdj
	passInlinableFuncToNestedIndCallAdj
	returnFeedsConstToIfAdj
	returnFeedsFuncToIndCallAdj
	returnFeedsInlinableFuncToIndCallAdj
	returnFeedsConcreteToInterfaceCallAdj
)

// AdjTypAndScore holds information about a specific score adjustment
// that we can make to a callsite if we find that some beneficial or
// detrimental condition applies. For heuristics where there is "may"
// and "must" version (ex: "param passes constant to if condition" vs
// "param passes constant to nested/conditional if") we store the
// sibling "may" type/adj in addition to the type/adj for the "must"
// version.
type AdjTypAndScore struct {
	AdjTyp    scoreAdjustTyp // kind of adjustment
	AdjVal    int            // value (positive or negative) added to score
	MayAdjTyp scoreAdjustTyp // "may apply" sibling kind
	MayAdjVal int            // "may apply" sibling score adj value
}

// ScoreAdjustmentTable type holds a table of flag/adj entries.
type ScoreAdjustmentTable struct {
	Entries []AdjTypAndScore
}

// This table records the specific values we use to adjust call
// site scores in a given scenario. Notes:
//
//   - the debug option WriteInlScoreAdjTab=<file> can be used to write
//     out the jsonified contents of this table to a file
//
//   - the debug option ReadInlScoreAdjTab=<file> will read json
//     from the specified file and use it to replace the contents of this
//     table (for experimentation purposes)
//     out the jsonified contents of this table to a file
//
//   - the numbers in this table are chosen very arbitrarily; ideally
//     we will go through some sort of turning process to decide
//     what value for each one produces the best performance.
var scoreAdjTab = ScoreAdjustmentTable{
	Entries: []AdjTypAndScore{

		// Entries based on call site flags.
		AdjTypAndScore{
			AdjTyp: panicPathAdj,
			AdjVal: 40,
		},
		AdjTypAndScore{
			AdjTyp: initFuncAdj,
			AdjVal: 20,
		},
		AdjTypAndScore{
			AdjTyp: inLoopAdj,
			AdjVal: -5,
		},

		// Entries based on values passed to specific params at calls.
		AdjTypAndScore{
			AdjTyp:    passConstToIfAdj,
			AdjVal:    -20,
			MayAdjTyp: passConstToNestedIfAdj,
			MayAdjVal: -15,
		},
		AdjTypAndScore{
			AdjTyp:    passConcreteToItfCallAdj,
			AdjVal:    -30,
			MayAdjTyp: passConcreteToNestedItfCallAdj,
			MayAdjVal: -25,
		},
		AdjTypAndScore{
			AdjTyp:    passFuncToIndCallAdj,
			AdjVal:    -25,
			MayAdjTyp: passFuncToNestedIndCallAdj,
			MayAdjVal: -20,
		},
		AdjTypAndScore{
			AdjTyp:    passInlinableFuncToIndCallAdj,
			AdjVal:    -45,
			MayAdjTyp: passInlinableFuncToNestedIndCallAdj,
			MayAdjVal: -40,
		},

		// Entries based on func results feeding into interesting contexts.
		AdjTypAndScore{
			AdjTyp: returnFeedsConstToIfAdj,
			AdjVal: -15,
		},
		AdjTypAndScore{
			AdjTyp: returnFeedsFuncToIndCallAdj,
			AdjVal: -25,
		},
		AdjTypAndScore{
			AdjTyp: returnFeedsInlinableFuncToIndCallAdj,
			AdjVal: -40,
		},
		AdjTypAndScore{
			AdjTyp: returnFeedsConcreteToInterfaceCallAdj,
			AdjVal: -25,
		},
	},
}

var expectedScoreAdjTabNumEntries = len(scoreAdjTab.Entries)

func readScoreAdjTabFromFile(path string, tab *ScoreAdjustmentTable) error {
	const me = "readScoreAdjTabFromFile"
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %q: %v", path, err)
	}
	if err := json.Unmarshal(data, tab); err != nil {
		return fmt.Errorf("json.Unmarshal error on contents of %q: %v",
			path, err)
	}
	// With the current implementation we expect a fully populated
	// entries list (e.g. to disable a specific heuristic, instead of
	// deleting its entry, just set its adjustment value(s) to zero).
	if len(tab.Entries) != expectedScoreAdjTabNumEntries {
		return fmt.Errorf("%s: error: in file %q wanted %d entries got %d",
			me, path, expectedScoreAdjTabNumEntries, len(tab.Entries))
	}
	return nil
}

func buildAdjMapFromTab(tab *ScoreAdjustmentTable) (map[scoreAdjustTyp]int, error) {
	m := make(map[scoreAdjustTyp]int)
	for _, e := range tab.Entries {
		if _, ok := m[e.AdjTyp]; ok {
			return nil, fmt.Errorf("duplicate entry for adj type %d\n", e.AdjTyp)
		}
		m[e.AdjTyp] = e.AdjVal
		if e.MayAdjTyp == 0 {
			continue
		}
		if _, ok := m[e.MayAdjTyp]; ok {
			return nil, fmt.Errorf("duplicate entry for adj type %d\n", e.MayAdjTyp)
		}
		m[e.MayAdjTyp] = e.MayAdjVal
	}
	return m, nil
}

func writeScoreAdjTabToFile(path string, tab *ScoreAdjustmentTable) error {
	data, err := json.MarshalIndent(*tab, "", "\t")
	if err != nil {
		return fmt.Errorf("marshal ScoreAdjustmentTable: %v", err)
	}
	if err := os.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("error writing %s: %v", path, err)
	}
	return nil
}

func SetupScoreAdjTable() {
	if base.Debug.WriteInlScoreAdjTab != "" {
		if err := writeScoreAdjTabToFile(base.Debug.WriteInlScoreAdjTab, &scoreAdjTab); err != nil {
			base.Fatalf("SetupScoreAdjTable: error on write: %v", err)
		}
		base.Debug.WriteInlScoreAdjTab = ""
	}
	if base.Debug.ReadInlScoreAdjTab != "" {
		var tab ScoreAdjustmentTable
		if err := readScoreAdjTabFromFile(base.Debug.ReadInlScoreAdjTab, &tab); err != nil {
			base.Fatalf("error reading score adj table: %v", err)
		}
		scoreAdjTab = tab
	}
	if m, err := buildAdjMapFromTab(&scoreAdjTab); err != nil {
		base.Fatalf("internal error in SetupScoreAdjTable: %v", err)
	} else {
		adjValues = m
	}
}

// This table records the specific values we use to adjust call
// site scores in a given scenario.
var adjValues map[scoreAdjustTyp]int

func adjValue(x scoreAdjustTyp) int {
	if val, ok := adjValues[x]; ok {
		return val
	} else {
		panic("internal error unregistered adjustment type")
	}
}

var mayMust = [...]struct{ may, must scoreAdjustTyp }{
	{may: passConstToNestedIfAdj, must: passConstToIfAdj},
	{may: passConcreteToNestedItfCallAdj, must: passConcreteToItfCallAdj},
	{may: passFuncToNestedIndCallAdj, must: passFuncToNestedIndCallAdj},
	{may: passInlinableFuncToNestedIndCallAdj, must: passInlinableFuncToIndCallAdj},
}

func isMay(x scoreAdjustTyp) bool {
	return mayToMust(x) != 0
}

func isMust(x scoreAdjustTyp) bool {
	return mustToMay(x) != 0
}

func mayToMust(x scoreAdjustTyp) scoreAdjustTyp {
	for _, v := range mayMust {
		if x == v.may {
			return v.must
		}
	}
	return 0
}

func mustToMay(x scoreAdjustTyp) scoreAdjustTyp {
	for _, v := range mayMust {
		if x == v.must {
			return v.may
		}
	}
	return 0
}

// computeCallSiteScore takes a given call site whose ir node is 'call' and
// callee function is 'callee' and with previously computed call site
// properties 'csflags', then computes a score for the callsite that
// combines the size cost of the callee with heuristics based on
// previously parameter and function properties.
func computeCallSiteScore(callee *ir.Func, calleeProps *FuncProps, call ir.Node, csflags CSPropBits) (int, scoreAdjustTyp) {
	// Start with the size-based score for the callee.
	score := int(callee.Inl.Cost)
	var tmask scoreAdjustTyp

	if debugTrace&debugTraceScoring != 0 {
		fmt.Fprintf(os.Stderr, "=-= scoring call to %s at %s , initial=%d\n",
			callee.Sym().Name, fmtFullPos(call.Pos()), score)
	}

	// First some score adjustments to discourage inlining in selected cases.
	if csflags&CallSiteOnPanicPath != 0 {
		score, tmask = adjustScore(panicPathAdj, score, tmask)
	}
	if csflags&CallSiteInInitFunc != 0 {
		score, tmask = adjustScore(initFuncAdj, score, tmask)
	}

	// Then adjustments to encourage inlining in selected cases.
	if csflags&CallSiteInLoop != 0 {
		score, tmask = adjustScore(inLoopAdj, score, tmask)
	}

	// Walk through the actual expressions being passed at the call.
	calleeRecvrParms := callee.Type().RecvParams()
	ce := call.(*ir.CallExpr)
	for idx := range ce.Args {
		// ignore blanks
		if calleeRecvrParms[idx].Sym == nil ||
			calleeRecvrParms[idx].Sym.IsBlank() {
			continue
		}
		arg := ce.Args[idx]
		pflag := calleeProps.ParamFlags[idx]
		if debugTrace&debugTraceScoring != 0 {
			fmt.Fprintf(os.Stderr, "=-= arg %d of %d: val %v flags=%s\n",
				idx, len(ce.Args), arg, pflag.String())
		}
		_, islit := isLiteral(arg)
		iscci := isConcreteConvIface(arg)
		fname, isfunc, _ := isFuncName(arg)
		if debugTrace&debugTraceScoring != 0 {
			fmt.Fprintf(os.Stderr, "=-= isLit=%v iscci=%v isfunc=%v for arg %v\n", islit, iscci, isfunc, arg)
		}

		if islit {
			if pflag&ParamMayFeedIfOrSwitch != 0 {
				score, tmask = adjustScore(passConstToNestedIfAdj, score, tmask)
			}
			if pflag&ParamFeedsIfOrSwitch != 0 {
				score, tmask = adjustScore(passConstToIfAdj, score, tmask)
			}
		}

		if iscci {
			// FIXME: ideally here it would be nice to make a
			// distinction between the inlinable case and the
			// non-inlinable case, but this is hard to do. Example:
			//
			//    type I interface { Tiny() int; Giant() }
			//    type Conc struct { x int }
			//    func (c *Conc) Tiny() int { return 42 }
			//    func (c *Conc) Giant() { <huge amounts of code> }
			//
			//    func passConcToItf(c *Conc) {
			//        makesItfMethodCall(c)
			//    }
			//
			// In the code above, function properties will only tell
			// us that 'makesItfMethodCall' invokes a method on its
			// interface parameter, but we don't know whether it calls
			// "Tiny" or "Giant". If we knew if called "Tiny", then in
			// theory in addition to converting the interface call to
			// a direct call, we could also inline (in which case
			// we'd want to decrease the score even more).
			//
			// One thing we could do (not yet implemented) is iterate
			// through all of the methods of "*Conc" that allow it to
			// satisfy I, and if all are inlinable, then exploit that.
			if pflag&ParamMayFeedInterfaceMethodCall != 0 {
				score, tmask = adjustScore(passConcreteToNestedItfCallAdj, score, tmask)
			}
			if pflag&ParamFeedsInterfaceMethodCall != 0 {
				score, tmask = adjustScore(passConcreteToItfCallAdj, score, tmask)
			}
		}

		if isfunc {
			mayadj := passFuncToNestedIndCallAdj
			mustadj := passFuncToIndCallAdj
			if fn := fname.Func; fn != nil && typecheck.HaveInlineBody(fn) {
				mayadj = passInlinableFuncToNestedIndCallAdj
				mustadj = passInlinableFuncToIndCallAdj
			}
			if pflag&ParamMayFeedIndirectCall != 0 {
				score, tmask = adjustScore(mayadj, score, tmask)
			}
			if pflag&ParamFeedsIndirectCall != 0 {
				score, tmask = adjustScore(mustadj, score, tmask)
			}
		}
	}

	return score, tmask
}

func adjustScore(typ scoreAdjustTyp, score int, mask scoreAdjustTyp) (int, scoreAdjustTyp) {

	if isMust(typ) {
		if mask&typ != 0 {
			return score, mask
		}
		may := mustToMay(typ)
		if mask&may != 0 {
			// promote may to must, so undo may
			score -= adjValue(may)
			mask &^= may
		}
	} else if isMay(typ) {
		must := mayToMust(typ)
		if mask&(must|typ) != 0 {
			return score, mask
		}
	}
	if mask&typ == 0 {
		if debugTrace&debugTraceScoring != 0 {
			fmt.Fprintf(os.Stderr, "=-= applying adj %d for %s\n",
				adjValue(typ), typ.String())
		}
		score += adjValue(typ)
		mask |= typ
	}
	return score, mask
}

// BudgetExpansion returns the amount to relax/expand the base
// inlining budget when the new inliner is turned on. With the new
// inliner, the score for a given callsite can be adjusted down by
// some amount due to heuristics, however we won't know whether this
// is going to happen until much later after the CanInline call. This
// function returns the amount to relax the budget initially (to allow
// for a large score adjustment); later on in RevisitInlinability
// we'll look at each individual function to demote it if needed.
// Note that there is a compile time cost associated with increasing
// the budget: larger budgets mean CanInline takes longer to bail on
// non-inlinable functions, and we'll wind up copying the IR for
// functions that may later on turn out to not be candidates after all.
func BudgetExpansion(maxbud int32) int32 {
	// In the default case, double the budget. This should be good enough
	// for most cases.
	if base.Debug.InlBudgetRelaxAmt != 0 {
		return int32(base.Debug.InlBudgetRelaxAmt)
	}
	return maxbud
}

// largestScoreAdjustment tries to estimate the largest possible
// negative score adjustment that could be applied to a call of the
// function with the specified props. Example:
//
//	func foo() {                  func bar(x int, p *int) int {
//	   ...                          if x < 0 { *p = x }
//	}                               return 99
//	                              }
//
// Function 'foo' above on the left has no interesting properties,
// thus as a result the most we'll adjust any call to is the value for
// "call in loop". If the calculated cost of the function is 150, and
// the in-loop adjustment is 5 (for example), then there is not much
// point treating it as inlinable. On the other hand "bar" has a param
// property (parameter "x" feeds unmodified to an "if" statement") and
// a return property (always returns same constant) meaning that a
// given call _could_ be rescored down as much as -35 points-- thus if
// the size of "bar" is 100 (for example) then there is at least a
// chance that scoring will enable inlining.
func largestScoreAdjustment(fn *ir.Func, props *FuncProps) int {
	var tmask scoreAdjustTyp
	score := adjValues[inLoopAdj] // any call can be in a loop
	for _, pf := range props.ParamFlags {
		if pf == ParamMayFeedInterfaceMethodCall {
			score, tmask = adjustScore(passConcreteToNestedItfCallAdj, score, tmask)
		}
		if pf == ParamFeedsInterfaceMethodCall {
			score, tmask = adjustScore(passConcreteToItfCallAdj, score, tmask)
		}
		if pf == ParamMayFeedIndirectCall {
			score, tmask = adjustScore(passFuncToNestedIndCallAdj, score, tmask)
		}
		if pf == ParamFeedsIndirectCall {
			score, tmask = adjustScore(passFuncToIndCallAdj, score, tmask)
		}
		if pf == ParamMayFeedIfOrSwitch {
			score, tmask = adjustScore(passConstToNestedIfAdj, score, tmask)
		}
		if pf == ParamFeedsIfOrSwitch {
			score, tmask = adjustScore(passConstToIfAdj, score, tmask)
		}
	}
	for _, rf := range props.ResultFlags {
		if rf == ResultAlwaysSameConstant {
			score, tmask = adjustScore(returnFeedsConstToIfAdj, score, tmask)
		}
		if rf == ResultIsConcreteTypeConvertedToInterface {
			score, tmask = adjustScore(returnFeedsConcreteToInterfaceCallAdj, score, tmask)
		}
		if rf == ResultAlwaysSameInlinableFunc {
			score, tmask = adjustScore(returnFeedsInlinableFuncToIndCallAdj, score, tmask)
		}
	}

	if debugTrace&debugTraceScoring != 0 {
		fmt.Fprintf(os.Stderr, "=-= largestScore(%v) is %d\n",
			fn, score)
	}

	return score
}

// DumpInlCallSiteScores is invoked by the inliner if the debug flag
// "-d=dumpinlcallsitescores" is set; it dumps out a human-readable
// summary of all (potentially) inlinable callsites in the package,
// along with info on call site scoring and the adjustments made to a
// given score. Here profile is the PGO profile in use (may be
// nil), budgetCallback is a callback that can be invoked to find out
// the original pre-adjustment hairyness limit for the function, and
// inlineHotMaxBudget is the constant of the same name used in the
// inliner. Sample output lines:
//
// Score  Adjustment  Status  Callee  CallerPos ScoreFlags
// 115    40          DEMOTED cmd/compile/internal/abi.(*ABIParamAssignment).Offset     expand_calls.go:1679:14|6       panicPathAdj
// 76     -5n         PROMOTED runtime.persistentalloc   mcheckmark.go:48:45|3   inLoopAdj
// 201    0           --- PGO  unicode.DecodeRuneInString        utf8.go:312:30|1
// 7      -5          --- PGO  internal/abi.Name.DataChecked     type.go:625:22|0        inLoopAdj
//
// In the dump above, "Score" is the final score calculated for the
// callsite, "Adjustment" is the amount added to or subtracted from
// the original hairyness estimate to form the score. "Status" shows
// whether anything changed with the site -- did the adjustment bump
// it down just below the threshold ("PROMOTED") or instead bump it
// above the threshold ("DEMOTED"); this will be blank ("---") if no
// threshold was crossed as a result of the heuristics. Note that
// "Status" also shows whether PGO was involved. "Callee" is the name
// of the function called, "CallerPos" is the position of the
// callsite, and "ScoreFlags" is a digest of the specific properties
// we used to make adjustments to callsite score via heuristics.
func DumpInlCallSiteScores(profile *pgo.Profile, budgetCallback func(fn *ir.Func, profile *pgo.Profile) (int32, bool)) {

	fmt.Fprintf(os.Stdout, "# scores for package %s\n", types.LocalPkg.Path)

	var indirectlyDueToPromotion func(cs *CallSite) bool
	indirectlyDueToPromotion = func(cs *CallSite) bool {
		bud, _ := budgetCallback(cs.Callee, profile)
		hairyval := cs.Callee.Inl.Cost
		score := int32(cs.Score)
		if hairyval > bud && score <= bud {
			return true
		}
		if cs.parent != nil {
			return indirectlyDueToPromotion(cs.parent)
		}
		return false
	}

	genstatus := func(cs *CallSite) string {
		hairyval := cs.Callee.Inl.Cost
		bud, isPGO := budgetCallback(cs.Callee, profile)
		score := int32(cs.Score)
		st := "---"
		expinl := false
		switch {
		case hairyval <= bud && score <= bud:
			// "Normal" inlined case: hairy val sufficiently low that
			// it would have been inlined anyway without heuristics.
			expinl = true
		case hairyval > bud && score > bud:
			// "Normal" not inlined case: hairy val sufficiently high
			// and scoring didn't lower it.
		case hairyval > bud && score <= bud:
			// Promoted: we would not have inlined it before, but
			// after score adjustment we decided to inline.
			st = "PROMOTED"
			expinl = true
		case hairyval <= bud && score > bud:
			// Demoted: we would have inlined it before, but after
			// score adjustment we decided not to inline.
			st = "DEMOTED"
		}
		inlined := cs.aux&csAuxInlined != 0
		indprom := false
		if cs.parent != nil {
			indprom = indirectlyDueToPromotion(cs.parent)
		}
		if inlined && indprom {
			st += "|INDPROM"
		}
		if inlined && !expinl {
			st += "|[NI?]"
		} else if !inlined && expinl {
			st += "|[IN?]"
		}
		if isPGO {
			st += "|PGO"
		}
		return st
	}

	if base.Debug.DumpInlCallSiteScores != 0 {
		sl := make([]*CallSite, 0, len(fpmap))
		for _, fih := range fpmap {
			for _, cs := range fih.cstab {
				sl = append(sl, cs)
			}
		}
		sort.Slice(sl, func(i, j int) bool {
			if sl[i].Score != sl[j].Score {
				return sl[i].Score < sl[j].Score
			}
			fni := ir.PkgFuncName(sl[i].Callee)
			fnj := ir.PkgFuncName(sl[j].Callee)
			if fni != fnj {
				return fni < fnj
			}
			ecsi := EncodeCallSiteKey(sl[i])
			ecsj := EncodeCallSiteKey(sl[j])
			return ecsi < ecsj
		})

		mkname := func(fn *ir.Func) string {
			var n string
			if fn == nil || fn.Nname == nil {
				return "<nil>"
			}
			if fn.Sym().Pkg == types.LocalPkg {
				n = "·" + fn.Sym().Name
			} else {
				n = ir.PkgFuncName(fn)
			}
			// don't try to print super-long names
			if len(n) <= 64 {
				return n
			}
			return n[:32] + "..." + n[len(n)-32:len(n)]
		}

		if len(sl) != 0 {
			fmt.Fprintf(os.Stdout, "Score  Adjustment  Status  Callee  CallerPos Flags ScoreFlags\n")
		}
		for _, cs := range sl {
			hairyval := cs.Callee.Inl.Cost
			adj := int32(cs.Score) - hairyval
			nm := mkname(cs.Callee)
			ecc := EncodeCallSiteKey(cs)
			fmt.Fprintf(os.Stdout, "%d  %d\t%s\t%s\t%s\t%s\n",
				cs.Score, adj, genstatus(cs),
				nm, ecc,
				cs.ScoreMask.String())
		}
	}
}
