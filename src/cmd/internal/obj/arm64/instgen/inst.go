// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync/atomic"
)

type docVar struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type classesintro struct {
	Count string `xml:"count,attr"`
}

type iclassintro struct {
	Count string `xml:"count,attr"`
}

type archVariant struct {
	Name    string `xml:"name,attr"`
	Feature string `xml:"feature,attr"`
}

type c struct {
	Value   string `xml:",chardata"`
	ColSpan string `xml:"colspan,attr"`
}

type box struct {
	HiBit    string `xml:"hibit,attr"`
	Width    string `xml:"width,attr"`
	Name     string `xml:"name,attr"`
	UseName  string `xml:"usename,attr"`
	Settings string `xml:"settings,attr"`
	PsBits   string `xml:"psbits,attr"`
	Cs       []c    `xml:"c"`
}

type regdiagram struct {
	Boxes  []box  `xml:"box"`
	binary uint32 // instruction encoding binary
	mask   uint32 // instruction decoding mask
}

type textA struct {
	Value string `xml:",chardata"`
	Link  string `xml:"link,attr"`
	Hover string `xml:"hover,attr"` // contains possible values
}

type asmtemplate struct {
	// <asmtemplate> contains two kinds of sub-element, <text> and <a>.
	// <text> contains some string literals, <a> contains a symbol and
	// two attributes link and hover. The order of <text> and <a> matters,
	// so we save both of them into the following structure, then we get
	// an ordered sub-element slice.
	TextA []textA `xml:",any"`
}

type element struct {
	Name   string // rule name
	Link   string // link to symbol
	Symbol string // asm template
}

type operand struct {
	name  string // asm template
	typ   string
	rules []element
}

type encoding struct {
	Name        string      `xml:"name,attr"`
	Label       string      `xml:"label,attr"`
	DocVars     []docVar    `xml:"docvars>docvar"`
	Boxes       []box       `xml:"box"`
	Asmtemplate asmtemplate `xml:"asmtemplate"`
	parsed      bool
	binary      uint32 // more specific instruction encoding than regdiagram.binary
	mask        uint32
	asm         string // asm template
	goOp        string // opcode in Go
	arm64Op     string // arm64 opcode
	operands    []operand
	class       int  // instruction class
	size        int  // instruction width
	invalid     bool // indicate if this is a valid encoding that need to print
	alias       bool // whether it is an alias
	prefix      string
	suffix      string
}

type iclass struct {
	Name        string      `xml:"name,attr"`
	OneOf       string      `xml:"oneof,attr"`
	ID          string      `xml:"id,attr"`
	NoEncodings string      `xml:"no_encodings,attr"`
	ISA         string      `xml:"isa,attr"`
	DocVars     []docVar    `xml:"docvars>docvar"`
	Iclassintro iclassintro `xml:"iclassintro"`
	ArchVariant archVariant `xml:"arch_variants>arch_variant"`
	Regdiagram  regdiagram  `xml:"regdiagram"`
	Encodings   []encoding  `xml:"encoding"`
}

type classes struct {
	Classesintro classesintro `xml:"classesintro"`
	Iclass       []iclass     `xml:"iclass"`
}

type symbol struct {
	Value string `xml:",chardata"`
	Link  string `xml:"link,attr"`
}

type account struct {
	Encodedin string   `xml:"encodedin,attr"`
	DocVars   []docVar `xml:"docvars>docvar"`
	Intro     string   `xml:"intro>para"`
}

type entry struct {
	Value string `xml:",chardata"`
	Class string `xml:"class,attr"`
}

type row struct {
	Entries []entry `xml:"entry"`
}

type tHead struct {
	Row row `xml:"row"`
}

type tBody struct {
	Row []row `xml:"row"`
}

type tGroup struct {
	Cols  string `xml:"cols,attr"`
	THead tHead  `xml:"thead"`
	TBody tBody  `xml:"tbody"`
}

type table struct {
	Class  string `xml:"class,attr"`
	TGroup tGroup `xml:"tgroup"`
}

type definition struct {
	Encodedin string `xml:"encodedin,attr"`
	Intro     string `xml:"intro"`
	Table     table  `xml:"table"`
}

type explanation struct {
	Symbol     symbol     `xml:"symbol"`
	Account    account    `xml:"account"`
	Definition definition `xml:"definition"`
}

type explanations struct {
	Scope        string        `xml:"scope,attr"`
	Explanations []explanation `xml:"explanation"`
}

