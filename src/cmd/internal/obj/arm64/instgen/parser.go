// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func parseXMLFiles(dir string) {
	log.Printf("start parsing the xml files\n")
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, file := range files {
		fileName := file.Name()
		if ext := path.Ext(fileName); ext != ".xml" {
			continue
		}
		wg.Add(1)
		fileName = path.Join(dir, fileName)
		go func(name string) {
			defer wg.Done()
			parse(name)
		}(fileName)
	}
	wg.Wait()
	log.Printf("Finish parsing the xml files\n")
}

// parse parses a xml file and adds the parsing result to insts.
func parse(f string) {
	xmlFile, err := os.Open(f)
	if err != nil {
		log.Fatalf("Open file %s failed: %v\n", f, err)
	}
	defer xmlFile.Close()
	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Fatalf("io.ReadAll %s failed: %v\n", f, err)
	}

	// Use heap memory, stack memory may be moved around.
	var inst = new(instruction)
	if err = xml.Unmarshal(byteValue, inst); err != nil {
		// Ignore non-instruction files.
		if strings.HasPrefix(err.Error(), "expected element type <instructionsection>") {
			return
		}
		log.Fatalf("Unmarshal %s failed: %v\n", f, err)
	}
	if inst.Type != "instruction" && inst.Type != "alias" {
		return
	}
	// A special case. The file contains some useless encodings, only the first one is valid.
	// TODO: remove it once the issue is fixed.
	if inst.Title == "MOV (SIMD&FP scalar, unpredicated)" {
		inst.Classes.Iclass[0].Encodings = inst.Classes.Iclass[0].Encodings[:1]
	}
	// MOV (inverted wide immediate) is not used in Go, we use MOVN.
	if inst.Title == "MOV (inverted wide immediate) -- A64" {
		return
	}

	if false { // for debugging
		inst.print()
	}

	for !instsLock.CompareAndSwap(0, 1) {
		runtime.Gosched()
	}
	insts = append(insts, inst)
	instsLock.Store(0)
}

func setBinary(code, bitVal uint32, value string) uint32 {
	switch value {
	case "0":
		code &^= bitVal
	case "1":
		code |= bitVal
	}
	return code
}

func setMask(code, bitVal uint32, value string) uint32 {
	switch value {
	case "0", "1":
		code |= bitVal
	}
	return code
}

func boxEncoding(b box, callBack func(uint32, uint32, string) uint32) uint32 {
	code := uint32(0)
	hi, err := strconv.Atoi(b.HiBit)
	if err != nil {
		log.Fatalf("convert HiBit to int failed, HiBit = %s\n", b.HiBit)
	}
	for _, c := range b.Cs {
		if c.ColSpan != "" {
			if c.Value != "" && !strings.HasPrefix(c.Value, "!=") {
				log.Fatalf("malformed c, c.ColSpan = %v, c.Value = %v\n", c.ColSpan, c.Value)
			}
			colSpan, err := strconv.Atoi(c.ColSpan)
			if err != nil {
				log.Fatalf("convert ColSpan to int failed, ColSpan = %s\n", c.ColSpan)
			}
			hi -= colSpan
			continue
		}
		code = callBack(code, uint32(1<<hi), c.Value)
		hi--
	}
	return code
}

// extractBinary extracts the known bits of instruction encoding in regdiagram,
// and assign the binary to inst.regdiagram.binary.
func (inst *instruction) extractBinary() {
	for i, iclass := range inst.Classes.Iclass {
		bin, mask := uint32(0), uint32(0)
		for _, box := range iclass.Regdiagram.Boxes {
			bin |= boxEncoding(box, setBinary)
			mask |= boxEncoding(box, setMask)
		}
		inst.Classes.Iclass[i].Regdiagram.binary = bin
		inst.Classes.Iclass[i].Regdiagram.mask = mask
	}
}

