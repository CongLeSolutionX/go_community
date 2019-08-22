// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"log"
	"sort"

	"lockcheck/lockgraph"
	"lockcheck/symbolize"
)

// A GraphBuilder dynamically constructs a LockGraph.
type GraphBuilder struct {
	ms      map[*Thread]*mState
	procs   map[*Proc]*procState
	classes map[string]*lockClass // Label -> lock class

	// stacks is the stack table against which Stacks are symbolized.
	stacks *lockgraph.StackTable
}

// NewGraphBuilder returns a new builder for a lockgraph.Graph.
func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{
		ms:      make(map[*Thread]*mState),
		procs:   make(map[*Proc]*procState),
		classes: make(map[string]*lockClass),
		stacks:  new(lockgraph.StackTable),
	}
}

// Acquire updates the graph for a lock acquisition.
func (gb *GraphBuilder) Acquire(thr *Thread, lockID uint64, stack lockgraph.Stack) {
	m := gb.getM(thr)
	l2 := gb.staticLock(thr, lockID, stack)
	// Add edges from all currently held locks.
	for _, l1 := range m.held {
		l1.addOut(l2)
	}
	// Add the lock to the set of held locks.
	m.held = append(m.held, l2)
}

// Release updates the graph for a lock release.
func (gb *GraphBuilder) Release(thr *Thread, lockID uint64) {
	m := gb.getM(thr)
	// Remove lockID from the held set.
	for i := range m.held {
		if m.held[i].lockID == lockID {
			m.held[i] = m.held[len(m.held)-1]
			m.held = m.held[:len(m.held)-1]
			return
		}
	}
	log.Printf("release without acquire (%s, lock %#x)", thr, lockID)
}

// MayAcquire updates the graph for an acquire/release of a lock.
func (gb *GraphBuilder) MayAcquire(thr *Thread, lockID uint64, stack lockgraph.Stack) {
	m := gb.getM(thr)
	if len(m.held) == 0 {
		// The lock can't participate in the lock order, so
		// don't bother recording it.
		return
	}
	l2 := gb.staticLock(thr, lockID, stack)
	for _, l1 := range m.held {
		l1.addOut(l2)
	}
}

// AcquireLabeled is like Acquire, but with an explicit class and
// rank.
func (gb *GraphBuilder) AcquireLabeled(thr *Thread, lockID uint64, stack lockgraph.Stack, class string, rank uint64) {
	// Record the class.
	cls, ok := gb.classes[class]
	if !ok {
		cls = &lockClass{label: class}
		gb.classes[cls.label] = cls
	}
	// Add edges from all currently held locks.
	m := gb.getM(thr)
	l2 := heldLock{lockID, rank, cls, stack}
	for _, l1 := range m.held {
		l1.addOut(l2)
	}
	// Add the lock to the set of held locks.
	m.held = append(m.held, l2)
}

// Finish finalizes the graph being build in gb into a lockgraph.Graph.
func (gb *GraphBuilder) Finish() *lockgraph.Graph {
	// Assign stable indexes to lock classes.
	var classes []*lockClass
	nOut := 0
	for _, cls := range gb.classes {
		classes = append(classes, cls)
		nOut += len(cls.out)
	}
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].label < classes[j].label
	})
	// Map classes back to indexes.
	indexes := make(map[*lockClass]int, len(classes))
	for i, cls := range classes {
		indexes[cls] = i
	}
	// Map class graph to integer out-edges.
	labels := make([]string, 0, len(classes))
	edges := make([][]lockgraph.Edge, 0, len(classes))
	allOut := make([]int, nOut)
	out := make([][]int, 0, len(classes))
	for _, cls1 := range classes {
		labels = append(labels, cls1.label)
		edgesOut := make([]lockgraph.Edge, 0, len(cls1.out))
		clsOut := allOut[:0:len(cls1.out)]
		allOut = allOut[len(cls1.out):]
		for cls2, edge2 := range cls1.out {
			clsOut = append(clsOut, indexes[cls2])
			edgesOut = append(edgesOut, edge2.finish())
		}
		edges = append(edges, edgesOut)
		out = append(out, clsOut)
	}
	return &lockgraph.Graph{labels, edges, out, gb.stacks}
}

// getM returns the mState for thr, creating it if necessary.
func (gb *GraphBuilder) getM(thr *Thread) *mState {
	m, ok := gb.ms[thr]
	if !ok {
		m = &mState{
			proc: gb.getProc(thr),
		}
		gb.ms[thr] = m
	}
	return m
}

// getProc returns the procState for thr, creating it if necessary.
func (gb *GraphBuilder) getProc(thr *Thread) *procState {
	proc, ok := gb.procs[thr.Proc]
	if !ok {
		proc = &procState{
			staticClasses: make(map[uint64]*lockClass),
		}
		gb.procs[thr.Proc] = proc
	}
	return proc
}

