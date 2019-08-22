// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"debug/dwarf"
	"fmt"

	"lockcheck/cache"
	"lockcheck/lockgraph"
	"lockcheck/symbolize"
)

// Proc represents a process.
type Proc struct {
	DWARF *symbolize.DWARF

	path    string
	pieDisp uint64
	threads map[int]*Thread

	// stacks caches whole stacks. It is indexed by a hash of the
	// PC stack.
	stacks map[uint64][]stackCacheEntry

	// frags caches symbolic stack fragments, indexed by PC.
	frags map[uint64][]lockgraph.SrcPos
}

// Thread represents a thread in a process.
type Thread struct {
	Proc *Proc
	ID   int
}

func (t *Thread) String() string {
	return fmt.Sprintf("%s [M %d]", t.Proc.path, t.ID)
}

// dwarfs is a cache mapping binary path to *DWARF data for that
// binary.
var dwarfs = cache.Cache{New: func(key interface{}) interface{} {
	path := key.(string)
	dw, err := symbolize.NewDWARF(path)
	if err != nil {
		return fmt.Errorf("loading DWARF for %s: %w", path, err)
	}
	return dw
}}

// NewProc creates a new Proc, using symbolic information from the
// binary at path. pieSym and piePos are a symbol and its loaded
// location in memory, used to determine the PIE loading offset of the
// binary.
func NewProc(path string, pieSym string, piePos uint64) (*Proc, error) {
	// Load symbol table. This is in a shared cache so it can be
	// shared across different instances of the same binary (which
	// is extremely common when testing the runtime).
	var dw *symbolize.DWARF
	switch val := dwarfs.Get(path).(type) {
	case *symbolize.DWARF:
		dw = val
	case error:
		return nil, val
	default:
		panic("bad type")
	}

	// Lookup the file address of the PIE symbol to compute the
	// PIE displacement.
	pieFunc := dw.Lines.FuncByName(pieSym)
	if pieFunc == nil {
		return nil, fmt.Errorf("failed to resolve %s for PIE displacement in %s", pieSym, path)
	}
	entry, ok := pieFunc.Val(dwarf.AttrEntrypc).(uint64)
	if !ok {
		entry, ok = pieFunc.Val(dwarf.AttrLowpc).(uint64)
	}
	if !ok {
		return nil, fmt.Errorf("failed to get entry of %s for PIE displacement in %s", pieSym, path)
	}
	pieDisp := piePos - entry

	return &Proc{
		DWARF:   dw,
		path:    path,
		pieDisp: pieDisp,
		threads: make(map[int]*Thread),
		stacks:  make(map[uint64][]stackCacheEntry),
		frags:   make(map[uint64][]lockgraph.SrcPos),
	}, nil
}

func (proc *Proc) Thread(mID int) *Thread {
	thr, ok := proc.threads[mID]
	if !ok {
		thr = &Thread{proc, mID}
		proc.threads[mID] = thr
	}
	return thr
}

type stackCacheEntry struct {
	pcs   []uint64
	stack lockgraph.Stack
}

// Stack symbolizes the stack given by pcs.
func (proc *Proc) Stack(st *lockgraph.StackTable, pcs []uint64) lockgraph.Stack {
	// Check cached whole stacks first.

	// Hash the stack.
	var h uint64
	for _, pc := range pcs {
		h += pc
		h += h << 10
		h ^= h >> 6
	}
	h += h << 3
	h ^= h >> 11
	// Check the map.
	stacks := proc.stacks[h]
	for _, s2 := range stacks {
		if stackEq(pcs, s2.pcs) {
			return s2.stack
		}
	}

	// Not found. Symbolize the stack.
	var out lockgraph.Stack
	lines := proc.DWARF.Lines
	for _, pc := range pcs {
		// These are all return PCs, so back up to the call.
		pc--

		// Check the fragment cache.
		if frag, ok := proc.frags[pc]; ok {
			out = append(out, frag...)
			continue
		}

		// Resolve the PC to a fragment.
		var frag []lockgraph.SrcPos
		pcStk, err := lines.Lookup(proc.MemToFile(pc))
		if err != nil {
			frag = append(frag, lockgraph.SrcPos{Error: fmt.Sprintf("%#x: %s", pc, err)})
		} else if pcStk == nil {
			frag = append(frag, lockgraph.SrcPos{Error: fmt.Sprintf("%#x: PC not found", pc)})
		} else {
			for _, frame := range pcStk {
				file := st.Files.Add(frame.File.Name)
				fun := st.Funcs.Add(lines.FuncName(frame.Func))
				frag = append(frag, lockgraph.SrcPos{File: file, Func: fun, Line: frame.Line})
			}
		}
		proc.frags[pc] = frag
		out = append(out, frag...)
	}

	// Update the stack cache.
	proc.stacks[h] = append(proc.stacks[h], stackCacheEntry{pcs, out})
	return out
}

func stackEq(s1, s2 []uint64) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// MemToFile translates an address in the in-memory process image to
// an address in the on-disk file.
func (proc *Proc) MemToFile(addr uint64) uint64 {
	return addr - proc.pieDisp
}
