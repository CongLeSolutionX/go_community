// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Writes dwarf information to object files.
package obj

import (
	"cmd/internal/dwarf"
	"cmd/internal/src"
	"fmt"
	"os"
	"strings"
)

/*
 * Generate a sequence of opcodes that is as short as possible.
 * See section 6.2.5
 */
const (
	LINE_BASE   = -4
	LINE_RANGE  = 10
	PC_RANGE    = (255 - OPCODE_BASE) / LINE_RANGE
	OPCODE_BASE = 11
)

// writeDebugLines uses 's' to generate  a DWARF v2 debug_lines section into the 'lines' symbol.
func (ctxt *Link) writeDebugLines(s, lines *LSym) {
	is_stmt := uint8(1) // initially = recommended default_is_stmt = 1, tracks is_stmt toggles.

	// Set up to write dwarf output.
	dctxt := dwCtxt{ctxt}

	// Write the header.
	// The header includes the length of the section, and we store the locations
	// we need to overwrite once we determine the length.
	unitLengthOffset := dctxt.Pos(lines)
	if dctxt.IsDwarf64() {
		dctxt.AddInt(lines, 4, 0xFFFFFFFF)
		dctxt.AddInt(lines, 8, 0)
	} else {
		dctxt.AddInt(lines, 4, 0)
	}
	unitstart := dctxt.Pos(lines)
	dctxt.AddInt(lines, 2, 2) // dwarf version (appendix F) -- version 3 is incompatible w/ XCode 9.0's dsymutil, latest supported on OSX 10.12 as of 2018-05
	headerLengthOffset := dctxt.Pos(lines)
	if dctxt.IsDwarf64() {
		dctxt.AddInt(lines, 8, 0)
	} else {
		dctxt.AddInt(lines, 4, 0)
	}
	headerstart := dctxt.Pos(lines)

	dctxt.AddUint8(lines, 1)              // minimum_indctxttruction_length
	dctxt.AddUint8(lines, is_stmt)        // default_is_stmt
	dctxt.AddUint8(lines, LINE_BASE&0xFF) // line_badctxte
	dctxt.AddUint8(lines, LINE_RANGE)     // line_range
	dctxt.AddUint8(lines, OPCODE_BASE)    // opcode_badctxte
	dctxt.AddUint8(lines, 0)              // dctxttandard_opcode_lengthdctxt[1]
	dctxt.AddUint8(lines, 1)              // dctxttandard_opcode_lengthdctxt[2]
	dctxt.AddUint8(lines, 1)              // dctxttandard_opcode_lengthdctxt[3]
	dctxt.AddUint8(lines, 1)              // dctxttandard_opcode_lengthdctxt[4]
	dctxt.AddUint8(lines, 1)              // dctxttandard_opcode_lengthdctxt[5]
	dctxt.AddUint8(lines, 0)              // dctxttandard_opcode_lengthdctxt[6]
	dctxt.AddUint8(lines, 0)              // dctxttandard_opcode_lengthdctxt[7]
	dctxt.AddUint8(lines, 0)              // dctxttandard_opcode_lengthdctxt[8]
	dctxt.AddUint8(lines, 1)              // dctxttandard_opcode_lengthdctxt[9]
	dctxt.AddUint8(lines, 0)              // dctxttandard_opcode_lengthdctxt[10]
	dctxt.AddUint8(lines, 0)              // include_directoriedctxt  (lines, empty)

	// Write the filenames
	for _, filename := range s.Func.Pcln.File {
		// Filenames might have a prefix, and ${GOROOT} appended.
		filename = strings.TrimPrefix(filename, src.FileSymPrefix)
		filename = os.ExpandEnv(filename)
		dctxt.AddString(lines, filename)
		dctxt.AddUint8(lines, 0) // number representing directory
		dctxt.AddUint8(lines, 0) // timestamp for modified time
		dctxt.AddUint8(lines, 0) // length of file in bytes
	}
	dctxt.AddUint8(lines, 0) // done filename entries.
	headerend := dctxt.Pos(lines)

	dctxt.AddUint8(lines, 0) // start extended opcode
	dctxt.AddUleb128(lines, 1+int64(ctxt.Arch.PtrSize))
	dctxt.AddUint8(lines, dwarf.DW_LNE_set_address)

	pc := s.Func.Text.Pc
	line := 1
	file := 1
	// TODO(jfaller): We'll need to set the reachability of the symbol when linking.
	//lines.AddAddr(ctxt.Arch, s)

	// Generate the actual line information.
	pcfile := NewPCIter(uint32(ctxt.Arch.Arch.MinLC))
	pcline := NewPCIter(uint32(ctxt.Arch.Arch.MinLC))
	pcstmt := NewPCIter(uint32(ctxt.Arch.Arch.MinLC))
	pcfile.Init(s.Func.Pcln.Pcfile.P)
	pcline.Init(s.Func.Pcln.Pcline.P)
	var pctostmtData Pcdata
	funcpctab(ctxt, &pctostmtData, s, "pctostmt", pctostmt, nil)
	pcstmt.Init(pctostmtData.P)
	var thispc uint32

	for !pcfile.Done && !pcline.Done {
		// Only changed if it advanced
		if int32(file) != pcfile.Value {
			dctxt.AddUint8(lines, dwarf.DW_LNS_set_file)
			dctxt.AddUleb128(lines, int64(pcline.Value))
			file = int(pcfile.Value)
		}

		// Only changed if it advanced
		if is_stmt != uint8(pcstmt.Value) {
			new_stmt := uint8(pcstmt.Value)
			switch new_stmt &^ 1 {
			case PrologueEnd:
				dctxt.AddUint8(lines, uint8(dwarf.DW_LNS_set_prologue_end))
			case EpilogueBegin:
				// TODO if there is a use for this, add it.
				// Don't forget to increase OPCODE_BASE by 1 and add entry for standard_opcode_lengths[11]
			}
			new_stmt &= 1
			if is_stmt != new_stmt {
				is_stmt = new_stmt
				dctxt.AddUint8(lines, uint8(dwarf.DW_LNS_negate_stmt))
			}
		}

		// putpcldelta makes a row in the DWARF matrix, always, even if line is unchanged.
		putpclcdelta(ctxt, dctxt, lines, uint64(s.Func.Text.Pc+int64(thispc)-pc), int64(pcline.Value)-int64(line))

		pc = s.Func.Text.Pc + int64(thispc)
		line = int(pcline.Value)

		// Take the minimum step forward for the three iterators
		thispc = pcfile.Nextpc
		if pcline.Nextpc < thispc {
			thispc = pcline.Nextpc
		}
		if !pcstmt.Done && pcstmt.Nextpc < thispc {
			thispc = pcstmt.Nextpc
		}

		if pcfile.Nextpc == thispc {
			pcfile.Next()
		}
		if !pcstmt.Done && pcstmt.Nextpc == thispc {
			pcstmt.Next()
		}
		if pcline.Nextpc == thispc {
			pcline.Next()
		}
	}
	// Add a section termination to reset the counters.
	dctxt.AddUint8(lines, uint8(dwarf.DW_LNS_negate_stmt))

	// Terminate line section.
	dctxt.AddUint8(lines, 0)
	dctxt.AddUleb128(lines, 1)
	dctxt.AddUint8(lines, dwarf.DW_LNE_end_sequence)

	if dctxt.IsDwarf64() {
		dctxt.SetInt(lines, unitLengthOffset+4, 8, int64(dctxt.Pos(lines)-unitstart)) // +4 because of 0xFFFFFFFF
		dctxt.SetInt(lines, headerLengthOffset, 8, int64(headerend-headerstart))
	} else {
		dctxt.SetInt(lines, unitLengthOffset, 4, int64(dctxt.Pos(lines)-unitstart))
		dctxt.SetInt(lines, headerLengthOffset, 4, int64(headerend-headerstart))
	}
}