// staticLock returns a heldLock for the static lock at address lockID
// in thr's process. It constructs the lock's class using DWARF.
func (gb *GraphBuilder) staticLock(thr *Thread, lockID uint64, stk lockgraph.Stack) heldLock {
	proc := gb.getProc(thr)
	cls, ok := proc.staticClasses[lockID]
	if !ok {
		// Construct unique symbolic lock class name. This is
		// derived from the lock address, but the class may
		// already exist from another process.
		var label string
		addr := thr.Proc.MemToFile(lockID)
		if v, ok := thr.Proc.DWARF.Vars.Lookup(addr); ok {
			label, ok = gb.prettyOffset(thr.Proc.DWARF, v, addr-v.Addr)
			if !ok && addr-v.Addr < (1<<20) {
				label = fmt.Sprintf("%s+%#x", v.Name, addr-v.Addr)
			}
		}
		if label == "" {
			gb.warnUnlabeled(thr, lockID, stk)
			label = fmt.Sprintf("%#x", lockID)
		}
		cls, ok = gb.classes[label]
		if !ok {
			cls = &lockClass{label: label}
			gb.classes[cls.label] = cls
		}
		// Record the class of this lock ID so we can quickly
		// look it up in the future.
		proc.staticClasses[lockID] = cls
	}
	return heldLock{lockID, 0, cls, stk}
}

// prettyOffset pretty-prints byte offset off from variable v.
func (gb *GraphBuilder) prettyOffset(dwarf *symbolize.DWARF, v symbolize.Var, off uint64) (string, bool) {
	typ, err := dwarf.Vars.VarType(v)
	if err != nil {
		return "", false
	}
	path := dwarf.OffsetPath(typ, off)
	if path == nil {
		return "", false
	}
	label := v.Name
	for _, off := range path {
		if off.StructName == "runtime.mutex" {
			// We found the mutex. Don't walk into its fields.
			break
		}
		if off.FieldName != "" {
			label += "." + off.FieldName
		} else {
			label += "[]"
		}
	}
	return label, true
}

// warnUnlabeled prints a warning that we couldn't resolve the lock
// class label of lockID.
func (gb *GraphBuilder) warnUnlabeled(thr *Thread, lockID uint64, stk lockgraph.Stack) {
	log.Printf("unlabeled lock %#x in %s at:\n%s", lockID, thr, gb.stacks.StringStack(stk))
}

type mState struct {
	proc *procState

	// held stores the locks currently held by this M and the
	// stacks at which they were acquired.
	held []heldLock
}

type heldLock struct {
	lockID uint64
	rank   uint64
	cls    *lockClass
	stack  lockgraph.Stack // Acquire stack
}

// addOut records that l2 was acquired while l1 was held.
func (l1 heldLock) addOut(l2 heldLock) {
	// If this is a lock of the same class, only create
	// down-rank edges. This way rank violations will
	// appear as self-cycles.
	if l1.cls == l2.cls && l1.rank < l2.rank {
		return
	}

	if l1.cls.out == nil {
		l1.cls.out = make(map[*lockClass]edgeBuilder)
	}
	// Add to the edge set.
	edges, ok := l1.cls.out[l2.cls]
	if !ok {
		edges = edgeBuilder{}
		l1.cls.out[l2.cls] = edges
	}
	edges.add(l1.stack, l2.stack)
}

type procState struct {
	// staticClasses is the set of classes for static locks in
	// this process.
	staticClasses map[uint64]*lockClass
}

// lockClass represents a lock class in the GraphBuilder.
type lockClass struct {
	label string

	// out lists the out-edges from this lock. That is, the locks
	// acquired with this lock held. For each out-edge, it tracks
	// the stacks at which this lock was acquired and the target
	// lock was acquired.
	out map[*lockClass]edgeBuilder
}

// A edgeBuilder accumulates information about an edge in the lock
// graph. Specifically, it is a counted set of stack pairs.
type edgeBuilder map[uint64][]lockgraph.StackPair

func (s edgeBuilder) add(l1s, l2s lockgraph.Stack) {
	hash := l1s.Hash() ^ l2s.Hash()
	for i, edge := range s[hash] {
		if edge.S1.Equals(l1s) && edge.S2.Equals(l2s) {
			// Found it.
			s[hash][i].Count++
			return
		}
	}
	s[hash] = append(s[hash], lockgraph.StackPair{l1s, l2s, 1})
}

func (s edgeBuilder) finish() lockgraph.Edge {
	var out []lockgraph.StackPair
	for _, edges := range s {
		out = append(out, edges...)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Count > out[j].Count
	})
	return lockgraph.Edge{Stacks: out}
}
