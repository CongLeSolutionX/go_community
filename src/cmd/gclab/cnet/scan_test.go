// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"fmt"
	"math/bits"
	"math/rand/v2"
	"slices"
	"sync"
	"syscall"
	"testing"
	"unsafe"
)

func mkMem(t testing.TB, nPages int) ([]heap.VAddr, func()) {
	mem, err := syscall.Mmap(-1, 0, int(heap.PageBytes.Mul(nPages)), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE|syscall.MAP_ANON)
	if err != nil {
		t.Fatalf("mmap failed: %s", err)
	}
	free := func() {
		syscall.Munmap(mem)
	}
	return unsafe.Slice((*heap.VAddr)(unsafe.Pointer(unsafe.SliceData(mem))), len(mem)/8), free
}

func TestScanSpanPacked(t *testing.T) {
	t.Run("impl=Go", func(t *testing.T) {
		testScanSpanPacked(t, scanSpanPackedGo)
	})
	t.Run("impl=AVX512", func(t *testing.T) {
		testScanSpanPacked(t, scanSpanPackedAVX512)
	})
	t.Run("impl=AVX512Lzcnt", func(t *testing.T) {
		testScanSpanPacked(t, scanSpanPackedAVX512Lzcnt)
	})
}

type scanFunc func(mem unsafe.Pointer, bufp *heap.VAddr, objDarts *objMask, sizeClass uintptr, ptrMask *ptrMask) (count int32)

func testScanSpanPacked(t *testing.T, scan scanFunc) {
	// Construct a fake memory
	mem, free := mkMem(t, 1)
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = heap.VAddr(int(heap.PageBytes) + i + 1)
	}

	// Construct a random pointer mask
	rnd := rand.New(rand.NewPCG(42, 42))
	var ptrs ptrMask
	for i := range ptrs {
		ptrs[i] = rnd.Uint64()
	}

	// Scan a few objects near i to test boundary conditions.
	const scanMask = 0x101

	const sizeClass = 3 // TODO: Test multiple size classes
	buf := make([]heap.VAddr, heap.PageWords)
	buf2 := make([]heap.VAddr, heap.PageWords)
	nObj := heap.PageBytes.Div(heap.Bytes(class_to_size[sizeClass]))
	for i := range nObj - (bits.Len(scanMask) - 1) {
		var objDarts objMask
		objDarts[i/64] = scanMask << (i % 64)

		n := scan(unsafe.Pointer(&mem[0]), &buf[0], &objDarts, sizeClass, &ptrs)
		n2 := scanSpanPackedRef(unsafe.Pointer(&mem[0]), buf2, &objDarts, sizeClass, &ptrs)

		if n2 != n {
			t.Errorf("object %d: want %d count, got %d", i, n2, n)
		} else if !slices.Equal(buf[:n], buf2[:n2]) {
			t.Errorf("object %d: want scanned pointers %d, got %d", i, buf2[:n2], buf[:n])
		}
	}
}

func scanSpanPackedRef(mem unsafe.Pointer, buf []heap.VAddr, objDarts *objMask, sizeClass uintptr, ptrMask *ptrMask) (count int32) {
	expandBy := heap.Bytes(class_to_size[sizeClass]).Words()
	for word := range heap.PageWords {
		objI := word.Div(expandBy)
		if objDarts[objI/64]&(1<<(objI%64)) == 0 {
			continue
		}
		if ptrMask[word/64]&(1<<(word%64)) == 0 {
			continue
		}
		ptr := *(*heap.VAddr)(unsafe.Add(mem, word.Bytes()))
		buf[count] = ptr
		count++
	}
	return count
}

var dataCacheSizes = sync.OnceValue(func() []heap.Bytes {
	cs := getDataCacheSizes()
	for i, c := range cs {
		fmt.Printf("# L%d cache: %s (%d Go pages)\n", i+1, c, c.Div(heap.PageBytes))
	}
	return cs
})

func benchmarkCacheSizes(b *testing.B, fn func(b *testing.B, heapPages int)) {
	cacheSizes := dataCacheSizes()
	b.Run("cache=tiny/pages=1", func(b *testing.B) {
		fn(b, 1)
	})
	for i, cacheBytes := range cacheSizes {
		pages := (cacheBytes * 3 / 4).Div(heap.PageBytes)
		b.Run(fmt.Sprintf("cache=L%d/pages=%d", i+1, pages), func(b *testing.B) {
			fn(b, pages)
		})
	}
	ramPages := (cacheSizes[len(cacheSizes)-1] * 3 / 2).Div(heap.PageBytes)
	b.Run(fmt.Sprintf("cache=ram/pages=%d", ramPages), func(b *testing.B) {
		fn(b, ramPages)
	})
}

func BenchmarkScanSpanPacked(b *testing.B) {
	benchmarkCacheSizes(b, benchmarkScanSpanPacked)
}

