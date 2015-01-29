// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Lock-free stack.
// The following code runs only on g0 stack.

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

func lfstackpush(head *uint64, node *lfnode) {
	node.pushcnt++
	new := lfstackPack(node, node.pushcnt)
	if node1, _ := lfstackUnpack(new); node1 != node {
		println("runtime: lfstackpush invalid packing: node=", node, " cnt=", _core.Hex(node.pushcnt), " packed=", _core.Hex(new), " -> node=", node1, "\n")
		_lock.Gothrow("lfstackpush")
	}
	for {
		old := Atomicload64(head)
		node.next = old
		if Cas64(head, old, new) {
			break
		}
	}
}

func lfstackpop(head *uint64) unsafe.Pointer {
	for {
		old := Atomicload64(head)
		if old == 0 {
			return nil
		}
		node, _ := lfstackUnpack(old)
		next := Atomicload64(&node.next)
		if Cas64(head, old, next) {
			return unsafe.Pointer(node)
		}
	}
}