// processEncoding handles each encoding element of a inst.
func (inst *instruction) processEncodings() {
	for i := range inst.Classes.Iclass {
		iclass := &inst.Classes.Iclass[i]
		// Don't use range because we may append new encodings.
		for j := 0; j < len(iclass.Encodings); j++ {
			enc := &iclass.Encodings[j]
			if !enc.parsed {
				// Set instruction class.
				enc.instClass()
				// Set instruction width.
				enc.instSize(iclass)
				// Refine the known bits and mask of the binary.
				bin, mask := iclass.Regdiagram.binary, iclass.Regdiagram.mask
				for _, box := range enc.Boxes {
					bin |= boxEncoding(box, setBinary)
					mask |= boxEncoding(box, setMask)
				}
				enc.binary = bin
				enc.mask = mask

				// Determine the Operands.
				// Check if a new element needs to be defined.
				if !enc.parseOperands() {
					continue
				}
				enc.combineOperands()
				enc.parsed = true

				// Split some WHILE* instructions into 32-bit and 64-bit forms, because the width
				// of the general purpose registers of these instructions matters, and the document
				// does not distinguish them by different encodings, but by <R>. For example:
				// WHILEGE  <Pd>.<T>, <R><n>, <R><m>
				// <R> can be W or X.
				// For the 32-bit form, we'll add a "W" suffix to its Go opcode.
				// We did not adjust the binary, mask and rules here, leaving it to the assembler.
				title := inst.Title
				switch title {
				case "WHILEGE (predicate)", "WHILEGT (predicate)", "WHILEHI (predicate)",
					"WHILEHS (predicate)", "WHILELE (predicate)", "WHILELO (predicate)",
					"WHILELS (predicate)", "WHILELT (predicate)",
					"CTERMEQ, CTERMNE":
					clone := enc.clone()
					clone.size = 32
					iclass.Encodings = append(iclass.Encodings, clone)
					enc = &iclass.Encodings[j]
				}
				// Split the optional pattern specifier <pattern> into separate operands.
				// All possible forms:
				// <Pd>.<T>{, <pattern>} => <Pd>.<T> and <Pd>.<T>, <pattern>
				// <Xd>{, <pattern>{, MUL #<imm>}} => <Xd> and <Xd>, <pattern>{, MUL #<imm>}
				// <Zdn>.D{, <pattern>{, MUL #<imm>}} => <Zdn>.D and <Zdn>.D, <pattern>{, MUL #<imm>}
				if asm := enc.asm; strings.Contains(asm, "<pattern>") {
					m := 1
					for ; m < len(enc.operands); m++ {
						if strings.Contains(enc.operands[m].name, "<pattern>") {
							break
						}
					}

					opr := &enc.operands[m]
					name := opr.name
					copr2 := operand{name: name}
					copr2.rules = make([]element, len(opr.rules))
					copy(copr2.rules, opr.rules)

					loc := strings.Index(name, "{,")       // index of the first "{,"
					opr.name = name[:loc]                  // <Pd>.<T> or <Xd> or <Zdn>.D
					copr2.name = name[loc+3 : len(name)-1] // <pattern> or <pattern>{, MUL #<imm>}
					if strings.Contains(name, "MUL") {
						// <Xd>{, <pattern>{, MUL #<imm>}} or <Zdn>.D{, <pattern>{, MUL #<imm>}}
						opr.rules = opr.rules[:1]
						copr2.rules = copr2.rules[1:]
					} else {
						// <Pd>.<T>{, <pattern>}
						opr.rules = opr.rules[:2]
						copr2.rules = copr2.rules[2:]
					}
					clone := enc.clone()
					// Set copr2 as the m+1 operand of clone.
					clone.operands = append(clone.operands, copr2)
					p, q := m+1, len(clone.operands)-1
					// All known cases so far m+1 is equal to len(clone.operands)-1,
					// so the following loop will not actually be executed.
					for p < q {
						clone.operands[p], clone.operands[q] = clone.operands[q], clone.operands[p]
					}
					iclass.Encodings = append(iclass.Encodings, clone)
					enc = &iclass.Encodings[j]
				}
			}
			// If opcode contains rules, then the opcode has different forms. For example:
			// B.<cond> can be BEQ, BGT etc. ADDHN{2} can be ADDHN and ADDHN2. We need to
			// split this encoding into different encodings by decoding the rules it contains.
			// Currently we only find three forms: B.<cond>, BFMLAL<bt> and <opcode>{2}.
			if len(enc.operands[0].rules) > 0 {
				opcode := enc.mnemonic()
				switch opcode {
				case "B", "BC": // B.<cond> and BC.<cond>
					if len(enc.operands[0].rules) != 1 || enc.operands[0].rules[0].Symbol != "<cond>" {
						log.Fatalf("unrecognized opcode rule: %v\n", enc.operands[0].rules)
					}
					suffixes := []string{
						"EQ", "NE", "CS", "HS", "CC", "LO", "MI", "PL",
						"VS", "VC", "HI", "LS", "GE", "LT", "GT", "LE",
					}
					// "<cond>" is encoded in "cond", bit 0-3.
					// cond        <cond>
					//-----------------
					// 0000   |    EQ
					// 0001   |    NE
					// 0010   |    CS or HS
					// 0011   |    CC or LO
					// 0100   |    MI
					// 0101   |    PL
					// 0110   |    VS
					// 0111   |    VC
					// 1000   |    HI
					// 1001   |    LS
					// 1010   |    GE
					// 1011   |    LT
					// 1100   |    GT
					// 1101   |    LE
					cond := uint32(1)
					enc.operands[0].rules = enc.operands[0].rules[:0] // clear opcode rules
					enc.mask |= 0xf
					enc.operands[0].name = opcode + suffixes[0]
					for k := 1; k < len(suffixes); k++ {
						enc = &iclass.Encodings[j] // enc may be expired due to stack growth.
						present := enc.clone()
						present.binary |= cond
						if k != 2 && k != 4 {
							// CS and HS have same encoding, CC and LO have same encoding.
							cond++
						}
						present.operands[0].name = opcode + suffixes[k]
						iclass.Encodings = append(iclass.Encodings, present)
					}
				case "BFMLAL": // BFMLAL<bt>
					if len(enc.operands[0].rules) != 1 || enc.operands[0].rules[0].Symbol != "<bt>" {
						log.Fatalf("unrecognized opcode rule: %v\n", enc.operands[0].rules)
					}
					suffixAbsent, suffixPresent := "B", "T"
					// "<bt>" is encoded in "Q", bit 30.
					// Q        <bt>
					//-----------------
					// 0   |    B
					// 1   |    T
					enc.operands[0].rules = enc.operands[0].rules[:0] // clear opcode rules
					enc.operands[0].name = opcode + suffixAbsent
					enc.mask |= 1 << 30

					present := enc.clone()
					present.binary |= 1 << 30
					present.operands[0].name = opcode + suffixPresent
					iclass.Encodings = append(iclass.Encodings, present)
				default: // <opcode>{2}, such as ADDHN{2}, FCVTL{2} etc.
					if len(enc.operands[0].rules) != 1 || enc.operands[0].rules[0].Symbol != "{2}" {
						log.Fatalf("unrecognized opcode rule: %v\n", enc.operands[0].rules)
					}
					suffix := "2"
					// "{2}" is encoded in "Q", bit 30.
					// Q        2
					//-----------------
					// 0   |    [absent]
					// 1   |    [present]
					enc.operands[0].rules = enc.operands[0].rules[:0] // clear opcode rules
					enc.operands[0].name = opcode
					enc.mask |= 1 << 30

					present := enc.clone()
					present.binary |= 1 << 30
					present.operands[0].name = opcode + suffix
					iclass.Encodings = append(iclass.Encodings, present)
				}
				// enc may be expired due to stack growth.
				enc = &iclass.Encodings[j]
			}

			enc.arm64Opcode()
			enc.template()
			enc.operandsType()
			enc.sortOperands()
			enc.goOpcode(iclass)
		}
	}
}

func (enc *encoding) clone() encoding {
	backup := *enc
	// Copy DocVars
	backup.DocVars = make([]docVar, len(enc.DocVars))
	copy(backup.DocVars, enc.DocVars)

	// Copy Boxes
	backup.Boxes = make([]box, len(enc.Boxes))
	copy(backup.Boxes, enc.Boxes)
	for i := 0; i < len(backup.Boxes); i++ {
		b := &backup.Boxes[i]
		b.Cs = make([]c, len(enc.Boxes[i].Cs))
		copy(b.Cs, enc.Boxes[i].Cs)
	}

	// Copy Asmtemplate
	backup.Asmtemplate.TextA = make([]textA, len(enc.Asmtemplate.TextA))
	copy(backup.Asmtemplate.TextA, enc.Asmtemplate.TextA)

	// Copy operands
	backup.operands = make([]operand, len(enc.operands))
	copy(backup.operands, enc.operands)
	for i := 0; i < len(enc.operands); i++ {
		opd := &backup.operands[i]
		opd.rules = make([]element, len(enc.operands[i].rules))
		copy(opd.rules, enc.operands[i].rules)
	}
	return backup
}

