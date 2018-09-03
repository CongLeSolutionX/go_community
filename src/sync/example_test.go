// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"sync"
)

type httpPkg struct{}

func (httpPkg) Get(url string) {}

var http httpPkg

// This example fetches several URLs concurrently,
// using a WaitGroup to block until all the fetches are complete.
func ExampleWaitGroup() {
	var wg sync.WaitGroup
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
}

func ExampleOnce() {
	var once sync.Once
	onceBody := func() {
		fmt.Println("Only once")
	}
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			once.Do(onceBody)
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	// Output:
	// Only once
}

func ExampleMap() {
	var sm sync.Map

	_, ok := sm.Load("non-existent-key")
	fmt.Println(ok)

	sm.Store("mykey", "myvalue")
	v, ok := sm.Load("mykey")
	fmt.Printf("%q, %v\n", v, ok)
	// Output:
	// false
	// "myvalue", true
}

func ExampleMap_Range() {
	var sm sync.Map
	sm.Store("0", "00")
	sm.Store("1", "01")
	sm.Store("2", "10")
	sm.Store("3", "11")

	sm.Range(func(k, v interface{}) bool {
		fmt.Printf("%s in binary is %s\n", k, v)
		return true
	})
	// Unordered output:
	// 2 in binary is 10
	// 3 in binary is 11
	// 0 in binary is 00
	// 1 in binary is 01
}
