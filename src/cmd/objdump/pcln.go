// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type decoder struct {
	order   binary.ByteOrder
	data    []byte
	pos     uint64
	ptrSize int
}

func (d *decoder) Uint8() uint8 {
	v := d.data[d.pos]
	d.pos++
	return v
}

func (d *decoder) Int8() int8 {
	return int8(d.Uint8())
}

func (d *decoder) Uint32() uint32 {
	v := d.order.Uint32(d.data[d.pos:])
	d.pos += 4
	return v
}

func (d *decoder) Int32() int32 {
	return int32(d.Uint32())
}

func (d *decoder) Uint64() uint64 {
	v := d.order.Uint64(d.data[d.pos:])
	d.pos += 8
	return v
}

func (d *decoder) Ptr() uint64 {
	var v uint64
	switch d.ptrSize {
	case 4:
		v = uint64(d.order.Uint32(d.data[d.pos:]))
	case 8:
		v = d.order.Uint64(d.data[d.pos:])
	default:
		panic("bad ptrSize")
	}
	d.pos += uint64(d.ptrSize)
	return v
}

func (d *decoder) CString() string {
	start := d.pos
	for d.data[d.pos] != 0 {
		d.pos++
	}
	d.pos++
	return string(d.data[start : d.pos-1])
}

type _func struct {
	entry   uintptr // start pc
	nameoff int32   // function name

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      int32
	pcfile    int32
	pcln      int32
	npcdata   int32
	funcID    uint8   // set for certain special runtime functions
	_unused   [2]int8 // unused
	nfuncdata uint8   // must be last
}

// findNextOffset streamlines reading inlined function names.
// Pcln data has inlined function names interleaved with _func objects. There
// is no real way to tell if the function names are there while reading pcln
// besides looking at all position values, and seeing if the current scan
// position would let us read one of them.  If the current position value isn't
// in the set of positions we know about, then we know there's an inlined
// function name, and we need to read it.
//
// This funciton with return either 0, or the index of the next position where
// we know there's data.
func findNextOffset(pos uint64, f _func, more ...int32) uint64 {
	p := []int{
		int(f.nameoff),
		int(f.pcsp),
		int(f.pcfile),
		int(f.pcln),
	}
	for _, v := range more {
		p = append(p, int(v))
	}
	sort.Ints(p)
	i := sort.SearchInts(p, int(pos))
	if i == len(p) {
		return 0
	}
	return uint64(p[i])
}