func (enc *encoding) parseOperands() bool {
	// This is the most vulnerable part.
	//
	// The mnemonic and operands of an instruction are sequentially recorded
	// in TextA, and we need to parse them out. According to the following rules:
	// 1. The mnemonic and the operand are separated by " ".
	// 2. The operands are separated by ", ". Symbols without intervals belong to
	//    the same operand.
	// 3, An operand may contain [] and {}, and the brackets contained in the
	//    operand must be in pairs. For example <R><m>{, <extend> {#<amount>}}
	//    is one operand not two.
	//
	// After this step we'll get all operands of this instruction encoding, the
	// operand interval symbol ", " will be discarded.
	asm, oprAsm := "", ""
	leftCurly, leftSquare := 0, 0
	elems := []element{}

	for m, ta := range enc.Asmtemplate.TextA {
		val := ta.Value
		if link := ta.Link; link != "" { // An <a> element
			encodedIn, candidates := processHover(ta.Hover)
			ruleName := link
			if encodedIn != "" {
				ruleName += "__" + encodedIn
			}
			// Some *_op symbols (such as at_op, dc_op) have a lot of candidates,
			// making the rule name extremely long. To avoid this situation,
			// the candidate values are not concatenated. Just make sure the rules
			// are unique.
			if candidates != "" && !strings.HasSuffix(link, "_op") && link != "sa_prfop" && link != "sa_pstatefield" {
				ruleName += "__" + candidates
			}
			val = strings.ReplaceAll(val, "&lt;", "<")
			val = strings.ReplaceAll(val, "&gt;", ">")
			elem := element{Name: ruleName, Link: link, Symbol: val}

			for !instsLock.CompareAndSwap(0, 1) {
				runtime.Gosched()
			}
			// Check if a new element needs to be defined.
			if _, ok := rules[ruleName]; !ok {
				rules[ruleName] = elem
			}
			instsLock.Store(0)
			elems = append(elems, elem)
		}
		asm += val

		// Parse operands
		for n := 0; n < len(val); n++ {
			ch := val[n]
			switch ch {
			case ',':
				if leftCurly == 0 && leftSquare == 0 {
					// This "," is an interval.
					continue
				}
			case ' ':
				if leftCurly == 0 && leftSquare == 0 {
					if oprAsm == "" {
						// Consecutive space separators.
						continue
					}
					if oprAsm == "MSL" && len(elems) == 0 {
						// Special cases:
						// MOVI <Vd>.<T>, #<imm8>, MSL #<amount>
						// MVNI <Vd>.<T>, #<imm8>, MSL #<amount>
						oprAsm += string(ch)
						continue
					}

					// This first one is mnemonic, followed by operands.
					rules := make([]element, len(elems))
					copy(rules, elems)
					opr := operand{name: oprAsm, rules: rules}
					enc.operands = append(enc.operands, opr)

					oprAsm = ""
					elems = elems[:0]
					continue
				}
			case '{':
				leftCurly++
			case '[':
				leftSquare++
			case '}':
				leftCurly--
			case ']':
				leftSquare--
			}
			oprAsm += string(ch)
		}
		// The last operand.
		if m == len(enc.Asmtemplate.TextA)-1 && leftCurly == 0 && leftSquare == 0 && oprAsm != "" {
			// Some instructions separate out the LSL extension format, such as:
			// LDRB  <Wt>, [<Xn|SP>, (<Wm>|<Xm>), <extend> {<amount>}]
			// LDRB  <Wt>, [<Xn|SP>, <Xm>{, LSL <amount>}]
			// But this will cause one Prog to match multiple instructions.
			// In order to avoid this situation, the two formats are merged.
			if oprAsm == "[<Xn|SP>, <Xm>{, LSL <amount>}]" {
				enc.invalid = true
				return false
			}
			// This first one is mnemonic, followed by operands.
			rules := make([]element, len(elems))
			copy(rules, elems)
			opr := operand{name: oprAsm, rules: rules}
			enc.operands = append(enc.operands, opr)
			oprAsm = ""
			elems = elems[:0]
		}
	}
	if oprAsm != "" || len(elems) != 0 {
		log.Fatalf("malformed Asmtemplate, oprAsm: %v, elems: %v\n", oprAsm, elems)
	}
	enc.asm = asm

	if false { // for debugging
		log.Printf("  Operands:\n")
		for k, op := range enc.operands {
			log.Printf("    Operands[%d]:\n", k)
			log.Printf("      Name: %v\n", op.name)
			log.Printf("      Rules:\n")
			for _, r := range op.rules {
				log.Printf("        Name: %v\n", r.Name)
				log.Printf("        Link: %v\n", r.Link)
				log.Printf("        Symbol: %v\n\n", r.Symbol)
			}
		}
	}
	return true
}

func combineTwoOperands(enc *encoding, i int) {
	enc.operands[i].name += ", " + enc.operands[i+1].name
	enc.operands[i].rules = append(enc.operands[i].rules, enc.operands[i+1].rules...)
	copy(enc.operands[i+1:len(enc.operands)-1], enc.operands[i+2:len(enc.operands)])
	enc.operands = enc.operands[:len(enc.operands)-1]
}

func combineRegPair(enc *encoding) {
	combined := false
	// The first operand is opcode
	for i := 1; i < len(enc.operands)-1; i++ {
		opr1, opr2 := enc.operands[i].name, enc.operands[i+1].name
		switch {
		case opr1 == "<Ws>" && opr2 == "<W(s+1)>",
			opr1 == "<Wt>" && opr2 == "<W(t+1)>",
			opr1 == "<Xs>" && opr2 == "<X(s+1)>",
			opr1 == "<Xt>" && opr2 == "<X(t+1)>",
			opr1 == "<Xt>" && opr2 == "<Xt+1>",
			opr1 == "<St1>" && opr2 == "<St2>",
			opr1 == "<Dt1>" && opr2 == "<Dt2>",
			opr1 == "<Qt1>" && opr2 == "<Qt2>",
			opr1 == "<Wt1>" && opr2 == "<Wt2>",
			opr1 == "<Xt1>" && opr2 == "<Xt2>":
			combineTwoOperands(enc, i)
			combined = true
		}
	}
	if !combined {
		log.Fatalf("Bad regex match: %v\n", enc.asm)
	}
}

func combineAddrOffset(enc *encoding) {
	size := len(enc.operands)
	if opr1 := enc.operands[size-2].name; opr1 != "[<Xn|SP>]" {
		log.Fatalf("Bad regex match: %v\n", enc.asm)
	}
	combineTwoOperands(enc, size-2)
}

func (enc *encoding) combineOperands() {
	// Special cases:
	// SYSP  #<op1>, <Cn>, <Cm>, #<op2>{, <Xt1>, <Xt2>}
	// TLBIP  <tlbip_op>{, <Xt1>, <Xt2>}
	switch enc.mnemonic() {
	case "SYSP", "TLBIP":
		return
	}
	var combineRules = []struct {
		reStr   string
		combine func(*encoding)
	}{
		// <Ws>, <W(s+1)>
		// <Wt>, <W(t+1)>
		// <Xs>, <X(s+1)>
		// <Xt>, <X(t+1)>
		// <Xt>, <Xt+1>
		{`<[WX][st]>,[\s]+<[WX][\(]?[st]\+1[\)]?`, combineRegPair},
		// <St1>, <St2>
		// <Dt1>, <Dt2>
		// <Qt1>, <Qt2>
		// <Wt1>, <Wt2>
		// <Xt1>, <Xt2>
		{`<[SDQWX][st]1>,[\s]+<[SDQWX][st]2>`, combineRegPair},
		// , [<Xn|SP>], #1
		// , [<Xn|SP>], #24
		// , [<Xn|SP>], <Xm>
		// , [<Xn|SP>], <imm>
		// , [<Xn|SP>], #<imm>
		// , [<Xn|SP>], #<simm>
		{`,[\s]+\[<Xn\|SP>\],[\s]+(#[1-9]+|<Xm>|[#]?<[s]?imm>)[\s]*$`, combineAddrOffset},
	}
	for i := 0; i < len(combineRules); i++ {
		if regExp := regexp.MustCompile(combineRules[i].reStr); regExp.MatchString(enc.asm) {
			combineRules[i].combine(enc)
		}
	}
}

// template resets the arm64 assembly template of an encoding, to make it cleaner.
func (enc *encoding) template() {
	asm := enc.operands[0].name
	if len(enc.operands) > 1 { // Has operands
		asm += "  "
		i := 1
		for ; i < len(enc.operands)-1; i++ {
			asm += enc.operands[i].name + ", "
		}
		asm += enc.operands[i].name
	}
	enc.asm = asm
}

func (enc *encoding) mnemonic() string {
	for _, docVar := range enc.DocVars {
		if docVar.Key == "alias_mnemonic" {
			return docVar.Value
		} else if docVar.Key == "mnemonic" {
			return docVar.Value
		}
	}
	log.Fatalf("Miss mnemonic: %v\n", enc)
	return ""
}

