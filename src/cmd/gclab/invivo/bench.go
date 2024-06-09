// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package invivo

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Benchmark struct {
	name string

	reportAll bool

	once    sync.Once
	runPool sync.Pool // of *runInternal, with metrics that are children of rootMetrics

	lock        sync.Mutex
	runs        int
	invalid     int
	rootMetrics []metricAccum // root metrics, allocated under once.
}

type Run struct {
	runInternal *runInternal
	poolSeq     uint32
}

type runInternal struct {
	poolSeq uint32

	b     *Benchmark
	accum time.Duration
	start time.Time

	validSeq uint64
	invalid  bool

	metrics []metricAccum
}

var allBenchmarks []*Benchmark

var valid atomic.Uint64

var numRuns atomic.Int64

func NewBenchmark(name string) *Benchmark {
	b := &Benchmark{name: name}
	allBenchmarks = append(allBenchmarks, b)
	return b
}

func (b *Benchmark) ReportAll() *Benchmark {
	b.reportAll = true
	return b
}

func (b *Benchmark) Start() Run {
	b.once.Do(func() {
		// Set up the metrics the first time we start a benchmark. We can't do
		// this in NewBenchmark because allMetrics may not be fully populated at
		// that time. (There are other ways to do this, like reinitializing on
		// every registerMetric, or initializing the root metrics for ALL
		// benchmarks on the first Start.)
		root := make([]metricAccum, len(allMetrics))
		for i := range root {
			root[i] = allMetrics[i].new()
		}
		b.rootMetrics = root
		b.runPool.New = func() any {
			accums := make([]metricAccum, len(root))
			for i := range accums {
				accums[i] = root[i].new()
			}
			return &runInternal{
				b:       b,
				metrics: accums,
			}
		}
	})

	// We try really hard to avoid allocation as part of Run. Hence, we use a
	// pool for the internals of Run. This is hidden from the user, but of
	// course we can't prevent the user from copying the Run itself, so we also
	// use a sequence number to protect against use-after-free bugs.
	numRuns.Add(1)
	internal := b.runPool.Get().(*runInternal)
	r := Run{internal, internal.poolSeq}
	r.StartTimer()
	return r
}

func (r Run) internal() *runInternal {
	if r.poolSeq != r.runInternal.poolSeq {
		panic("Run reused after Done")
	}
	return r.runInternal
}

func (r Run) StopTimer() {
	ri := r.internal()
	if ri.start.IsZero() {
		return
	}
	ri.accum += time.Since(ri.start)
	ri.start = time.Time{}
	ri.invalid = ri.invalid || ri.validSeq != valid.Load()
}

func (r Run) StartTimer() {
	ri := r.internal()
	if !ri.start.IsZero() {
		return
	}
	ri.validSeq = valid.Load()
	ri.start = time.Now()
}

func (r Run) Elapsed() time.Duration {
	ri := r.internal()
	e := ri.accum
	if !ri.start.IsZero() {
		e += time.Since(ri.start)
	}
	return e
}

func (r Run) Done() {
	r.doneInternal(false, "")
}

// DoneImmediate is like Done, but immediately reports this iteration's results,
// optionally with a sub-benchmark name.
func (r Run) DoneImmediate(subBenchmark string) {
	r.doneInternal(true, subBenchmark)
}

func (r Run) doneInternal(report bool, subBenchmark string) {
	if r.runInternal == nil {
		panic("Done already called")
	}

	r.StopTimer()

	ri := r.internal()

	metricNSPerOp.Set(r, float64(ri.accum))

	ri.b.lock.Lock()
	defer ri.b.lock.Unlock()

	ri.b.runs++
	if ri.invalid {
		ri.b.invalid++
	}

	if !ri.invalid {
		if report {
			reportOne(ri.b.name, subBenchmark, 1, ri.metrics)
		} else if ri.b.reportAll {
			reportOne(ri.b.name+"One", "", 1, ri.metrics)
		}
	}

	// Merge our metrics into the benchmark
	for _, m := range ri.metrics {
		m.commit()
	}

	// Clear fields and return to the pool
	r.runInternal = nil
	ri.accum = 0
	ri.invalid = false
	ri.poolSeq++
	ri.b.runPool.Put(ri)
	numRuns.Add(-1)
}

// Invalidate invalidates the results of all currently running benchmarks.
func Invalidate() {
	valid.Add(1)
}

func Report() {
	if v := numRuns.Load(); v != 0 {
		panic(fmt.Sprintf("%d runs still pending", v))
	}

	for _, b := range allBenchmarks {
		if b.runs == 0 {
			// No runs
			continue
		}
		if b.reportAll {
			// We reported the individual runs
			continue
		}

		if b.runs == b.invalid {
			fmt.Printf("# Warning: All runs invalid\n# ")
		} else if b.invalid > 0 {
			fmt.Printf("# Warning: %d runs invalid\n", b.invalid)
		}

		reportOne(b.name, "", b.runs, b.rootMetrics)

		// Clear the benchmark
		b.runs = 0
		b.invalid = 0
		for i := range b.rootMetrics {
			b.rootMetrics[i].reset()
		}
	}
}

func reportOne(name, subName string, runs int, metrics []metricAccum) {
	// Report missing metrics first.
	for i, d := range metrics {
		if d.count() != 0 && d.count() != runs {
			fmt.Printf("# Warning: %q has samples from %d runs of %d\n", allMetrics[i].name, d.count(), runs)
		}
	}

	if subName == "" {
		fmt.Printf("Benchmark%s\t%d", name, runs)
	} else {
		fmt.Printf("Benchmark%s/%s\t%d", name, subName, runs)
	}

	for i, d := range metrics {
		if d.count() == 0 {
			continue
		}
		fmt.Printf("\t%f %s", d.report(), allMetrics[i].name)
	}

	fmt.Printf("\n")
}
