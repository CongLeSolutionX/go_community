// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// This file defines [TypeMap], a mapping whose keys are Types, whose
// hash function is [Hash], and whose equivalence relation is [Identical].

import (
	"bytes"
	"fmt"
	"iter"
	"reflect"

	_ "unsafe" // for linkname hack
)

// TypeMap[V] is a generic map from [Type] to values of type V.
//
// Keys (types) are considered equivalent if they are [Identical].
//
// Just as with map[K]V, a nil *TypeMap is a valid empty map.
type TypeMap[V any] struct {
	table  map[uint][]typeMapEntry[V] // maps hash to bucket; entry.key==nil means unused
	length int                        // number of map entries
}

type typeMapEntry[V any] struct {
	key   Type
	hash  uint
	value V
}

// All returns an iterator over the key/value entries of the map in
// undefined order.
func (m *TypeMap[V]) All() iter.Seq2[Type, V] {
	return func(yield func(Type, V) bool) {
		if m != nil {
			for _, bucket := range m.table {
				for _, e := range bucket {
					if e.key != nil {
						if !yield(e.key, e.value) {
							return
						}
					}
				}
			}
		}
	}
}

// Keys returns an iterator over the map keys in undefined order.
func (m *TypeMap[V]) Keys() iter.Seq[Type] {
	// TODO(adonovan): opt: avoid double iteration.
	return func(yield func(Type) bool) {
		for k := range m.All() {
			if !yield(k) {
				break
			}
		}
	}
}

// Delete removes the entry with the given key, if any.
// It returns true if the entry was found.
func (m *TypeMap[V]) Delete(key Type) bool {
	if m != nil && m.table != nil {
		hash := Hash(key)
		bucket := m.table[hash]
		for i, e := range bucket {
			if e.key != nil && Identical(key, e.key) {
				// We can't compact the bucket as it
				// would disturb iterators.
				bucket[i] = typeMapEntry[V]{}
				m.length--
				return true
			}
		}
	}
	return false
}

// At returns the map entry for the given key.
// The result is zero if the entry is not present.
func (m *TypeMap[V]) At(key Type) V {
	if m != nil && m.table != nil {
		for _, e := range m.table[Hash(key)] {
			if e.key != nil && Identical(key, e.key) {
				return e.value
			}
		}
	}
	return *new(V)
}

// Set sets the map entry for key to val,
// and returns the previous entry, if any.
func (m *TypeMap[V]) Set(key Type, value V) (prev V) {
	if m.table != nil {
		hash := Hash(key)
		bucket := m.table[hash]
		var hole *typeMapEntry[V]
		for i, e := range bucket {
			if e.key == nil {
				hole = &bucket[i]
			} else if Identical(key, e.key) {
				prev = e.value
				bucket[i].value = value
				return
			}
		}

		if hole != nil {
			*hole = typeMapEntry[V]{key, hash, value} // overwrite deleted entry
		} else {
			m.table[hash] = append(bucket, typeMapEntry[V]{key, hash, value})
		}
	} else {
		hash := Hash(key)
		m.table = map[uint][]typeMapEntry[V]{hash: {typeMapEntry[V]{key, hash, value}}}
	}

	m.length++
	return
}

// Len returns the number of map entries.
func (m *TypeMap[V]) Len() int {
	if m != nil {
		return m.length
	}
	return 0
}

// String returns a string representation of the map's entries.
// Values are printed using [fmt.Sprint].
// Order is unspecified.
func (m *TypeMap[V]) String() string {
	return m.toString(true)
}

// KeysString returns a string representation of the map's key set.
// Order is unspecified.
func (m *TypeMap[V]) KeysString() string {
	return m.toString(false)
}

