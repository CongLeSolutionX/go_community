// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"fmt"
	"sort"
)

//!!! unexport
type Batch struct {
	events   []*Event
	selected bool
}

type Frontier struct {
	ev    *Event
	batch int
	g     uint64
	curr  gState
	next  gState
}

type gState struct {
	seq    uint64
	status int
}

type frontierList []Frontier

func (l frontierList) Len() int {
	return len(l)
}

func (l frontierList) Less(i, j int) bool {
	return l[i].ev.Ts < l[j].ev.Ts
}

func (l frontierList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type eventList []*Event

func (l eventList) Len() int {
	return len(l)
}

func (l eventList) Less(i, j int) bool {
	return l[i].Ts < l[j].Ts
}

func (l eventList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func order(m map[int][]*Event) (events []*Event, err error) {
	pending := 0
	var batches []*Batch
	for _, v := range m {
		pending += len(v)
		batches = append(batches, &Batch{v, false})
	}
	gs := make(map[uint64]gState)
	var frontier []Frontier
	for ; pending != 0; pending-- {
		const (
			gDead = iota
			gRunnable
			gRunning
			gWaiting

			unordered = ^uint64(0)
			garbage   = ^uint64(0) - 1
			noseq     = ^uint64(0)
			seqinc    = ^uint64(0) - 1
		)
		for i, b := range batches {
			if b.selected || len(b.events) == 0 {
				continue
			}
			ev := b.events[0]
			g := unordered
			var curr, next gState
			switch ev.Type {
			case EvGoCreate:
				g = ev.Args[0]
				curr = gState{0, gDead}
				next = gState{1, gRunnable}
			case EvGoWaiting, EvGoInSyscall:
				g = ev.Args[0]
				curr = gState{1, gRunnable}
				next = gState{2, gWaiting}
			case EvGoStart:
				g = ev.G
				curr = gState{ev.Args[1], gRunnable}
				next = gState{ev.Args[1] + 1, gRunning}

			case EvGoStartLocal:
				ev.Type = EvGoStart
				g = ev.G
				curr = gState{noseq, gRunnable}
				next = gState{seqinc, gRunning}

			case EvGoBlock, EvGoBlockSend, EvGoBlockRecv, EvGoBlockSelect,
				EvGoBlockSync, EvGoBlockCond, EvGoBlockNet, EvGoSleep, EvGoSysBlock:
				g = ev.G
				curr = gState{noseq, gRunning}
				next = gState{noseq, gWaiting}
			case EvGoSched, EvGoPreempt:
				g = ev.G
				curr = gState{noseq, gRunning}
				next = gState{noseq, gRunnable}
			case EvGoUnblock, EvGoSysExit:
				g = ev.Args[0]
				curr = gState{ev.Args[1], gWaiting}
				next = gState{ev.Args[1] + 1, gRunnable}
			case EvGoUnblockLocal, EvGoSysExitLocal:
				if ev.Type == EvGoUnblockLocal {
					ev.Type = EvGoUnblock
				} else {
					ev.Type = EvGoSysExit
				}
				g = ev.Args[0]
				curr = gState{noseq, gWaiting}
				next = gState{seqinc, gRunnable}
			case EvGCStart:
				g = garbage
				curr = gState{ev.Args[0], gDead}
				next = gState{ev.Args[0] + 1, gDead}
			}
			if g != unordered {
				state := gs[g]
				if curr.seq != noseq && curr.seq != state.seq || curr.status != state.status {
					continue
				}
			}
			frontier = append(frontier, Frontier{ev, i, g, curr, next})
			b.events = b.events[1:]
			b.selected = true
		}
		if len(frontier) == 0 {
			return nil, fmt.Errorf("no consistent ordering of events possible")
		}
		sort.Sort(frontierList(frontier))
		f := frontier[0]
		frontier[0] = frontier[len(frontier)-1]
		frontier = frontier[:len(frontier)-1]
		events = append(events, f.ev)
		if f.g != unordered {
			state := gs[f.g]
			if f.curr.seq != noseq && f.curr.seq != state.seq || f.curr.status != state.status {
				panic("event sequences are broken")
			}
			switch f.next.seq {
			case noseq:
				f.next.seq = state.seq
			case seqinc:
				f.next.seq = state.seq + 1
			}
			gs[f.g] = f.next
		}
		if !batches[f.batch].selected {
			panic("frontier batch is not selected")
		}
		batches[f.batch].selected = false
	}

	if true {
		// Make sure time stamps respect sequence numbers.
		// The tests will skip (not fail) the test case if they see this error,
		// so check everything else that could possibly be wrong first.
		if !sort.IsSorted(eventList(events)) {
			return nil, ErrTimeOrder
		}
	}

	lastSysBlock := make(map[uint64]int64)
	for _, ev := range events {
		switch ev.Type {
		case EvGoSysBlock:
			lastSysBlock[ev.G] = ev.Ts
		case EvGoSysExit:
			// EvGoSysExit emission is delayed until the thread has a P.
			// Give it the real sequence number and time stamp.
			ts := int64(ev.Args[2])
			if ts == 0 {
				continue
			}
			if ts < lastSysBlock[ev.G] {
				return nil, ErrTimeOrder
			}
			ev.Ts = ts
		}
	}
	sort.Sort(eventList(events))

	return
}
