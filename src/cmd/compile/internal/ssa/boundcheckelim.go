// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This optimization eliminates unnecessary bound checks for cases:
//
// if len(arr) > CONST_2 { arr[len(arr) - CONST_1] }
// if len(arr) > CONST_2 { arr[CONST_1] }
//
// where CONST_2 > CONST_1

package ssa

// Operation classes
const (
	OTHER = iota
	ADD
	SUB
	CONST
	NEQ
	LEQ
	LE
	EQ
)

// Result of bound check
const (
	UNKNOWN = iota
	IN_BOUND
	OUT_OF_BOUND
)

// Find all bound checks and blocks where they are used
func findBoundChecks(f *Func) map[*Value][]*Block {
	var vals map[*Value][]*Block
	vals = make(map[*Value][]*Block)

	for _, b := range f.Blocks {

		if b.Kind != BlockIf {
			continue
		}

		checkVal := b.Controls[0]

		if checkVal.Op != OpIsInBounds {
			continue
		}

		// The save value can be used in few different blocks as control value
		vals[checkVal] = append(vals[checkVal], b)
	}

	return vals
}

// Get predicate of IF branch for pass to checkBlock.
// If there is more than one way between checkBlock and ifBlock
// the second return value will be false
func getBranchPredicate(checkBlock *Block, ifBlock *Block) (bool, bool) {
	var pred = checkBlock

	for {
		if pred == nil {
			return false, false
		}

		if len(pred.Preds) != 1 {
			return false, false
		}

		parent := pred.Preds[0].b

		if parent != ifBlock {
			pred = parent
			continue
		}

		if parent.Succs[0].b == pred {
			return true, true
		} else {
			return false, true
		}
	}
}

// Checks if elimination of bound check can be done
// Returns index difference and known upper bound
func getCheckResult(v *Value) int {
	b := v.Block // Block where bound check is located

	lenVal := v.Args[1]
	indexCountVal := v.Args[0]
	countType := getValType(indexCountVal)

	// NOTE Did not find examples of other operations here
	// Possible operation is ZeroExt8to64 but we can not work with it
	if countType != ADD && countType != CONST {
		return UNKNOWN
	}

	var indConst *Value
	if countType == ADD {
		if indexCountVal.Args[0] == lenVal {
			indConst = indexCountVal.Args[1]
		}

		if indConst == nil && indexCountVal.Args[1] == lenVal {
			indConst = indexCountVal.Args[0]
		}
	} else if countType == CONST {
		indConst = indexCountVal
	}

	if indConst == nil {
		return UNKNOWN
	}

	indexArgType := getValType(indConst)
	if indexArgType != CONST {
		// TODO probably we can extend optimization to work not only with CONST:
		// if b < len(arr) { use arr[b] }
		return UNKNOWN
	}

	ifBlock := indConst.Block
	if ifBlock.Kind != BlockIf {
		return UNKNOWN
	}

	predicate, isOnlyPath := getBranchPredicate(b, ifBlock)

	if !isOnlyPath {
		return UNKNOWN
	}

	cmpVal := ifBlock.Controls[0]

	if cmpVal != b.Preds[0].b.Controls[0] {
		return UNKNOWN
	}

	cmpType := getValType(cmpVal)

	if cmpType != LE && cmpType != LEQ && cmpType != EQ {
		return UNKNOWN
	}

	if cmpVal.Args[1] != lenVal {
		return UNKNOWN
	}

	upperBoundVal := cmpVal.Args[0]
	upperBoundValType := getValType(upperBoundVal)

	if upperBoundValType != CONST {
		return UNKNOWN
	}

	var indDiff int64 = -1

	if countType == ADD {
		// if len(arr) (==|>=|>) upperBoundVal { arr[len(arr) - indConst] }
		indDiff = upperBoundVal.AuxInt + indConst.AuxInt
	} else if countType == CONST {
		// if len(arr) (==|>=|>) upperBoundVal { arr[indConst] }
		indDiff = upperBoundVal.AuxInt - indConst.AuxInt

		if cmpType == EQ || cmpType == LEQ {
			// if len(arr) (==|>=) upperBoundVal { arr[indConst] }
			// assume upperBoundVal - indConst == 0 => out of range
			indDiff--
		}
	}

	if indDiff >= 0 && indDiff < upperBoundVal.AuxInt {
		// Lesser than upper bound
		if predicate {
			// Bound check is true branch
			return IN_BOUND
		} else {
			// Bound check is false branch
			return OUT_OF_BOUND
		}
	} else {
		// Greater than upper bound
		if cmpType == EQ {
			if predicate {
				// Bound check is true branch
				// Always out of bounds
				return OUT_OF_BOUND
			} else {
				return IN_BOUND
			}
		}
	}

	return UNKNOWN
}

// returns type of operation
func getValType(v *Value) int {
	switch v.Op {
	case OpAdd64, OpAdd32, OpAdd16, OpAdd8:
		return ADD
	case OpConst64, OpConst32, OpConst16, OpConst8:
		return CONST
	case OpLeq64, OpLeq32, OpLeq16, OpLeq8:
		return LEQ
	case OpLess64, OpLess32, OpLess16, OpLess8:
		return LE
	case OpEq64, OpEq32, OpEq16, OpEq8:
		return EQ
	default:
		return OTHER
	}

	return OTHER
}

// For block with IsInBounds operation returns succ block with PanicBounds
func getPanicBlock(b *Block) (*Block, int) {
	panicBlock := b.Succs[0].b
	if panicBlock.Kind == BlockExit {
		return panicBlock, 0
	}

	return b.Succs[1].b, 1
}

// For block with IsInBounds operation returns succ block without PanicBounds
func getNonPanicBlock(b *Block) (*Block, int) {
	nonPanicBlock := b.Succs[0].b
	if nonPanicBlock.Kind != BlockExit {
		return nonPanicBlock, 0
	}

	return b.Succs[1].b, 1
}

// Transform IF blocks with check to plain blocks
func eliminateChecks(v *Value, checkType int, blockList []*Block) {
	for _, b := range blockList {
		var blockToRemove *Block
		var succInd int

		if checkType == IN_BOUND {
			// Always in bounds
			blockToRemove, succInd = getPanicBlock(b)
		} else {
			// Always out of bounds
			blockToRemove, succInd = getNonPanicBlock(b)
		}

		// Transform to PlaneBlock
		b.removeSucc(succInd)
		b.Controls[0] = nil
		b.Controls[1] = nil
		b.Kind = BlockPlain
		v.Uses--

		blockToRemove.removePred(0)
	}
}

// Eliminates bound checks if they are always in (or out of) bounds
func boundcheckelim(f *Func) {

	valMap := findBoundChecks(f)

	for v, blockList := range valMap {
		checkType := getCheckResult(v)

		if checkType == UNKNOWN {
			continue
		}

		eliminateChecks(v, checkType, blockList)
	}
}
