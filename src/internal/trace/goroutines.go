// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

import (
	"strings"
)

// GDesc contains statistics and execution details of a single goroutine.
type GDesc struct {
	ID           uint64
	Name         string
	PC           uint64
	CreationTime int64
	StartTime    int64
	EndTime      int64

	// List of regions in the goroutine, sorted based on the start time.
	Regions []*UserRegionDesc

	// Statistics of execution time during the goroutine execution.
	GExecutionStat

	*gdesc // private part.
}

// UserRegionDesc represents a region and goroutine execution stats
// while the region was active.
type UserRegionDesc struct {
	TaskID uint64
	Name   string

	// Region start event. Normally EvUserRegion start event or nil,
	// but can be EvGoCreate event if the region is a synthetic
	// region representing task inheritance from the parent goroutine.
	Start *EventV1

	// Region end event. Normally EvUserRegion end event or nil,
	// but can be EvGoStop or EvGoEnd event if the goroutine
	// terminated without explicitly ending the region.
	End *EventV1

	GExecutionStat
}

// GExecutionStat contains statistics about a goroutine's execution
// during a period of time.
type GExecutionStat struct {
	ExecTime      int64
	SchedWaitTime int64
	IOTime        int64
	BlockTime     int64
	SyscallTime   int64
	GCTime        int64
	SweepTime     int64
	TotalTime     int64
}

// sub returns the stats v-s.
func (s GExecutionStat) sub(v GExecutionStat) (r GExecutionStat) {
	r = s
	r.ExecTime -= v.ExecTime
	r.SchedWaitTime -= v.SchedWaitTime
	r.IOTime -= v.IOTime
	r.BlockTime -= v.BlockTime
	r.SyscallTime -= v.SyscallTime
	r.GCTime -= v.GCTime
	r.SweepTime -= v.SweepTime
	r.TotalTime -= v.TotalTime
	return r
}

// snapshotStat returns the snapshot of the goroutine execution statistics.
// This is called as we process the ordered trace event stream. lastTs and
// activeGCStartTime are used to process pending statistics if this is called
// before any goroutine end event.
func (g *GDesc) snapshotStat(lastTs, activeGCStartTime int64) (ret GExecutionStat) {
	ret = g.GExecutionStat

	if g.gdesc == nil {
		return ret // finalized GDesc. No pending state.
	}

	if activeGCStartTime != 0 { // terminating while GC is active
		if g.CreationTime < activeGCStartTime {
			ret.GCTime += lastTs - activeGCStartTime
		} else {
			// The goroutine's lifetime completely overlaps
			// with a GC.
			ret.GCTime += lastTs - g.CreationTime
		}
	}

	if g.TotalTime == 0 {
		ret.TotalTime = lastTs - g.CreationTime
	}

	if g.lastStartTime != 0 {
		ret.ExecTime += lastTs - g.lastStartTime
	}
	if g.blockNetTime != 0 {
		ret.IOTime += lastTs - g.blockNetTime
	}
	if g.blockSyncTime != 0 {
		ret.BlockTime += lastTs - g.blockSyncTime
	}
	if g.blockSyscallTime != 0 {
		ret.SyscallTime += lastTs - g.blockSyscallTime
	}
	if g.blockSchedTime != 0 {
		ret.SchedWaitTime += lastTs - g.blockSchedTime
	}
	if g.blockSweepTime != 0 {
		ret.SweepTime += lastTs - g.blockSweepTime
	}
	return ret
}

// finalize is called when processing a goroutine end event or at
// the end of trace processing. This finalizes the execution stat
// and any active regions in the goroutine, in which case trigger is nil.
func (g *GDesc) finalize(lastTs, activeGCStartTime int64, trigger *EventV1) {
	if trigger != nil {
		g.EndTime = trigger.Ts
	}
	finalStat := g.snapshotStat(lastTs, activeGCStartTime)

	g.GExecutionStat = finalStat

	// System goroutines are never part of regions, even though they
	// "inherit" a task due to creation (EvGoCreate) from within a region.
	// This may happen e.g. if the first GC is triggered within a region,
	// starting the GC worker goroutines.
	if !IsSystemGoroutine(g.Name) {
		for _, s := range g.activeRegions {
			s.End = trigger
			s.GExecutionStat = finalStat.sub(s.GExecutionStat)
			g.Regions = append(g.Regions, s)
		}
	}
	*(g.gdesc) = gdesc{}
}

// gdesc is a private part of GDesc that is required only during analysis.
type gdesc struct {
	lastStartTime    int64
	blockNetTime     int64
	blockSyncTime    int64
	blockSyscallTime int64
	blockSweepTime   int64
	blockGCTime      int64
	blockSchedTime   int64

	activeRegions []*UserRegionDesc // stack of active regions
}

func IsSystemGoroutine(entryFn string) bool {
	// This mimics runtime.isSystemGoroutine as closely as
	// possible.
	// Also, locked g in extra M (with empty entryFn) is system goroutine.
	return entryFn == "" || entryFn != "runtime.main" && strings.HasPrefix(entryFn, "runtime.")
}
