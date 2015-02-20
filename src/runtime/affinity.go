// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

// System topology information is represented as a tree of scheduling domains.
// The root of the tree is schedDomainWorld. The next layer is schedDomainNode,
// then schedDomainProcessor, then schedDomainCore and finally the leafs are
// schedDomainThread. Only schedDomainThread nodes represent actual scheduling
// resources, all other nodes serve merely grouping role. If necessary, later
// this model can be extended with schedDomainCache nodes.
//
// OS-specific code needs to define:
// type schedAffinity
//	an opaque representation of thread affinity with print and add methods.
// func affinityInitOS() *schedDomain
//	initialization function that returns topology tree.

// A schedDomain is one node in the topology tree.
type schedDomain struct {
	kind     schedDomainKind // Node kind.
	affinity schedAffinity   // Affinity representing subtree starting at this node.
	nthread  int             // Number of threads in subtree starting at this node.
	// Abstract representation of distance between this node
	// and all other subnodes of the parent node.
	// If distance == nil, then all nodes are equally distant
	// (or we don't have information about distance).
	// Otherwise, len(distance) is equal to len(sub) of the parent node.
	distance []int
	sub      []*schedDomain // Child nodes.
}

type schedDomainKind int

const (
	schedDomainWorld schedDomainKind = iota
	schedDomainNode
	schedDomainProcessor
	schedDomainCore
	schedDomainThread
)

var schedWorld *schedDomain // Root of the topology tree.

func affinityInit() {
	d := affinityInitOS()
	verifyDomain(d, nil)
	if debug.affinity > 0 {
		println("================")
		println("system topology:")
		printDomain(d, -1, 0)
		println("================")
	}
	schedWorld = d
}

func verifyDomain(d, p *schedDomain) {
	switch d.kind {
	case schedDomainWorld:
		if p != nil {
			throw("schedDomainWorld: bad parent")
		}
	case schedDomainNode:
		if p == nil || p.kind != schedDomainWorld {
			throw("schedDomainNode: bad parent")
		}
	case schedDomainProcessor:
		if p == nil || p.kind != schedDomainNode {
			throw("schedDomainProcessor: bad parent")
		}
	case schedDomainCore:
		if p == nil || p.kind != schedDomainProcessor {
			throw("schedDomainCore: bad parent")
		}
	case schedDomainThread:
		if p == nil || p.kind != schedDomainCore {
			throw("schedDomainThread: bad parent")
		}
		if len(d.sub) != 0 {
			throw("schedDomainThread: has sub")
		}
		d.nthread = 1
	default:
		throw("unknown sched domain")
	}

	if d.distance != nil && len(d.distance) != len(p.sub) {
		throw("bad distance len")
	}

	if d.kind != schedDomainThread {
		for _, s := range d.sub {
			verifyDomain(s, d)
			d.nthread += s.nthread
			d.affinity = d.affinity.add(s.affinity)
		}
	}
}

func printDomain(d *schedDomain, id, ident int) {
	for i := 0; i < ident; i++ {
		print("  ")
	}
	print(domainKindName(d.kind))
	if id >= 0 {
		print(id)
	}
	print(" thr=", d.nthread, " aff=")
	d.affinity.print()
	if len(d.distance) > 0 {
		print(" dist=[")
		for i, p := range d.distance {
			if i != 0 {
				print(",")
			}
			print(p)
		}
		print("]")
	}
	print("\n")
	for i, s := range d.sub {
		printDomain(s, i, ident+1)
	}
}

func domainKindName(k schedDomainKind) string {
	switch k {
	case schedDomainWorld:
		return "world"
	case schedDomainNode:
		return "node"
	case schedDomainProcessor:
		return "processor"
	case schedDomainCore:
		return "core"
	case schedDomainThread:
		return "thread"
	default:
		return "unknown"
	}
}

//go:linkname testing_setAffinity testing.runtime_setAffinity
func testing_setAffinity(gather bool) {
	a := calculateAffinity(schedWorld, int(gomaxprocs), gather)
	if debug.affinity > 0 {
		print("setting affinity to ")
		a.print()
		print("\n")
	}
}

func calculateAffinity(d *schedDomain, procs int, gather bool) schedAffinity {
	if procs >= d.nthread || procs == 1 && d.kind == schedDomainWorld {
		return d.affinity
	}
	if gather && d.kind != schedDomainProcessor {
		// See if it fits into one of the subdomains.
		for _, s := range d.sub {
			if procs <= s.nthread {
				return calculateAffinity(s, procs, gather)
			}
		}
		var a schedAffinity
		remain := procs
		for _, s := range d.sub {
			if remain > 0 && s.nthread > 0 {
				n := s.nthread
				if n > remain {
					n = remain
				}
				remain -= n
				a = a.add(calculateAffinity(s, n, gather))
			}
		}
		return a
	}
	n := make([]int, len(d.sub))
	for i, s := range d.sub {
		n[i] = s.nthread
	}
	remain := d.nthread - procs
	for remain > 0 {
		max := 0
		maxi := 0
		for i := range d.sub {
			if max <= n[i] {
				max = n[i]
				maxi = i
			}
		}
		n[maxi]--
		remain--
	}
	var a schedAffinity
	for i, s := range d.sub {
		if n[i] > 0 {
			a = a.add(calculateAffinity(s, n[i], gather))
		}
	}
	return a
}
