package telemetry

import (
	"fmt"
	"io"
	"runtime"
	"time"
)

type GCControl int

const (
	NoGC GCControl = iota
	GC
)

type Metrics struct {
	gc    GCControl
	marks []*mark
	mark  *mark
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
//   tele := teleteletry.New(teleteleetry.GC)
//   defer tele.Report(os.Stdout)
//   // etc
//   tele.Mark("foo")
//   foo()
//   tele.Mark("bar")
//   bar()
// }
func New(gc GCControl) *Metrics {
	if gc == GC {
		runtime.GC()
	}
	return &Metrics{
		gc: gc,
	}
}

// Report reports the metrics.
func (m *Metrics) Report(w io.Writer) {
	m.closeMark()
	writeMem := func(name string, gc GCControl, start, end, afterGC uint64) {
		io.WriteString(w, fmt.Sprintf("\t%s:\t%d\n", name, end))
		io.WriteString(w, fmt.Sprintf("\t∆%s:\t%d\n", name, end-start))
		if gc == GC {
			io.WriteString(w, fmt.Sprintf("\t%s(afterGC):\t%d\n", name, afterGC))
			io.WriteString(w, fmt.Sprintf("\t∆%s(afterGC):\t%d\n", name, end-afterGC))
		}
	}
	var totTime time.Duration
	for _, curMark := range m.marks {
		totTime += curMark.endT.Sub(curMark.startT)
	}
	for _, curMark := range m.marks {
		dur := curMark.endT.Sub(curMark.startT)
		io.WriteString(w, fmt.Sprintf("%v:\n", curMark.name))
		io.WriteString(w, fmt.Sprintf("\ttime:\t%s\t%.2f%%\n", dur.String(), float64(dur*100)/float64(totTime)))
		writeMem("heap alloc", m.gc, curMark.startM.HeapAlloc, curMark.endM.HeapAlloc, curMark.gcM.HeapAlloc)
		writeMem("heap in use", m.gc, curMark.startM.HeapInuse, curMark.endM.HeapInuse, curMark.gcM.HeapInuse)
		writeMem("heap objects", m.gc, curMark.startM.HeapObjects, curMark.endM.HeapObjects, curMark.gcM.HeapObjects)
	}
}

// Mark starts a new mark in the telemetry table.
func (m *Metrics) Mark(name string) {
	m.closeMark()
	m.mark = &mark{name: name}
	// Unlikely we need to a GC here, as one was likely just done in closeMark.
	runtime.ReadMemStats(&m.mark.startM)
	m.mark.startT = time.Now()
}

func (m *Metrics) closeMark() {
	if m.mark == nil {
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
