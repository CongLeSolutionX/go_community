package ssa

// maxUnrollSize is the limit of new instructions unroll is willing to create.
// Hopefully most of them will be optimised away.
// TODO: figure out a better mettric
const maxUnrollSize = 32

// unroll perform trivial loop unrolling
// this helps with loops of known fixed size
func unroll(ind indVar, f *Func) bool {
	// match trivial ? -> header -> {body -> header, ?} cases
	body := ind.entry
	if body.Kind != BlockPlain {
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because body isn't plain", body)
		}
		return false
	}

	header := body.Succs[0]
	if header.b.Kind != BlockIf {
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because header is not an If block", body)
		}
		return false
	}
	_ = header.b.Succs[1] // assert the If block has two successors
	var exit Edge
	for i, e := range header.b.Succs {
		if e.b == body {
			exit = header.b.Succs[1-i]
			goto FoundHeader
		}
	}
	if body.Func.pass.debug > 3 {
		body.Func.Warnl(body.Pos, "body=%s unrolling failed because body was not found in the next block succs", body)
	}
	return false
FoundHeader:
	if header.b != ind.ind.Block {
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because induction variable (%v) is not in the header block", body, ind.ind)
		}
		return false
	}

	var delta int64
	switch ind.max.Op {
	case OpConst64, OpConst32, OpConst16, OpConst8:
		switch ind.min.Op {
		case OpConst64, OpConst32, OpConst16, OpConst8:
			delta = ind.max.AuxInt - ind.min.AuxInt
		default:
			// TODO: handle downward loops ?
			if body.Func.pass.debug > 3 {
				body.Func.Warnl(body.Pos, "body=%s unrolling failed because body was not found in the next block succs", body)
			}
			return false
		}
	default:
		var w *Value
		w, delta = isConstDelta(ind.max)
		if w != ind.min {
			if body.Func.pass.debug > 3 {
				body.Func.Warnl(body.Pos, "body=%s unrolling failed because the base of ind.max != ind.min", body)
			}
			return false
		}
	}
	if delta%ind.inc != 0 {
		// INVESTIGATE: Should not happen ?
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because delta%index.inc != 0", body)
		}
		return false
	}

	count := delta / ind.inc
	visitor := &unrollUsesVisitor{
		visited: make([]*Value, 0, maxUnrollSize),
		blocks:  [2]*Block{body, header.b},
	}
	for _, v := range header.b.Values {
		if v.Op != OpPhi {
			continue
		}

		if !visitor.visit(v) {
			if body.Func.pass.debug > 3 {
				body.Func.Warnl(body.Pos, "body=%s unrolling failed because it was rejected by visitor", body)
			}
			return false
		}
	}

	if len(visitor.visited) > maxUnrollSize {
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because the visitor visited more things than allowed by limit", body)
		}
		return false
	}

	// shrink count until we found a compatible unrollSize
	// we might not be able to unroll 4 times, but maybe we can do 2 times
	// (and so do a partial unroll)
	for count*int64(len(visitor.visited)) > maxUnrollSize ||
		delta%(count*ind.inc) != 0 {
		// FIXME: this is a stupid algorithm, there must be algorithms to find
		// compatible multiples faster.
		count--
	}
	if count == 0 {
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because it didn't found a solution to the unroll size", body)
		}
		return false
	}

	full := count*ind.inc == delta
	if !full {
		// TODO: support partial unrolling
		if body.Func.pass.debug > 3 {
			body.Func.Warnl(body.Pos, "body=%s unrolling failed because the unrolling isn't full", body)
		}
		return false
	}
	// perform unrolling
	origs := visitor.visited
	// relations is a slice of slice of offets telling us which arguments
	// should we use.
	// if it is O then this is a loop invariant.
	// in all other cases we store idx+1 or ^(idx+1) in order to keep 0 free.
	// if it is positive this is pointing to variables in the current iteration
	// of the loop.
	// if it is negative, after being NOTed it will point to variables in the
	// previous iteration of the loop.
	relations := make([][]int, len(origs))
