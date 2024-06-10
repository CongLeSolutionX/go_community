// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gclab uses the GC scan experiment trace data to simulate different scanning
// algorithms.

package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"internal/trace"
	"internal/trace/testdata/cmd/gclab/heap"
	"internal/trace/testdata/cmd/gclab/shortest"
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
	Heap  = heap.Heap
	VAddr = heap.VAddr
	Bytes = heap.Bytes
)

type Eventer interface {
	GCStart(gomaxprocs int)
	GCEnd()

	NewSpan(base VAddr, sc *heap.SizeClass, nPages int, noScan bool)

	Scan(base VAddr, typ scanType, offs []uint64, ptrs []VAddr, found []bool)

	ScanWB(ev trace.Event, value VAddr, found bool)

	AllocBlack(ev trace.Event, base VAddr)
}

type Logger struct {
	l   *log.Logger
	gen int
}

func (l *Logger) GCStart(gomaxprocs int) {
	l.l = log.New(os.Stdout, fmt.Sprintf("[GC %d] ", l.gen), 0)
	l.l.Printf("start GOMAXPROCS=%d", gomaxprocs)
}

func (l *Logger) GCEnd() {
	l.l.Printf("end")
	l.gen++
}

func (l *Logger) NewSpan(base VAddr, sc *heap.SizeClass, nPages int, noScan bool) {
	var objBytes heap.Bytes
	if sc != nil {
		objBytes = sc.ObjectBytes
	}
	l.l.Printf("new span [%s,%s) %s objects", base, base.Plus(heap.PageBytes.Mul(nPages)), objBytes)
}

func (l *Logger) Scan(base VAddr, typ scanType, offs []uint64, ptrs []VAddr, founds []bool) {
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

	l.l.Printf("scan %s %s", base, typStr)
	for i, off := range offs {
		l.l.Printf("  +%#08x => %s %v", off, ptrs[i], founds[i])
	}
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
	h.AtEnd = append(h.AtEnd, shortest.Scanner)

	expBatchIDs := make(map[*trace.ExperimentalData]bool)
	var expBatches []trace.ExperimentalBatch
	gomaxprocs := -1

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
		case trace.EventMetric:
			em := ev.Metric()
			if em.Name == "/sched/gomaxprocs:threads" {
				gomaxprocs = int(em.Value.Uint64())
			}

		case trace.EventRangeBegin:
			// Find the start of the GC.
			if ev.Range().Name == "GC concurrent mark phase" {
				if gomaxprocs == -1 {
					panic("unknown GOMAXPROCS")
				}
				for _, s := range eventers {
					s.GCStart(gomaxprocs)
				}
			}

		case trace.EventExperimental:
			ex := ev.Experimental()

			if ex.Data != nil && !expBatchIDs[ex.Data] {
				expBatchIDs[ex.Data] = true
				expBatches = append(expBatches, ex.Data.Batches...)
			}

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

			case "GCScanGCDone":
				numgc := int(ex.Args[0])

				// Process queued experimental batches.
				processExpBatches(eventers, expBatches, numgc)

				for _, s := range eventers {
					s.GCEnd()
				}

				clear(expBatchIDs)
				expBatches = nil
			}
		}
	}
}

func processExpBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	var r *bytes.Reader
	varint := func() uint64 {
		val, err := binary.ReadUvarint(r)
		if err != nil {
			panic(err)
		}
		return val
	}
	var offs []uint64
	var ptrs []VAddr
	var founds []bool
	for _, b := range bs {
		r = bytes.NewReader(b.Data)
		gen := int(varint())
		if gen != numgc {
			continue
		}

		//log.Printf("batch %d bytes", len(b.Data))
		for r.Len() > 0 {
			baseTyp := varint()
			base := VAddr(baseTyp &^ 0b11)
			typ := scanType(baseTyp & 0b11)
			n := varint()
			prevOff := uint64(0)
			offs, ptrs, founds = offs[:0], ptrs[:0], founds[:0]
			for range n {
				off := prevOff + uint64(varint())
				ptr := VAddr(varint())
				found := ptr&0b1 != 0
				ptr &^= 0b11
				offs = append(offs, off)
				ptrs = append(ptrs, ptr)
				founds = append(founds, found)
				prevOff = off
				//log.Printf("  %d %x %s", i, off, ptr)
			}

			for _, s := range eventers {
				s.Scan(base, typ, offs, ptrs, founds)
			}
		}
	}
}
