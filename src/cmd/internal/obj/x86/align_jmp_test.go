// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64

package x86_test

import (
	"bufio"
	"bytes"
	"cmd/internal/objabi"
	"internal/testenv"
	"strconv"
	"strings"
	"testing"
)

const asmDataPrefix = `
TEXT ·testASM(SB),4,$0
start:
`

const nineNOPs = `
        BYTE $0x66
        BYTE $0x0F
        BYTE $0x1F
        BYTE $0x84
        BYTE $0x00
        BYTE $0x00
        BYTE $0x00
        BYTE $0x00
        BYTE $0x00
`

const sixNOPs = `
        BYTE $0x66
        BYTE $0x0F
        BYTE $0x1F
        BYTE $0x44
        BYTE $0x00
        BYTE $0x00
`

const fiveNOPs = `
        BYTE $0x0F
        BYTE $0x1F
        BYTE $0x44
        BYTE $0x00
        BYTE $0x00
`

const threeNOPs = `
        BYTE $0x0F
        BYTE $0x1F
        BYTE $0x00
`

const twoNOPs = `
        BYTE $0x66
        BYTE $0x90
`

const oneNOP = `
        BYTE $0x90
`

// The JMP instruction crosses a 32 byte boundary
const testJmp = asmDataPrefix +
	nineNOPs + sixNOPs + `
        JMP start` +
	nineNOPs + fiveNOPs + `
        JMP start
        RET
`

// The CALL instruction crosses a 32 byte boundary
const testCall = asmDataPrefix +
	nineNOPs + sixNOPs + `
        CALL target(SB)` +
	nineNOPs + twoNOPs + `
        CALL target(SB)
        RET
TEXT target(SB),0,$0
       RET
`

// The CALL instruction crosses a 32 byte boundary.
// It will be padded with prefixes.  We use real
// instructions to create the initial alignments so
// there's something to attach a prefix to.
const testCallPadPrefixes = asmDataPrefix +
	nineNOPs + `
        MOVQ AX, AX
        MOVQ AX, AX
        CALL target(SB)` +
	nineNOPs + `
        XCHGQ AX, AX
        CALL target(SB)
        RET
TEXT target(SB),0,$0
       RET
`

// The CMPQ instruction crosses a 32 byte boundary
const testFused1 = asmDataPrefix +
	nineNOPs + sixNOPs + `
        CMPQ AX, $0
        JL start
` + nineNOPs + oneNOP + `
        CMPQ AX, $0
        JL start
        RET
`

// The JMP instruction ends on a 32 byte boundary
const testEndOn32 = asmDataPrefix +
	nineNOPs + fiveNOPs + `
        JMP start` +
	nineNOPs + fiveNOPs + `
        JMP start
        RET
`

// The CMPQ ends on a 32 byte boundary and is followed by
// a condtional jump with which it is fused.
const testFused2 = asmDataPrefix +
	nineNOPs + threeNOPs + `
        CMPQ AX, $0
        JL start
` + nineNOPs + oneNOP + `
        CMPQ AX, $0
        JL start
        RET
`

// The conditional jump crosses a 32 byte boundary and it
// is fused with the preceeding comparison instruction.
const testFused3 = asmDataPrefix +
	nineNOPs + twoNOPs + `
        CMPQ AX, $0
        JL start
` + nineNOPs + oneNOP + `
        CMPQ AX, $0
        JL start
        RET
`

// The CMPQ ends on a 32 byte boundary and is followed by
// a condtional jump with which it is fused.  The CMPQ and
// the JUMP instruction are separated by assembler directives
// that don't consume any bytes in the object code.
const testFused4 = asmDataPrefix +
	nineNOPs + threeNOPs + `
        CMPQ AX, $0
        PCDATA  $2, $1
        JL start
` + nineNOPs + oneNOP + `
        CMPQ AX, $0
        PCDATA  $2, $1
        JL start
        RET
`

type decodedInstr struct {
	address uint64
	size    uint64
	text    string
}

func assembleDisassemble(t *testing.T, source string) []decodedInstr {
	var ins []decodedInstr

	testenv.MustHaveGoBuild(t)
	objout := objdumpOutput(t, "alignjmp", source)
	data := bytes.Split(objout, []byte("\n"))
	for _, row := range data {
		var di decodedInstr
		var err error
		scan := bufio.NewScanner(bytes.NewReader(row))
		scan.Split(bufio.ScanWords)
		if !scan.Scan() {
			continue
		}
		if !scan.Scan() {
			continue
		}
		addr := scan.Text()
		if len(addr) < 3 {
			continue
		}
		di.address, err = strconv.ParseUint(addr[2:], 16, 64)
		if err != nil {
			continue
		}
		if !scan.Scan() {
			continue
		}
		di.size = uint64(len(scan.Text()) / 2)
		if !scan.Scan() {
			continue
		}
		di.text = scan.Text()
		ins = append(ins, di)
	}

	return ins
}

func checkAlignment(t *testing.T, ins []decodedInstr) {
	for i := 0; i < len(ins); i++ {
		di := ins[i]

		// Check the unfused jumps
		if strings.Contains(di.text, "JMP") || strings.Contains(di.text, "CALL") {
			if (di.address&0x1f)+di.size >= 32 {
				t.Errorf("Instruction %s address %x size %d crosses or ends on 32 byte boundary",
					di.text, di.address, di.size)
			}
		} else if strings.Contains(di.text, "CMP") {
			if i+1 < len(ins) {
				di2 := ins[i+1]
				if strings.Contains(di2.text, "JL") {
					if (di.address&0x1f)+di.size+di2.size >= 32 {
						t.Errorf("Fused instructions %s %s address %x size %d crosses or ends on 32 byte boundary",
							di.text, di2.text, di.address, di.size+di2.size)
					}
					i++
				}
			}
		}
	}
}

func TestAlignJMP(t *testing.T) {
	if objabi.GOAMD64 != "alignedjumps" {
		t.Skip("Skipping TestAlignJMP as GOAMD64 != alignedjumps")
	}

	// Check assembler does not place stand alone or fused jumps on or across 32 byte boundaries

	for _, d := range []string{testJmp, testCall, testCallPadPrefixes, testEndOn32, testFused1, testFused2, testFused3, testFused4} {
		dd := assembleDisassemble(t, d)
		if len(dd) == 0 {
			t.Errorf("Disassembling %s produced no instructions", d)
		}
		checkAlignment(t, dd)
	}
}
