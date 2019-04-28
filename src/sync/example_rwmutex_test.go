// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"sync"
)

// This example shows the basic usage of sync.RWMutex to protect a map
// shared between reader and writer go routines.
func ExampleRWMutex() {
	sharedMap := NewSafeMap()
	keys := []string{"zero", "one", "two"}
	done := make(chan bool)
	go func() {
		for i, key := range keys {
			sharedMap.Store(key, i)
		}
		done <- true
	}()
	<-done
	for _, key := range keys {
		if value, ok := sharedMap.Lookup(key); ok {
			fmt.Println("Key:", key, "Value:", value)
		}
	}

	// Output:
	// Key: zero Value: 0
	// Key: one Value: 1
	// Key: two Value: 2
}

type SafeMap struct {
	table map[string]int
	// sync.RWMutex pointer because mutex must not be
	// copied after first use.
	mux *sync.RWMutex
}

func NewSafeMap() SafeMap {
	return SafeMap{
		table: make(map[string]int),
		mux:   &sync.RWMutex{},
	}
}

func (s SafeMap) Lookup(key string) (int, bool) {
	// Lookup only reads the state of map without change it, so
	// read lock can be used.
	s.mux.RLock()
	defer s.mux.RUnlock()
	val, ok := s.table[key]
	return val, ok
}

func (s SafeMap) Store(key string, value int) {
	// Store method changes the state of map, so write lock
	// must be used.
	s.mux.Lock()
	defer s.mux.Unlock()
	s.table[key] = value
}
