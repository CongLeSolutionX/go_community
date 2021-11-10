// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// An Context is an opaque type checking context. It may be used to share
// identical type instances across type-checked packages or calls to
// Instantiate.
//
// It is safe for concurrent use.
type Context struct {
	mu      sync.Mutex
	typeMap map[string][]entry // type hash -> instances
	nextID  int                // next unique ID
	seen    map[*Named]int     // assigned unique IDs
}

type entry struct {
	targs    []Type
	orig     Type
	instance Type
}

// NewContext creates a new Context.
func NewContext() *Context {
	return &Context{
		typeMap: make(map[string][]entry),
		seen:    make(map[*Named]int),
	}
}

// typeHash returns a string representation of typ, which can be used as an exact
// type hash: types that are identical produce identical string representations.
// If typ is a *Named type and targs is not empty, typ is printed as if it were
// instantiated with targs. The result is guaranteed to not contain blanks (" ").
func (ctxt *Context) typeHash(typ Type, targs []Type) string {
	assert(ctxt != nil)
	assert(typ != nil)
	var buf bytes.Buffer

	h := newTypeHasher(&buf, ctxt)
	h.typ(typ)
	h.typeList(targs)

	return strings.Replace(buf.String(), " ", "#", -1) // ReplaceAll is not available in Go1.4
}

// instance returns an existing instantiation of orig with targs, if it exists.
// Otherwise, it returns nil.
func (ctxt *Context) instance(h string, orig Type, targs []Type) Type {
	if existing := ctxt.typeMap[h]; len(existing) > 0 {
		for _, e := range existing {
			if identicalInstance2(orig, targs, e.orig, e.targs) {
				return e.instance
			}
			if debug {
				// While debugging or fuzzing, we want to know if non-identical types
				// have the same hash.
				panic(fmt.Sprintf("non-identical instances: orig: %s, targs: %v and %s", orig, targs, e.instance))
			}
		}
	}
	return nil
}

// typeForHash de-duplicates n against previously seen types with the hash h.
// If an identical type is found with the type hash h, the previously seen type
// is returned. Otherwise, n is returned, and recorded in the Context for the
// hash h.
func (ctxt *Context) typeForHash(h string, orig Type, targs []Type, inst Type) Type {
	assert(inst != nil)

	ctxt.mu.Lock()
	defer ctxt.mu.Unlock()

	if existing := ctxt.typeMap[h]; len(existing) > 0 {
		for _, e := range existing {
			if inst == nil || Identical(inst, e.instance) {
				return e.instance
			}
			if debug && inst != nil {
				panic(fmt.Sprintf("%s and %s are not identical", inst, e.instance))
			}
		}
	}

	if inst != nil {
		ctxt.typeMap[h] = append(ctxt.typeMap[h], entry{
			orig:     orig,
			targs:    targs,
			instance: inst,
		})
	}

	return inst
}

// idForType returns a unique ID for the pointer n.
func (ctxt *Context) idForType(n *Named) int {
	ctxt.mu.Lock()
	defer ctxt.mu.Unlock()
	id, ok := ctxt.seen[n]
	if !ok {
		id = ctxt.nextID
		ctxt.seen[n] = id
		ctxt.nextID++
	}
	return id
}
