// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

import "fmt"

const (
	LEAF_RANK        = 1
	ZERO_RANK        = 0
	NOT_KEY32        = int32(-0x80000000)
	ENTRY_CACHE_SIZE = 5
)

// Red-black tree with data stored at internal nodes,
// following Tarjan, Data Structures and Network Algorithms,
// pp 48-52, using explicit rank instead of red and black.
// Deletion is not yet implemented because it is not yet needed.
// Extra operations glb, lub, glbEq, lubEq are provided for
// use in sparse lookup algorithms.

type RBTint32 struct {
	root    *node32
	compare rbComparer
	aCache  node32Cache
	// An extra-clever implementation will have special cases
	// for small sets, but we are not extra-clever today.
}

type rbComparer func(x, y int32) Cmp

type node32 struct {
	// Standard conventions hold for left = smaller, right = larger
	left, right, parent *node32
	data                fmt.Stringer
	key                 int32
	rank                int8 // From Tarjan pp 48-49:
	// IF x is a node with a parent, then x.rank <= x.parent.rank <= x.rank+1
	// IF x is a node with a grandparent, then x.rank < x.parent.parent.rank
	// IF x is an "external [null] node", then x.rank = 0 && x.parent.rank = 1
	// Translating from Tarjan, any node with one or more null children should have rank = 1
}

type node32Cache struct {
	nodeCache      *[ENTRY_CACHE_SIZE]node32
	nodeCacheIndex int
}

// makeNode returns a new leaf node with the given key and nil data.
func (t *node32Cache) makeNode(key int32) *node32 {
	if t.nodeCache == nil || t.nodeCacheIndex >= ENTRY_CACHE_SIZE {
		t.nodeCache = &[ENTRY_CACHE_SIZE]node32{}
		t.nodeCacheIndex = 0
	}
	x := &t.nodeCache[t.nodeCacheIndex]
	t.nodeCacheIndex++
	// x := &node32{}
	x.key = key
	x.rank = LEAF_RANK
	return x
}

// IsSingle returns true iff t is empty.
func (t *RBTint32) IsEmpty() bool {
	return t.root == nil
}

// IsSingle returns true iff t is a singleton (leaf).
func (t *RBTint32) IsSingle() bool {
	return t.root != nil && t.root.isLeaf()
}

// VisitInOrder applies f to the key and data pairs in t,
// with keys ordered from smallest to largest.
func (t *RBTint32) VisitInOrder(f func(int32, fmt.Stringer)) {
	if t.root == nil {
		return
	}
	t.root.visitInOrder(f)
}

func (n *node32) nilOrData() fmt.Stringer {
	if n == nil {
		return nil
	}
	return n.data
}

func (n *node32) nilOrKeyAndData() (k int32, d fmt.Stringer) {
	if n == nil {
		k = NOT_KEY32
		d = nil
	} else {
		k = n.key
		d = n.data
	}
	return
}

func (n *node32) zeroOrRank() int8 {
	if n == nil {
		return 0
	}
	return n.rank
}

// Find returns the data associated with x in the tree, or
// nil if x is not in the tree.
func (t *RBTint32) Find(x int32) fmt.Stringer {
	return t.root.find(t.compare, x).nilOrData()
}

// Insert either adds x to the tree if x was not previously
// a key in the tree, or updates the data for x in the tree if
// x was already a key in the tree.  The previous data associated
// with x is returned, and is nil if x was not previously a
// key in the tree.
func (t *RBTint32) Insert(x int32, data fmt.Stringer) fmt.Stringer {
	n := t.root
	var newroot *node32
	if n == nil {
		n = t.aCache.makeNode(x)
		newroot = n
	} else {
		newroot, n = n.insert(t.compare, x, &t.aCache)
	}
	r := n.data
	n.data = data
	t.root = newroot
	return r
}

// Min returns the minimum element of t.
// If t is empty, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) Min() (k int32, d fmt.Stringer) {
	return t.root.min().nilOrKeyAndData()
}

// Max returns the maximum element of t.
// If t is empty, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) Max() (k int32, d fmt.Stringer) {
	return t.root.max().nilOrKeyAndData()
}

// Glb returns the greatest-lower-bound-exclusive of x and the associated
// data.  If x has no glb in the tree, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) Glb(x int32) (k int32, d fmt.Stringer) {
	return t.root.glb(t.compare, x, false).nilOrKeyAndData()
}

// GlbEq returns the greatest-lower-bound-inclusive of x and the associated
// data.  If x has no glbEQ in the tree, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) GlbEq(x int32) (k int32, d fmt.Stringer) {
	return t.root.glb(t.compare, x, true).nilOrKeyAndData()
}

// Lub returns the least-upper-bound-exclusive of x and the associated
// data.  If x has no lub in the tree, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) Lub(x int32) (k int32, d fmt.Stringer) {
	return t.root.lub(t.compare, x, false).nilOrKeyAndData()
}

// LubEq returns the least-upper-bound-inclusive of x and the associated
// data.  If x has no lubEq in the tree, then (NOT_KEY32, nil) is returned.
func (t *RBTint32) LubEq(x int32) (k int32, d fmt.Stringer) {
	return t.root.lub(t.compare, x, true).nilOrKeyAndData()
}

func (t *node32) isLeaf() bool {
	return t.left == nil && t.right == nil && t.rank == LEAF_RANK
}

