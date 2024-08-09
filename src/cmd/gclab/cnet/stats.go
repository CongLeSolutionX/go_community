// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"bytes"
	"cmd/gclab/invivo"
	"cmd/gclab/stats"
	"fmt"
	"reflect"
	"time"
)

const (
	scanStats      = false
	dartboardStats = false
)

type Stats struct {
	RegionBitDensity          stats.Dist[float64] `region bit density`
	RegionObjectDensity       stats.Dist[float64] `region object density`
	RegionObjectMarkedDensity stats.Dist[float64] `fraction of newly marked objected per region scan`
	DartboardDupBits          stats.Dist[float64] `fraction of dartboard region already set per dartboard flush`
	DartboardNewBits          stats.Dist[float64] `fraction of dartboard region newly set per dartboard flush`
	DartboardAddrs            stats.Dist[int]     `count of addresses per flush to dartboard`

	SpanQueuedWordDensity   stats.Dist[float64] `density of queued words per span scan`
	SpanQueuedObjectDensity stats.Dist[float64] `density of queued objects per span scan`
	SpanGreyObjectDensity   stats.Dist[float64] `density of grey objects per span scan`

	RegionScanCount stats.Dist[int] `number of times each mark region is scanned`

	LAddr32s stats.Dist[int] `LAddr32 count per buffer->buffer flush`
	LAddr64s stats.Dist[int] `LAddr64 count per buffer->buffer flush`
}

var gStats Stats

var benchmarkDrain = invivo.NewBenchmark("Drain")
var metricMarkedMBPerSec = invivo.NewMetricRate("marked-MB/sec")
var metricSpanMBPerSec = invivo.NewMetricRate("span-MB/sec")

var benchmarkScanRegion = invivo.NewBenchmark("ScanRegion") // Excludes flushing
var benchmarkScanBuf = invivo.NewBenchmark("ScanBuf")       // Excludes flushing

var benchmarkFlush = invivo.NewBenchmark("Flush") // Only at root, includes all sub-flushing
var metricAddrsPerSec = invivo.NewMetricRate("addrs/sec")
var metricAddrsPerOp = invivo.NewMetricPerOp("addrs")

var benchmarkFlushLayer = mkBenchmarkFlushLayer() // Excludes sub-flushes

func mkBenchmarkFlushLayer() []*invivo.Benchmark {
	res := make([]*invivo.Benchmark, 16)
	for i := range res {
		res[i] = invivo.NewBenchmark(fmt.Sprintf("FlushLayer/layer=%d", i))
	}
	return res
}

var scanCount int

// var traceScanOnce bool
const traceScanOnce = traceScan

type scanRegionStats struct {
	spans        int
	fullSpans    int
	partialSpans int
	largeSpans   int

	objects int

	objectsScanned int
	wordsScanned   int

	pagesScanned int
	pagesSkipped int

	flushes int

	Ns time.Duration
}

type statsPerP struct {
	scanStats       scanRegionStats
	benchScanRegion invivo.Run
	benchScanBuf    invivo.Run

	benchStack []invivo.Run
}

// pushBenchmark pauses the current benchmark run, starts a new run for the given benchmark, and pushes that run onto the stack.
func (p *statsPerP) pushBenchmark(b *invivo.Benchmark) invivo.Run {
	if len(p.benchStack) > 0 {
		p.benchStack[len(p.benchStack)-1].StopTimer()
	}
	r := b.Start()
	p.benchStack = append(p.benchStack, r)
	return r
}

// popBenchmark marks the current benchmark run done, pops it off the stack, and
// unpauses the new current benchmark.
func (p *statsPerP) popBenchmark() {
	p.benchStack[len(p.benchStack)-1].Done()
	p.benchStack = p.benchStack[:len(p.benchStack)-1]
	if len(p.benchStack) > 0 {
		p.benchStack[len(p.benchStack)-1].StartTimer()
	}
}

func (p *statsPerP) startScanRegion() func() {
	p.scanStats = scanRegionStats{}

	p.benchScanRegion = p.pushBenchmark(benchmarkScanRegion)
	return func() {
		bench := p.benchScanRegion
		bench.StopTimer()
		markedBytes := p.scanStats.wordsScanned * 8
		metricMarkedMBPerSec.Set(bench, float64(markedBytes)/1e6, bench.Elapsed().Seconds())
		p.scanStats.Ns = bench.Elapsed()
		scanCount++

		reportStats("ScanRegionOne", &p.scanStats)

		p.popBenchmark()
	}
}

func (p *statsPerP) startScanBuf() func() {
	p.scanStats = scanRegionStats{}

	p.benchScanBuf = p.pushBenchmark(benchmarkScanBuf)
	return func() {
		bench := p.benchScanBuf
		bench.StopTimer()
		markedBytes := p.scanStats.wordsScanned * 8
		metricMarkedMBPerSec.Set(bench, float64(markedBytes)/1e6, bench.Elapsed().Seconds())
		p.scanStats.Ns = bench.Elapsed()
		scanCount++

		reportStats("ScanBufOne", &p.scanStats)

		p.popBenchmark()
	}
}

func incStat(stat *int, by int) {
	// We update the stats regardless of scanStats because some are used for benchmark reporting.
	*stat += by
}

func reportStats(name string, stats any) {
	if !scanStats {
		return
	}

	invivo.Invalidate()

	var buf bytes.Buffer

	rv := reflect.ValueOf(stats).Elem()
	ns := rv.FieldByName("Ns").Interface().(time.Duration)

	fmt.Fprintf(&buf, "Benchmark%s/ns=%d\t1", name, ns)

	rt := rv.Type()
	for i := range rt.NumField() {
		f := rv.Field(i)
		name := rt.Field(i).Name
		if name == "Ns" {
			continue
		}
		fmt.Fprintf(&buf, "\t%v %s", f, name)
	}

	fmt.Printf("%s\n", buf.Bytes())
}
