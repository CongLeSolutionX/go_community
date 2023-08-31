// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package escape

import (
	"cmd/compile/internal/base"
	"math"
	"strings"
)

const numEscResults = 7

// A leaks represents a set of assignment flows from a parameter:
//   - to the heap (the leaks zeroth element), or
//   - potentially to the heap due to use as an interface receiver (the
//     leaks first element), or
//   - to any of its function's first numEscResults result parameters.
type leaks [2 + numEscResults]uint8

// Empty reports whether l is an empty set (i.e., no assignment flows).
func (l leaks) Empty() bool { return l == leaks{} }

// Heap returns the minimum deref count of any assignment flow from l
// to the heap. If no such flows exist, Heap returns -1.
func (l leaks) Heap() int { return l.get(0) }

// IfaceRecv returns the minimum deref count of any assignment flow from l
// to use as an interface receiver in a call. If no such flows exist,
// IfaceRecv returns -1.
// TODO: we could use boolean flag go122UseIfaceRecvEscapeAnalysis
// to keep the format the same if disabled.
func (l leaks) IfaceRecv() int { return l.get(1) }

// Result returns the minimum deref count of any assignment flow from
// l to its function's i'th result parameter. If no such flows exist,
// Result returns -1.
func (l leaks) Result(i int) int { return l.get(2 + i) }

// AddHeap adds an assignment flow from l to the heap.
func (l *leaks) AddHeap(derefs int) { l.add(0, derefs) }

// AddHeap adds an assignment flow from l to use as an
// interface receiver in a call.
func (l *leaks) AddIfaceRecv(derefs int) { l.add(1, derefs) }

// AddResult adds an assignment flow from l to its function's i'th
// result parameter.
func (l *leaks) AddResult(i, derefs int) { l.add(2+i, derefs) }

func (l *leaks) setResult(i, derefs int) { l.set(2+i, derefs) }

func (l leaks) get(i int) int { return int(l[i]) - 1 }

func (l *leaks) add(i, derefs int) {
	if old := l.get(i); old < 0 || derefs < old {
		l.set(i, derefs)
	}
}

func (l *leaks) set(i, derefs int) {
	v := derefs + 1
	if v < 0 {
		base.Fatalf("invalid derefs count: %v", derefs)
	}
	if v > math.MaxUint8 {
		// 254 dereferences ought to be enough for anyone.
		v = math.MaxUint8
	}

	l[i] = uint8(v)
}

// Optimize removes result flow paths that are equal in length or
// longer than the shortest heap flow path.
func (l *leaks) Optimize() {
	// If we have a path to the heap, then there's no use in
	// keeping equal or longer paths elsewhere.
	if x := l.Heap(); x >= 0 {
		// First, remove any leaks to results.
		for i := 0; i < numEscResults; i++ {
			if l.Result(i) >= x {
				l.setResult(i, -1)
			}
		}
		// Second, remove any iface recv.
		if l.IfaceRecv() >= x {
			// TODO: check if we have a test that hits this
			l.set(1, -1)
		}
	}
}

var leakTagCache = map[leaks]string{}

// Encode converts l into a binary string for export data.
func (l leaks) Encode() string {
	if l.Heap() == 0 {
		// Space optimization: empty string encodes more
		// efficiently in export data.
		return ""
	}
	if s, ok := leakTagCache[l]; ok {
		return s
	}

	n := len(l)
	for n > 0 && l[n-1] == 0 {
		n--
	}
	// TODO: for now, we try to break any readers we don't know about
	s := "es2:" + string(l[:n])
	leakTagCache[l] = s
	return s
}

// parseLeaks parses a binary string representing a leaks.
func parseLeaks(s string) leaks {
	var l leaks
	if !strings.HasPrefix(s, "es2:") {
		l.AddHeap(0)
		return l
	}
	copy(l[:], s[4:])
	return l
}
