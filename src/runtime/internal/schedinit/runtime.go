// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_sched "runtime/internal/sched"
)

// TODO: Move to parfor.go when parfor.c becomes parfor.go.
func parforalloc(nthrmax uint32) *_sched.Parfor {
	return &_sched.Parfor{
		Thr:     &make([]_sched.Parforthread, nthrmax)[0],
		Nthrmax: nthrmax,
	}
}

var Envs []string
var Argslice []string