// arm64Opcode sets the arm64 opcode of an encoding.
func (enc *encoding) arm64Opcode() {
	if len(enc.operands) == 0 {
		log.Fatalf("Miss mnemonic: %v\n", enc)
	}
	// Add a prefix "A64_".
	enc.arm64Op = "A64_" + enc.operands[0].name
}

const (
	C_NONE int = iota
	C_GENERAL
	C_ADVSIMD
	C_FPSIMD
	C_FLOAT
	C_SVE
	C_SVE2
	C_MORTLACH
	C_MORTLACH2
	C_SYSTEM
)

func (enc *encoding) className() string {
	switch enc.class {
	case C_GENERAL:
		return "general"
	case C_ADVSIMD:
		return "advsimd"
	case C_FPSIMD:
		return "fpsimd"
	case C_FLOAT:
		return "float"
	case C_SVE:
		return "sve"
	case C_SVE2:
		return "sve2"
	case C_MORTLACH:
		return "mortlach"
	case C_MORTLACH2:
		return "mortlach2"
	case C_SYSTEM:
		return "system"
	}
	return ""
}

func (enc *encoding) instClass() {
	val := ""
	for _, d := range enc.DocVars {
		if d.Key == "instr-class" {
			val = d.Value
			break
		}
	}
	class := C_NONE
	switch val {
	case "general":
		class = C_GENERAL
	case "advsimd":
		class = C_ADVSIMD
	case "fpsimd":
		class = C_FPSIMD
	case "float":
		class = C_FLOAT
	case "sve":
		class = C_SVE
	case "sve2":
		class = C_SVE2
	case "mortlach":
		class = C_MORTLACH
	case "mortlach2":
		class = C_MORTLACH2
	case "system":
		class = C_SYSTEM
	case "":
		// Special cases. LDRAA and LDRAB miss the instr-class info.
		if opcode := enc.mnemonic(); opcode == "LDRAA" || opcode == "LDRAB" {
			class = C_GENERAL
		}
	default:
		fmt.Printf("unknow inst class %v\n", class)
	}
	enc.class = class
}

func getSize(attr string) int {
	size := 0
	switch {
	case strings.HasPrefix(attr, "8-bit"):
		size = 8
	case strings.HasPrefix(attr, "16-bit"),
		strings.HasPrefix(attr, "Half-precision"),
		strings.HasSuffix(attr, "16-bit elements"),
		strings.HasSuffix(attr, "half precision elements"):
		size = 16
	case strings.HasPrefix(attr, "32-bit"),
		strings.HasPrefix(attr, "Single-precision"),
		strings.HasSuffix(attr, "32-bit elements"),
		strings.HasSuffix(attr, "single precision elements"):
		size = 32
	case strings.HasPrefix(attr, "64-bit"),
		strings.HasPrefix(attr, "Double-precision"),
		strings.HasSuffix(attr, "64-bit elements"),
		strings.HasSuffix(attr, "double precision elements"):
		size = 64
	case strings.HasPrefix(attr, "128-bit"):
		size = 128
	}
	return size
}

// instSize judges the width of the instruction operation according to the Label and Name attributes.
func (enc *encoding) instSize(iclass *iclass) {
	size := getSize(enc.Label)
	if size == 0 {
		size = getSize(iclass.Name)
	}
	enc.size = size
}

func (enc *encoding) isSpecialOp() (string, string, string, bool) {
	spOp := true
	prefix, op, suffix := "", enc.operands[0].name, ""
	switch op {
	case "LDSETB", "LDSETAB", "LDSETALB", "LDSETLB",
		"LDSETH", "LDSETAH", "LDSETALH", "LDSETLH",
		"LDSETP", "LDSETAP", "LDSETALP", "LDSETLP",
		"STSETB", "STSETLB",
		"STSETH", "STSETLH":
		op = strings.Replace(op, "SET", "OR", 1) // Rename LDSET* as LDOR*
	case "LDSET", "LDSETA", "LDSETAL", "LDSETL",
		"STSET", "STSETL":
		op = strings.Replace(op, "SET", "OR", 1)
		fallthrough
	case "LDADD", "LDADDA", "LDADDAL", "LDADDL",
		"LDCLR", "LDCLRA", "LDCLRAL", "LDCLRL",
		"LDEOR", "LDEORA", "LDEORAL", "LDEORL",
		"LDSMAX", "LDSMAXA", "LDSMAXAL", "LDSMAXL",
		"LDSMIN", "LDSMINA", "LDSMINAL", "LDSMINL",
		"LDUMAX", "LDUMAXA", "LDUMAXAL", "LDUMAXL",
		"LDUMIN", "LDUMINA", "LDUMINAL", "LDUMINL",
		"SWP", "SWPA", "SWPAL", "SWPL",
		"CAS", "CASA", "CASAL", "CASL",
		"CASP", "CASPA", "CASPAL", "CASPL",
		"STADD", "STADDL",
		"STCLR", "STCLRL",
		"STEOR", "STEORL",
		"STSMAX", "STSMAXL",
		"STSMIN", "STSMINL",
		"STUMAX", "STUMAXL",
		"STUMIN", "STUMINL":
		// Add a "D" and "W" suffix for part of 64-bit and 32-bit atomic instructions, respectively.
		// ST64B, ST64BV, ST64BV0 and FEAT_THE instructions are not suffixed.
		if enc.size == 64 {
			suffix = "D"
		} else {
			suffix = "W"
		}
	case "LDR", "STR", "LDUR", "STUR":
		class := enc.class
		if class == C_SVE || class == C_SVE2 || class == C_MORTLACH || class == C_MORTLACH2 {
			if enc.hasPREG() {
				prefix = "P"
			} else {
				prefix = "Z"
			}
			break
		}
		op = "MOV"
		switch enc.size {
		case 8:
			prefix, suffix = "F", "B"
		case 16:
			prefix, suffix = "F", "H"
		case 32:
			if class == C_GENERAL {
				suffix = "WU"
			} else if class == C_ADVSIMD || class == C_FPSIMD || class == C_FLOAT {
				prefix, suffix = "F", "S"
			}
		case 64:
			if class == C_GENERAL {
				suffix = "D"
			} else if class == C_ADVSIMD || class == C_FPSIMD || class == C_FLOAT {
				prefix, suffix = "F", "D"
			}
		case 128:
			prefix, suffix = "F", "Q"
		}
	case "LDRB", "LDURB":
		op, suffix = "MOV", "BU"
	case "LDRH", "LDURH":
		op, suffix = "MOV", "HU"
	case "LDRSB":
		op = "MOV" // 32-bit
		if enc.size == 64 {
			suffix = "B"
		} else {
			suffix = "BW"
		}
	case "LDRSH":
		op = "MOV" // 32-bit
		if enc.size == 64 {
			suffix = "H"
		} else {
			suffix = "HW"
		}
	case "LDRSW":
		op, suffix = "MOV", "W"
	case "STRB", "STURB":
		op, suffix = "MOV", "B"
	case "STRH", "STURH":
		op, suffix = "MOV", "H"
	case "MOV":
		size := enc.size
		switch enc.class {
		case C_GENERAL:
			if size == 64 {
				suffix = "D"
			} else {
				suffix = "W"
			}
		default:
			spOp = false
		}
	case "MOVI":
		// MOVI instruction has different widths. It would be best if different
		// opcode suffixes are set according to the instruction width, but since
		// the old code did not do this, in order to maintain backward compatibility,
		// we have to keep it as it is.
		prefix = "V"
	case "NOP":
		op = "NOOP"
	case "UMOV":
		op = "VMOV"
	case "UXTB", "UXTH":
		// The general version of the UXTB and UXTH instructions have a "32-bit" attribute,
		// treat them as special cases, otherwise they will be suffixed with "W".
		if enc.class != C_GENERAL {
			spOp = false
		}
	case // Leave cryptographic extension instructions unchanged.
		"AESD", "AESE", "AESIMC", "AESMC",
		"SHA1C", "SHA1H", "SHA1M", "SHA1P", "SHA1SU0", "SHA1SU1",
		"SHA256H", "SHA256H2", "SHA256SU0", "SHA256SU1",
		"SHA512H", "SHA512H2", "SHA512SU0", "SHA512SU1",
		"SM3PARTW1", "SM3PARTW2", "SM3SS1", "SM3TT1A", "SM3TT1B", "SM3TT2A", "SM3TT2B",
		"SM4E", "SM4EKEY",
		// Leave SVE predicate instructions unchanged.
		"PEXT", "PFALSE", "PFIRST", "PMOV", "PNEXT", "PSEL", "PTEST", "PTRUE", "PTRUES", "PUNPKHI", "PUNPKLO":
	case "WHILEGE", "WHILEGT", "WHILEHI", "WHILEHS",
		"WHILELE", "WHILELO", "WHILELS", "WHILELT",
		"CTERMEQ", "CTERMNE":
		if enc.size == 32 {
			suffix = "W"
		}
	default:
		spOp = false
	}
	return prefix, op, suffix, spOp
}