func (m *TypeMap[V]) toString(values bool) string {
	var qf Qualifier
	var buf bytes.Buffer
	buf.WriteByte('{')
	sep := ""
	for k, v := range m.All() {
		buf.WriteString(sep)
		sep = ", "
		WriteType(&buf, k, qf)
		if values {
			fmt.Fprintf(&buf, ": %v", v)
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

// -- Hash --

// Hash computes a hash value for the given type t
// such that Identical(x, y) implies Hash(x) == Hash(y).
func Hash(t Type) uint {
	// See Identical for rationale.
	switch t := t.(type) {
	case *Basic:
		return uint(t.Kind())

	case *Alias:
		return Hash(t.Rhs())

	case *Array:
		return 9043 + 2*uint(t.Len()) + 3*Hash(t.Elem())

	case *Slice:
		return 9049 + 2*Hash(t.Elem())

	case *Struct:
		var hash uint = 9059
		for i, n := 0, t.NumFields(); i < n; i++ {
			f := t.Field(i)
			if f.Anonymous() {
				hash += 8861
			}
			hash += hashString(t.Tag(i))
			hash += hashString(f.Name()) // (ignore f.Pkg)
			hash += Hash(f.Type())
		}
		return hash

	case *Pointer:
		return 9067 + 2*Hash(t.Elem())

	case *Signature:
		var hash uint = 9091
		if t.Variadic() {
			hash *= 8863
		}

		// FIXME all wrong.
		// Use a separate hasher for types inside of the signature, where type
		// parameter identity is modified to be (index, constraint). We must use a
		// new memo for this hasher as type identity may be affected by this
		// masking. For example, in func[T any](*T), the identity of *T depends on
		// whether we are mapping the argument in isolation, or recursively as part
		// of hashing the signature.
		//
		// We should never encounter a generic signature while hashing another
		// generic signature, but defensively set sigTParams only if mask is
		// unset.
		tparams := t.TypeParams()
		for i := 0; i < tparams.Len(); i++ {
			tparam := tparams.At(i)
			hash += 7 * Hash(tparam.Constraint())
		}

		return hash + 3*hashTuple(t.Params()) + 5*hashTuple(t.Results())

	case *Union:
		return hashUnion(t)

	case *Interface:
		// Interfaces are identical if they have the same set of methods, with
		// identical names and types, and they have the same set of type
		// restrictions. See Identical for more details.
		var hash uint = 9103

		// Hash methods.
		for i, n := 0, t.NumMethods(); i < n; i++ {
			// Method order is not significant.
			// Ignore m.Pkg().
			m := t.Method(i)
			// Use shallow hash on method signature to
			// avoid anonymous interface cycles.
			hash += 3*hashString(m.Name()) + 5*shallowHash(m.Type())
		}

		/// TODO terms
		// // Hash type restrictions.
		// terms, err := typeparams.InterfaceTermSet(t) // FIXME what do we do in here?
		// // if err != nil t has invalid type restrictions.
		// if err == nil {
		// 	hash += hashTermSet(terms)
		// }

		return hash

	case *Map:
		return 9109 + 2*Hash(t.Key()) + 3*Hash(t.Elem())

	case *Chan:
		return 9127 + 2*uint(t.Dir()) + 3*Hash(t.Elem())

	case *Named:
		hash := hashPtr(t.Obj())
		targs := t.TypeArgs()
		for i := 0; i < targs.Len(); i++ {
			targ := targs.At(i)
			hash += 2 * Hash(targ)
		}
		return hash

	case *TypeParam:
		// FIXME think
		return 9876543 * uint(t.Index())

	case *Tuple:
		return hashTuple(t)
	}

	panic(fmt.Sprintf("%T: %v", t, t))
}

// hashString computes the Fowler–Noll–Vo hash of s.
func hashString(s string) uint {
	var h uint
	for i := 0; i < len(s); i++ {
		h ^= uint(s[i])
		h *= 16777619
	}
	return h
}

func hashTuple(tuple *Tuple) uint {
	// See go/identicalTypes for rationale.
	n := tuple.Len()
	hash := 9137 + 2*uint(n)
	for i := 0; i < n; i++ {
		hash += 3 * Hash(tuple.At(i).Type())
	}
	return hash
}

func hashUnion(t *Union) uint {
	return 123
	// // Hash type restrictions.
	// terms, err := typeparams.UnionTermSet(t) // FIXME what do we do here?
	// // if err != nil t has invalid type restrictions. Fall back on a non-zero
	// // hash.
	// if err != nil {
	// 	return 9151
	// }
	// return hashTermSet(terms)
}

func hashTermSet(terms []*Term) uint {
	hash := 9157 + 2*uint(len(terms))
	for _, term := range terms {
		// term order is not significant.
		termHash := Hash(term.Type())
		if term.Tilde() {
			termHash *= 9161
		}
		hash += 3 * termHash
	}
	return hash
}

func hashPtr(ptr any) uint {
	if heapObjectsCanMove() {
		panic("Hasher.hashPtr assumes a non-moving GC")
	}
	// TODO(adonovan): use maphash.Comparable (#54670, CL 609761),
	// but it's ultimately the same approach.
	return uint(reflect.ValueOf(ptr).Pointer())
}

//go:linkname heapObjectsCanMove runtime.heapObjectsCanMove
func heapObjectsCanMove() bool

// shallowHash computes a hash of t without looking at any of its
// element Types, to avoid potential anonymous cycles in the types of
// interface methods.
//
// When an unnamed non-empty interface type appears anywhere among the
// arguments or results of an interface method, there is a potential
// for endless recursion. Consider:
//
//	type X interface { m() []*interface { X } }
//
// The problem is that the Methods of the interface in m's result type
// include m itself; there is no mention of the named type X that
// might help us break the cycle.
// (See comment in go/identical, case *Interface, for more.)
func shallowHash(t Type) uint {
	// t is the type of an interface method (Signature),
	// its params or results (Tuples), or their immediate
	// elements (mostly Slice, Pointer, Basic, Named),
	// so there's no need to optimize anything else.
	switch t := t.(type) {
	case *Alias:
		return shallowHash(t.Rhs())

	case *Signature:
		var hash uint = 604171
		if t.Variadic() {
			hash *= 971767
		}
		// The Signature/Tuple recursion is always finite
		// and invariably shallow.
		return hash + 1062599*shallowHash(t.Params()) + 1282529*shallowHash(t.Results())

	case *Tuple:
		n := t.Len()
		hash := 9137 + 2*uint(n)
		for i := 0; i < n; i++ {
			hash += 53471161 * shallowHash(t.At(i).Type())
		}
		return hash

	case *Basic:
		return 45212177 * uint(t.Kind())

	case *Array:
		return 1524181 + 2*uint(t.Len())

	case *Slice:
		return 2690201

	case *Struct:
		return 3326489

	case *Pointer:
		return 4393139

	case *Union:
		return 562448657

	case *Interface:
		return 2124679 // no recursion here

	case *Map:
		return 9109

	case *Chan:
		return 9127

	case *Named:
		return hashPtr(t.Obj())

	case *TypeParam:
		return hashPtr(t.Obj())
	}
	panic(fmt.Sprintf("shallowHash: %T: %v", t, t))
}
