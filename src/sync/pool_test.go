// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Pool is no-op under race detector, so all these tests do not work.
// +build !race

package sync_test

import (
	"runtime"
	"runtime/debug"
	. "sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	// disable GC so we can control when it happens.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	var p Pool
	if p.Get() != nil {
		t.Fatal("expected empty")
	}

	// Make sure that the goroutine doesn't migrate to another P
	// between Put and Get calls.
	Runtime_procPin()
	p.Put("a")
	p.Put("b")
	if g := p.Get(); g != "a" {
		t.Fatalf("got %#v; want a", g)
	}
	if g := p.Get(); g != "b" {
		t.Fatalf("got %#v; want b", g)
	}
	if g := p.Get(); g != nil {
		t.Fatalf("got %#v; want nil", g)
	}
	Runtime_procUnpin()

	p.Put("c")
	debug.SetGCPercent(100) // to allow following GC to actually run
	runtime.GC()
	runtime.GC() // we now keep some objects until two consecutive GCs
	if g := p.Get(); g != nil {
		t.Fatalf("got %#v; want nil after GC", g)
	}
}

func TestPoolNew(t *testing.T) {
	// disable GC so we can control when it happens.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))

	i := 0
	p := Pool{
		New: func() interface{} {
			i++
			return i
		},
	}
	if v := p.Get(); v != 1 {
		t.Fatalf("got %v; want 1", v)
	}
	if v := p.Get(); v != 2 {
		t.Fatalf("got %v; want 2", v)
	}

	// Make sure that the goroutine doesn't migrate to another P
	// between Put and Get calls.
	Runtime_procPin()
	p.Put(42)
	if v := p.Get(); v != 42 {
		t.Fatalf("got %v; want 42", v)
	}
	Runtime_procUnpin()

	if v := p.Get(); v != 3 {
		t.Fatalf("got %v; want 3", v)
	}
}

// Test that Pool does not hold pointers to previously cached resources.
func TestPoolGC(t *testing.T) {
	testPool(t, true)
}

// Test that Pool releases resources on GC.
func TestPoolRelease(t *testing.T) {
	testPool(t, false)
}

func testPool(t *testing.T, drain bool) {
	var p Pool
	const N = 100
loop:
	for try := 0; try < 3; try++ {
		var fin, fin1 uint32
		for i := 0; i < N; i++ {
			v := new(string)
			runtime.SetFinalizer(v, func(_ *string) {
				atomic.AddUint32(&fin, 1)
			})
			p.Put(v)
		}
		if drain {
			for i := 0; i < N; i++ {
				p.Get()
			}
		}
		for i := 0; i < 5; i++ {
			runtime.GC()
			time.Sleep(time.Duration(i*100+10) * time.Millisecond)
			// 1 pointer can remain on stack or elsewhere
			if fin1 = atomic.LoadUint32(&fin); fin1 >= N-1 {
				continue loop
			}
		}
		t.Fatalf("only %v out of %v resources are finalized on try %v", fin1, N, try)
	}
}

// TestPoolPartialRelease tests that after a GC cycle half of the poolLocals
// have been dropped.
func TestPoolPartialRelease(t *testing.T) {
	if runtime.GOMAXPROCS(-1) <= 1 {
		t.Skip("pool partial release test is only stable when GOMAXPROCS > 1")
	}

	// disable GC so we can control when it happens.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	runtime.GC() // run GC now so that any pending GC (triggered by a previous test) does not affect this test

	Ps := runtime.GOMAXPROCS(-1)
	Gs := Ps * 10
	Gobjs := 10000

loop:
	p := Pool{}
	wg := WaitGroup{}
	for i := 0; i < Gs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < Gobjs; j++ {
				p.Put(new(int))
			}
		}()
	}
	wg.Wait()
	if total, empty := p.NumShards(); total != Ps || empty != 0 {
		// we could not fill correctly all shards; retry
		for i := 0; i < 4; i++ {
			runtime.GC()
		}
		goto loop
	}

	runtime.GC()
	if total, empty := p.NumShards(); total != Ps || (empty != Ps/2 && empty != (Ps/2+Ps&1)) {
		// After the first GC half of the shards should be empty. Note that, when Ps is odd,
		// depending on the GC cycle we may get either Ps/2 or Ps/2+1 empty shards.
		t.Fatalf("after first GC: shards total %d/%d, empty %d/%d", total, Ps, empty, Ps/2)
	}

	runtime.GC()
	if total, empty := p.NumShards(); total != Ps || empty != Ps {
		// After the second GC all shards should be empty.
		t.Fatalf("after second GC: shards total %d/%d, empty %d/%d", total, Ps, empty, Ps)
	}
}

// TestPoolCleanup tests that Pools are fully GCed within 4 GC cycles (see the
// comments in poolCleanup).
func TestPoolCleanup(t *testing.T) {
	// disable GC so we can control when it happens.
	defer debug.SetGCPercent(debug.SetGCPercent(-1))
	runtime.GC() // run GC now so that any pending GC (triggered by a previous test) does not affect this test

	var finalized int32
	wg := WaitGroup{}

	for j := 0; j < 1000; j++ {
		p := new(Pool)
		runtime.SetFinalizer(p, func(_ *Pool) {
			atomic.AddInt32(&finalized, 1)
		})
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p.Put(new(int))
			}()
		}
	}
	wg.Wait()

	for i := 0; i < 4; i++ {
		runtime.GC()
		if atomic.LoadInt32(&finalized) == 1000 {
			return
		}
	}
	t.Fatalf("Pool not collected after 4 GC cycles: %d collected", atomic.LoadInt32(&finalized))
}

func TestPoolStress(t *testing.T) {
	const P = 10
	N := int(1e6)
	if testing.Short() {
		N /= 100
	}
	var p Pool
	done := make(chan bool)
	for i := 0; i < P; i++ {
		go func() {
			var v interface{} = 0
			for j := 0; j < N; j++ {
				if v == nil {
					v = 0
				}
				p.Put(v)
				v = p.Get()
				if v != nil && v.(int) != 0 {
					t.Errorf("expect 0, got %v", v)
					break
				}
			}
			done <- true
		}()
	}
	for i := 0; i < P; i++ {
		<-done
	}
}

func BenchmarkPool(b *testing.B) {
	b.ReportAllocs()
	var p Pool
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.Put(1)
			p.Get()
		}
	})
}

func BenchmarkPoolOverflow(b *testing.B) {
	b.ReportAllocs()
	var p Pool
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for b := 0; b < 100; b++ {
				p.Put(1)
			}
			for b := 0; b < 100; b++ {
				p.Get()
			}
		}
	})
}

func BenchmarkPoolWithGC(b *testing.B) {
	b.ReportAllocs()
	p := Pool{New: func() interface{} { return new(int) }}
	var f int32
	b.RunParallel(func(pb *testing.PB) {
		first := atomic.CompareAndSwapInt32(&f, 0, 1)
		var inuse []interface{}
		for pb.Next() {
			for i := 0; i < 10000; i++ {
				inuse = append(inuse, p.Get())
			}
			for _, v := range inuse {
				p.Put(v)
			}
			inuse = inuse[:0]
			if first {
				runtime.GC()
			}
		}
	})
}
