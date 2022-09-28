// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
)

var atExitFuncs []func()

func AtExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

func Exit(code int) {
	for i := len(atExitFuncs) - 1; i >= 0; i-- {
		f := atExitFuncs[i]
		atExitFuncs = atExitFuncs[:i]
		f()
	}
	os.Exit(code)
}

// To enable tracing support (-t flag), set EnableTrace to true.
const EnableTrace = false

func init() {
	if os.Getenv("GOGC") != "" {
		// do not interfere with requested GOGC.
		return
	}
	logHeapTweaks := os.Getenv("GO_gc_LOG_HEAP_TWEAKS") != ""
	const allocs = "/gc/heap/allocs:bytes"
	const frees = "/gc/heap/frees:bytes"
	const goal = "/gc/heap/goal:bytes"
	const startHeapGoal = uint64(4 * 1024 * 1024) // first GC has this goal and it does not adjust
	const threshold = uint64(64 * 1024 * 1024)    // 64M -- Once heap size (live + new) exceeds threshold, back off

	sample := []metrics.Sample{metrics.Sample{Name: allocs}, metrics.Sample{Name: frees}, metrics.Sample{Name: goal}}
	metrics.Read(sample)
	for _, s := range sample {
		if s.Value.Kind() == metrics.KindBad {
			// fmt.Fprintf(os.Stderr, "Unexpected kind-bad for metric %s\n", s.Name)
			return
		}
	}

	// Tinker with GOGC to make the heap grow rapidly at first.
	currentGoal := sample[2].Value.Uint64() // Believe this will be 4MByte or less, perhaps 512k
	// Recall goal = live (assumed 2M at start) plus GOC * live, hence initially 4M

	myGogc := (100 + 100*threshold) / currentGoal //

	// Note that notification is driven by finalization, so expect heap to hit desired goal, GC, then finalizer, and allocs-frees will be near-to-above threshold.
	debug.SetGCPercent(int(myGogc))

	// GC quirk -- despite increasing GOGC, for the first collection that does not affect the goal.
	// therefore make the finalizer tolerate running very early.
	var adjustFunc func(pMyGogc *uint64)
	adjustFunc = func(pMyGogc *uint64) {
		metrics.Read(sample)
		inUse := sample[0].Value.Uint64() - sample[1].Value.Uint64()
		if inUse < startHeapGoal {
			inUse = startHeapGoal
		}
		goal := sample[2].Value.Uint64()
		if inUse < 2*threshold/3 {
			oldfactor := *pMyGogc
			*pMyGogc = 100 * threshold / inUse
			if logHeapTweaks {
				fmt.Fprintf(os.Stderr, "Retry GOGC adjust, current inuse %d, current goal %d, gogc was %d, is now %d\n",
					inUse, goal, oldfactor, *pMyGogc)
			}
			debug.SetGCPercent(int(*pMyGogc))
			runtime.SetFinalizer(pMyGogc, adjustFunc) // Repeat, hoping for a better result.
			return
		}
		debug.SetGCPercent(100)
		if logHeapTweaks {
			fmt.Fprintf(os.Stderr, "Reset GC, current inuse %d, old goal %d, gogc was %d\n",
				inUse, sample[2].Value.Uint64(), *pMyGogc)
		}
	}

	runtime.SetFinalizer(&myGogc, adjustFunc)
}

func Compiling(pkgs []string) bool {
	if Ctxt.Pkgpath != "" {
		for _, p := range pkgs {
			if Ctxt.Pkgpath == p {
				return true
			}
		}
	}

	return false
}

// The racewalk pass is currently handled in three parts.
//
// First, for flag_race, it inserts calls to racefuncenter and
// racefuncexit at the start and end (respectively) of each
// function. This is handled below.
//
// Second, during buildssa, it inserts appropriate instrumentation
// calls immediately before each memory load or store. This is handled
// by the (*state).instrument method in ssa.go, so here we just set
// the Func.InstrumentBody flag as needed. For background on why this
// is done during SSA construction rather than a separate SSA pass,
// see issue #19054.
//
// Third we remove calls to racefuncenter and racefuncexit, for leaf
// functions without instrumented operations. This is done as part of
// ssa opt pass via special rule.

// TODO(dvyukov): do not instrument initialization as writes:
// a := make([]int, 10)

// Do not instrument the following packages at all,
// at best instrumentation would cause infinite recursion.
var NoInstrumentPkgs = []string{
	"runtime/internal/atomic",
	"runtime/internal/math",
	"runtime/internal/sys",
	"runtime/internal/syscall",
	"runtime",
	"runtime/race",
	"runtime/msan",
	"runtime/asan",
	"internal/cpu",
}

// Don't insert racefuncenter/racefuncexit into the following packages.
// Memory accesses in the packages are either uninteresting or will cause false positives.
var NoRacePkgs = []string{"sync", "sync/atomic"}
