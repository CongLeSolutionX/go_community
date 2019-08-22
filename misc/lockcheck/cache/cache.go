// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache provides caching structures.
package cache

import "sync"

// Cache is a concurrent-safe map that is automatically populated.
type Cache struct {
	// New is a function that returns the value of key k when k
	// isn't already in the map.
	New func(k interface{}) interface{}

	dataLock sync.Mutex
	data     map[interface{}]*cacheVal
}

type cacheVal struct {
	fill sync.Once
	val  interface{}
}

// Get returns the value for key, creating it by calling c.New if
// necessary.
func (c *Cache) Get(key interface{}) interface{} {
	c.dataLock.Lock()
	if c.data == nil {
		c.data = make(map[interface{}]*cacheVal)
	}
	cval, ok := c.data[key]
	if !ok {
		cval = new(cacheVal)
		c.data[key] = cval
	}
	c.dataLock.Unlock()

	cval.fill.Do(func() {
		cval.val = c.New(key)
	})
	return cval.val
}