origsRelationLoop:
	for i, v := range origs {
		// because of our trivial case, all interiteration variables go
		// through phis.
		if v.Op == OpPhi {
			a := v.Args[header.i]
			if a.Block != header.b && a.Block != body {
				// the value is outside of the loop
				relations[i] = []int{0}
				continue
			}
			for j, w := range origs {
				if w == a {
					relations[i] = []int{^(j + 1)}
					continue origsRelationLoop
				}
			}
			panic("can't find phi's argument")
		}

		r := make([]int, len(v.Args))
	ArgLoop:
		for i, a := range v.Args {
			if a.Block != header.b && a.Block != body {
				// the value is outside of the loop
				r[i] = 0
				continue
			}
			for j, w := range origs {
				if w == a {
					r[i] = j + 1
					continue ArgLoop
				}
			}
			panic("can't find argument")
		}
		relations[i] = r
	}

	// create front and back buffers, we will need origs later so save it
	previousIteration := append([]*Value(nil), origs...)
	currentIteration := make([]*Value, len(origs))
	originalValuesLengthOfBody := len(body.Values)
	// Only runs up to 1 (if we need for 4 count, we must duplicate the body
	// 3 times because the body already contain the content once).
	for i := count; i != 1; i-- {
		// Actually unroll into the body now.
		for i, orig := range previousIteration {
			// First copy the previous body as-is.
			var c *Value
			if orig.Op == OpPhi {
				c = body.NewValue1(orig.Pos, OpCopy, orig.Type, orig)
			} else {
				c = body.NewValue0(orig.Pos, orig.Op, orig.Type)
				c.Aux = orig.Aux
				c.AuxInt = orig.AuxInt
				c.AddArgs(orig.Args...)
			}
			currentIteration[i] = c
		}
		for i, v := range currentIteration {
			// Then rewire the arguments using the relation table.
			for i, dest := range relations[i] {
				if dest == 0 {
					continue
				}
				pick := currentIteration
				if dest < 0 {
					dest = ^dest
					pick = previousIteration
				}
				dest--
				v.SetArg(i, pick[dest])
			}
		}
		// Swap front and back buffers (discard previousIteration).
		previousIteration, currentIteration = currentIteration, previousIteration
	}
	if full {
		// So we fully unrolled the loop, we are gonna rewrite the body
		// to jump to the exit block.
		// The strategy is if the exit block has only one predecessor
		// (rather no phis) we will move our original phis there and
		// if we only have one remaining predecessor we will replace
		// entry inputs directly.
		// This avoid having to scan usage in header's dependency.
		// else (if we have multiple predecessors still) we are gonna
		// recreate new phis.

		// If the exit block has multiple predecessors then we just
		// append our leaking values to thoses.

		// We actually apply this for each phi indepently.
		var totalPhis int
		for _, v := range header.b.Values {
			if v.Op == OpPhi {
				totalPhis++
			}
		}
		moves := make([]unrollPhiMove, totalPhis)
		var j int
		for i, v := range header.b.Values {
			if v.Op != OpPhi {
				continue
			}

			moves[j] = unrollPhiMove{
				old:           v,
				oldIdxInBlock: i,
			}
			j++
		}

	ExitPhisComparingToMovesLoop:
		for _, v := range exit.b.Values {
			if v.Op != OpPhi {
				continue
			}

			for i, m := range moves {
				if v.Args[exit.i] == m.old {
					moves[i].new = v
					continue ExitPhisComparingToMovesLoop
				}
			}
		}

		// iterate backward to keep indexes correct when we shrink header.b.Values
		for i := len(moves); i != 0; {
			i--
			m := moves[i]
			var valueToReplaceOldPhiWith, phiInExit *Value
			if m.new == nil {
				// no already existing phi node, we will have to move ours
				if len(header.b.Preds) == 2 {
					// we don't need to create a phi in header because
					// there will only be one remaining predecessor.
					valueToReplaceOldPhiWith = m.old.Args[1-header.i]
					header.b.Values[m.oldIdxInBlock] = header.b.Values[len(header.b.Values)-1]
					header.b.Values[len(header.b.Values)-1] = nil
					header.b.Values = header.b.Values[:len(header.b.Values)-1]
				} else {
					// there will be other predecessors remaining in header,
					// this require us to create a new phi in header to deal
					// with thoses.
					v := m.old
					c := header.b.Func.newFreeValue(OpPhi, v.Type, v.Pos)
					c.Block = header.b
					c.Aux = v.Aux
					c.AuxInt = v.AuxInt
					c.AddArgs(v.Args[:header.i]...)
					c.AddArgs(v.Args[header.i+1:]...)
					header.b.Values[m.oldIdxInBlock] = c
					valueToReplaceOldPhiWith = c
				}
				phiInExit = m.old
				phiInExit.Block = exit.b
				exit.b.Values = append(exit.b.Values, phiInExit)
				// So at this point we have a value usable in header,
				// the only way for exit to not have a phi for this value already
				// is because all predecessors of exit already make use of header's
				// value, however when will add the edge from body to exit, exit's
				// phi will use the result from body instead.
				// All of this to say, we just have to setup the exit's phi to use
				// header's value from all current edges.
				phiInExit.resetArgs()
				for len(phiInExit.Args) < len(exit.b.Preds) {
					phiInExit.AddArg(valueToReplaceOldPhiWith)
				}
			} else {
				if len(header.b.Preds) == 2 {
					// we don't need to fix the phi in header because
					// there will only be one remaining predecessor.
					valueToReplaceOldPhiWith = m.old.Args[1-header.i]
					m.old.reset(OpInvalid)
				} else {
					// fix header's phi
					v := m.old
					v.Args[header.i].Uses--
					oldLen := len(v.Args)
					v.Args = append(v.Args[:header.i], v.Args[header.i+1:]...)
					v.Args[:oldLen][oldLen-1] = nil
				}
				phiInExit = m.new
			}

			// Now replace all usage of the old phi in header and body.
			for _, v := range header.b.Values {
				if v.Op == OpPhi {
					// Don't replace usage of Phis because they might still use it
					// in nested loops cases.
					// We fixed them earlier anyway so they are good.
					continue
				}
				for i, a := range v.Args {
					if a == m.old {
						v.SetArg(i, valueToReplaceOldPhiWith)
					}
				}
			}
			// small optimisation, we know that unrolled values can't use the phi
			// because they would point to previous iteration's values instead.
			for _, v := range body.Values[:originalValuesLengthOfBody] {
				for i, a := range v.Args {
					if a == m.old {
						v.SetArg(i, valueToReplaceOldPhiWith)
					}
				}
			}

			// Now add body's output to the phi in exit
			var idxInRelations int
			for i, v := range origs {
				if v == phiInExit {
					idxInRelations = i
					goto AddOutputFromBody
				}
			}
			panic("should have found value")
		AddOutputFromBody:

			dest := relations[idxInRelations][0]
			var v *Value
			if dest == 0 {
				// dests of zero indicate outside of loop value
				v = valueToReplaceOldPhiWith
			} else {
				if dest > 0 {
					// destinations of output values should never reffer to the
					// current iteration, there is no current iteration
					panic("output value's relation table reffer to current iteration")
				}
				dest = ^dest
				dest--
				v = previousIteration[dest]
			}
			phiInExit.AddArg(v)
		}
		// rewrite blocks edges
		header.b.removePred(header.i)
		body.removeSucc(0)
		body.AddEdgeTo(exit.b)

		if body.Func.pass.debug > 0 {
			body.Func.Warnl(body.Pos, "body=%s loop unrolled %d times", body, count)
		}

		return true
	}
	// TODO: implement partial unrolls
	panic("unreachable") // should have returned earlier
}

type unrollPhiMove struct {
	old           *Value
	oldIdxInBlock int
	new           *Value
}

type unrollUsesVisitor struct {
	visited []*Value
	blocks  [2]*Block
}

func (vis *unrollUsesVisitor) visit(v *Value) bool {
	switch v.Op {
	case OpStaticCall, OpClosureCall, OpInterCall, OpTailCall,
		OpStaticLECall, OpClosureLECall, OpInterLECall, OpTailLECall:
		// it's rarely worth duplicating calls
		return false
	}

	for _, b := range vis.blocks {
		if b == v.Block {
			goto ScanValue
		}
	}
	return true // value is not relevent anymore

ScanValue:
	for _, vv := range vis.visited {
		if vv == v {
			return true
		}
	}
	if len(vis.visited) == maxUnrollSize/2 {
		// too big wont unroll
		return false
	}
	vis.visited = append(vis.visited, v)

	for _, a := range v.Args {
		if !vis.visit(a) {
			return false
		}
	}
	return true
}
