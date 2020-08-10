// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package par

import "fmt"

// Queue manages a set of work items to be executed in parallel. The number of
// active work items is limited, and excess items are queued sequentially.
type Queue struct {
	maxActive int
	st        chan queueState
	idle      chan struct{} // closed when st.active == 0
}

type queueState struct {
	active  int // number of goroutines processing work; always nonzero when len(backlog) > 0
	backlog []func()
}

// NewQueue returns a Queue that executes up to maxActive items in parallel.
//
// maxActive must be positive.
func NewQueue(maxActive int) *Queue {
	if maxActive < 1 {
		panic(fmt.Sprintf("par.NewQueue called with nonpositive limit (%d)", maxActive))
	}

	idle := make(chan struct{})
	close(idle)
	q := &Queue{
		maxActive: maxActive,
		idle:      idle,
		st:        make(chan queueState, 1),
	}
	q.st <- queueState{}
	return q
}

// Add adds f as a work item in the queue.
//
// Add returns immediately, but the queue will be marked as non-idle until after
// f (and any subsequently-added work) has completed.
func (q *Queue) Add(f func()) {
	st := <-q.st
	if st.active >= q.maxActive {
		st.backlog = append(st.backlog, f)
		q.st <- st
		return
	}
	if st.active == 0 {
		// Mark q as non-idle.
		q.idle = make(chan struct{})
	}
	st.active++
	q.st <- st

	go func() {
		for {
			f()

			st := <-q.st
			if len(st.backlog) == 0 {
				if st.active--; st.active == 0 {
					close(q.idle)
				}
				q.st <- st
				return
			}
			f, st.backlog = st.backlog[0], st.backlog[1:]
			q.st <- st
		}
	}()
}

// Idle returns a channel that will be closed when q has no (active or enqueued)
// work outstanding.
//
// Idle must not be called concurrently with Add if the Add call may cause the
// queue to transition from idle to non-idle.
func (q *Queue) Idle() <-chan struct{} {
	return q.idle
}
