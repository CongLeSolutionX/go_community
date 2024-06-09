// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package invivo

// metricAccum is a metric accumulator.
type metricAccum interface {
	// new returns a new child metricAccum that will commit its results to this
	// metricAccum.
	new() metricAccum
	// commit accumulates the value of this metricAccum into its parent and
	// resets this accumulator.
	commit()
	// count returns the total number of samples accumulated into this metric.
	count() int
	// report returns the accumulated value of this metric.
	report() float64
	// reset resets this accumulator to its zero state.
	reset()
}

// metric is a registered metric.
type metric struct {
	name string
	new  func() metricAccum // return a new root accumulator
}

var (
	metricNSPerOp = &MetricPerOp{0}
	allMetrics    = []*metric{{"ns/op", func() metricAccum { return &metricPerOpAccum{id: 0} }}}
)

func registerMetric(name string, new func() metricAccum) int {
	id := len(allMetrics)
	allMetrics = append(allMetrics, &metric{name, new})
	return id
}

// MetricPerOp is a metric that reports the per-Run average.
type MetricPerOp struct {
	id int
}

func NewMetricPerOp(name string) *MetricPerOp {
	m := new(MetricPerOp)
	m.id = registerMetric(name+"/op", func() metricAccum {
		return &metricPerOpAccum{id: m.id}
	})
	return m
}

func (m *MetricPerOp) Set(r Run, val float64) {
	accum := r.internal().metrics[m.id].(*metricPerOpAccum)
	accum.n = 1
	accum.total = val
}

type metricPerOpAccum struct {
	id     int
	parent *metricPerOpAccum

	n     int
	total float64
}

func (m *metricPerOpAccum) new() metricAccum {
	return &metricPerOpAccum{id: m.id, parent: m}
}

func (m *metricPerOpAccum) commit() {
	p := m.parent
	p.n += m.n
	p.total += m.total
	m.reset()
}

func (m *metricPerOpAccum) reset() {
	m.n, m.total = 0, 0
}

func (m *metricPerOpAccum) count() int {
	return m.n
}

func (m *metricPerOpAccum) report() float64 {
	return m.total / float64(m.n)
}

// MetricSum is a metric that reports a total sum. This is rarely what you want
// except in single iteration benchmarks.
type MetricSum struct {
	id int
}

func NewMetricSum(name string) *MetricSum {
	m := new(MetricSum)
	m.id = registerMetric(name, func() metricAccum {
		return &MetricSumAccum{id: m.id}
	})
	return m
}

func (m *MetricSum) Add(r Run, val float64) {
	accum := r.internal().metrics[m.id].(*MetricSumAccum)
	accum.total += val
	accum.used = true
}

type MetricSumAccum struct {
	id     int
	parent *MetricSumAccum

	total float64
	used  bool
}

func (m *MetricSumAccum) new() metricAccum {
	return &MetricSumAccum{id: m.id, parent: m}
}

func (m *MetricSumAccum) commit() {
	p := m.parent
	p.total += m.total
	m.reset()
}

func (m *MetricSumAccum) reset() {
	m.total = 0
	m.used = false
}

func (m *MetricSumAccum) count() int {
	if m.used {
		return 1
	}
	return 0
}

func (m *MetricSumAccum) report() float64 {
	return m.total
}

// MetricRate is a metric that accumulates a total rate.
type MetricRate struct {
	id int
}

func NewMetricRate(name string) *MetricRate {
	m := new(MetricRate)
	m.id = registerMetric(name, func() metricAccum {
		return &metricRateAccum{id: m.id}
	})
	return m
}

func (m *MetricRate) Set(r Run, numer, denom float64) {
	if denom == 0 {
		if numer == 0 {
			return
		}
		panic("divide by zero")
	}

	accum := r.internal().metrics[m.id].(*metricRateAccum)
	accum.n = 1
	accum.numer = numer
	accum.denom = denom
}

type metricRateAccum struct {
	id     int
	parent *metricRateAccum

	n     int
	numer float64
	denom float64
}

func (m *metricRateAccum) new() metricAccum {
	return &metricRateAccum{id: m.id, parent: m}
}

func (m *metricRateAccum) commit() {
	p := m.parent
	p.n += m.n
	p.numer += m.numer
	p.denom += m.denom
	m.reset()
}

func (m *metricRateAccum) reset() {
	m.n, m.numer, m.denom = 0, 0, 0
}

func (m *metricRateAccum) count() int {
	return m.n
}

func (m *metricRateAccum) report() float64 {
	if m.denom == 0 {
		return 0
	}
	return m.numer / m.denom
}
