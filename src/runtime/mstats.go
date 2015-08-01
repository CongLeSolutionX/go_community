// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory statistics

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	"unsafe"
)

// A MemStats records statistics about the memory allocator.
type MemStats struct {
	// General statistics.
	Alloc      uint64 // bytes allocated and not yet freed
	TotalAlloc uint64 // bytes allocated (even if freed)
	Sys        uint64 // bytes obtained from system (sum of XxxSys below)
	Lookups    uint64 // number of pointer lookups
	Mallocs    uint64 // number of mallocs
	Frees      uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    uint64 // bytes allocated and not yet freed (same as Alloc above)
	HeapSys      uint64 // bytes obtained from system
	HeapIdle     uint64 // bytes in idle spans
	HeapInuse    uint64 // bytes in non-idle span
	HeapReleased uint64 // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  uint64 // bytes used by stack allocator
	StackSys    uint64
	MSpanInuse  uint64 // mspan structures
	MSpanSys    uint64
	MCacheInuse uint64 // mcache structures
	MCacheSys   uint64
	BuckHashSys uint64 // profiling bucket hash table
	GCSys       uint64 // GC metadata
	OtherSys    uint64 // other system allocations

	// Garbage collector statistics.
	NextGC       uint64 // next collection will happen when HeapAlloc ≥ this amount
	LastGC       uint64 // end time of last collection (nanoseconds since 1970)
	PauseTotalNs uint64
	PauseNs      [256]uint64 // circular buffer of recent GC pause durations, most recent at [(NumGC+255)%256]
	PauseEnd     [256]uint64 // circular buffer of recent GC pause end times
	NumGC        uint32
	EnableGC     bool
	DebugGC      bool

	// Per-size allocation statistics.
	// 61 is NumSizeClasses in the C code.
	BySize [61]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	}
}

// Size of the trailing by_size array differs between Go and C,
// and all data after by_size is local to runtime, not exported.
// NumSizeClasses was changed, but we can not change Go struct because of backward compatibility.
// sizeof_C_MStats is what C thinks about size of Go struct.
var sizeof_C_MStats = unsafe.Offsetof(_base.Memstats.By_size) + 61*unsafe.Sizeof(_base.Memstats.By_size[0])

func init() {
	var memStats MemStats
	if sizeof_C_MStats != unsafe.Sizeof(memStats) {
		println(sizeof_C_MStats, unsafe.Sizeof(memStats))
		_base.Throw("MStats vs MemStatsType size mismatch")
	}
}

// ReadMemStats populates m with memory allocator statistics.
func ReadMemStats(m *MemStats) {
	stopTheWorld("read mem stats")

	_base.Systemstack(func() {
		readmemstats_m(m)
	})

	startTheWorld()
}

func readmemstats_m(stats *MemStats) {
	updatememstats(nil)

	// Size of the trailing by_size array differs between Go and C,
	// NumSizeClasses was changed, but we can not change Go struct because of backward compatibility.
	_base.Memmove(unsafe.Pointer(stats), unsafe.Pointer(&_base.Memstats), sizeof_C_MStats)

	// Stack numbers are part of the heap numbers, separate those out for user consumption
	stats.StackSys += stats.StackInuse
	stats.HeapInuse -= stats.StackInuse
	stats.HeapSys -= stats.StackInuse
}

//go:linkname readGCStats runtime/debug.readGCStats
func readGCStats(pauses *[]uint64) {
	_base.Systemstack(func() {
		readGCStats_m(pauses)
	})
}

func readGCStats_m(pauses *[]uint64) {
	p := *pauses
	// Calling code in runtime/debug should make the slice large enough.
	if cap(p) < len(_base.Memstats.Pause_ns)+3 {
		_base.Throw("short slice passed to readGCStats")
	}

	// Pass back: pauses, pause ends, last gc (absolute time), number of gc, total pause ns.
	_base.Lock(&_base.Mheap_.Lock)

	n := _base.Memstats.Numgc
	if n > uint32(len(_base.Memstats.Pause_ns)) {
		n = uint32(len(_base.Memstats.Pause_ns))
	}

	// The pause buffer is circular. The most recent pause is at
	// pause_ns[(numgc-1)%len(pause_ns)], and then backward
	// from there to go back farther in time. We deliver the times
	// most recent first (in p[0]).
	p = p[:cap(p)]
	for i := uint32(0); i < n; i++ {
		j := (_base.Memstats.Numgc - 1 - i) % uint32(len(_base.Memstats.Pause_ns))
		p[i] = _base.Memstats.Pause_ns[j]
		p[n+i] = _base.Memstats.Pause_end[j]
	}

	p[n+n] = _base.Memstats.Last_gc
	p[n+n+1] = uint64(_base.Memstats.Numgc)
	p[n+n+2] = _base.Memstats.Pause_total_ns
	_base.Unlock(&_base.Mheap_.Lock)
	*pauses = p[:n+n+3]
}