func (enc *encoding) hasZREG() bool {
	rExp := `<Z[A-Za-z1-9]+>`
	regExp := regexp.MustCompile(rExp)
	return regExp.MatchString(enc.asm)
}

func (enc *encoding) hasPREG() bool {
	rExp := `<P[A-Za-z1-9]+>`
	regExp := regexp.MustCompile(rExp)
	return regExp.MatchString(enc.asm)
}

// For historical reasons, we set different suffixes for some conversion instructions.
func specialSuffix(attr string) string {
	suffix := ""
	switch attr {
	case "Half-precision", "Half-precision, zero":
		suffix = "H"
	case "Single-precision", "Single-precision, zero":
		suffix = "S"
	case "Double-precision", "Double-precision, zero":
		suffix = "D"
	case "Half-precision to 16-bit":
		suffix = "HH"
	case "Half-precision to 32-bit":
		suffix = "HW"
	case "Half-precision to 64-bit":
		suffix = "H"
	case "Single-precision to 32-bit":
		suffix = "SW"
	case "Single-precision to 64-bit":
		suffix = "S"
	case "Double-precision to 32-bit":
		suffix = "DW"
	case "Double-precision to 64-bit":
		suffix = "D"
	case "16-bit to half-precision":
		suffix = "HH"
	case "32-bit to half-precision":
		suffix = "WH"
	case "32-bit to single-precision":
		suffix = "WS"
	case "32-bit to double-precision":
		suffix = "WD"
	case "64-bit to half-precision":
		suffix = "H"
	case "64-bit to single-precision":
		suffix = "S"
	case "64-bit to double-precision":
		suffix = "D"
	case "Half-precision to single-precision":
		suffix = "HS"
	case "Half-precision to double-precision":
		suffix = "HD"
	case "Single-precision to half-precision":
		suffix = "SH"
	case "Single-precision to double-precision":
		suffix = "SD"
	case "Double-precision to half-precision":
		suffix = "DH"
	case "Double-precision to single-precision":
		suffix = "DS"
	}
	return suffix
}

func (enc *encoding) goOpcodeSuffix(iclass *iclass) string {
	suffix := ""
	switch enc.class {
	case C_GENERAL:
		if enc.size == 32 {
			suffix = "W"
		}
	default:
		// Special cases.
		suffix = specialSuffix(enc.Label)
		if suffix == "" {
			suffix = specialSuffix(iclass.Name)
		}
		if suffix == "" {
			switch enc.size {
			case 8:
				suffix = "B"
			case 16:
				suffix = "H"
			case 32:
				suffix = "S"
			case 64:
				suffix = "D"
			case 128:
				suffix = "Q"
			}
		}
	}
	// Add a "H" suffix for the FEAT_FP16 feature because some FEAT_FP16
	// instruction formats are exactly the same as the non FEAT_FP16
	// instruction format, it's impossible to distinguish them.
	if suffix == "" && iclass.ArchVariant.Feature == "FEAT_FP16" {
		suffix = "H"
	}
	return suffix
}

func (enc *encoding) goOpcodePrefix(iclass *iclass) string {
	prefix := ""
	switch enc.class {
	case C_ADVSIMD:
		prefix = "V"
	case C_FPSIMD:
		prefix = "F"
	case C_SVE, C_SVE2, C_MORTLACH, C_MORTLACH2:
		if enc.hasZREG() {
			prefix = "Z"
		} else if enc.hasPREG() {
			prefix = "P"
		}
	}
	return prefix
}

// goOpcode determines the Go opcode representation of an encoding.
func (enc *encoding) goOpcode(iclass *iclass) {
	if len(enc.operands) == 0 {
		log.Fatalf("Miss mnemonic: %v\n", enc)
	}
	prefix, opcode, suffix := "A", "", ""
	// Special cases.
	if pre, op, suf, spOp := enc.isSpecialOp(); spOp {
		prefix += pre
		opcode = op
		suffix = suf
	} else {
		prefix += enc.goOpcodePrefix(iclass)
		opcode = enc.operands[0].name
		suffix = enc.goOpcodeSuffix(iclass)
	}
	enc.goOp = prefix + opcode + suffix
	enc.prefix, enc.suffix = prefix, suffix
}

