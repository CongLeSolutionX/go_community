// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build nodebuglog

package runtime

type debugLogger struct{}

type debugLog struct{}

func dlog() *debugLog {
	return nil
}

func printDebugLog() {
}

func (l *debugLog) int64(x int64) *debugLog           { return nil }
func (l *debugLog) int(x int) *debugLog               { return nil }
func (l *debugLog) uint64(x uint64) *debugLog         { return nil }
func (l *debugLog) hex(x uintptr) *debugLog           { return nil }
func (l *debugLog) pc(pc uintptr) *debugLog           { return nil }
func (l *debugLog) traceback(pcs []uintptr) *debugLog { return nil }
func (l *debugLog) callers(skip, limit int) *debugLog { return nil }
func (l *debugLog) string(x string) *debugLog         { return nil }
func (l *debugLog) sp() *debugLog                     { return nil }

func (l *debugLog) end() {}
