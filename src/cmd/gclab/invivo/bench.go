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

	once        sync.Once
	metricsPool sync.Pool // of []metricI, children of rootMetrics

	lock        sync.Mutex
	runs        int
	invalid     int
	rootMetrics []metricAccum // root metrics, allocated under once.
}

type Run struct {
	b     *Benchmark
	accum time.Duration
	start time.Time

	validSeq uint64
	invalid  bool

	metrics []metricAccum
}

var allBenchmarks []*Benchmark

var valid atomic.Uint64

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
		root := make([]metricAccum, len(allMetrics))
		for i := range root {
			root[i] = allMetrics[i].new()
		}
		b.rootMetrics = root
		b.metricsPool.New = func() any {
			accums := make([]metricAccum, len(root))
			for i := range accums {
				accums[i] = root[i].new()
			}
			return accums
		}
	})

	r := Run{
		b:       b,
		metrics: b.metricsPool.Get().([]metricAccum),
	}
	r.StartTimer()
	return r
}

func (r *Run) StopTimer() {
	if r.start.IsZero() {
		return
	}
	r.accum += time.Since(r.start)
	r.start = time.Time{}
	r.invalid = r.invalid || r.validSeq != valid.Load()
}

func (r *Run) StartTimer() {
	if !r.start.IsZero() {
		return
	}
	r.validSeq = valid.Load()
	r.start = time.Now()
}

func (r Run) Elapsed() time.Duration {
	e := r.accum
	if !r.start.IsZero() {
		e += time.Since(r.start)
	}
	return e
}

func (r *Run) Done() {
	r.StopTimer()

	if r.b == nil {
		panic("Done already called")
	}

	metricNSPerOp.Set(r, float64(r.accum))

	r.b.lock.Lock()
	defer r.b.lock.Unlock()

	r.b.runs++
	if r.invalid {
		r.b.invalid++
	}

	if r.b.reportAll && !r.invalid {
		reportOne(r.b.name+"One", 1, r.metrics)
	}

	// Merge our metrics into the benchmark
	for _, m := range r.metrics {
		m.commit()
	}
	r.b.metricsPool.Put(r.metrics)
	r.metrics = nil

	r.b = nil
}

// Invalidate invalidates the results of all currently running benchmarks.
func Invalidate() {
	valid.Add(1)
}

func Report() {
	for _, b := range allBenchmarks {
		if b.runs == 0 {
			// No runs
			continue
		}

		if b.runs == b.invalid {
			fmt.Printf("# Warning: All runs invalid\n# ")
		} else if b.invalid > 0 {
			fmt.Printf("# Warning: %d runs invalid\n", b.invalid)
		}

		reportOne(b.name, b.runs, b.rootMetrics)
	}
}

func reportOne(name string, runs int, metrics []metricAccum) {
	// Report missing metrics first.
	for i, d := range metrics {
		if d.count() != 0 && d.count() != runs {
			fmt.Printf("# Warning: %q has samples from %d runs of %d\n", allMetrics[i].name, d.count, runs)
		}
	}

	fmt.Printf("Benchmark%s\t%d", name, runs)

	for i, d := range metrics {
		if d.count() == 0 {
			continue
		}
		fmt.Printf("\t%v %s", d.report(), allMetrics[i].name)
	}

	fmt.Printf("\n")
}
