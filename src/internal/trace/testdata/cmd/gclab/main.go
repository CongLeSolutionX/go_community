// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gclab uses the GC scan experiment trace data to simulate different scanning
// algorithms.

package main

import (
	"flag"
	"fmt"
	"internal/trace"
	"internal/trace/testdata/cmd/gclab/heap"
	"io"
	"log"
	"os"
	"runtime/pprof"
)

var (
	logRawEvents = flag.Bool("log-raw-events", false, "whether to log raw trace events")
	logEvents    = flag.Bool("log-events", false, "whether to log high level events")
	cpuProfile   = flag.String("cpuprofile", "", "write CPU profile to `pprof`")
)

type scanType uint8

const (
	scanTypeBlock scanType = iota
	scanTypeObject
	scanTypeConservative
	scanTypeTinyAllocs
)

var sizeClasses []heap.SizeClass

type (
	Heap   = heap.Heap
	Object = heap.Object
	VAddr  = heap.VAddr
	Bytes  = heap.Bytes
)

type Eventer interface {
	GCStart()
	GCEnd()

	NewSpan(base VAddr, sc *heap.SizeClass, nPages int, noScan bool)

	Scan(ev trace.Event, base VAddr, typ scanType)
	ScanEnd(ev trace.Event)

	ScanPointer(ev trace.Event, value VAddr, found bool, offset Bytes)
	ScanWB(ev trace.Event, value VAddr, found bool)

	AllocBlack(ev trace.Event, base VAddr)
}

type Logger struct {
	l   *log.Logger
	gen int
}

func (l *Logger) GCStart() {
	l.l = log.New(os.Stdout, fmt.Sprintf("[GC %d] ", l.gen), 0)
	l.l.Printf("start")
}

func (l *Logger) GCEnd() {
	l.l.Printf("end")
	l.gen++
}

func (l *Logger) NewSpan(base VAddr, sc *heap.SizeClass, nPages int, noScan bool) {}

func (l *Logger) Scan(ev trace.Event, base VAddr, typ scanType) {
	typStr := "UNKNOWN"
	switch typ {
	case scanTypeBlock:
		typStr = "block"
	case scanTypeObject:
		typStr = "object"
	case scanTypeConservative:
		typStr = "conservative"
	case scanTypeTinyAllocs:
		typStr = "tiny allocs"
	}

	l.l.Printf("P %2d scan %s %s", ev.Proc(), base, typStr)
}

func (l *Logger) ScanEnd(ev trace.Event) {
	l.l.Printf("P %2d end scan", ev.Proc())
}

func (l *Logger) ScanPointer(ev trace.Event, value VAddr, found bool, offset Bytes) {
	l.l.Printf("P %2d   +%#08x => %s %v", ev.Proc(), offset, value, found)
}

func (l *Logger) ScanWB(ev trace.Event, value VAddr, found bool) {
	l.l.Printf("P %2d WB %s %v", ev.Proc(), value, found)
}

func (l *Logger) AllocBlack(ev trace.Event, value VAddr) {
	l.l.Printf("P %2d alloc %s", ev.Proc(), value)
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	traceFile, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	var eventers []Eventer
	if *logEvents {
		eventers = append(eventers, new(Logger))
	}
	h := new(Heaper)
	eventers = append(eventers, h)

	r, err := trace.NewReader(traceFile)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ev, err := r.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if *logRawEvents {
			log.Println(ev.String())
		}

		switch ev.Kind() {
		case trace.EventRangeBegin, trace.EventRangeEnd:
			// Find the start of the GC.
			if ev.Range().Name == "GC concurrent mark phase" {
				if ev.Kind() == trace.EventRangeBegin {
					for _, s := range eventers {
						s.GCStart()
					}
				} else {
					for _, s := range eventers {
						s.GCEnd()
					}
				}
			}

		case trace.EventExperimental:
			ex := ev.Experimental()
			switch ex.Name {
			case "GCScanClass":
				sc := heap.SizeClass{ObjectBytes: Bytes(ex.Args[0]), SpanPages: int(ex.Args[1])}
				if len(sizeClasses) > 0 && sizeClasses[len(sizeClasses)-1].ObjectBytes >= sc.ObjectBytes {
					panic("size classes out of order or duplicate size class event")
				}
				sizeClasses = append(sizeClasses, sc)

			case "GCScanSpan":
				base := VAddr(ex.Args[0])
				spanClass := ex.Args[1] & 0xff
				noScan := (spanClass & 1) != 0
				sizeClass := spanClass >> 1
				if sizeClass == 0 {
					nPages := ex.Args[1] >> 8
					for _, s := range eventers {
						s.NewSpan(base, nil, int(nPages), noScan)
					}
				} else {
					sc := &sizeClasses[sizeClass]
					for _, s := range eventers {
						s.NewSpan(base, sc, sc.SpanPages, noScan)
					}
				}

			case "GCScan":
				base := VAddr(ex.Args[0] &^ 0b11)
				typ := scanType(ex.Args[0] & 0b11)
				n := ex.Args[1]
				_ = n // TODO: Drop from event?

				for _, s := range eventers {
					s.Scan(ev, base, typ)
				}

			case "GCScanEnd":
				for _, s := range eventers {
					s.ScanEnd(ev)
				}

			case "GCScanPointer":
				value := VAddr(ex.Args[0] &^ 0b11)
				found := (ex.Args[0] & 0b1) != 0
				offset := Bytes(ex.Args[1])

				for _, s := range eventers {
					s.ScanPointer(ev, value, found, offset)
				}

			case "GCScanWB":
				value := VAddr(ex.Args[0] &^ 0b11)
				found := (ex.Args[0] & 0b1) != 0

				for _, s := range eventers {
					s.ScanWB(ev, value, found)
				}

			case "GCScanAllocBlack":
				base := VAddr(ex.Args[0])
				for _, s := range eventers {
					s.AllocBlack(ev, base)
				}
			}
		}
	}
}
