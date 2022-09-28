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

type finalizable struct {
	a [4]uint64 // 32 bytes, enough to ensure prompt finalization.
}

// AdjustStartingHeap modifies GOGC so that the heap will grow to the
// requested size before GC occurs.  Once this size is reached
// (approximately) GOGC will be reset to 100.
//
// Setting the environment variable GOgc_ASH_LOG to a non-empty
// string will log GC adjustments to os.Stderr.
//
// NOTE: If you think this code would help startup time in your own
// application and you decide to use it, please benchmark first to see if it
// actually works for you (it may not: the Go compiler is not typical), and
// whatever the outcome, please leave a comment on bug #VWXYZ.  This code
// uses supported interfaces, but depends more than we like on observed
// behavior of the garbage collector, so if many people need this feature, we
// should consider/propose a better way to accomplish it.
func AdjustStartingHeap(requestedHeapGoal uint64) {
	logHeapTweaks := os.Getenv("GOgc_ASH_LOG") != ""

	mp := runtime.GOMAXPROCS(0)

	const (
		allocs = "/gc/heap/allocs:bytes"
		frees  = "/gc/heap/frees:bytes"
		goal   = "/gc/heap/goal:bytes"
		count  = "/gc/cycles/total:gc-cycles"
	)

	// Go's GC has to guess at its initial heap limit (the point at which its
	// first GC should complete) and currently that guess is 4MiB.  It is
	// currently possible that it could be as low as 0.5MiB. The first GC's
	// goal is also, currently, not adjustable, and it is not entirely clear
	// when it occurs, so this code checks after each GC and readjusts GOGC
	// until the goal is approximately reached.
	const startHeapGoal = uint64(4 * 1024 * 1024)

	sample := []metrics.Sample{{Name: allocs}, {Name: frees}, {Name: goal}, {Name: count}}
	const (
		ALLOCS = 0
		FREES  = 1
		GOAL   = 2
		COUNT  = 3
	)
	metrics.Read(sample)
	for _, s := range sample {
		if s.Value.Kind() == metrics.KindBad {
			// Just return, a slightly slower compilation is a tolerable outcome.
			if logHeapTweaks {
				fmt.Fprintf(os.Stderr, "GOgc_ASH_Regret: unexpected KindBad for metric %s\n", s.Name)
			}
			return
		}
	}

	// Tinker with GOGC to make the heap grow rapidly at first.
	currentGoal := sample[GOAL].Value.Uint64() // Believe this will be 4MByte or less, perhaps 512k
	myGogc := 100 * requestedHeapGoal / currentGoal
	if myGogc <= 150 {
		return
	}

	if logHeapTweaks {
		AtExit(func() {
			metrics.Read(sample)
			goal := sample[GOAL].Value.Uint64()
			count := sample[COUNT].Value.Uint64()
			inUse := sample[ALLOCS].Value.Uint64() - sample[FREES].Value.Uint64()
			oldGogc := debug.SetGCPercent(100)
			fmt.Fprintf(os.Stderr, "GOgc_ASH_Result: inuse %d, goal %d gogc %d count %d maxprocs %d\n", inUse, goal, oldGogc, count, mp)
		})
	}

	debug.SetGCPercent(int(myGogc))

	// Set up a finalizer to detect garbage collection, and when that occurs,
	// check to see what new value for GOGC is most appropriate.  If the
	// memory in use is close-to-or-above the goal heap size, set GOGC to 100
	// and stop fiddling with it.

	var adjustFunc func(px *finalizable)
	adjustFunc = func(px *finalizable) {
		metrics.Read(sample)
		goal := sample[GOAL].Value.Uint64()
		count := sample[COUNT].Value.Uint64()

		// This calculated value of inUse is NOT the same as goal, per experiments.
		// If the heap is growing, steady-state assumptions don't hold.
		inUse := sample[ALLOCS].Value.Uint64() - sample[FREES].Value.Uint64()
		if inUse < startHeapGoal {
			inUse = startHeapGoal
		}

		diff := goal - requestedHeapGoal
		if diff < 0 {
			diff = -diff
		}

		if diff < requestedHeapGoal/8 { // Stay the course
			if logHeapTweaks {
				fmt.Fprintf(os.Stderr, "GOgc_ASH_Reuse: GOGC adjust, current inuse %d, current goal %d, count is %d\n",
					inUse, goal, count)
			}
			runtime.SetFinalizer(px, adjustFunc) // Repeat, hoping for a better result.
			return
		}

		candidateGogc := 100 * requestedHeapGoal / inUse

		if candidateGogc > 150 {
			// Not done growing the heap.
			oldGogc := debug.SetGCPercent(int(candidateGogc))
			if logHeapTweaks {
				fmt.Fprintf(os.Stderr, "GOgc_ASH_Retry: GOGC adjust, current inuse %d, current goal %d, gogc was %d, is now %d, count is %d\n",
					inUse, goal, oldGogc, candidateGogc, count)
			}
			runtime.SetFinalizer(px, adjustFunc) // Repeat, hoping for a better result.
			return
		}

		// In this case we're done boosting GOGC, set it to 100 and don't set a new finalizer.
		oldGogc := debug.SetGCPercent(100)
		if logHeapTweaks {
			fmt.Fprintf(os.Stderr, "GOgc_ASH_Reset: GC, current inuse %d, old goal %d, gogc was %d, count is %d\n",
				inUse, goal, oldGogc, count)
		}
	}

	var x finalizable // the only purpose for x is to have a finalizer attached to it.
	runtime.SetFinalizer(&x, adjustFunc)
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
