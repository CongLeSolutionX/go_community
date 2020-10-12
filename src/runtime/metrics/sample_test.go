// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics_test

import (
	"runtime"
	"runtime/metrics"
	"strings"
	"testing"
)

func prepareAllMetricsSamples() (map[string]metrics.Description, []metrics.Sample) {
	all := metrics.All()
	samples := make([]metrics.Sample, len(all))
	descs := make(map[string]metrics.Description)
	for i := range all {
		samples[i].Name = all[i].Name
		descs[all[i].Name] = all[i]
	}
	return descs, samples
}

func TestReadMetricsConsistency(t *testing.T) {
	// Tests whether readMetrics produces consistent, sensible values.
	// The values are read concurrently with the runtime doing other
	// things (e.g. allocating) so what we read can't reasonably compared
	// to runtime values.

	// Run a few GC cycles to get some of the stats to be non-zero.
	runtime.GC()
	runtime.GC()
	runtime.GC()

	// Read all the supported metrics through the metrics package.
	descs, samples := prepareAllMetricsSamples()
	metrics.Read(samples)

	// Check to make sure the values we read make sense.
	var totalVirtual struct {
		got, want uint64
	}
	var objects struct {
		alloc, free *metrics.Float64Histogram
		total       uint64
	}
	var gc struct {
		numGC  uint64
		pauses uint64
	}
	for i := range samples {
		kind := samples[i].Value.Kind()
		if want := descs[samples[i].Name].Kind; kind != want {
			t.Errorf("supported metric %q has unexpected kind: got %d, want %d", samples[i].Name, kind, want)
			continue
		}
		if samples[i].Name != "/memory/classes/total:bytes" && strings.HasPrefix(samples[i].Name, "/memory/classes") {
			v := samples[i].Value.Uint64()
			totalVirtual.want += v

			// None of these stats should ever get this big.
			// If they do, there's probably overflow involved,
			// usually due to bad accounting.
			if int64(v) < 0 {
				t.Errorf("%q has high/negative value: %d", samples[i].Name, v)
			}
		}
		switch samples[i].Name {
		case "/memory/classes/total:bytes":
			totalVirtual.got = samples[i].Value.Uint64()
		case "/gc/heap/objects:objects":
			objects.total = samples[i].Value.Uint64()
		case "/gc/heap/allocs-by-size:objects":
			objects.alloc = samples[i].Value.Float64Histogram()
		case "/gc/heap/frees-by-size:objects":
			objects.free = samples[i].Value.Float64Histogram()
		case "/gc/cycles:gc-cycles":
			gc.numGC = samples[i].Value.Uint64()
		case "/gc/pauses:seconds":
			h := samples[i].Value.Float64Histogram()
			gc.pauses = 0
			for i := range h.Counts {
				gc.pauses += h.Counts[i]
			}
		case "/sched/goroutines:goroutines":
			if samples[i].Value.Uint64() < 1 {
				t.Error("number of goroutines is less than one")
			}
		}
	}
	if totalVirtual.got != totalVirtual.want {
		t.Errorf(`"/memory/classes/total:bytes" does not match sum of /memory/classes/**: got %d, want %d`, totalVirtual.got, totalVirtual.want)
	}
	if len(objects.alloc.Buckets) != len(objects.free.Buckets) {
		t.Error("allocs-by-size and frees-by-size buckets don't match in length")
	} else if len(objects.alloc.Counts) != len(objects.free.Counts) {
		t.Error("allocs-by-size and frees-by-size counts don't match in length")
	} else {
		for i := range objects.alloc.Buckets {
			ba := objects.alloc.Buckets[i]
			bf := objects.free.Buckets[i]
			if ba != bf {
				t.Errorf("bucket %d is different for alloc and free hists: %f != %f", i, ba, bf)
			}
		}
		if !t.Failed() {
			got, want := uint64(0), objects.total
			for i := range objects.alloc.Counts {
				if objects.alloc.Counts[i] < objects.free.Counts[i] {
					t.Errorf("found more allocs than frees in object dist bucket %d", i)
					continue
				}
				got += objects.alloc.Counts[i] - objects.free.Counts[i]
			}
			if got != want {
				t.Errorf("object distribution counts don't match count of live objects: got %d, want %d", got, want)
			}
		}
	}
	// The current GC has at least 2 pauses per GC.
	// Check to see if that value makes sense.
	if gc.pauses < gc.numGC*2 {
		t.Errorf("fewer pauses than expected: got %d, want at least %d", gc.pauses, gc.numGC*2)
	}
}
