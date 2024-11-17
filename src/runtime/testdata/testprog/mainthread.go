// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"runtime/mainthread"
)

func MainThread() {
	mainthread.Do(func() { println("Ok") })
}
func init() {
	register("MainThread", func() {
		MainThread()
	})
	register("MainThread2", func() {
		MainThread2()
	})
}
func MainThread2() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Print(err)
		}
	}()
	mainthread.Do(func() {
		print("hello,")
		mainthread.Do(func() {
			print("world")
		})
	})
}