func (enc *encoding) specialArgs() bool {
	special := true
	switch enc.operands[0].name {
	case "CBNZ", "CBZ",
		"GCSSTR", "GCSSTTR", "ST2G",
		"ST64B", "ST64BV", "ST64BV0",
		"STADD", "STADDL", "STADDB", "STADDLB", "STADDH", "STADDLH",
		"STCLR", "STCLRL", "STCLRB", "STCLRLB", "STCLRH", "STCLRLH",
		"STEOR", "STEORL", "STEORB", "STEORLB", "STEORH", "STEORLH",
		"STG", "STGM", "STGP",
		"STILP", "STLLR", "STLLRB", "STLLRH",
		"STLR", "STLRB", "STLRH",
		"STLUR", "STLURB", "STLURH",
		"STNP", "STP",
		"STR", "STRB", "STRH",
		"STTR", "STTRB", "STTRH",
		"STUR", "STURB", "STURH",
		"STZ2G", "STZG", "STZGM",
		"ST1", "ST2", "ST3", "ST4", "STL1",
		"ST1B", "ST1D", "ST1H", "ST1Q", "ST1W",
		"ST2B", "ST2D", "ST2H", "ST2Q", "ST2W",
		"ST3B", "ST3D", "ST3H", "ST3Q", "ST3W",
		"ST4B", "ST4D", "ST4H", "ST4Q", "ST4W",
		"STNT1B", "STNT1D", "STNT1H", "STNT1W":
		// Argument order is the same as the ARM64 syntax:
		// cbz, cbnz and store instructions, do nothing.
	case "STLXR", "STLXRB", "STLXRH", "STXR", "STXRB", "STXRH", "STLXP", "STXP":
		// stlxr w16, xzr, [x15] <=> STLXR ZR, (R15), R16
		// stlxp w5, x17, x19, [x4] <=> STLXP (R17, R19), (R4), R5
		enc.operands[1], enc.operands[2] = enc.operands[2], enc.operands[1]
		enc.operands[2], enc.operands[3] = enc.operands[3], enc.operands[2]
	case "MADD", "MSUB", "SMADDL", "SMSUBL", "UMADDL", "UMSUBL",
		"FMADD", "FMSUB", "FNMADD", "FNMSUB":
		// madd x1, x2, x3, x4 <=> MADD R3, R4, R2, R1
		// fmadd d1, d2, d3, d4 <=> FMADDD F3, F4, F2, F1
		enc.operands[1], enc.operands[3] = enc.operands[3], enc.operands[1]
		enc.operands[2], enc.operands[4] = enc.operands[4], enc.operands[2]
		enc.operands[3], enc.operands[4] = enc.operands[4], enc.operands[3]
	case "BFI", "BFM", "BFXIL",
		"SBFM", "SBFIZ", "SBFX",
		"UBFM", "UBFIZ", "UBFX":
		// bfi w0, w20, #16, #6 <=> BFIW $16, R20, $6, R0
		enc.operands[1], enc.operands[3] = enc.operands[3], enc.operands[1]
		enc.operands[3], enc.operands[4] = enc.operands[4], enc.operands[3]
	case "FCCMP", "FCCMPE":
		// fccmp d26, d8, #0x0, al <=> FCCMPD AL, F8, F26, $0
		enc.operands[1], enc.operands[4] = enc.operands[4], enc.operands[1]
		enc.operands[3], enc.operands[4] = enc.operands[4], enc.operands[3]
	case "CCMP", "CCMN":
		// ccmp w19, w14, #0xb, cs <=> CCMPW HS, R19, R14, $11
		enc.operands[1], enc.operands[2] = enc.operands[2], enc.operands[1]
		enc.operands[3], enc.operands[4] = enc.operands[4], enc.operands[3]
		enc.operands[1], enc.operands[3] = enc.operands[3], enc.operands[1]
	case "CSEL", "CSINC", "CSINV", "CSNEG", "FCSEL":
		// csel x1, x0, x19, gt <=> CSEL GT, R0, R19, R1
		enc.operands[1], enc.operands[4] = enc.operands[4], enc.operands[1]
	case "TBNZ", "TBZ":
		// tbz x1, #4, loop <=> TBZ $4, R1, loop
		enc.operands[1], enc.operands[2] = enc.operands[2], enc.operands[1]
	case "MOVI", "MVNI":
		if strings.Contains(enc.asm, "MSL") {
			// movi <Vd>.<T>, #<imm8>, MSL #<amount> <=> MOVI #<imm8>, MSL #<amount>, <Vd>.<T>
			enc.operands[1], enc.operands[2] = enc.operands[2], enc.operands[1]
			enc.operands[2], enc.operands[3] = enc.operands[3], enc.operands[2]
		} else {
			special = false
		}
	default:
		// atomic instructions with 3 operands, ST64BV and ST64BV0.
		// swpa x5, x7, [x6]  <=> SWPAD	R5, (R6), R7
		// cas  w5, w6, [x7]  <=> CASW	R5, (R7), R6
		regStr := `<[XW]s>,.*<Xn\|SP>`
		if regExp := regexp.MustCompile(regStr); regExp.MatchString(enc.asm) && enc.class == C_GENERAL && len(enc.operands) > 3 {
			enc.operands[2], enc.operands[3] = enc.operands[3], enc.operands[2]
		} else {
			special = false
		}
	}
	return special
}

// sortOperands reorders the operands of an encoding according to Go assembly syntax.
func (enc *encoding) sortOperands() {
	if enc.specialArgs() {
		return
	}
	// Reverse args, placing dest last.
	for i, j := 1, len(enc.operands)-1; i < j; i, j = i+1, j-1 {
		enc.operands[i], enc.operands[j] = enc.operands[j], enc.operands[i]
	}
}

