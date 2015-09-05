// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"hash/crc32"
	"runtime"
	"testing"
)

var logbuf = make([]uint64, 10000)

func testLog(t *testing.T, read func([]uint64) int, f func([]uint64)) {
	for count := 0; ; count++ {
		n := read(logbuf)
		if n == 0 {
			break
		}
		if count > 0 {
			continue // just drain log - no allocations!
		}
		log := logbuf[:n]
		for len(log) > 0 {
			n = int(log[0])
			if n < 3 || n > len(log) {
				t.Errorf("bad record %x !!!", log)
				break
			}
			t.Logf("%#x", log[:n])
			f(log[:n])
			log = log[n:]
		}
	}
}

func TestProfCPU(t *testing.T) {
	buf := make([]byte, 100000)
	runtime.SetProfTag(0x123)
	runtime.EnableProfCPU(100, 200)
	for i := 0; i < 20000; i++ {
		crc32.Update(0, crc32.IEEETable, buf)
	}
	runtime.EnableProfCPU(0, 0)
	found := false
	testLog(t, runtime.ReadProfCPU, func(entry []uint64) {
		if entry[1] == 0x123 && !found {
			t.Logf("found goroutine tag")
			found = true
		}
	})
	if !found {
		t.Fatal("did not find profiling entry with goroutine tag")
	}
}

var g interface{}

func TestProfMem(t *testing.T) {
	old := runtime.MemProfileRate
	runtime.MemProfileRate = 1
	defer func() {
		runtime.MemProfileRate = old
		runtime.GC()
	}()

	drain := func() { testLog(t, runtime.ReadProfMem, func([]uint64){}) }

	t.Logf("phase 1")
	drain()
	runtime.EnableProfMem(200)
	for i := 0; i < 1000; i++ {
		g = new(int)
		runtime.SetProfTag(uint64(i) + 0x234)
	}
	runtime.GC()
	g = nil
	runtime.GC()

	found := false
	want := uint64(0x234)
	testLog(t, runtime.ReadProfMem, func(entry []uint64) {
		if entry[1] == want {
			want++
			if want == 0x234+20 {
				found = true
			}
		}
	})
	if !found {
		t.Errorf("did not find profiling entry with goroutine tag %#x", want)
	}

	t.Logf("phase 2")
	drain()
	g = new([64]int) // allocate, to trigger log write, to trigger overflow report
	runtime.GC()
	g = nil
	runtime.GC()

	found = false
	testLog(t, runtime.ReadProfMem, func(entry []uint64) {
		if entry[1] == ^uint64(0) {
			found = true
		}
	})
	if !found {
		t.Errorf("did not find overflow profiling entry")
	}

	t.Logf("phase 3")
	drain()
	runtime.SetProfTag(0x345)
	for i := 0; i < 1000; i++ {
		g = new([64]int)
	}
	runtime.GC()
	g = nil
	runtime.GC()

	found = false
	want = uint64(0x345)
	testLog(t, runtime.ReadProfMem, func(entry []uint64) {
		if entry[1] == want {
			found = true
		}
	})
	if !found {
		t.Errorf("did not find profiling entry with goroutine tag %#x", want)
	}
}