//go:nowritebarrier
func updatememstats(stats *_base.Gcstats) {
	if stats != nil {
		*stats = _base.Gcstats{}
	}
	for mp := _base.Allm; mp != nil; mp = mp.Alllink {
		if stats != nil {
			src := (*[unsafe.Sizeof(_base.Gcstats{}) / 8]uint64)(unsafe.Pointer(&mp.Gcstats))
			dst := (*[unsafe.Sizeof(_base.Gcstats{}) / 8]uint64)(unsafe.Pointer(stats))
			for i, v := range src {
				dst[i] += v
			}
			mp.Gcstats = _base.Gcstats{}
		}
	}

	_base.Memstats.Mcache_inuse = uint64(_base.Mheap_.Cachealloc.Inuse)
	_base.Memstats.Mspan_inuse = uint64(_base.Mheap_.Spanalloc.Inuse)
	_base.Memstats.Sys = _base.Memstats.Heap_sys + _base.Memstats.Stacks_sys + _base.Memstats.Mspan_sys +
		_base.Memstats.Mcache_sys + _base.Memstats.Buckhash_sys + _base.Memstats.Gc_sys + _base.Memstats.Other_sys

	// Calculate memory allocator stats.
	// During program execution we only count number of frees and amount of freed memory.
	// Current number of alive object in the heap and amount of alive heap memory
	// are calculated by scanning all spans.
	// Total number of mallocs is calculated as number of frees plus number of alive objects.
	// Similarly, total amount of allocated memory is calculated as amount of freed memory
	// plus amount of alive heap memory.
	_base.Memstats.Alloc = 0
	_base.Memstats.Total_alloc = 0
	_base.Memstats.Nmalloc = 0
	_base.Memstats.Nfree = 0
	for i := 0; i < len(_base.Memstats.By_size); i++ {
		_base.Memstats.By_size[i].Nmalloc = 0
		_base.Memstats.By_size[i].Nfree = 0
	}

	// Flush MCache's to MCentral.
	_base.Systemstack(_gc.Flushallmcaches)

	// Aggregate local stats.
	_gc.Cachestats()

	// Scan all spans and count number of alive objects.
	_base.Lock(&_base.Mheap_.Lock)
	for i := uint32(0); i < _base.Mheap_.Nspan; i++ {
		s := _gc.H_allspans[i]
		if s.State != _base.XMSpanInUse {
			continue
		}
		if s.Sizeclass == 0 {
			_base.Memstats.Nmalloc++
			_base.Memstats.Alloc += uint64(s.Elemsize)
		} else {
			_base.Memstats.Nmalloc += uint64(s.Ref)
			_base.Memstats.By_size[s.Sizeclass].Nmalloc += uint64(s.Ref)
			_base.Memstats.Alloc += uint64(s.Ref) * uint64(s.Elemsize)
		}
	}
	_base.Unlock(&_base.Mheap_.Lock)

	// Aggregate by size class.
	smallfree := uint64(0)
	_base.Memstats.Nfree = _base.Mheap_.Nlargefree
	for i := 0; i < len(_base.Memstats.By_size); i++ {
		_base.Memstats.Nfree += _base.Mheap_.Nsmallfree[i]
		_base.Memstats.By_size[i].Nfree = _base.Mheap_.Nsmallfree[i]
		_base.Memstats.By_size[i].Nmalloc += _base.Mheap_.Nsmallfree[i]
		smallfree += uint64(_base.Mheap_.Nsmallfree[i]) * uint64(_iface.Class_to_size[i])
	}
	_base.Memstats.Nfree += _base.Memstats.Tinyallocs
	_base.Memstats.Nmalloc += _base.Memstats.Nfree

	// Calculate derived stats.
	_base.Memstats.Total_alloc = uint64(_base.Memstats.Alloc) + uint64(_base.Mheap_.Largefree) + smallfree
	_base.Memstats.Heap_alloc = _base.Memstats.Alloc
	_base.Memstats.Heap_objects = _base.Memstats.Nmalloc - _base.Memstats.Nfree
}
