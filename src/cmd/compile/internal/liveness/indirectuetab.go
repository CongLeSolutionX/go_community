// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package liveness

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/ssa"
	"fmt"
	"sort"
	"strings"
)

// IndUETab is a table that maps ssa.ID's to a list of indirect
// upwards-exposed uses at the inst defined by that ID. It can be
// thought of as a more space-efficient implementation of
// "map[ssa.ID][]*ir.Name".
type IndUETab struct {
	names         []*ir.Name
	nameToSlot    map[*ir.Name]int
	tmp           []int
	tmpid         ssa.ID
	ueListAtValue map[ssa.ID]int
	nameLog       []int
}

// MakeIndUETab creates/initializes an object of type IndUETab.
func MakeIndUETab(names []*ir.Name, sizeHint int) *IndUETab {
	nameToSlot := make(map[*ir.Name]int)
	for slot, n := range names {
		nameToSlot[n] = slot
	}
	// FIXME: replace with slices.Clone
	nnames := make([]*ir.Name, len(names))
	copy(nnames, names)
	return &IndUETab{
		names:         nnames,
		nameToSlot:    nameToSlot,
		nameLog:       make([]int, 0, sizeHint),
		ueListAtValue: make(map[ssa.ID]int),
		tmpid:         -1,
	}
}

// Add is used to record that we see an indirect upwards-exposed use
// name n at the instruction that defines SSA value id. A sequence of
// calls to Add (on a given instruction) should be followed by a call
// to Finalize for that ssa value.
// is that if
func (t *IndUETab) Add(id ssa.ID, n *ir.Name) error {
	if t.tmpid == -1 {
		t.tmpid = id
	} else if t.tmpid != id {
		return fmt.Errorf("successive calls to Add with different ids: v%d v%d",
			t.tmpid, id)
	}
	sl, ok := t.nameToSlot[n]
	if !ok {
		return fmt.Errorf("internal error: no slot for name %q", n.Sym().Name)
	}
	t.tmp = append(t.tmp, sl)
	return nil
}

// Finalize informs the table that we're done recording new upwards
// exposes uses at value
func (t *IndUETab) Finalize(id ssa.ID) error {
	if len(t.tmp) == 0 {
		return nil
	}
	if id != t.tmpid {
		return fmt.Errorf("Add / Finalize id clash: v%d v%d", t.tmpid, id)
	}
	pos := len(t.nameLog)
	t.ueListAtValue[id] = pos
	for _, sl := range t.tmp {
		t.nameLog = append(t.nameLog, sl)
	}
	t.tmp = t.tmp[:0]
	t.tmpid = -1
	t.nameLog = append(t.nameLog, -1)
	return nil
}

// Get returns a list of the names recorded as having indirect upwards
// exposes uses at the inst that defines ssa value id.
func (t *IndUETab) Get(id ssa.ID, tmp []*ir.Name) ([]*ir.Name, error) {
	if len(t.tmp) != 0 {
		return tmp, fmt.Errorf("pending adds on Get")
	}
	tmp = tmp[:0]
	logPos, ok := t.ueListAtValue[id]
	if !ok {
		return tmp, nil
	}
	nameSlot := t.nameLog[logPos]
	for nameSlot != -1 {
		tmp = append(tmp, t.names[nameSlot])
		logPos++
		nameSlot = t.nameLog[logPos]
	}
	return tmp, nil
}

func (t *IndUETab) String() string {
	if t == nil {
		return "<nil>"
	}
	if len(t.tmp) != 0 {
		return fmt.Sprintf("pending adds: %+v", t.tmp)
	}
	var sb strings.Builder
	ids := make([]ssa.ID, 0, len(t.ueListAtValue))
	for k := range t.ueListAtValue {
		ids = append(ids, k)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var tmp []*ir.Name
	for k, id := range ids {
		fmt.Fprintf(&sb, " v%d:", id)
		tmp, _ = t.Get(id, tmp)
		for _, n := range tmp {
			fmt.Fprintf(&sb, " %s", n.Sym().Name)
		}
		if k != len(ids)-1 {
			fmt.Fprintf(&sb, "\n")
		}
	}
	return sb.String()
}
