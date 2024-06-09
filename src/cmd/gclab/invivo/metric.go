// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package invivo

// metricAccum is a metric accumulator.
type metricAccum interface {
	// new returns a new child metricAccum that will commit its results to this
	// metricAccum.
	new() metricAccum
	// commit accumulates the value of this metricAccum into its parent.
	commit()
	// count returns the total number of samples accumulated into this metric.
	count() int
	// report returns the accumulated value of this metric.
	report() float64
}

// metric is a registered metric.
type metric struct {
	name string
	new  func() metricAccum // return a new root accumulator
}

var (
	metricNSPerOp = &MetricAvg{0}
	allMetrics    = []*metric{{"ns/op", func() metricAccum { return &metricAvgAccum{id: 0} }}}
)

func registerMetric(name string, new func() metricAccum) int {
	id := len(allMetrics)
	allMetrics = append(allMetrics, &metric{name, new})
	return id
}

// MetricAvg is a metric that takes the average of its samples.
type MetricAvg struct {
	id int
}

func NewMetricAvg(name string) *MetricAvg {
	m := new(MetricAvg)
	m.id = registerMetric(name, func() metricAccum {
		return &metricAvgAccum{id: m.id}
	})
	return m
}

func (m *MetricAvg) Set(r *Run, val float64) {
	accum := r.metrics[m.id].(*metricAvgAccum)
	accum.n = 1
	accum.total = val
}

type metricAvgAccum struct {
	id     int
	parent *metricAvgAccum

	n     int
	total float64
}

func (m *metricAvgAccum) new() metricAccum {
	return &metricAvgAccum{id: m.id, parent: m}
}

func (m *metricAvgAccum) commit() {
	p := m.parent
	p.n += m.n
	p.total += m.total
	m.n, m.total = 0, 0
}

func (m *metricAvgAccum) count() int {
	return m.n
}

func (m *metricAvgAccum) report() float64 {
	return m.total / float64(m.n)
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

func (m *MetricRate) Set(r *Run, numer, denom float64) {
	if denom == 0 {
		if numer == 0 {
			return
		}
		panic("divide by zero")
	}

	accum := r.metrics[m.id].(*metricRateAccum)
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
