// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parallel for algorithm.

package schedinit

import (
	_sched "runtime/internal/sched"
)

func parforalloc(nthrmax uint32) *_sched.Parfor {
	return &_sched.Parfor{
		Thr: make([]_sched.Parforthread, nthrmax),
	}
}