func (t *node32) visitInOrder(f func(int32, fmt.Stringer)) {
	if t.left != nil {
		t.left.visitInOrder(f)
	}
	f(t.key, t.data)
	if t.right != nil {
		t.right.visitInOrder(f)
	}
}

func (t *node32) maxChildRank() int8 {
	if t.left == nil {
		if t.right == nil {
			return ZERO_RANK
		}
		return t.right.rank
	}
	if t.right == nil {
		return t.left.rank
	}
	if t.right.rank > t.left.rank {
		return t.right.rank
	}
	return t.left.rank
}

func (t *node32) minChildRank() int8 {
	if t.left == nil || t.right == nil {
		return ZERO_RANK
	}
	if t.right.rank < t.left.rank {
		return t.right.rank
	}
	return t.left.rank
}

func (t *node32) find(compare rbComparer, key int32) *node32 {
	for t != nil {
		c := compare(key, t.key)
		if c < CMPeq {
			t = t.left
		} else if c > CMPeq {
			t = t.right
		} else {
			return t
		}
	}
	return nil
}

func (t *node32) min() *node32 {
	if t == nil {
		return t
	}
	for t.left != nil {
		t = t.left
	}
	return t
}

func (t *node32) max() *node32 {
	if t == nil {
		return t
	}
	for t.right != nil {
		t = t.right
	}
	return t
}

func (t *node32) glb(compare rbComparer, key int32, allow_eq bool) *node32 {
	var best *node32 = nil
	for t != nil {
		c := compare(key, t.key)
		if c <= CMPeq {
			if allow_eq && c == CMPeq {
				return t
			}
			// t is too big, glb is to left.
			t = t.left
		} else {
			// t is a lower bound, record it and seek a better one.
			best = t
			t = t.right
		}
	}
	return best
}

func (t *node32) lub(compare rbComparer, key int32, allow_eq bool) *node32 {
	var best *node32 = nil
	for t != nil {
		c := compare(key, t.key)
		if c >= CMPeq {
			if allow_eq && c == CMPeq {
				return t
			}
			// t is too small, lub is to right.
			t = t.right
		} else {
			// t is a upper bound, record it and seek a better one.
			best = t
			t = t.left
		}
	}
	return best
}

func (t *node32) insert(compare rbComparer, x int32, w *node32Cache) (newroot, newnode *node32) {
	// defaults
	newroot = t
	newnode = t
	c := compare(x, t.key)
	if c == CMPeq {
		return
	}
	if c < CMPeq {
		if t.left == nil {
			n := w.makeNode(x)
			n.parent = t
			t.left = n
			newnode = n
			return
		}
		var new_l *node32
		new_l, newnode = t.left.insert(compare, x, w)
		t.left = new_l
		new_l.parent = t
		newrank := 1 + new_l.maxChildRank()
		if newrank > t.rank {
			if newrank > 1+t.right.zeroOrRank() { // rotations required
				if new_l.left.zeroOrRank() < new_l.right.zeroOrRank() {
					// double rotation
					t.left = new_l.rightToRoot()
				}
				newroot = t.leftToRoot()
				return
			} else {
				t.rank = newrank
			}
		}
	} else { // x > t.key
		if t.right == nil {
			n := w.makeNode(x)
			n.parent = t
			t.right = n
			newnode = n
			return
		}
		var new_r *node32
		new_r, newnode = t.right.insert(compare, x, w)
		t.right = new_r
		new_r.parent = t
		newrank := 1 + new_r.maxChildRank()
		if newrank > t.rank {
			if newrank > 1+t.left.zeroOrRank() { // rotations required
				if new_r.right.zeroOrRank() < new_r.left.zeroOrRank() {
					// double rotation
					t.right = new_r.leftToRoot()
				}
				newroot = t.rightToRoot()
				return
			} else {
				t.rank = newrank
			}
		}
	}
	return
}

func (t *node32) rightToRoot() *node32 {
	//    this
	// left  right
	//      rl   rr
	//
	// becomes
	//
	//       right
	//    this   rr
	// left  rl
	//
	right := t.right
	rl := right.left
	right.parent = t.parent
	right.left = t
	t.parent = right
	// parent's child ptr fixed in caller
	t.right = rl
	if rl != nil {
		rl.parent = t
	}
	return right
}

func (t *node32) leftToRoot() *node32 {
	//     this
	//  left  right
	// ll  lr
	//
	// becomes
	//
	//    left
	//   ll  this
	//      lr  right
	//
	left := t.left
	lr := left.right
	left.parent = t.parent
	left.right = t
	t.parent = left
	// parent's child ptr fixed in caller
	t.left = lr
	if lr != nil {
		lr.parent = t
	}
	return left
}

// next returns the successor of t in a left-to-right
// walk of the tree in which t is embedded.
func (t *node32) next() *node32 {
	// If there is a right child, it is to the right
	r := t.right
	if r != nil {
		return r.min()
	}
	// if t is p.left, then p, else repeat.
	p := t.parent
	for p != nil {
		if p.left == t {
			return p
		}
		t = p
		p = t.parent
	}
	return nil
}

// prev returns the predecessor of t in a left-to-right
// walk of the tree in which t is embedded.
func (t *node32) prev() *node32 {
	// If there is a left child, it is to the left
	l := t.left
	if l != nil {
		return l.max()
	}
	// if t is p.right, then p, else repeat.
	p := t.parent
	for p != nil {
		if p.right == t {
			return p
		}
		t = p
		p = t.parent
	}
	return nil
}
