// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !testgo

package main

import (
	"cmd/go/internal/invariant"
	"fmt"
	"log"
	"runtime"
	"strings"
)

func init() {
	invariant.ReportViolationsTo(func(desc string) {
		var (
			callers = make([]uintptr, 1)
			caller  runtime.Frame
		)
		const skip = 3 // runtime.Callers, Check itself, and this hook.
		if runtime.Callers(skip, callers) > 0 {
			caller, _ = runtime.CallersFrames(callers).Next()
		}

		var buf strings.Builder
		buf.WriteString("internal invariant violated")
		if caller.File != "" {
			fmt.Fprintf(&buf, " at %s", caller.File)
			if caller.Line != 0 {
				fmt.Fprintf(&buf, ":%d", caller.Line)
			}
		}
		if len(desc) > 0 {
			fmt.Fprintf(&buf, ": %s", desc)
		}
		buf.WriteString("\nPlease report this, ideally with steps to reproduce, at https://golang.org/issue.")

		log.Print(buf.String())
	})
}
