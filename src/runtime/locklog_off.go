// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !locklog

// No-op implementation of lock logging when logging is disabled.

package runtime

type lockClass struct {
	name string
}

func lockLabeled(l *mutex, cls *lockClass, rank uint64) {
	lock(l)
}

func lockLogAcquire(l *mutex)    {}
func lockLogRelease(l *mutex)    {}
func lockLogMayAcquire(l *mutex) {}
func lockLogFlushAll()           {}

type lockLogPerM struct{}

func (lo *lockLogPerM) flush() {}
