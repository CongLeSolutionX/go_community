// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync_test

import (
	"fmt"
	"sync"
)

func ExampleMutex() {
	sharedData := 0
	mutex := sync.Mutex{}
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			mutex.Lock()
			defer mutex.Unlock()
			sharedData++
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	fmt.Println("SharedData value:", sharedData)

	// Output:
	// SharedData value: 10
}
