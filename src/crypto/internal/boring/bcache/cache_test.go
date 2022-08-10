// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bcache

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

var registeredCache Cache[int]

func init() {
	registeredCache.Register()
}

func TestCache(t *testing.T) {
	// Use unregistered cache for functionality tests,
	// to keep the runtime from clearing behind our backs.
	c := new(Cache[int])

	// Create many entries.
	seq := uint32(0)
	nextKey := func() unsafe.Pointer {
		x := new(int)
		*x = int(atomic.AddUint32(&seq, 1))
		return unsafe.Pointer(x)
	}
	nextValue := func() *int {
		x := new(int)
		*x = int(atomic.AddUint32(&seq, 1))
		return x
	}
	m := make(map[unsafe.Pointer]*int)
	for i := 0; i < 10000; i++ {
		k := nextKey()
		v := nextValue()
		m[k] = v
		c.Put(k, v)
	}

	// Overwrite a random 20% of those.
	n := 0
	for k := range m {
		v := nextValue()
		m[k] = v
		c.Put(k, v)
		if n++; n >= 2000 {
			break
		}
	}

	// Check results.
	str := func(p unsafe.Pointer) string {
		if p == nil {
			return "nil"
		}
		return fmt.Sprint(*(*int)(p))
	}
	for k, v := range m {
		if cv := c.Get(k); cv != v {
			t.Fatalf("c.Get(%v) = %v, want %v", str(k), *cv, *v)
		}
	}

	c.Clear()
	for k := range m {
		if cv := c.Get(k); cv != nil {
			t.Fatalf("after GC, c.Get(%v) = %v, want nil", str(k), *cv)
		}
	}

	// Check that registered cache is cleared at GC.
	c = &registeredCache
	for k, v := range m {
		c.Put(k, v)
	}
	runtime.GC()
	for k := range m {
		if cv := c.Get(k); cv != nil {
			t.Fatalf("after Clear, c.Get(%v) = %v, want nil", str(k), *cv)
		}
	}

	// Check that cache works for concurrent access.
	// Lists are discarded if they reach 1000 entries,
	// and there are cacheSize list heads, so we should be
	// able to do 100 * cacheSize entries with no problem at all.
	c = new(Cache[int])
	var barrier, wg sync.WaitGroup
	const N = 100
	barrier.Add(N)
	wg.Add(N)
	var lost int32
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()

			m := make(map[unsafe.Pointer]*int)
			for j := 0; j < cacheSize; j++ {
				k, v := nextKey(), nextValue()
				m[k] = v
				c.Put(k, v)
			}
			barrier.Done()
			barrier.Wait()

			for k, v := range m {
				if cv := c.Get(k); cv != v {
					t.Errorf("c.Get(%v) = %v, want %v", str(k), *cv, *v)
					atomic.AddInt32(&lost, +1)
				}
			}
		}()
	}
	wg.Wait()
	if lost != 0 {
		t.Errorf("lost %d entries", lost)
	}
}