func printPclnStats(out io.Writer, encoding binary.ByteOrder, data []byte) {
	d := decoder{encoding, data, 0, 0}
	d.pos += 7 // Skip most of the header
	d.ptrSize = int(d.Uint8())

	// Read the size of the func table
	nfunc := int(d.Uint64())
	size := (2*nfunc + 1) * d.ptrSize // nfunc entries + end pc

	// Read the offset table.
	begin := d.pos
	pcs := make([]uint64, nfunc)
	funcOffsets := make([]uint64, nfunc)
	for i := 0; i < nfunc; i++ {
		pcs[i] = d.Ptr()
		funcOffsets[i] = d.Ptr()
	}
	d.Ptr() // end pc
	if d.pos != begin+uint64(size) {
		panic(fmt.Sprintf("%d %d %d", d.pos, begin, (nfunc+1)*d.ptrSize))
	}
	ftabOffset := d.Uint32()
	// Depending on the architecture, there's padding here before we start the _func array.

	// Read the func entries
	begin = d.pos
	funcTot, fnameTot, pcspTot, pclnTot, pcfileTot := 0, 0, 0, 0, 0
	pcdata := [3]int{}
	names := make(map[int32]int)
	pclookup := make(map[int32]int)
	readData := func(off int32, skip int, lookup map[int32]int, countCached bool) int {
		if v := lookup[off]; off == 0 || v != 0 {
			// The caller might want to double count the data. If so, return it.
			if countCached {
				return v
			}
			return 0
		}
		if d.pos != uint64(off) {
			panic("data sync error")
		}
		d.pos += uint64(skip)
		s := len(d.CString()) + skip + 1 // add the skipped amount, and null terminator
		lookup[off] = s
		return s
	}

	lastPos := funcOffsets[0]
	for i := 0; i < nfunc; i++ {
		d.pos = funcOffsets[i]
		f := _func{
			uintptr(d.Ptr()),            // entry
			d.Int32(),                   // nameoff
			d.Int32(),                   // args
			d.Uint32(),                  // deferreturn
			d.Int32(),                   // pcsp
			d.Int32(),                   // pcfile
			d.Int32(),                   // pcln
			d.Int32(),                   // npcdata
			d.Uint8(),                   // funcID
			[2]int8{d.Int8(), d.Int8()}, // _
			d.Uint8(),                   // nfuncdata
		}

		// Sanity check out the previously read pc and the entry.
		if uintptr(pcs[i]) != f.entry {
			panic(fmt.Sprintf("pcs: %d %d %d", i, pcs[i], f.entry))
		}

		// Read the pcdata offsets.
		pcdataOffsets := []int32{}
		for j := 0; j < int(f.npcdata); j++ {
			pcdataOffsets = append(pcdataOffsets, d.Int32())
		}

		// funcdata is ptr aligned
		d.pos += (-d.pos) & uint64(d.ptrSize-1)
		for j := 0; j < int(f.nfuncdata); j++ {
			d.Ptr()
		}
		funcTot += int(d.pos - funcOffsets[i])

		// Read the name.
		fnameTot += readData(f.nameoff, 0, names, false)

		for nextPos := findNextOffset(d.pos, f, pcdataOffsets...); d.pos < nextPos; {
			t := d.pos
			s := d.CString()
			names[int32(t)] = len(s) + 1
			fnameTot += len(s) + 1
		}

		pcspTot += readData(f.pcsp, 0, pclookup, true)
		pcfileTot += readData(f.pcfile, 0, pclookup, true)
		pclnTot += readData(f.pcln, 0, pclookup, true)

		// Read PCData
		for i, off := range pcdataOffsets {
			pcdata[i] += readData(off, 1, pclookup, true)
		}

		lastPos = d.pos
		lastPos += -lastPos & uint64(d.ptrSize-1)
		funcTot += int(lastPos - d.pos)
	}
	if lastPos != uint64(ftabOffset) {
		panic("pcln table didn't end at file table")
	}

	// Report the sizes.
	// Note that there's some double counting involved in the pc data, see countCached in
	// the readData func above.
	io.WriteString(out, fmt.Sprintf("\tPCLNTAB\t\t%d\n", len(data)))
	io.WriteString(out, fmt.Sprintf("\t  func table:\t\t% 10d\t%d entries\n", size, nfunc))
	io.WriteString(out, fmt.Sprintf("\t  _func structs:\t% 10d % 10d\n", funcTot, funcTot-nfunc*12))
	io.WriteString(out, fmt.Sprintf("\t  pcsp:\t\t\t% 10d\n", pcspTot))
	io.WriteString(out, fmt.Sprintf("\t  pcfile:\t\t% 10d\n", pcfileTot))
	io.WriteString(out, fmt.Sprintf("\t  pcln:\t\t\t% 10d\n", pclnTot))
	pctot := 0
	for _, v := range pcdata {
		pctot += v
	}
	io.WriteString(out, fmt.Sprintf("\t  pcdata:\t\t% 10d\n", pctot))
	io.WriteString(out, fmt.Sprintf("\t    [0]:\t\t% 10d\n", pcdata[0]))
	io.WriteString(out, fmt.Sprintf("\t    [1]:\t\t% 10d\n", pcdata[1]))
	io.WriteString(out, fmt.Sprintf("\t    [2]:\t\t% 10d\n", pcdata[2]))

	// The rest is the file table.
	fileSize := len(data) - int(lastPos)
	totString := fnameTot + fileSize
	io.WriteString(out, fmt.Sprintf("\t  Strings:\t\t% 10d\n", totString))
	io.WriteString(out, fmt.Sprintf("\t    func names:\t\t% 10d\n", fnameTot))
	io.WriteString(out, fmt.Sprintf("\t    files:\t\t% 10d\n", fileSize))
}