func benchmarkScanSpanPacked(b *testing.B, nPages int) {
	const sizeClass = 3 // TODO: Sweep a few size classes

	rnd := rand.New(rand.NewPCG(42, 42))

	// Construct a fake memory
	mem, free := mkMem(b, nPages)
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = heap.VAddr(int(heap.PageBytes) + i + 1)
	}

	// Construct a random pointer mask
	ptrs := make([]ptrMask, nPages)
	for i := range ptrs {
		for j := range ptrs[i] {
			ptrs[i][j] = rnd.Uint64()
		}
	}

	// Visit the pages in a random order
	pageOrder := rnd.Perm(nPages)

	// Create the scan buffer.
	buf := make([]heap.VAddr, heap.PageWords)

	// Sweep from 0 darts to all darts. We'll use the same darts for each page
	// because I don't think that predictability matters.
	objBytes := heap.Bytes(class_to_size[sizeClass])
	nObj := heap.PageBytes.Div(objBytes)
	dartOrder := rnd.Perm(nObj)
	const steps = 11
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		// Set frac darts.
		nDarts := int(float64(len(dartOrder))*frac + 0.5)
		var objDarts objMask
		for _, dart := range dartOrder[:nDarts] {
			objDarts[dart/64] |= 1 << (dart % 64)
		}

		// The AVX-512 versions work on clusters of 64 bytes, so really they're
		// sensitive to how many clusters have marks in them far more than how
		// many objects are marked.
		greyClusters, totalClusters := 0, 0
		for page := range ptrs {
			greyClusters += countGreyClusters(sizeClass, &objDarts, &ptrs[page])
			totalClusters += heap.PageBytes.Div(64)
		}
		pctClusters := 100 * float64(greyClusters) / float64(totalClusters)

		// Report MB/s of how much memory they're actually hitting. This assumes
		// 64 byte cache lines (TODO: Should it assume 128 byte cache lines?)
		// and expands each access to the whole cache line. This is useful for
		// comparing against memory bandwidth.
		//
		// TODO: Add a benchmark that just measures single core memory bandwidth
		// for comparison. (See runtime memcpy benchmarks.)
		//
		// TODO: Should there be a separate measure where we don't expand to
		// cache lines?
		avgBytes := int64(greyClusters * 64 / len(ptrs))

		b.Run(fmt.Sprintf("pct=%d", int(100*frac)), func(b *testing.B) {
			b.Run("impl=Go", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scanSpanPackedGo(unsafe.Pointer(&mem[heap.PageWords.Mul(page)]), &buf[0], &objDarts, sizeClass, &ptrs[page])
				}
			})
			b.Run("impl=AVX512", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scanSpanPackedAVX512(unsafe.Pointer(&mem[heap.PageWords.Mul(page)]), &buf[0], &objDarts, sizeClass, &ptrs[page])
				}
				b.ReportMetric(pctClusters, "clusters-pct")
			})
			b.Run("impl=AVX512Lzcnt", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scanSpanPackedAVX512Lzcnt(unsafe.Pointer(&mem[heap.PageWords.Mul(page)]), &buf[0], &objDarts, sizeClass, &ptrs[page])
				}
				b.ReportMetric(pctClusters, "clusters-pct")
			})
			b.Run("impl=Ref", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scanSpanPackedRef(unsafe.Pointer(&mem[heap.PageWords.Mul(page)]), buf, &objDarts, sizeClass, &ptrs[page])
				}
			})
		})
	}
}

func countGreyClusters(sizeClass int, objDarts *objMask, ptrMask *ptrMask) int {
	clusters := 0
	lastCluster := -1

	expandBy := heap.Bytes(class_to_size[sizeClass]).Words()
	for word := range heap.PageWords {
		objI := word.Div(expandBy)
		if objDarts[objI/64]&(1<<(objI%64)) == 0 {
			continue
		}
		if ptrMask[word/64]&(1<<(word%64)) == 0 {
			continue
		}
		c := word.Bytes().Div(64)
		if c != lastCluster {
			lastCluster = c
			clusters++
		}
	}
	return clusters
}

func BenchmarkScanMaxBandwidth(b *testing.B) {
	// Measure the theoretical "maximum" bandwidth of scanning by reproducing
	// the memory access pattern of a full page scan, but using memcpy as the
	// kernel instead of scanning.
	benchmarkCacheSizes(b, func(b *testing.B, heapPages int) {
		mem, free := mkMem(b, heapPages)
		defer free()
		for i := range mem {
			mem[i] = heap.VAddr(int(heap.PageBytes) + i + 1)
		}

		buf := make([]heap.VAddr, heap.PageWords)
		// TODO: On my laptop, it's quite a bit faster if buf is in mmap'ed
		// memory. WHY? Does this apply to scan, too?
		//buf, freeBuf := mkMem(b, 1)
		//defer freeBuf()

		// Visit the pages in a random order
		rnd := rand.New(rand.NewPCG(42, 42))
		pageOrder := rnd.Perm(heapPages)

		b.SetBytes(int64(heap.PageBytes))

		b.ResetTimer()
		for i := range b.N {
			page := pageOrder[i%len(pageOrder)]
			copy(buf, mem[heap.PageWords.Mul(page):])
		}
	})
}
