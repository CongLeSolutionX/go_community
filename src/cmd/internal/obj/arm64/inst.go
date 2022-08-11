// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// The file contains the arm64 instruction table, which is get by parsing
// the xml document https://developer.arm.com/downloads/-/exploration-tools

type arg struct {
	aType int       // such as REG, RSP, IMM, COND, PCREL...
	elms  []elmType // the elements that this arg includes
}

// inst describes the format of an arm instruction.
type inst struct {
	goOp     string // Go opcode mnemonic
	armOp    string // Arm opcode mnemonic
	feature  string // such as "FEAT_LSE", "FEAT_CSSC"
	skeleton uint32 // known bits
	mask     uint32 // mask for disassembly, 1 for known bits, 0 for unknown bits
	args     []arg  // args, in Go order
}

type icmp []inst

func (x icmp) Len() int {
	return len(x)
}

func (x icmp) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x icmp) Less(i, j int) bool {
	p1 := &x[i]
	p2 := &x[j]
	if p1.goOp != p2.goOp {
		return p1.goOp < p2.goOp
	}
	if len(p1.args) != len(p2.args) {
		return len(p1.args) < len(p2.args)
	}
	if p1.skeleton != p2.skeleton {
		return p1.skeleton < p2.skeleton
	}
	if p1.mask != p2.mask {
		return p1.mask < p2.mask
	}
	return false
}

// Arm64 Instruction table.
var instTab = []inst{}
