// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
)

func TestReadMetrics(t *testing.T) {
	// Tests whether readMetrics produces values aligning
	// with ReadMemStats while the world is stopped.
	var mstats runtime.MemStats
	samples := runtime.PrepareAllMetrics()
	runtime.ReadMetricsSlow(&mstats, samples)

	checkUint64 := func(t *testing.T, m string, got, want uint64) {
		t.Helper()
		if got != want {
			t.Errorf("metric %q: got %d, want %d", m, got, want)
		}
	}

	// Check to make sure the values we read line up with other values we read.
	for i := range samples {
		switch name := samples[i].Name(); name {
		case "/memory/classes/heap/free:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.HeapIdle-mstats.HeapReleased)
		case "/memory/classes/heap/released:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.HeapReleased)
		case "/memory/classes/heap/objects:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.HeapAlloc)
		case "/memory/classes/heap/unused:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.HeapInuse-mstats.HeapAlloc)
		case "/memory/classes/heap/stacks:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.StackInuse)
		case "/memory/classes/metadata/mcache/free:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.MCacheSys-mstats.MCacheInuse)
		case "/memory/classes/metadata/mcache/inuse:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.MCacheInuse)
		case "/memory/classes/metadata/mspan/free:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.MSpanSys-mstats.MSpanInuse)
		case "/memory/classes/metadata/mspan/inuse:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.MSpanInuse)
		case "/memory/classes/metadata/other:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.GCSys)
		case "/memory/classes/os-stacks:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.StackSys-mstats.StackInuse)
		case "/memory/classes/other:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.OtherSys)
		case "/memory/classes/profiling/buckets:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.BuckHashSys)
		case "/memory/classes/total:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.Sys)
		case "/gc/heap/objects:objects":
			checkUint64(t, name, samples[i].Uint64(), mstats.HeapObjects)
		case "/gc/heap/goal:bytes":
			checkUint64(t, name, samples[i].Uint64(), mstats.NextGC)
		case "/gc/cycles/automatic:gc-cycles":
			checkUint64(t, name, samples[i].Uint64(), uint64(mstats.NumGC-mstats.NumForcedGC))
		case "/gc/cycles/forced:gc-cycles":
			checkUint64(t, name, samples[i].Uint64(), uint64(mstats.NumForcedGC))
		case "/gc/cycles/total:gc-cycles":
			checkUint64(t, name, samples[i].Uint64(), uint64(mstats.NumGC))
		}
	}
}
