// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package base

var Finlock Mutex // protects the following variables
var Fing *G       // goroutine that runs finalizers
var Fingwait bool
var Fingwake bool

func wakefing() *G {
	var res *G
	Lock(&Finlock)
	if Fingwait && Fingwake {
		Fingwait = false
		Fingwake = false
		res = Fing
	}
	Unlock(&Finlock)
	return res
}

var (
	FingRunning bool
)
