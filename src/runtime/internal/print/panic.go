// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package print

import (
	_base "runtime/internal/base"
)

// Print all currently active panics.  Used when crashing.
func Printpanics(p *_base.Panic) {
	if p.Link != nil {
		Printpanics(p.Link)
		print("\t")
	}
	print("panic: ")
	Printany(p.Arg)
	if p.Recovered {
		print(" [recovered]")
	}
	print("\n")
}
