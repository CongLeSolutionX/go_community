// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objfile

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/arch/arm/armasm"
	"golang.org/x/arch/x86/x86asm"
)

// Disasm is a disassembler for a given File.
type Disasm struct {
	syms      []Sym            //symbols in file, sorted by address
	goarch    string           // GOARCH string
	disasm    disasmFunc       // disassembler function for goarch
	byteOrder binary.ByteOrder // byte order for goarch
	f         *File            // underlying file
}

// Disasm returns a disassembler for the file f.
func (f *File) Disasm() (*Disasm, error) {
	syms, err := f.Symbols()
	if err != nil {
		return nil, err
	}

	goarch := f.GOARCH()
	disasm := disasms[goarch]
	byteOrder := byteOrders[goarch]
	if disasm == nil || byteOrder == nil {
		return nil, fmt.Errorf("unsupported architecture")
	}

	// Filter out section symbols, overwriting syms in place.
	keep := syms[:0]
	for _, sym := range syms {
		switch sym.Name {
		case "runtime.text", "text", "_text", "runtime.etext", "etext", "_etext":
			// drop
		default:
			keep = append(keep, sym)
		}
	}
	syms = keep
	d := &Disasm{
		syms:      syms,
		goarch:    goarch,
		disasm:    disasm,
		byteOrder: byteOrder,
		f:         f,
	}

	return d, nil
}

// lookup finds the symbol name containing addr.
func (d *Disasm) lookup(addr uint64) (name string, base uint64) {
	i := sort.Search(len(d.syms), func(i int) bool { return addr < d.syms[i].Addr })
	if i > 0 {
		s := d.syms[i-1]
		if s.Addr != 0 && s.Addr <= addr && addr < s.Addr+uint64(s.Size) {
			return s.Name, s.Addr
		}
	}
	return "", 0
}

// base returns the final element in the path.
// It works on both Windows and Unix paths,
// regardless of host operating system.
func base(path string) string {
	path = path[strings.LastIndex(path, "/")+1:]
	path = path[strings.LastIndex(path, `\`)+1:]
	return path
}

// Print prints a disassembly of the file to w.
// If filter is non-nil, the disassembly only includes functions with names matching filter.
// The disassembly only includes functions that overlap the range [start, end).
func (d *Disasm) Print(w io.Writer, filter *regexp.Regexp, start, end uint64) {
	printed := false
	bw := bufio.NewWriter(w)
	for i := range d.syms {
		sym := &d.syms[i]
		if sym.Code != 'T' && sym.Code != 't' ||
			sym.Addr+uint64(sym.Size) <= start || end <= sym.Addr ||
			filter != nil && !filter.MatchString(sym.Name) {
			continue
		}
		if printed {
			fmt.Fprintf(bw, "\n")
		}
		printed = true

		file, _ := d.f.PC2Line(sym, sym.Addr)
		fmt.Fprintf(bw, "TEXT %s(SB) %s\n", sym.Name, file)

		tw := tabwriter.NewWriter(bw, 1, 8, 1, '\t', 0)
		d.Decode(sym, func(pc uint64, code []byte, file string, line int, text string) {
			fmt.Fprintf(tw, "\t%s:%d\t%#x\t", base(file), line, pc)
			if len(code)%4 != 0 || d.goarch == "386" || d.goarch == "amd64" {
				// Print instruction as bytes.
				fmt.Fprintf(tw, "%x", code)
			} else {
				// Print instruction as 32-bit words.
				for j := 0; j < len(code); j += 4 {
					if j > 0 {
						fmt.Fprintf(tw, " ")
					}
					fmt.Fprintf(tw, "%08x", d.byteOrder.Uint32(code[j:]))
				}
			}
			fmt.Fprintf(tw, "\t%s\n", text)
		})
		tw.Flush()
	}
	bw.Flush()
}

// Decode disassembles all the symbols in the file, calling f for each instruction.
func (d *Disasm) DecodeAll(f func(pc uint64, code []byte, file string, line int, text string)) {
	for i := range d.syms {
		d.Decode(&d.syms[i], f)
	}
}

// Decode disassembles all the instructions in sym, calling f for each instruction.
func (d *Disasm) Decode(sym *Sym, f func(pc uint64, code []byte, file string, line int, text string)) {
	start := sym.Addr
	end := sym.Addr + uint64(sym.Size)
	code, err := d.f.GetText(sym)
	if err != nil {
		//TODO
		fmt.Printf("decode err %s\n", err)
		return
	}
	relocs := d.f.Relocs(sym)

	lookup := d.lookup
	for pc := start; pc < end; {
		i := pc - start
		text, size := d.disasm(code[i:], pc, lookup)
		text += "\t"
		first := true
		for len(relocs) > 0 && relocs[0].Offset < int(i)+size {
			if first {
				first = false
			} else {
				text += " "
			}
			text += relocs[0].String(i)
			relocs = relocs[1:]
		}
		file, line := d.f.PC2Line(sym, pc)
		f(pc, code[i:i+uint64(size)], file, line, text)
		pc += uint64(size)
	}
}

type lookupFunc func(addr uint64) (sym string, base uint64)
type disasmFunc func(code []byte, pc uint64, lookup lookupFunc) (text string, size int)

func disasm_386(code []byte, pc uint64, lookup lookupFunc) (string, int) {
	return disasm_x86(code, pc, lookup, 32)
}

func disasm_amd64(code []byte, pc uint64, lookup lookupFunc) (string, int) {
	return disasm_x86(code, pc, lookup, 64)
}

func disasm_x86(code []byte, pc uint64, lookup lookupFunc, arch int) (string, int) {
	inst, err := x86asm.Decode(code, 64)
	var text string
	size := inst.Len
	if err != nil || size == 0 || inst.Op == 0 {
		size = 1
		text = "?"
	} else {
		text = x86asm.GoSyntax(inst, pc, lookup)
	}
	return text, size
}

type textReader struct {
	code []byte
	pc   uint64
}

func (r textReader) ReadAt(data []byte, off int64) (n int, err error) {
	if off < 0 || uint64(off) < r.pc {
		return 0, io.EOF
	}
	d := uint64(off) - r.pc
	if d >= uint64(len(r.code)) {
		return 0, io.EOF
	}
	n = copy(data, r.code[d:])
	if n < len(data) {
		err = io.ErrUnexpectedEOF
	}
	return
}

func disasm_arm(code []byte, pc uint64, lookup lookupFunc) (string, int) {
	inst, err := armasm.Decode(code, armasm.ModeARM)
	var text string
	size := inst.Len
	if err != nil || size == 0 || inst.Op == 0 {
		size = 4
		text = "?"
	} else {
		text = armasm.GoSyntax(inst, pc, lookup, textReader{code, pc})
	}
	return text, size
}

var disasms = map[string]disasmFunc{
	"386":   disasm_386,
	"amd64": disasm_amd64,
	"arm":   disasm_arm,
}

var byteOrders = map[string]binary.ByteOrder{
	"386":     binary.LittleEndian,
	"amd64":   binary.LittleEndian,
	"arm":     binary.LittleEndian,
	"ppc64":   binary.BigEndian,
	"ppc64le": binary.LittleEndian,
	"s390x":   binary.BigEndian,
}
