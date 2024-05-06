// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.rangefunc

package main

import (
	"fmt"
	"iter"
	"runtime"
)

func init() {
	register("CoroLockOSThreadIterLock", func() { CoroLockOSThread(callerExhaust, iterLock) })
	register("CoroLockOSThreadIterLockYield", func() { CoroLockOSThread(callerExhaust, iterLockYield) })
	register("CoroLockOSThreadLock", func() { CoroLockOSThread(callerExhaustLocked, iterSimple) })
	register("CoroLockOSThreadLockIterNested", func() { CoroLockOSThread(callerExhaustLocked, iterNested) })
	register("CoroLockOSThreadLockIterLock", func() { CoroLockOSThread(callerExhaustLocked, iterLock) })
	register("CoroLockOSThreadLockIterLockYield", func() { CoroLockOSThread(callerExhaustLocked, iterLockYield) })
	register("CoroLockOSThreadLockIterYieldNewG", func() { CoroLockOSThread(callerExhaustLocked, iterYieldNewG) })
	register("CoroLockOSThreadLockAfterPull", func() { CoroLockOSThread(callerLockAfterPull, iterSimple) })
	register("CoroLockOSThreadStopLocked", func() { CoroLockOSThread(callerStopLocked, iterSimple) })
	register("CoroLockOSThreadStopLockedIterNested", func() { CoroLockOSThread(callerStopLocked, iterNested) })
}

func CoroLockOSThread(driver func(iter.Seq[int]) error, seq iter.Seq[int]) {
	if err := driver(seq); err != nil {
		println("error:", err.Error())
		return
	}
	println("OK")
}

func callerExhaust(i iter.Seq[int]) error {
	next, _ := iter.Pull(i)
	for {
		v, ok := next()
		if !ok {
			break
		}
		if v != 5 {
			return fmt.Errorf("bad iterator: wanted value %d, got %d", 5, v)
		}
	}
	return nil
}

func callerExhaustLocked(i iter.Seq[int]) error {
	runtime.LockOSThread()
	next, _ := iter.Pull(i)
	for {
		v, ok := next()
		if !ok {
			break
		}
		if v != 5 {
			return fmt.Errorf("bad iterator: wanted value %d, got %d", 5, v)
		}
	}
	runtime.UnlockOSThread()
	return nil
}

func callerLockAfterPull(i iter.Seq[int]) error {
	n := 0
	next, _ := iter.Pull(i)
	for {
		runtime.LockOSThread()
		n++
		v, ok := next()
		if !ok {
			break
		}
		if v != 5 {
			return fmt.Errorf("bad iterator: wanted value %d, got %d", 5, v)
		}
	}
	for range n {
		runtime.UnlockOSThread()
	}
	return nil
}

func callerStopLocked(i iter.Seq[int]) error {
	runtime.LockOSThread()
	next, stop := iter.Pull(i)
	v, _ := next()
	stop()
	if v != 5 {
		return fmt.Errorf("bad iterator: wanted value %d, got %d", 5, v)
	}
	runtime.UnlockOSThread()
	return nil
}

func iterSimple(yield func(int) bool) {
	for range 3 {
		if !yield(5) {
			return
		}
	}
}

func iterNested(yield func(int) bool) {
	next, stop := iter.Pull(iterSimple)
	for {
		v, ok := next()
		if ok {
			if !yield(v) {
				stop()
			}
		} else {
			return
		}
	}
}

func iterLock(yield func(int) bool) {
	for range 3 {
		runtime.LockOSThread()
		runtime.UnlockOSThread()

		if !yield(5) {
			return
		}
	}
}

func iterLockYield(yield func(int) bool) {
	for range 3 {
		runtime.LockOSThread()
		ok := yield(5)
		runtime.UnlockOSThread()
		if !ok {
			return
		}
	}
}

func iterYieldNewG(yield func(int) bool) {
	for range 3 {
		done := make(chan struct{})
		var ok bool
		go func() {
			ok = yield(5)
			done <- struct{}{}
		}()
		<-done
		if !ok {
			return
		}
	}
}
