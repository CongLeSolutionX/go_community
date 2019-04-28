// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"sync"
)

// This example shows the basic usage of sync.RWMutex to protect a map
// shared between reader and writer goroutines.
func ExampleRWMutex() {
	sharedMap := NewSafeMap()
	keys := []string{"zero", "one", "two"}
	go func() {
		for i, key := range keys {
			sharedMap.Store(key, i)
		}
	}()
	go func() {
		for _, key := range keys {
			if value, ok := sharedMap.Lookup(key); ok {
				fmt.Println("Key:", key, "Value:", value)
			}
		}
	}()
}

type SafeMap struct {
	// sync.RWMutex pointer because mutex must not be
	// copied after first use.
	sync.RWMutex
	table map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		table: make(map[string]int),
	}
}

func (s *SafeMap) Lookup(key string) (int, bool) {
	// Lookup only reads the state of the map without
	// changing it, so read lock can be used.
	s.RLock()
	defer s.RUnlock()
	val, ok := s.table[key]
	return val, ok
}

func (s *SafeMap) Store(key string, value int) {
	// Store method changes the state of the map, so write lock
	// must be used.
	s.Lock()
	defer s.Unlock()
	s.table[key] = value
}