type instruction struct {
	XMLName      xml.Name     `xml:"instructionsection"`
	Title        string       `xml:"title,attr"`
	Type         string       `xml:"type,attr"`
	DocVars      []docVar     `xml:"docvars>docvar"`
	Classes      classes      `xml:"classes"`
	Explanations explanations `xml:"explanations"`
}

func (i instruction) print() {
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("  ", "    ")
	if err := enc.Encode(i); err != nil {
		fmt.Printf("Encode error in print(): %v\n", err)
	}
}

var (
	tab1  = "\t"
	tab2  = "\t\t"
	tab3  = "\t\t\t"
	next1 = "\n"
	next2 = "\n\n"
)

func generateElements(outPutDir string) {
	path := path.Join(outPutDir, "elements.go")
	if _, err := os.Stat(path); err == nil {
		log.Fatalf("elements.go exists in %v\n", outPutDir)
	}
	elemGo, err := os.Create(path)
	if err != nil {
		log.Fatalf("Create file %s failed: %v\n", path, err)
	}
	defer func() {
		elemGo.Close()
		cmd := exec.Command("gofmt", "-w", path)
		if err := cmd.Run(); err != nil {
			log.Fatalf("gofmt -w %s failed: %v\n", path, err)
		}
	}()
	// Get sorted keys
	keys := make([]string, len(rules))
	i := 0
	for k := range rules {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	w := bufio.NewWriter(elemGo)
	header := `// Code generated by instgen. DO NOT EDIT.

package arm64

type elmType uint16

//go:generate stringer -type elmType -trimprefix sa_`
	fmt.Fprintf(w, "%s%s", header, next1)
	fmt.Fprintf(w, "%s%s", "const(", next1)
	fmt.Fprintf(w, "%s%s%s", tab1, "sa_None elmType = iota", next1)
	for _, r := range keys {
		fmt.Fprintf(w, "%s// %s%s", tab1, rules[r].Symbol, next1)
		fmt.Fprintf(w, "%s%s%s", tab1, r, next1)
	}
	fmt.Fprintf(w, ")%s", next1)
	w.Flush()
}

func generateInstructions(outPutDir string) {
	path := path.Join(outPutDir, "instructions.go")
	if _, err := os.Stat(path); err == nil {
		log.Fatalf("instruction.go exists in %v\n", outPutDir)
	}
	instGo, err := os.Create(path)
	if err != nil {
		log.Fatalf("Create file %s failed: %v\n", path, err)
	}
	defer func() {
		instGo.Close()
		cmd := exec.Command("gofmt", "-w", path)
		if err := cmd.Run(); err != nil {
			log.Fatalf("gofmt -w %s failed: %v\n", path, err)
		}
	}()

	featurs := make(map[string]struct{})
	arm64Ops := make(map[string]struct{})
	goOps := make(map[string]struct{})
	// Some Go custom OPs.
	goOps["AWORD"] = struct{}{}
	goOps["ADWORD"] = struct{}{}
	goOps["AREM"] = struct{}{} // div+msub
	goOps["AREMW"] = struct{}{}
	goOps["AUREM"] = struct{}{}
	goOps["AUREMW"] = struct{}{}
	goOps["AVMOVS"] = struct{}{} // load 32-bit from constant pool.
	goOps["AVMOVD"] = struct{}{} // load 64-bit from constant pool.
	goOps["AVMOVQ"] = struct{}{} // load 128-bit from constant pool.
	goOps["AUXTW"] = struct{}{}  // ubfx <Xd>, <Xn>, #0, #32
	goOps["AUXTH"] = struct{}{}  // ubfx <Xd>, <Xn>, #0, #16
	goOps["AUXTB"] = struct{}{}  // ubfx <Xd>, <Xn>, #0, #8
	w := bufio.NewWriter(instGo)

	// Write instTab
	header := `// Code generated by instgen. DO NOT EDIT.

// The file contains the arm64 instruction table, which is created by parsing
// the xml document https://developer.arm.com/downloads/-/exploration-tools.

package arm64

import "cmd/internal/obj"

type arg struct {
	aType oprType   // operand class, register, constant, memory operation etc.
	elms  []elmType // the elements that this arg includes
}

// inst describes the format of an Arm64 instruction.
type inst struct {
	goOp     obj.As  // Go opcode mnemonic
	armOp    a64Type // Arm64 opcode mnemonic
	feature  uint16  // such as "FEAT_LSE", "FEAT_CSSC"
	skeleton uint32  // known bits
	mask     uint32  // mask for disassembly, 1 for known bits, 0 for unknown bits
	alias    bool    // whether it is an alias
	args     []arg   // args, in Go order
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
	for k := 0; k < len(p1.args); k++ {
		if p1.args[k].aType != p2.args[k].aType {
			return p1.args[k].aType < p2.args[k].aType
		}
	}
	if p1.skeleton != p2.skeleton {
		return p1.skeleton < p2.skeleton
	}
	if p1.mask != p2.mask {
		return p1.mask < p2.mask
	}
	return false
}`

	//	opds := make(map[string]int)
	fmt.Fprintf(w, "%s%s", header, next2)
	fmt.Fprintf(w, "var instTab = []inst{%s", next1)
	for _, inst := range insts {
		for i := range inst.Classes.Iclass {
			iclass := &inst.Classes.Iclass[i]
			feature := iclass.ArchVariant.Feature
			if feature != "" {
				feature = strings.ReplaceAll(feature, " && ", "__") // Special case "FEAT_D128 && FEAT_THE"
				if _, ok := featurs[feature]; !ok {
					featurs[feature] = struct{}{}
				}
			} else {
				feature = "FEAT_NONE"
			}
			for j := range iclass.Encodings {
				enc := &iclass.Encodings[j]
				if enc.invalid {
					continue
				}

				// Records arm64 opcode and go opcode.
				aOp, gOp := enc.arm64Op, enc.goOp
				if _, ok := arm64Ops[aOp]; !ok {
					arm64Ops[aOp] = struct{}{}
				}
				// AB == obj.AJMP, ABR == ABL == ABLR == obj.ACALL, ARET == obj.ARET
				// CALL, JMP and RET are portable opcodes.
				if _, ok := goOps[gOp]; !ok && gOp != "AB" && gOp != "ABR" && gOp != "ABL" && gOp != "ABLR" && gOp != "ARET" {
					goOps[gOp] = struct{}{}
				}

				// Outputs asm template.
				fmt.Fprintf(w, "%s// %s%s", tab1, enc.asm, next1)

				if len(enc.operands) == 1 {
					fmt.Fprintf(w, "%s{%s, %s, %s, 0x%x, 0x%x, %t, nil,\n\t},%s",
						tab1, enc.goOp, enc.arm64Op, feature, enc.binary, enc.mask, enc.alias, next2)
					continue
				}

				// Outputs arm64 opcode, go opcode, feature, binary and mask.
				fmt.Fprintf(w, "%s{%s, %s, %s, 0x%x, 0x%x, %t, []arg{%s",
					tab1, enc.goOp, enc.arm64Op, feature, enc.binary, enc.mask, enc.alias, next1)

				// Outputs operands, this first one is mnemonic.
				for k := 1; k < len(enc.operands); k++ {
					opr := enc.operands[k]
					fmt.Fprintf(w, "%s{%s, []elmType{%s", tab2, opr.typ, next1) // operand type
					for _, rule := range opr.rules {                            // elements
						fmt.Fprintf(w, "%s%s,%s", tab3, rule.Name, next1)
					}
					fmt.Fprintf(w, "%s}},%s", tab2, next1)
				}
				fmt.Fprintf(w, "%s}},%s", tab1, next2)
			}
		}
	}
	fmt.Fprintf(w, "}%s", next2)

	// Outputs features to instruction.go.
	// Get sorted keys.
	keys := make([]string, len(featurs))
	i := 0
	for k := range featurs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	fmt.Fprintf(w, "const (%s", next1)
	fmt.Fprintf(w, "%s%s%s", tab1, "FEAT_NONE uint16 = iota", next1)
	for _, k := range keys {
		fmt.Fprintf(w, "%s%s%s", tab1, k, next1)
	}
	fmt.Fprintf(w, ")%s", next1)
	w.Flush()

	// Generates arm64ops.go and ops.go
	generateArm64Ops(outPutDir, arm64Ops)
	generateOps(outPutDir, goOps)
}

func generateArm64Ops(dir string, arm64Ops map[string]struct{}) {
	path := path.Join(dir, "arm64ops.go")
	if _, err := os.Stat(path); err == nil {
		log.Fatalf("arm64ops.go exists in %v\n", dir)
	}
	arm64OpsGo, err := os.Create(path)
	if err != nil {
		log.Fatalf("Create file %s failed: %v\n", path, err)
	}
	defer func() {
		arm64OpsGo.Close()
		cmd := exec.Command("gofmt", "-w", path)
		if err := cmd.Run(); err != nil {
			log.Fatalf("gofmt -w %s failed: %v\n", path, err)
		}
	}()

	// Get sorted keys
	keys := make([]string, len(arm64Ops))
	i := 0
	for k := range arm64Ops {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	w := bufio.NewWriter(arm64OpsGo)
	header := `// Code generated by instgen. DO NOT EDIT.

package arm64

type a64Type uint16

//go:generate stringer -type a64Type -trimprefix A64_`
	fmt.Fprintf(w, "%s%s", header, next1)
	fmt.Fprintf(w, "const (%s", next1)
	fmt.Fprintf(w, "%sA64_BEGIN a64Type = iota%s", tab1, next1)
	for _, opcode := range keys {
		fmt.Fprintf(w, "%s%s%s", tab1, opcode, next1)
	}
	fmt.Fprintf(w, ")%s", next1)
	w.Flush()
}

func generateOps(dir string, goOps map[string]struct{}) {
	path := path.Join(dir, "ops.go")
	if _, err := os.Stat(path); err == nil {
		log.Fatalf("ops.go exists in %v\n", dir)
	}

	opsGo, err := os.Create(path)
	if err != nil {
		log.Fatalf("Create file %s failed: %v\n", path, err)
	}
	defer func() {
		opsGo.Close()
		cmd := exec.Command("gofmt", "-w", path)
		if err := cmd.Run(); err != nil {
			log.Fatalf("gofmt -w %s failed: %v\n", path, err)
		}
	}()

	// Get sorted keys
	keys := make([]string, len(goOps))
	i := 0
	for k := range goOps {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	w := bufio.NewWriter(opsGo)
	header := `// Code generated by instgen. DO NOT EDIT.

package arm64

import "cmd/internal/obj"

//go:generate go run ../stringer.go -i ops.go -o anames.go -p arm64`
	fmt.Fprintf(w, "%s%s", header, next1)
	fmt.Fprintf(w, "const (%s", next1)
	// The first opcode.
	fmt.Fprintf(w, "%s%s = obj.ABaseARM64 + obj.A_ARCHSPECIFIC + iota%s", tab1, keys[0], next1)
	for i := 1; i < len(keys); i++ {
		fmt.Fprintf(w, "%s%s%s", tab1, keys[i], next1)
	}
	fmt.Fprintf(w, "%sALAST%s", tab1, next1)
	fmt.Fprintf(w, "%sAB = obj.AJMP%s", tab1, next1)
	fmt.Fprintf(w, "%sABR = obj.AJMP%s", tab1, next1)
	fmt.Fprintf(w, "%sABL = obj.ACALL%s", tab1, next1)
	fmt.Fprintf(w, "%sABLR = obj.ACALL%s", tab1, next1)
	fmt.Fprintf(w, "%sARET = obj.ARET%s", tab1, next1)
	fmt.Fprintf(w, ")%s", next1)
	w.Flush()
}

var orders = []string{
	"AC_REG",
	"AC_RSP",
	"AC_FREG",
	"AC_VREG",
	"AC_ZREG",
	"AC_PREG",
	"AC_PNREG",
	"AC_PREGM",
	"AC_PREGZ",
	"AC_ZTREG",
	"AC_PAIR",
	"AC_REGSHIFT",
	"AC_REGEXT",
	"AC_COND",
	"AC_SPR",
	"AC_SPOP",
	"AC_IMM",
	"AC_LABEL",
	"AC_REGLIST1",
	"AC_REGLIST2",
	"AC_REGLIST3",
	"AC_REGLIST4",
	"AC_LISTIDX",
	"AC_REGIDX",
	"AC_ARNG",
	"AC_ARNGIDX",
	"AC_MEMIMM",
	"AC_MEMIMMEXT",
	"AC_MEMPREIMM",
	"AC_MEMPOSTIMM",
	"AC_MEMPOSTREG",
	"AC_MEMEXT",
	"AC_ZAVECTORSEL",
	"AC_ZAHVTILEIDX",
	"AC_ZAVECTORSELVG2",
	"AC_ZAVECTORSELVG4",
	"AC_ZAVECTORIDXVG2",
	"AC_ZAVECTORIDXVG4",
	"AC_ZAHVTILESEL",
	"AC_ZAVECTORIDX",
	"AC_ANY",
}

// This function is for debugging
func check() {
	// var myOpcodes = make(map[string][]string)
	//      var binMasks = make(map[uint64]encoding)
	types := make(map[string]map[string]int)
	nameRules := make(map[string]map[string]int)

	for _, inst := range insts {
		for i := range inst.Classes.Iclass {
			iclass := &inst.Classes.Iclass[i]
			for j := range iclass.Encodings {
				enc := &iclass.Encodings[j]
				if enc.invalid {
					continue
				}
				// binMask := uint64(enc.binary)<<32 | uint64(enc.mask)
				// if _, ok := binMasks[binMask]; !ok {
				//      binMasks[binMask] = *enc
				// } else {
				//      if enc.alias && binMasks[binMask].alias || !enc.alias && !binMasks[binMask].alias {
				//              fmt.Printf("old: %s\n", binMasks[binMask].asm)
				//              fmt.Printf("new: %s, bin: 0x%x, mask: 0x%x\n\n", enc.asm, enc.binary, enc.mask)
				//      }
				// }

				for jj := 1; jj < len(enc.operands); jj++ {
					opr_type := enc.operands[jj].typ
					if opr_type == "" {
						log.Fatalf("empty operand type: %s\n", enc.asm)
					}
					opr_name := enc.operands[jj].name
					if _, ok := types[opr_type]; !ok {
						types[opr_type] = make(map[string]int)
						types[opr_type][opr_name] = 1
					} else {
						if _, ok2 := types[opr_type][opr_name]; !ok2 {
							types[opr_type][opr_name] = 1
						} else {
							types[opr_type][opr_name]++
						}
					}

					opr_type_name := opr_type + "__" + opr_name
					if _, ok := nameRules[opr_type_name]; !ok {
						nameRules[opr_type_name] = make(map[string]int)
						for _, rule := range enc.operands[jj].rules {
							nameRules[opr_type_name][rule.Name] = 1
						}
					} else {
						for _, rule := range enc.operands[jj].rules {
							if _, ok2 := nameRules[opr_type_name][rule.Name]; !ok2 {
								nameRules[opr_type_name][rule.Name] = 1
							} else {
								nameRules[opr_type_name][rule.Name]++
							}
						}
					}
				}
				// if _, ok := myInsts[classes]; !ok {
				//      myInsts[classes] = enc.asm
				// } else {
				// for jj := 1; jj < len(enc.operands); jj++ {
				//      if enc.operands[jj].typ == "AC_REGLIST2" || enc.operands[jj].typ == "AC_REGLIST4" {
				// if !arng.MatchString(enc.asm) {
				//      fmt.Printf("old: %s\n", myInsts[classes])
				//      fmt.Printf("new: %s, class: %s\n\n", enc.asm, enc.className())
				// }
				//      break
				// }
				//}
				//}

				// if enc.class == C_GENERAL {
				//      continue
				// }
				// myopcode := enc.operands[0].name
				// if _, ok := myOpcodes[myopcode]; !ok {
				//      myOpcodes[myopcode] = []string{enc.suffix}
				// } else {
				//      myOpcodes[myopcode] = append(myOpcodes[myopcode], enc.suffix)
				// }

				// for k := 1; k < len(enc.operands); k++ {
				// opr := enc.operands[k]
				// if _, ok := opds[opr.name]; ok {
				//      opds[opr.name]++
				// } else if opr.name != "" {
				//      opds[opr.name] = 1
				// }
				//}
			}
		}
	}
	// for k, v := range myOpcodes {
	//      if len(v) > 1 {
	//              hasSpace, hasOthers := false, false
	//              for _, op := range v {
	//                      if "" == op {
	//                              hasSpace = true
	//                      } else {
	//                              hasOthers = true
	//                      }
	//              }
	//              if hasOthers && hasSpace {
	//                      fmt.Printf("%s suffix = %v\n", k, v)
	//              }
	//      }
	// }

	// type sopr struct {
	//      name  string
	//      count int
	// }
	// sortedOprs := make([]sopr, 0, len(opds))
	// for k, v := range opds {
	//      sortedOprs = append(sortedOprs, sopr{k, v})
	// }
	// sort.Slice(sortedOprs, func(i, j int) bool {
	//      if sortedOprs[i].name != sortedOprs[j].name {
	//              return sortedOprs[i].name < sortedOprs[j].name
	//      }
	//      return sortedOprs[i].count > sortedOprs[j].count
	// })
	// for _, v := range sortedOprs {
	//      fmt.Printf("%-50s :  %v\n", v.name, v.count)
	// }
	if len(orders) != len(types) {
		log.Fatalf("len(orders) != len(types)\n")
	}
	for _, typ := range orders {
		fmt.Println(typ + ":")
		for k, c := range types[typ] {
			fmt.Printf("\t%-50s : %d\n", k, c)
			opr_type_name := typ + "__" + k
			for m, n := range nameRules[opr_type_name] {
				fmt.Printf("\t\t%-50s : %d\n", m, n)
			}
		}
		fmt.Println()
	}
}

func generate(outPutDir string) {
	// Check the results before output, for debugging.
	// check()
	generateElements(outPutDir)
	generateInstructions(outPutDir)
}

// Simple cas-lock to coordinate appending elements to insts, and processing.
var instsLock atomic.Int64
var insts []*instruction

var rules map[string]element
