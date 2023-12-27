// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"debug/dwarf"
	"debug/macho"
	"io"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unsafe"
)

// writeHeapProto writes the current heap profile in protobuf format to w.
func writeHeapProto(w io.Writer, p []runtime.MemProfileRecord, rate int64, defaultSampleType string) error {
	b := newProfileBuilder(w)
	b.pbValueType(tagProfile_PeriodType, "space", "bytes")
	b.pb.int64Opt(tagProfile_Period, rate)
	b.pbValueType(tagProfile_SampleType, "alloc_objects", "count")
	b.pbValueType(tagProfile_SampleType, "alloc_space", "bytes")
	b.pbValueType(tagProfile_SampleType, "inuse_objects", "count")
	b.pbValueType(tagProfile_SampleType, "inuse_space", "bytes")
	if defaultSampleType != "" {
		b.pb.int64Opt(tagProfile_DefaultSampleType, b.stringIndex(defaultSampleType))
	}

	values := []int64{0, 0, 0, 0}
	var locs []uint64
	for _, r := range p {
		hideRuntime := true
		for tries := 0; tries < 2; tries++ {
			stk := r.Stack()
			// For heap profiles, all stack
			// addresses are return PCs, which is
			// what appendLocsForStack expects.
			if hideRuntime {
				for i, addr := range stk {
					if f := runtime.FuncForPC(addr); f != nil && strings.HasPrefix(f.Name(), "runtime.") {
						continue
					}
					// Found non-runtime. Show any runtime uses above it.
					stk = stk[i:]
					break
				}
			}
			locs = b.appendLocsForStack(locs[:0], stk)
			if len(locs) > 0 {
				break
			}
			hideRuntime = false // try again, and show all frames next time.
		}

		values[0], values[1] = scaleHeapSample(r.AllocObjects, r.AllocBytes, rate)
		values[2], values[3] = scaleHeapSample(r.InUseObjects(), r.InUseBytes(), rate)
		if len(r.Roots) > 1 {
			// We aren't tracking how much each GC root is contributing to the
			// amount of in-use memory here so just evenly divide it
			for i, v := range values {
				// TODO: We're dividing up allocs too... but really we only
				// want to do this for in-use
				// So perhaps two samples, one for allocs and one for in-use,
				// and split up in-use only?
				values[i] = int64(float64(v) / float64(len(r.Roots)))
			}
		}
		var blockSize int64
		if r.AllocObjects > 0 {
			blockSize = r.AllocBytes / r.AllocObjects
		}
		roots := r.Roots
		for {
			b.pbSample(values, locs, func() {
				if blockSize != 0 {
					b.pbLabel(tagSample_Label, "bytes", "", blockSize)
				}
				if len(roots) > 0 {
					name := roots[0]
					roots = roots[1:]
					b.pbLabel(tagSample_Label, "gc root", name, 0)
				}
			})
			if len(roots) == 0 {
				break
			}
		}
	}
	b.build()
	return nil
}

// scaleHeapSample adjusts the data from a heap Sample to
// account for its probability of appearing in the collected
// data. heap profiles are a sampling of the memory allocations
// requests in a program. We estimate the unsampled value by dividing
// each collected sample by its probability of appearing in the
// profile. heap profiles rely on a poisson process to determine
// which samples to collect, based on the desired average collection
// rate R. The probability of a sample of size S to appear in that
// profile is 1-exp(-S/R).
func scaleHeapSample(count, size, rate int64) (int64, int64) {
	if count == 0 || size == 0 {
		return 0, 0
	}

	if rate <= 1 {
		// if rate==1 all samples were collected so no adjustment is needed.
		// if rate<1 treat as unknown and skip scaling.
		return count, size
	}

	avgSize := float64(size) / float64(count)
	scale := 1 / (1 - math.Exp(-avgSize/float64(rate)))

	return int64(float64(count) * scale), int64(float64(size) * scale)
}

func init() {
	getNamesForGlobals()
}

var globalMeta = map[uintptr]string{}
var nonzeroMeta = []int{1, 2, 3}

func getSymbolNameForPtr(p string) (string, bool) {
	addr, err := strconv.ParseUint(p, 0, 64)
	if err != nil {
		return "", false
	}
	name, ok := globalMeta[uintptr(addr)]
	return name, ok
}

func getNamesForGlobals() {
	me, err := os.Executable()
	if err != nil {
		panic(err)
	}
	d, err := macho.Open(me)
	if err != nil {
		println("not macho")
		return
		// TODO: elf support
		//d, err = elf.Open(me)
		//if err != nil {
		//	panic(err)
		//}
	}
	_ = d

	type mapping struct {
		lo, hi, offset uint64
	}
	mappings := []mapping{}
	machVMInfo(func(lo, hi, offset uint64, file, buildID string) {
		mappings = append(mappings, mapping{
			lo: lo, hi: hi, offset: offset},
		)
	})

	p := &globalMeta
	q := &nonzeroMeta
	var bssadjustment int64
	var dataadjustment int64
	var bsssect uint8
	var datasect uint8
	for _, s := range d.Symtab.Syms {
		if strings.HasSuffix(s.Name, ".globalMeta") {
			bsssect = s.Sect
			bssadjustment = int64(uintptr(unsafe.Pointer(p))) - int64(s.Value)
		}
		if strings.HasSuffix(s.Name, ".nonzeroMeta") {
			datasect = s.Sect
			dataadjustment = int64(uintptr(unsafe.Pointer(q))) - int64(s.Value)
		}
	}
	if bsssect == 0 {
		println("no bss")
		return
	}
	if datasect == 0 {
		println("no data segment")
		return
	}

	for _, s := range d.Symtab.Syms {
		var adj int64
		switch s.Sect {
		case bsssect:
			adj = bssadjustment
		case datasect:
			adj = dataadjustment
		default:
			continue
		}
		addr := uintptr(int64(s.Value) + adj)
		globalMeta[addr] = s.Name
	}
}