func putpclcdelta(linkctxt *Link, dctxt dwCtxt, s *LSym, deltaPC uint64, deltaLC int64) {
	// Choose a special opcode that minimizes the number of bytes needed to
	// encode the remaining PC delta and LC delta.
	var opcode int64
	if deltaLC < LINE_BASE {
		if deltaPC >= PC_RANGE {
			opcode = OPCODE_BASE + (LINE_RANGE * PC_RANGE)
		} else {
			opcode = OPCODE_BASE + (LINE_RANGE * int64(deltaPC))
		}
	} else if deltaLC < LINE_BASE+LINE_RANGE {
		if deltaPC >= PC_RANGE {
			opcode = OPCODE_BASE + (deltaLC - LINE_BASE) + (LINE_RANGE * PC_RANGE)
			if opcode > 255 {
				opcode -= LINE_RANGE
			}
		} else {
			opcode = OPCODE_BASE + (deltaLC - LINE_BASE) + (LINE_RANGE * int64(deltaPC))
		}
	} else {
		if deltaPC <= PC_RANGE {
			opcode = OPCODE_BASE + (LINE_RANGE - 1) + (LINE_RANGE * int64(deltaPC))
			if opcode > 255 {
				opcode = 255
			}
		} else {
			// Use opcode 249 (pc+=23, lc+=5) or 255 (pc+=24, lc+=1).
			//
			// Let x=deltaPC-PC_RANGE.  If we use opcode 255, x will be the remaining
			// deltaPC that we need to encode separately before emitting 255.  If we
			// use opcode 249, we will need to encode x+1.  If x+1 takes one more
			// byte to encode than x, then we use opcode 255.
			//
			// In all other cases x and x+1 take the same number of bytes to encode,
			// so we use opcode 249, which may save us a byte in encoding deltaLC,
			// for similar reasons.
			switch deltaPC - PC_RANGE {
			// PC_RANGE is the largest deltaPC we can encode in one byte, using
			// DW_LNS_const_add_pc.
			//
			// (1<<16)-1 is the largest deltaPC we can encode in three bytes, using
			// DW_LNS_fixed_advance_pc.
			//
			// (1<<(7n))-1 is the largest deltaPC we can encode in n+1 bytes for
			// n=1,3,4,5,..., using DW_LNS_advance_pc.
			case PC_RANGE, (1 << 7) - 1, (1 << 16) - 1, (1 << 21) - 1, (1 << 28) - 1,
				(1 << 35) - 1, (1 << 42) - 1, (1 << 49) - 1, (1 << 56) - 1, (1 << 63) - 1:
				opcode = 255
			default:
				opcode = OPCODE_BASE + LINE_RANGE*PC_RANGE - 1 // 249
			}
		}
	}
	if opcode < OPCODE_BASE || opcode > 255 {
		panic(fmt.Sprintf("produced invalid special opcode %d", opcode))
	}

	// Subtract from deltaPC and deltaLC the amounts that the opcode will add.
	deltaPC -= uint64((opcode - OPCODE_BASE) / LINE_RANGE)
	deltaLC -= (opcode-OPCODE_BASE)%LINE_RANGE + LINE_BASE

	// Encode deltaPC.
	if deltaPC != 0 {
		if deltaPC <= PC_RANGE {
			// Adjust the opcode so that we can use the 1-byte DW_LNS_const_add_pc
			// instruction.
			opcode -= LINE_RANGE * int64(PC_RANGE-deltaPC)
			if opcode < OPCODE_BASE {
				panic(fmt.Sprintf("produced invalid special opcode %d", opcode))
			}
			dctxt.AddUint8(s, dwarf.DW_LNS_const_add_pc)
		} else if (1<<14) <= deltaPC && deltaPC < (1<<16) {
			dctxt.AddUint8(s, dwarf.DW_LNS_fixed_advance_pc)
			dctxt.AddUint16(s, uint16(deltaPC))
		} else {
			dctxt.AddUint8(s, dwarf.DW_LNS_advance_pc)
			dctxt.AddUleb128(s, int64(deltaPC))
		}
	}

	// Encode deltaLC.
	if deltaLC != 0 {
		dctxt.AddUint8(s, dwarf.DW_LNS_advance_line)
		dctxt.AddUleb128(s, deltaLC)
	}

	// Output the special opcode.
	dctxt.AddUint8(s, uint8(opcode))
}