func operandType(opr operand) string {
	name := opr.name
	if strings.HasPrefix(name, "#") {
		return "AC_IMM" // Constant.
	}
	switch name {
	case "<Cm>", "<Cn>", "<const>", "MSL #<amount>", "{#<imm>}", "{ <mask> }":
		return "AC_IMM" // Constants in system instructions, such as SYS  #<op1>, <Cn>, <Cm>, #<op2>{, <Xt>}
	case "(<prfop>|#<imm5>)", "(<rprfop>|#<imm6>)", "<at_op>", "<brb_op>{, <Xt>}",
		"<dc_op>", "<prfop>", "<tlbi_op>{, <Xt>}", "<tlbip_op>{, <Xt1>, <Xt2>}", "<ic_op>{, <Xt>}",
		"<option>nXS", "<option>|#<imm>", "{<option>|#<imm>}", "{<option>}",
		"<pstatefield>", "<vl>", "CSYNC", "DSYNC", "RCTX", "{<targets>}",
		"<pattern>", "<pattern>{, MUL #<imm>}":
		return "AC_SPOP" // Special operands, such as system registers.
	case "(<systemreg>|S<op0>_<op1>_<Cn>_<Cm>_<op2>)":
		return "AC_SPR"
	case "<cond>":
		return "AC_COND"
	case "<label>":
		return "AC_LABEL"
	case "<Wm>{, <shift> #<amount>}", "<Xm>{, <shift> #<amount>}":
		return "AC_REGSHIFT"
	case "<Wm>{, <extend> {#<amount>}}", "<R><m>{, <extend> {#<amount>}}":
		return "AC_REGEXT"
	case "ZA[<Wv>, <offs>]":
		return "AC_ZAVECTORIDX"
	case "<Xn|SP>{, <Xm>}", "<Pm>.<T>[<Wv>, <imm>]":
		return "AC_ANY"
	}
	var operandsRules = []struct {
		reStr string
		class string
	}{
		// <R><d>
		// <Ws>
		// <Wt>
		// X16
		// {<Xn>}
		// <Xn>!
		{`^(<[WX][a-z]+>!?|<R><[a-z]+>|X[0-9]+|{<[WX][a-z]+>})$`, "AC_REG"},
		// <R><n|SP>
		// <Wd|WSP>
		// <Xd|SP>
		{`^<([WX][a-z]{1}|R><n)\|[W]?SP>$`, "AC_RSP"},
		// <Bt>
		// <Da>
		// <Ha>
		// <Qd>
		// <Sm>
		{`^<[BDHQS][a-z]{1}>$`, "AC_FREG"},
		// <V><d>
		// <V><dn>
		// <Va><d>
		{`^<V[a-z]?><[a-z]+>$`, "AC_VREG"},
		// <PNg>
		// <Pd>
		{`^<P[N]?[a-z]{1}>$`, "AC_PREG"},
		// <Pg>/M
		{`^<P[N]?[a-z]{1}>\/M$`, "AC_PREGM"},
		// <Pg>/<ZM>
		// <Pg>/Z
		{`^<P[N]?[a-z]{1}>\/(Z|<ZM>)$`, "AC_PREGZ"},
		// <PNn>[<imm>]
		// <Zd>[<imm>]
		// ZT0[<offs>]
		{`^(<[PZ][N]?[a-z]{1}>|ZT0)\[<[a-z]+>\]$`, "AC_REGIDX"},
		// <Zd>
		{`^<Z[a-z]{1}>$`, "AC_ZREG"},
		// ZT0
		// { ZT0 }
		{`^(ZT0|{ ZT0 })$`, "AC_ZTREG"},
		// <Wt>, <W(t+1)>
		// <Xs>, <X(s+1)>
		// <Xt>, <Xt+1>
		{`^<[WX][st]>,[\s]+<[WX]\(?[st]\+1\)?>$`, "AC_PAIR"},
		// <St1>, <St2>
		// <Dt1>, <Dt2>
		// <Qt1>, <Qt2>
		// <Wt1>, <Wt2>
		// <Xt1>, <Xt2>
		{`^<[SDQWX][st]1>,[\s]+<[SDQWX][st]2>$`, "AC_PAIR"},
		// <PNd>.<T>
		// <Pd>.<T>
		// <Pd>.B
		// <Va>.16B
		// <Vn>.2D
		// <Vd>.<Ta>
		// <ZAda>.D
		// <Zm>.<Tb>
		{`^<[PVZ][a-zA-Z]+>\.([1-9]*[BDHQS]|<T[a-z]*>)$`, "AC_ARNG"},
		// <Vd>.<Ts>[<index1>]
		// <Vd>.D[1]
		// <Vm>.2H[<index>]
		// <Vm>.S[<imm2>]
		// <Zm>.B[<index>]
		// <Zn>.<T>[<imm>]
		{`^<[VZ][a-zA-Z]*>\.([1-9]*[BDHQS]|<T[a-z]*>)\[(<(index|imm)[1-9]*>|[0-9]+)\]$`, "AC_ARNGIDX"},
		// { <Vn>.16B }
		// { <Zn>.<T> }
		// { <Zt>.B }
		{`^{[\s]+<[PVZ][a-z]+>\.([1-9]*[BDHQS]|<T[a-z]*>)[\s]+}$`, "AC_REGLIST1"},
		// { <Pd1>.<T>, <Pd2>.<T> }
		// { <Vn>.16B, <Vn+1>.16B }
		{`^{[\s]+(<[PVZ][a-z]+([1-2]|\+[1-2])*>\.([1-9]*[BDHQS]|<T[a-z]*>),*[\s]*){2}}$`, "AC_REGLIST2"},
		// { <Zd1>.B-<Zd2>.B }
		// { <Zm1>.<T>-<Zm2>.<T> }
		{`^{[\s]+(<[PVZ][a-z]+[1-2]*>\.([1-9]*[BDHQS]|<T[a-z]*>)-*[\s]*){2}}$`, "AC_REGLIST2C"},
		// { <Vn>.16B, <Vn+1>.16B, <Vn+2>.16B }
		// { <Vt>.<T>, <Vt2>.<T>, <Vt3>.<T> }
		// { <Zt1>.B, <Zt2>.B, <Zt3>.B }
		{`^{[\s]+(<[PVZ][a-z]+([1-3]|\+[1-3])*>\.([1-9]*[BDHQS]|<T[a-z]*>)(,|-)*[\s]*){3}}$`, "AC_REGLIST3"},
		// { <Vn>.16B, <Vn+1>.16B, <Vn+2>.16B, <Vn+3>.16B }
		// { <Zd1>.<T>, <Zd2>.<T>, <Zd3>.<T>, <Zd4>.<T> }
		// { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }
		{`^{[\s]+(<[PVZ][a-z]+([1-4]|\+[1-4])*>\.([1-9]*[BDHQS]|<T[a-z]*>),*[\s]*){4}}$`, "AC_REGLIST4"},
		// { <Zd1>.<T>-<Zd4>.<T> }
		// { <Zd1>.D-<Zd4>.D }
		// { <Zn1>.<Tb>-<Zn4>.<Tb> }
		{`^{[\s]+(<[PVZ][a-z]+[14]>\.([BDHQS]|<T[a-z]*>)-*[\s]*){2}}$`, "AC_REGLIST4C"},
		// { <Vt>.B }[<index>]
		// { <Vt>.B, <Vt2>.B, <Vt3>.B }[<index>]
		{`^{[\s]+(<[PVZ][a-z]+[1-4]*>\.[BDHQS],*[\s]*){1,4}}\[<index>\]$`, "AC_LISTIDX"},
		// [<Xn|SP> {,#0}]
		// [<Xn|SP>]
		// [<Xn|SP>{, #<imm>}]
		{`^\[<Xn\|SP>([\s]*\{,[\s]*#([0-9]+|<[a-z]+>)\})*\]$`, "AC_MEMIMM"},
		// [<Zn>.D{, #<imm>}]
		{`^\[<Z[a-z]+>\.[BDHQS]\{,[\s]*#<[a-z]+>\}\]$`, "AC_MEMIMM"},
		// [<Xn|SP>{, #<imm>, MUL VL}]
		{`^\[<Xn\|SP>[\s]*\{,[\s]*#<[a-z]+>,[\s]*MUL[\s]+VL[\s]*\}\]$`, "AC_MEMIMMEXT"},
		// [<Xn|SP>], <imm>
		// [<Xn|SP>], #16
		// [<Xn|SP>], #<imm>
		{`^\[<Xn\|SP>\][\s]*,[\s]*(#[0-9]+|#?<[a-z]+>)$`, "AC_MEMPOSTIMM"},
		// [<Xn|SP>], <Xm>
		{`^\[<Xn\|SP>\][\s]*,[\s]*<X[a-z]+>$`, "AC_MEMPOSTREG"},
		// [<Xn|SP>, <Xm>]
		// [<Xn|SP>, <Zm>.D]
		{`^\[<Xn\|SP>,[\s]*(<X[a-z]+>|<Z[a-z]+>\.[BDHQS])\]$`, "AC_MEMEXT"},
		// [<Xn|SP>, (<Wm>|<Xm>), <extend> {<amount>}]
		// [<Xn|SP>, (<Wm>|<Xm>){, <extend> {<amount>}}]
		{`^\[<Xn\|SP>,[\s]*\(<W[a-z]+>\|<X[a-z]+>\)\{?,[\s]*<extend>[\s]*\{<amount>\}\}?\]$`, "AC_MEMEXT"},
		// [<Xn|SP>, <Xm>, LSL #1]
		// [<Xn|SP>, <Xm>{, LSL <amount>}]
		// [<Xn|SP>{, <Xm>, LSL #1}]
		{`^\[<Xn\|SP>\{?,[\s]*<X[a-z]+>\{?,[\s]*LSL[\s]+(<amount>|#[0-9]+)\}?\]$`, "AC_MEMEXT"},
		// [<Xn|SP>{, <Xm>}]
		// [<Zn>.D{, <Xm>}]
		{`^\[(<Xn\|SP>|<Z[a-z]+>\.[BDHQS])\{,[\s]*<X[a-z]+>\}\]$`, "AC_MEMEXT"},
		// [<Xn|SP>, <Zm>.D, <mod> #1]
		// [<Xn|SP>, <Zm>.D, <mod>]
		// [<Xn|SP>, <Zm>.D, LSL #1]
		{`^\[<Xn\|SP>,[\s]*<Z[a-z]+>\.[BDHQS],[\s]*(<mod>([\s]+#[0-9]+)*|LSL[\s]+#[0-9]+)\]$`, "AC_MEMEXT"},
		// [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
		// [<Zn>.D, <Zm>.D, SXTW{ <amount>}]
		// [<Zn>.D, <Zm>.D, UXTW{ <amount>}]
		{`^\[<Z[a-z]+>\.(<T>|[BDHQS]),[\s]*<Z[a-z]+>\.(<T>|[BDHQS])\{?,[\s]*(<mod>|SXTW\{?|UXTW\{?)[\s]+<amount>\}?\]$`, "AC_MEMEXT"},
		// [<Xd>]!
		{`^\[<X[a-z]+>\]!$`, "AC_MEMPREIMM"},
		// [<Xn|SP>, #<imm>]!
		// [<Xn|SP>{, #<simm>}]!
		{`^\[<Xn\|SP>\{?,[\s]*#(<[a-z]+>|-?[0-9]+)\}?\]!$`, "AC_MEMPREIMM"},
		// <ZAd><HV>.D[<Ws>, <offs>]
		// ZA0<HV>.B[<Ws>, <offs>]
		// { <ZAt><HV>.D[<Ws>, <offs>] }
		// { ZA0<HV>.B[<Ws>, <offs>] }
		{`^\{?[\s]*(<ZA[a-z]+>|ZA[0-9]+)<HV>\.[BDHQS]\[<W[a-z]+>,[\s]*<[a-z]+>\][\s]*\}?$`, "AC_ZAHVTILEIDX"},
		// <ZAd><HV>.D[<Ws>, <offsf>:<offsl>]
		// ZA0<HV>.B[<Ws>, <offsf>:<offsl>]
		{`^(<ZA[a-z]+>|ZA[0-9]+)<HV>\.[BDHQS]\[<W[a-z]+>,[\s]*<[a-z]+>:<[a-z]+>\]$`, "AC_ZAHVTILESEL"},
		// ZA.<T>[<Wv>, <offs>{, VGx2}]
		// ZA.D[<Wv>, <offs>, VGx2]
		// ZA.H[<Wv>, <offs>{, VGx2}]
		{`^ZA\.(<T>|[BDHQS])\[<W[a-z]+>,[\s]*<[a-z]+>\{?,[\s]*VGx2\}?\]$`, "AC_ZAVECTORIDXVG2"},
		// ZA.<T>[<Wv>, <offs>{, VGx4}]
		// ZA.D[<Wv>, <offs>, VGx4]
		// ZA.H[<Wv>, <offs>{, VGx4}]
		{`^ZA\.(<T>|[BDHQS])\[<W[a-z]+>,[\s]*<[a-z]+>\{?,[\s]*VGx4\}?\]$`, "AC_ZAVECTORIDXVG4"},
		// ZA.<T>[<Wv>, <offsf>:<offsl>]
		// ZA.D[<Wv>, <offsf>:<offsl>]
		{`^ZA\.(<T>|[BDHQS])\[<W[a-z]+>,[\s]*<[a-z]+>:<[a-z]+>\]$`, "AC_ZAVECTORSEL"},
		// ZA.<T>[<Wv>, <offsf>:<offsl>{, VGx2}]
		// ZA.D[<Wv>, <offsf>:<offsl>{, VGx2}]
		{`^ZA\.(<T>|[BDHQS])\[<W[a-z]+>,[\s]*<[a-z]+>:<[a-z]+>\{?,[\s]*VGx2\}?\]$`, "AC_ZAVECTORSELVG2"},
		// ZA.<T>[<Wv>, <offsf>:<offsl>{, VGx4}]
		// ZA.D[<Wv>, <offsf>:<offsl>{, VGx4}]
		{`^ZA\.(<T>|[BDHQS])\[<W[a-z]+>,[\s]*<[a-z]+>:<[a-z]+>\{?,[\s]*VGx4\}?\]$`, "AC_ZAVECTORSELVG4"},
	}
	for i := 0; i < len(operandsRules); i++ {
		if regExp := regexp.MustCompile(operandsRules[i].reStr); regExp.MatchString(name) {
			return operandsRules[i].class
		}
	}
	log.Fatalf("unrecognized operand type: %s\n", opr.name)
	return "AC_NONE"
}

