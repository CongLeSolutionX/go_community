// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build ignore

package main

import "strings"

var regNames386 = []string{
	"AX",
	"CX",
	"DX",
	"BX",
	"SP",
	"BP",
	"SI",
	"DI",
	"X0",
	"X1",
	"X2",
	"X3",
	"X4",
	"X5",
	"X6",
	"X7",

	// pseudo-registers
	"SB",
	"FLAGS",
}

func init() {

	num := map[string]int{}
	for i, name := range regNames386 {
		num[name] = i
	}

	buildReg := func(s string) regMask {
		m := regMask(0)
		for _, r := range strings.Split(s, " ") {
			if n, ok := num[r]; ok {
				m |= regMask(1) << uint(n)
				continue
			}
			panic("register " + r + " not found")
		}
		return m
	}

	// Common individual register masks
	var (
		gp     = buildReg("AX CX DX BX BP SI DI")
		gpsp   = gp | buildReg("SP")
		gpspsb = gpsp | buildReg("SB")
		flags  = buildReg("FLAGS")
	)

	var (
		gp11sb = regInfo{inputs: []regMask{gpspsb}, outputs: []regMask{gp}}
		gp21sp = regInfo{inputs: []regMask{gpsp, gp}, outputs: []regMask{gp}, clobbers: flags}

		gpstore = regInfo{inputs: []regMask{gpspsb, gpsp, 0}}
	)

	ops := []opData{
		{name: "ADDL", argLength: 2, reg: gp21sp, asm: "ADDL", commutative: true}, // arg0 + arg1

		{name: "LEAL", argLength: 1, reg: gp11sb, aux: "SymOff", rematerializeable: true},

		{name: "MOVLstore", argLength: 3, reg: gpstore, asm: "MOVL", aux: "SymOff", typ: "Mem"}, // store 4 bytes in arg1 to arg0+auxint+aux. arg2=mem
	}

	blocks := []blockData{}

	archs = append(archs, arch{
		name:            "386",
		pkg:             "cmd/internal/obj/x86",
		genfile:         "../../x86/ssa.go",
		ops:             ops,
		blocks:          blocks,
		regnames:        regNames386,
		gpregmask:       gp,
		flagmask:        flags,
		framepointerreg: int8(num["BP"]),
	})

}
