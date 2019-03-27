// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "runtime/internal/atomic"

// preemptFlags is a bit set of tasks to perform at a goroutine preemption.
type preemptFlags uint8

const (
	// preemptScan indicates that a goroutine's stack should be
	// scanned at the next preemption point.
	preemptScan preemptFlags = 1 << iota

	// preemptSched indicates that a goroutine switch should
	// happen at the next preemption point.
	preemptSched
)

func (f *preemptFlags) set(p preemptFlags) {
	atomic.Or8((*uint8)(f), uint8(p))
}

func (f *preemptFlags) clear(p preemptFlags) {
	atomic.And8((*uint8)(f), ^uint8(p))
}
