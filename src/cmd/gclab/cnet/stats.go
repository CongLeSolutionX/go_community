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

var benchmarkScanRegion = invivo.NewBenchmark("ScanRegion")
var metricMBPerSec = invivo.NewMetricRate("MB/sec")

var benchmarkDrain = invivo.NewBenchmark("Drain")
var metricMarkedMBPerSec = invivo.NewMetricRate("marked-MB/sec")
var metricSpanMBPerSec = invivo.NewMetricRate("span-MB/sec")

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

var scanRegionStatsOne scanRegionStats
var scanRegionBench invivo.Run

func statsScanRegion() func() {
	scanRegionStatsOne = scanRegionStats{}

	scanRegionBench = benchmarkScanRegion.Start()
	return func() {
		bench := &scanRegionBench
		bench.StopTimer()
		// TODO: This is bytes of address space, while we usually look at marked
		// bytes. We don't have any way to compute that per scan.
		bytes := bytesPerRegion
		metricMBPerSec.Set(bench, float64(bytes)/1e6, bench.Elapsed().Seconds())
		scanRegionStatsOne.Ns = bench.Elapsed()
		scanCount++
		bench.Done()

		reportStats("ScanRegionOne", &scanRegionStatsOne)
	}
}

func incStat(stat *int, by int) {
	if scanStats {
		*stat += by
	}
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