// operandsType classifies all operands of an encoding.
func (enc *encoding) operandsType() {
	for i := 1; i < len(enc.operands); i++ {
		enc.operands[i].typ = operandType(enc.operands[i])
	}
}

func trimField(field string) string {
	field = strings.TrimSpace(field)
	field = strings.TrimPrefix(field, "\"")      // left &quot;
	field = strings.TrimSuffix(field, "\"")      // right &quot;
	field = strings.ReplaceAll(field, ":", "_")  // cmode<2:1>
	field = strings.ReplaceAll(field, "<", "_")  // cmode<2:1>
	field = strings.ReplaceAll(field, ">", "_")  // cmode<2:1>
	field = strings.ReplaceAll(field, "'", "_")  // D_'0'_Zd
	field = strings.ReplaceAll(field, "\"", "_") // "Rn" and "Rm"
	field = strings.ReplaceAll(field, " ", "_")  // "Rn" and "Rm"
	field = strings.TrimRight(field, "_")
	return field
}

func trimVals(vals string) string {
	vals = strings.TrimSpace(vals)
	// Values are separated by ",".
	fields := strings.Split(vals, ",")
	res := ""
	for _, field := range fields {
		field = strings.TrimSpace(field)
		field = strings.TrimPrefix(field, "#")      // #3
		field = strings.ReplaceAll(field, ":", "_") // H:L:M
		field = strings.ReplaceAll(field, "-", "_") // (32-Uint(immh:immb))
		field = strings.ReplaceAll(field, "(", "_") // (32-Uint(immh:immb))
		field = strings.ReplaceAll(field, ")", "_") // (32-Uint(immh:immb))
		field = strings.ReplaceAll(field, " ", "_") // LSL #12
		field = strings.ReplaceAll(field, "#", "_") // LSL #12
		res += field + "_"
	}
	res = strings.TrimRight(res, "_")
	return res
}

// processHover extracts the encodedin and candidate values information in hover.
func processHover(hover string) (field, vals string) {
	fieldStartMark := " (field "
	start := strings.LastIndex(hover, fieldStartMark)
	if start < 0 {
		return
	}
	hover = hover[start+len(fieldStartMark):]
	end := strings.Index(hover, ")") // field ends with ")"
	field = hover[:end]
	field = trimField(field)

	start = strings.Index(hover, "[") // vals starts with "["
	if start < 0 {
		return
	}
	end = strings.Index(hover, "]") // vals ends with "]"
	vals = hover[start+1 : end]
	vals = trimVals(vals)
	return
}

func processXMLFiles() {
	// Init rules.
	rules = make(map[string]element)

	var wg sync.WaitGroup
	for _, inst := range insts {
		wg.Add(1)
		go func(inst *instruction) {
			defer wg.Done()
			inst.extractBinary()
			inst.processEncodings()
		}(inst)
	}
	wg.Wait()
}
