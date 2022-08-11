// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

// The file contains the arm64 instruction table, which is derived from instFormats
// of https://github.com/golang/arch/blob/master/arm64/arm64asm/tables.go. In the
// future, we'd better make these two tables consistent.

type instArgs []argtype

// inst describes the format of an arm instruction.
type inst struct {
	skeleton uint32   // the known bits of the instruction
	as       string   // Go instruction name
	args     instArgs // instruction argument types
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
	if p1.as != p2.as {
		return p1.as < p2.as
	}
	if len(p1.args) != len(p2.args) {
		return len(p1.args) < len(p2.args)
	}
	return false
}

// arm instruction table.
var instTab = []inst{}
