// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package synctest provides support for testing concurrent code.
package synctest

import (
	_ "unsafe" // for go:linkname
)

// Run executes f in a new goroutine.
//
// The new goroutine and any goroutines transitively started by it form
// an isolated "bubble".
// Run waits for all goroutines in the bubble to exit before returning.
//
// Goroutines in the bubble use a synthetic time implementation.
// The initial time is midnight UTC 2000-01-01.
//
// Time advances when every goroutine in the bubble is idle.
// For example, a call to time.Sleep will block until all other
// goroutines are idle and return after the bubble's clock has
// advanced.
//
// If every goroutine is idle and there are no timers scheduled,
// Run panics.
//
// Channels, time.Timers, and time.Tickers created within the bubble
// are associated with it. Operating on a bubbled channel, timer, or ticker
// from outside the bubble panics.
//
//go:linkname Run
func Run(f func())

// Wait blocks until every goroutine within the current bubble,
// other than the current goroutine, is idle.
//
// A goroutine is idle if it is blocked on:
//   - a send or receive on a channel from within the bubble
//   - a select statement where every case is a channel within the bubble
//   - sync.Cond.Wait
//   - time.Sleep
//
// A goroutine executing a system call or waiting for
// an external event such as a network operation is never idle.
// For example, a goroutine blocked reading from an network connection
// is not idle, even if no data is currently available on the connection.
//
// A goroutine is not idle when blocked on a send or receive on a channel
// that was not created within its bubble.
//
//go:linkname Wait
func Wait()

//go:linkname Running
func Running() bool
