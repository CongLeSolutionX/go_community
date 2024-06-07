// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gclab uses the GC scan experiment trace data to simulate different scanning
// algorithms.

package main

import (
	"bytes"
	"cmd/gclab/heap"
	"encoding/binary"
	"flag"
	"fmt"
	"internal/trace"
	"io"
	"iter"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var (
	logRawEvents = flag.Bool("log-raw-events", false, "whether to log raw trace events")
	logEvents    = flag.Bool("log-events", false, "whether to log high level events")
	cpuProfile   = flag.String("cpuprofile", "", "write CPU profile to `pprof`")
	memProfile   = flag.String("memprofile", "", "write memory profile to `file`")
)

type batchHeader uint8

const (
	gcScanBatchSizes batchHeader = iota
	gcScanBatchSpans
	gcScanBatchScan
	gcScanBatchAllocs
	gcScanBatchTypes
)

type heapBitsType uint8

const (
	heapBitsNone   heapBitsType = iota // No scan
	heapBitsPacked                     // Packed at end of span
	heapBitsHeader                     // Pointed to by each object header
	heapBitsOOB                        // Pointed to by span struct (one object per span)
)

type scanType uint8

const (
	scanTypeNone scanType = iota
	scanTypeRoot
	scanTypeObject
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
	SpanInfo(base VAddr, allocBits []uint64, heapBitsType heapBitsType, heapBits []uint64)

	NewType(id uint64, size Bytes, ptrWords heap.Words, ptrData []uint64)

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

func (l *Logger) SpanInfo(base VAddr, allocBits []uint64, heapBitsType heapBitsType, heapBits []uint64) {
	var s string
	switch heapBitsType {
	case heapBitsNone:
		s = "none"
	case heapBitsPacked:
		s = fmt.Sprintf("packed %016x", heapBits)
	case heapBitsHeader:
		s = fmt.Sprintf("header %d", heapBits)
	case heapBitsOOB:
		s = fmt.Sprintf("OOB %d", heapBits)
	}
	l.l.Printf("span info %s alloc bits %x heap bits %s", base, allocBits, s)
}

func (l *Logger) NewType(id uint64, size Bytes, ptrWords heap.Words, ptrData []uint64) {
	l.l.Printf("new type %d size %s ptr words %v ptrMask %016x", id, size, ptrWords, ptrData)
}

func (l *Logger) Scan(base VAddr, typ scanType, offs []uint64, ptrs []VAddr, founds []bool) {
	typStr := "UNKNOWN"
	switch typ {
	case scanTypeNone:
		typStr = "none"
	case scanTypeRoot:
		typStr = "root"
	case scanTypeObject:
		typStr = "object"
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

	expBatchIDs := make(map[*trace.ExperimentalData]bool)
	var expBatches []trace.ExperimentalBatch
	gomaxprocs := -1

	// synced tracks whether we're in a clean state for GC events. The trace
	// can start in the middle of a GC, in which case we want to ignore
	// events. But we want to catch bad events between GCs, so this is set
	// to true if we see *either* a GC start or GC done event.
	synced := false

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
			if err.Error() == "no frequency event found" {
				// TODO: This just means a truncated trace, but it's unfortunate
				// we have to check the error string.
				break
			}
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

		case trace.EventExperimental:
			ex := ev.Experimental()

			if ex.Data != nil && !expBatchIDs[ex.Data] {
				expBatchIDs[ex.Data] = true
				expBatches = append(expBatches, ex.Data.Batches...)
			}

			switch ex.Name {
			case "GCScanGCStart":
				synced = true

				numgc := int(ex.Args[0])
				gmp := int(ex.Args[1])
				if gmp != gomaxprocs {
					log.Fatalf("GOMAXPROCS=%d, but GCScanGCStart has GOMAXPROCS %d", gomaxprocs, gmp)
				}

				processSizeBatches(eventers, expBatches, numgc)

				for _, s := range eventers {
					// TODO: Pass in size classes?
					s.GCStart(gomaxprocs)
				}

				processSpanBatches(eventers, expBatches, numgc)

			case "GCScanSpan":
				if !synced {
					break
				}
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
				if !synced {
					break
				}
				value := VAddr(ex.Args[0] &^ 0b11)
				found := (ex.Args[0] & 0b1) != 0

				for _, s := range eventers {
					s.ScanWB(ev, value, found)
				}

			case "GCScanAllocBlack":
				if !synced {
					break
				}
				base := VAddr(ex.Args[0])
				for _, s := range eventers {
					s.AllocBlack(ev, base)
				}

			case "GCScanGCDone":
				if !synced {
					synced = true
					break
				}
				numgc := int(ex.Args[0])

				processTypeBatches(eventers, expBatches, numgc)
				processAllocBatches(eventers, expBatches, numgc)

				// Process queued scan batches.
				processScanBatches(eventers, expBatches, numgc)

				for _, s := range eventers {
					s.GCEnd()
				}

				clear(expBatchIDs)
				expBatches = nil
			}
		}
	}

	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}

type myReader struct {
	*bytes.Reader
}

func (r *myReader) byte() byte {
	val, err := r.Reader.ReadByte()
	if err != nil {
		panic(err)
	}
	return val
}

func (r *myReader) bytes(n int) []byte {
	out := make([]byte, n)
	got, _ := r.Reader.Read(out)
	if got < n {
		panic("short read")
	}
	return out
}

func (r *myReader) varint() uint64 {
	val, err := binary.ReadUvarint(r.Reader)
	if err != nil {
		panic(err)
	}
	return val
}

func (r *myReader) uint64() uint64 {
	var val uint64
	for i := range 8 {
		val |= uint64(r.byte()) << (i * 8)
	}
	return val
}

type gcBatch struct {
	b trace.ExperimentalBatch

	typ batchHeader
}

func iterBatches(bs []trace.ExperimentalBatch, numgc int) iter.Seq2[gcBatch, myReader] {
	return func(yield func(gcBatch, myReader) bool) {
		for _, b := range bs {
			r := myReader{bytes.NewReader(b.Data)}

			typ := batchHeader(r.byte())

			gen := int(r.varint())
			if gen != numgc {
				continue
			}

			//log.Printf("batch type %d, %d bytes", typ, len(b.Data))

			if !yield(gcBatch{b, typ}, r) {
				break
			}
		}
	}
}

func processSizeBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	for b, r := range iterBatches(bs, numgc) {
		switch b.typ {
		case gcScanBatchSizes:
			sizeClasses = nil
			for r.Len() > 0 {
				id := len(sizeClasses)
				size := Bytes(r.varint())
				pages := int(r.varint())
				heapBits := heapBitsType(r.byte())

				sc := heap.SizeClass{ID: id, ObjectBytes: size, SpanPages: pages, HeapBitsType: heap.HeapBitsType(heapBits)}
				if len(sizeClasses) > 0 && sizeClasses[len(sizeClasses)-1].ObjectBytes >= sc.ObjectBytes {
					panic("size classes out of order or duplicate size class event")
				}
				sizeClasses = append(sizeClasses, sc)
			}
		}
	}
}

func processSpanBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	for b, r := range iterBatches(bs, numgc) {
		switch b.typ {
		case gcScanBatchSpans:
			for r.Len() > 0 {
				// XXX Dedup this with new spans.
				base := heap.VAddr(r.varint())
				f2 := r.varint()
				spanClass := f2 & 0xff
				noScan := (spanClass & 1) != 0
				sizeClass := spanClass >> 1
				if sizeClass == 0 {
					nPages := f2 >> 8
					for _, s := range eventers {
						s.NewSpan(base, nil, int(nPages), noScan)
					}
				} else {
					sc := &sizeClasses[sizeClass]
					for _, s := range eventers {
						s.NewSpan(base, sc, sc.SpanPages, noScan)
					}
				}
			}
		}
	}
}

func processTypeBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	var partial struct {
		id       uint64
		size     Bytes
		ptrWords heap.Words
		ptrData  []uint64
	}

	for b, r := range iterBatches(bs, numgc) {
		switch b.typ {
		case gcScanBatchTypes:
			for r.Len() > 0 {
				id := r.varint()
				size := Bytes(r.varint())
				if size == 0 {
					// Continuation
					if partial.id == 0 {
						log.Fatalf("continuation type %d following complete type", id)
					} else if id != partial.id {
						log.Fatalf("partial type %d followed by continuation type %d", partial.id, id)
					}
					size = partial.size
					ptrMaskOffset := r.varint()
					if uint64(len(partial.ptrData)) != ptrMaskOffset {
						log.Fatalf("have %d bytes of ptrMask, but continuation starts at %d byte offset", len(partial.ptrData), ptrMaskOffset)
					}
				} else {
					// Start
					if partial.id != 0 {
						log.Fatalf("partial type %d followed by type start %d", partial.id, id)
					}
					partial.id = id
					partial.size = size
					partial.ptrWords = heap.Words(r.varint())
				}

				ptrsLenAndPartial := r.varint()
				ptrsLen := ptrsLenAndPartial >> 1
				isPartial := (ptrsLenAndPartial & 1) != 0
				for range ptrsLen {
					partial.ptrData = append(partial.ptrData, r.uint64())
				}

				if !isPartial {
					for _, s := range eventers {
						s.NewType(id, size, partial.ptrWords, partial.ptrData)
					}
					partial.id = 0
					partial.size = size
					partial.ptrData = nil
				}
			}
		}
	}
}

func processAllocBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	for b, r := range iterBatches(bs, numgc) {
		switch b.typ {
		case gcScanBatchAllocs:
			for r.Len() > 0 {
				spanAddr := VAddr(r.varint())

				nElems := r.varint()
				allocBitsLen := (nElems + 63) / 64
				allocBits := make([]uint64, allocBitsLen)
				for i := range allocBits {
					allocBits[i] = r.uint64()
				}

				heapBitsType := heapBitsNone
				var heapBits []uint64

				nHeapBits := r.byte()
				if nHeapBits == 0 {
					// No scan
				} else if nHeapBits < 0xfe {
					// Packed heap bits
					heapBitsType = heapBitsPacked
					heapBits = make([]uint64, nHeapBits)
					for i := range heapBits {
						heapBits[i] = r.uint64()
					}
				} else {
					// Per-allocation type bits
					if nHeapBits == 0xfe {
						heapBitsType = heapBitsOOB
					} else {
						heapBitsType = heapBitsHeader
					}
					heapBits = make([]uint64, nElems)
					for i, aBits := range allocBits {
						for j := range 64 {
							if aBits&(1<<j) != 0 {
								heapBits[i*64+j] = r.varint()
							}
						}
					}
					if len(heapBits) == 1 && heapBits[0] == 0 {
						// Delayed zero large object. Treat like noscan.
						heapBitsType = heapBitsNone
						heapBits = nil
					}
				}

				for _, s := range eventers {
					s.SpanInfo(spanAddr, allocBits, heapBitsType, heapBits)
				}
			}
		}
	}
}

func processScanBatches(eventers []Eventer, bs []trace.ExperimentalBatch, numgc int) {
	var offs []uint64
	var ptrs []VAddr
	var founds []bool
	for b, r := range iterBatches(bs, numgc) {
		switch b.typ {
		case gcScanBatchScan:
			for r.Len() > 0 {
				baseTyp := r.varint()
				base := VAddr(baseTyp &^ 0b11)
				typ := scanType(baseTyp & 0b11)
				n := r.varint()
				prevOff := uint64(0)
				offs, ptrs, founds = offs[:0], ptrs[:0], founds[:0]
				for range n {
					off := prevOff + r.varint()
					ptr := VAddr(r.varint())
					found := ptr&0b1 != 0
					ptr &^= 0b11
					offs = append(offs, off)
					ptrs = append(ptrs, ptr)
					founds = append(founds, found)
					prevOff = off
				}

				for _, s := range eventers {
					s.Scan(base, typ, offs, ptrs, founds)
				}
			}
		}
	}
}
