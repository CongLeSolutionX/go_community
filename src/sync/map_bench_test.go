// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"golang.org/x/sync/syncmap"
)

type bench struct {
	setup func(*testing.B, mapInterface)
	perG  func(begin, end int, m mapInterface)
}

func benchMap(b *testing.B, bench bench) {
	ng := runtime.GOMAXPROCS(0)
	for _, ty := range [...]mapInterface{&DeepCopyMap{}, &RWMutexMap{}, &sync.Map{}} {
		b.Run(fmt.Sprintf("%T", ty), func(b *testing.B) {
			m := reflect.New(reflect.TypeOf(ty).Elem()).Interface().(mapInterface)
			if bench.setup != nil {
				bench.setup(b, m)
			}

			var wg sync.WaitGroup
			wg.Add(ng)

			begin := 0
			start := make(chan struct{})
			for g := ng; g > 0; g-- {
				end := begin + b.N

				go func(begin, end int) {
					<-start
					bench.perG(begin, end, m)
					wg.Done()
				}(begin, end)

				begin = end
			}

			b.ResetTimer()
			close(start)
			wg.Wait()
		})
	}
}

func BenchmarkLoadMostlyHits(b *testing.B) {
	const hits, misses = 1023, 1

	benchMap(b, bench{
		setup: func(_ *testing.B, m mapInterface) {
			for i := 0; i < hits; i++ {
				m.LoadOrStore(i, i)
			}
			// Prime the map to get it into a steady state.
			for i := 0; i < hits*2; i++ {
				m.Load(i % hits)
			}
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				j := i % (hits + misses)
				m.Load(j)
			}
		},
	})
}

func BenchmarkLoadMostlyMisses(b *testing.B) {
	const hits, misses = 1, 1023

	benchMap(b, bench{
		setup: func(_ *testing.B, m mapInterface) {
			for i := 0; i < hits; i++ {
				m.LoadOrStore(i, i)
			}
			// Prime the map to get it into a steady state.
			for i := 0; i < hits*2; i++ {
				m.Load(i % hits)
			}
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				j := i % (hits + misses)
				m.Load(j)
			}
		},
	})
}

func BenchmarkLoadOrStoreBalanced(b *testing.B) {
	const hits, misses = 2, 2

	benchMap(b, bench{
		setup: func(b *testing.B, m mapInterface) {
			if _, ok := m.(*DeepCopyMap); ok {
				b.Skip("DeepCopyMap has quadratic running time.")
			}
			for i := 0; i < hits; i++ {
				m.LoadOrStore(i, i)
			}
			// Prime the map to get it into a steady state.
			for i := 0; i < hits*2; i++ {
				m.Load(i % hits)
			}
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				j := i % (hits + misses)
				if j < hits {
					if _, ok := m.LoadOrStore(j, i); !ok {
						panic(fmt.Sprintf("unexpected miss for %v", j))
					}
				} else {
					if v, ok := m.Load(i); ok {
						panic(fmt.Sprintf("unexpected hit for %v: existing value %v", i, v))
					}
					if v, loaded := m.LoadOrStore(i, i); loaded {
						panic(fmt.Sprintf("failed to store %v: existing value %v", i, v))
					}
				}
			}
		},
	})
}

func BenchmarkLoadOrStoreUnique(b *testing.B) {
	benchMap(b, bench{
		setup: func(b *testing.B, m mapInterface) {
			if _, ok := m.(*DeepCopyMap); ok {
				b.Skip("DeepCopyMap has quadratic running time.")
			}
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				m.LoadOrStore(i, i)
			}
		},
	})
}

func BenchmarkLoadOrStoreCollision(b *testing.B) {
	benchMap(b, bench{
		setup: func(_ *testing.B, m mapInterface) {
			m.LoadOrStore(0, 0)
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				m.LoadOrStore(0, 0)
			}
		},
	})
}

// BenchmarkAdversarialAlloc tests performance when we store a new value
// immediately whenever the map is promoted to clean.
//
// This forces the Load calls to always acquire the map's mutex.
func BenchmarkAdversarialAlloc(b *testing.B) {
	var m syncmap.Map
	dirty := reflect.ValueOf(&m).Elem().FieldByName("dirty")

	for i := 0; i < b.N; i++ {
		m.Load(i)
		if dirty.IsNil() {
			m.LoadOrStore(i, i)
		}
	}
}

// BenchmarkAdversarialDelete tests performance when we delete and restore a
// value immediately after a large map has been promoted.
//
// This forces the Load calls to always acquire the map's mutex and periodically
// makes a full copy of the map despite changing only one entry.
func BenchmarkAdversarialDelete(b *testing.B) {
	const mapSize = 1 << 10

	benchMap(b, bench{
		setup: func(_ *testing.B, m mapInterface) {
			for i := 0; i < mapSize; i++ {
				m.Store(i, i)
			}
		},

		perG: func(begin, end int, m mapInterface) {
			for i := begin; i < end; i++ {
				m.Load(i)

				if i%mapSize == 0 {
					var key int
					m.Range(func(k, _ interface{}) bool {
						key = k.(int)
						return false
					})
					m.Delete(key)
					m.Store(key, key)
				}
			}
		},
	})
}
