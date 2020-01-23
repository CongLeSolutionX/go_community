// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package benchmark

import (
	"fmt"
	"io"
	"runtime"
	"time"
	"unicode"
)

type GCControl int

const (
	NoGC GCControl = iota
	GC
)

type Metrics struct {
	enabled bool
	gc      GCControl
	marks   []*mark
	mark    *mark
}

type mark struct {
	name              string
	startM, endM, gcM runtime.MemStats
	startT, endT      time.Time
}

// New creates a new Metrics object.
//
// Typical usage should look like:
//
// func main() {
//   tele := benchmark.New(benchmark.GC)
//   defer tele.Report(os.Stdout)
//   // etc
//   tele.Start("foo")
//   foo()
//   tele.Start("bar")
//   bar()
// }
func New(enabled bool, gc GCControl) *Metrics {
	if enabled && gc == GC {
		runtime.GC()
	}
	return &Metrics{
		enabled: enabled,
		gc:      gc,
	}
}

// Report reports the metrics.
// Closes the currently Start(ed) range, and writes the report to the given io.Writer.
func (m *Metrics) Report(w io.Writer) {
	if !m.enabled {
		return
	}

	m.closeMark()

	gcString := ""
	if m.gc == GC {
		gcString = "_GC"
	}

	var totTime time.Duration
	for _, curMark := range m.marks {
		totTime += curMark.endT.Sub(curMark.startT)
	}
	io.WriteString(w, fmt.Sprintf("%s 1 %d ms/op\n", makeBenchString("total time"+gcString), totTime.Milliseconds()))
	for _, curMark := range m.marks {
		dur := curMark.endT.Sub(curMark.startT)
		io.WriteString(w, fmt.Sprintf("%s 1 %d ms/op", makeBenchString(curMark.name+gcString), dur.Milliseconds()))
		io.WriteString(w, fmt.Sprintf("\t%d allocs/op", curMark.endM.TotalAlloc-curMark.startM.TotalAlloc))
		io.WriteString(w, fmt.Sprintf("\t%d mallocs/op", curMark.endM.Mallocs-curMark.startM.Mallocs))
		if m.gc == GC {
			io.WriteString(w, fmt.Sprintf("\t%d B", curMark.gcM.HeapAlloc))
		} else {
			io.WriteString(w, fmt.Sprintf("\t%d B", curMark.endM.HeapAlloc))
		}
		io.WriteString(w, "\n")
	}
}

// Starts a new mark in the telemetry table.
// Once a metric is started, it continues until either a Report is issued, or another Start is called.
func (m *Metrics) Start(name string) {
	if !m.enabled {
		return
	}
	m.closeMark()
	m.mark = &mark{name: name}
	// Unlikely we need to a GC here, as one was likely just done in closeMark.
	runtime.ReadMemStats(&m.mark.startM)
	m.mark.startT = time.Now()
}

func (m *Metrics) closeMark() {
	if !m.enabled || m.mark == nil {
		return
	}
	m.mark.endT = time.Now()
	runtime.ReadMemStats(&m.mark.endM)
	if m.gc == GC {
		runtime.GC()
		runtime.ReadMemStats(&m.mark.gcM)
	}
	m.marks = append(m.marks, m.mark)
	m.mark = nil
}

// makeBenchString makes a benchmark string consumable by Go's benchmarking tools.
func makeBenchString(name string) string {
	needCap := true
	ret := "Benchmark"
	for _, r := range name {
		if unicode.IsSpace(r) {
			needCap = true
			continue
		}
		if needCap {
			r = unicode.ToUpper(r)
			needCap = false
		}
		ret = ret + string(r)
	}
	return ret
}
