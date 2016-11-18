// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"sync"
	"sync/atomic"
)

// A dmap is an distributed atomic map.
// It supports only insert and lookup; there is no delete.
type dmap struct {
	mu      sync.Mutex   // held when writing to root
	root    atomic.Value // *node
	m       atomic.Value // map[interface{}]interface{}
	compare func(interface{}, interface{}) int
}

// A node is a node in the distributed map.
type node struct {
	key, val    interface{}  // never changed after creation
	mu          sync.Mutex   // held when writing to left or right
	left, right atomic.Value // *node
	count       uint32       // number of lookups
}

// New returns a new distribute map given a comparison function.
func newdmap(compare func(interface{}, interface{}) int) *dmap {
	return &dmap{
		compare: compare,
	}
}

// Insert inserts a key/value pair into a dmap.
func (d *dmap) insert(key, val interface{}) {
	var n *node
	for {
		root, _ := d.root.Load().(*node)
		if root != nil {
			n = root
			break
		}
		root = &node{
			key: key,
			val: val,
		}
		d.mu.Lock()
		if d.root.Load() == nil {
			d.root.Store(root)
			d.mu.Unlock()
			return
		}
		d.mu.Unlock()
	}

	for {
		cmp := d.compare(key, n.key)
		if cmp == 0 {
			if val != n.val {
				panic("invalid double-insert")
			}
			return
		}
		p := &n.left
		if cmp > 0 {
			p = &n.right
		}
		n2, _ := (*p).Load().(*node)
		if n2 != nil {
			n = n2
		} else {
			n2 = &node{
				key: key,
				val: val,
			}
			n.mu.Lock()
			if (*p).Load() == nil {
				(*p).Store(n2)
				n.mu.Unlock()
				return
			}
			n.mu.Unlock()
		}
	}
}

// Lookup looks up a key in the distributed map.
func (d *dmap) lookup(key interface{}) interface{} {
	// Common values are cached in a map held in the root.
	m, _ := d.m.Load().(map[interface{}]interface{})
	if val, ok := m[key]; ok {
		return val
	}

	n, _ := d.root.Load().(*node)
	for n != nil {
		cmp := d.compare(key, n.key)
		if cmp == 0 {
			count := atomic.AddUint32(&n.count, 1)

			// Add this key/val pair to the map in the root,
			// but only if it's worth copying the existing map.
			if count < 0 || (count > 1 && int(count) > len(m)) {
				newm := make(map[interface{}]interface{}, len(m)+1)
				for k, v := range m {
					newm[k] = v
				}
				newm[key] = n.val

				// It's possible that some other
				// goroutine has updated d.m since we
				// loaded it.  That means we did extra
				// work but it's otherwise OK.
				d.m.Store(newm)
			}

			return n.val
		}

		p := &n.left
		if cmp > 0 {
			p = &n.right
		}
		n, _ = (*p).Load().(*node)
	}

	return nil
}
