// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trace

// Implemented in src/internal/trace/parser.go.
// We can't directly declare this function in trace_test.go, because presence
// of external functions requires either (1) special casing in go tool, or
// (2) presence of C/asm sources. We can't special case a test package in
// go tool because it is not considered to be "standard". Adding empty_test.s
// does not work either because go tool does not support asm files in tests.
func breakTimestampsForTesting(v bool)

var BreakTimestampsForTesting = breakTimestampsForTesting
