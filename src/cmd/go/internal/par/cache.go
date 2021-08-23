// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package par

import (
	"sync"
	"sync/atomic"
)

// Cache runs an action once per key and caches the result.
type Cache[K, V any] struct {
	m sync.Map
}

type cacheEntry[V any] struct {
	done   uint32
	mu     sync.Mutex
	result V
}

// Do calls the function f if and only if Do is being called for the first time with this key.
// No call to Do with a given key returns until the one call to f returns.
// Do returns the value returned by the one call to f.
func (c *Cache[K, V]) Do(key K, f func() V) interface{} {
	entryIface, ok := c.m.Load(key)
	if !ok {
		entryIface, _ = c.m.LoadOrStore(key, new(cacheEntry[V]))
	}
	e := entryIface.(*cacheEntry[V])
	if atomic.LoadUint32(&e.done) == 0 {
		e.mu.Lock()
		if atomic.LoadUint32(&e.done) == 0 {
			e.result = f()
			atomic.StoreUint32(&e.done, 1)
		}
		e.mu.Unlock()
	}
	return e.result
}

// Get returns the cached result associated with key.
// It returns nil if there is no such result.
// If the result for key is being computed, Get does not wait for the computation to finish.
func (c *Cache[K, V]) Get(key K) (value V, ok bool) {
	entryIface, ok := c.m.Load(key)
	if !ok {
		return value, false
	}
	e := entryIface.(*cacheEntry[V])
	if atomic.LoadUint32(&e.done) == 0 {
		return value, false
	}
	return e.result, true
}

// Clear removes all entries in the cache.
//
// Concurrent calls to Get may return old values. Concurrent calls to Do
// may return old values or store results in entries that have been deleted.
//
// TODO(jayconrod): Delete this after the package cache clearing functions
// in internal/load have been removed.
func (c *Cache[K, V]) Clear() {
	c.m.Range(func(key, _ interface{}) bool {
		c.m.Delete(key)
		return true
	})
}

// Delete removes an entry from the map. It is safe to call Delete for an
// entry that does not exist. Delete will return quickly, even if the result
// for a key is still being computed; the computation will finish, but the
// result won't be accessible through the cache.
//
// TODO(jayconrod): Delete this after the package cache clearing functions
// in internal/load have been removed.
func (c *Cache[K, V]) Delete(key K) {
	c.m.Delete(key)
}

// DeleteIf calls pred for each key in the map. If pred returns true for a key,
// DeleteIf removes the corresponding entry. If the result for a key is
// still being computed, DeleteIf will remove the entry without waiting for
// the computation to finish. The result won't be accessible through the cache.
//
// TODO(jayconrod): Delete this after the package cache clearing functions
// in internal/load have been removed.
func (c *Cache[K, V]) DeleteIf(pred func(K) bool) {
	c.m.Range(func(key, _ interface{}) bool {
		if pred(key.(K)) {
			c.Delete(key.(K))
		}
		return true
	})
}
