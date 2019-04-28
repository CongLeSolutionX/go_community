// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"sync"
)

// This example shows the basic usage of sync.RWMutex to protect a slice
// shared between reader and writer goroutines.
func ExampleRWMutex() {
	sharedSlice := NewConcurrentSlice()
	values := []string{"zero", "one", "two"}
	go func() {
		for _, val := range values {
			sharedSlice.Add(val)
		}
	}()
	go func() {
		for i := 0; i < sharedSlice.Len(); i++ {
			val := sharedSlice.Index(i)
			fmt.Println("Value:", val)
		}
	}()
}

type ConcurrentSlice struct {
	sync.RWMutex
	slice []string
}

func NewConcurrentSlice() *ConcurrentSlice {
	return &ConcurrentSlice{}
}

func (c *ConcurrentSlice) Len() int {
	// Len only reads the state of the slice without
	// changing it, so read lock can be used.
	c.RLock()
	defer c.RUnlock()
	return len(c.slice)
}

func (c *ConcurrentSlice) Index(i int) string {
	// Index only reads the state of the slice without
	// changing it, so read lock can be used.
	c.RLock()
	defer c.RUnlock()
	val := c.slice[i]
	return val
}

func (c *ConcurrentSlice) Add(val string) {
	// Add method changes the state of the map, so write lock
	// must be used.
	c.Lock()
	defer c.Unlock()
	c.slice = append(c.slice, val)
}
