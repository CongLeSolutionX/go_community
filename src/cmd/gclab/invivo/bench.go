// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package invivo

import (
	"bytes"
	"fmt"
	"sync"
	"time"
)

type Benchmark struct {
	name string

	lock    sync.Mutex
	metrics []dist
}

type dist struct {
	count int
	total float64
}

type Run struct {
	b     *Benchmark
	accum time.Duration
	start time.Time

	metrics []dist
}

type Metric struct {
	name string
	id   int
}

var allBenchmarks []*Benchmark

var allMetrics = []*Metric{{"ns/op", 0}}
var metricPool = sync.Pool{
	New: func() any {
		return make([]dist, len(allMetrics))
	},
}

func NewBenchmark(name string) *Benchmark {
	b := &Benchmark{name: name}
	allBenchmarks = append(allBenchmarks, b)
	return b
}

func NewMetric(name string) *Metric {
	for _, m := range allMetrics {
		if m.name == name {
			return m
		}
	}

	m := &Metric{name, len(allMetrics)}
	allMetrics = append(allMetrics, m)
	return m
}

func (b *Benchmark) Start() Run {
	return Run{b, 0, time.Now(), metricPool.Get().([]dist)}
}

func (r *Run) StopTimer() {
	if r.start.IsZero() {
		return
	}
	r.accum += time.Since(r.start)
	r.start = time.Time{}
}

func (r *Run) StartTimer() {
	if !r.start.IsZero() {
		return
	}
	r.start = time.Now()
}

func (r Run) Elapsed() time.Duration {
	e := r.accum
	if !r.start.IsZero() {
		e += time.Since(r.start)
	}
	return e
}

func (r *Run) SetMetric(value float64, metric *Metric) {
	r.metrics[metric.id] = dist{count: 1, total: value}
}

func (r *Run) Done() {
	r.StopTimer()

	if r.b == nil {
		panic("Done already called")
	}

	r.metrics[0] = dist{count: 1, total: float64(r.accum)}

	r.b.lock.Lock()
	defer r.b.lock.Unlock()

	if r.b.metrics == nil {
		// Donate our metrics slice to the benchmark
		r.b.metrics = r.metrics
	} else {
		// Merge our metrics into the benchmark
		for i, m := range r.metrics {
			r.b.metrics[i].count += m.count
			r.b.metrics[i].total += m.total
		}
		metricPool.Put(r.metrics)
		r.metrics = nil
	}

	r.b = nil
}

func Report() {
	var buf bytes.Buffer
	for _, b := range allBenchmarks {
		if b.metrics == nil {
			// No runs
			continue
		}
		buf.Reset()

		fmt.Fprintf(&buf, "Benchmark%s\t%d", b.name, b.metrics[0].count)

		for i, d := range b.metrics {
			if d.count == 0 {
				continue
			}
			if d.count != b.metrics[0].count {
				fmt.Printf("# Warning: %q has samples from %d runs of %d\n", d.count, b.metrics[0].count)
			}
			fmt.Fprintf(&buf, "\t%v %s", d.total/float64(d.count), allMetrics[i].name)
		}

		fmt.Printf("%s\n", buf.Bytes())
	}
}
