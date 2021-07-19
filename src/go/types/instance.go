// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"go/token"
)

// An instance represents an instantiated generic type syntactically
// (without expanding the instantiation). Type instances appear only
// during type-checking and are replaced by their fully instantiated
// (expanded) types before the end of type-checking.
type instance struct {
	check   *Checker    // for lazy instantiation
	pos     token.Pos   // position of type instantiation; for error reporting only
	posList []token.Pos // position of each targ; for error reporting only
	verify  bool        // if set, constraint satisfaction is verified
}

// expand returns the instantiated (= expanded) type of t.
// The result is either an instantiated *Named type, or
// Typ[Invalid] if there was an error.
func (n *Named) complete() {
	if n.instance != nil && len(n.targs) > 0 && n.underlying == nil {
		check := n.instance.check
		// check.dump("*** completing %v", n)
		inst, _ := check.instantiate(n.instance.pos, n.orig.underlying, n.tparams, n.targs, n.instance.posList, n.instance.verify)
		n.underlying = inst
		n.fromRHS = inst
		n.methods = n.orig.methods
	}
}

// expand expands a type instance into its instantiated
// type and leaves all other types alone. expand does
// not recurse.
func expand(typ Type) Type {
	if t, _ := typ.(*Named); t != nil {
		t.complete()
	}
	return typ
}

// expandf is set to expand.
// Call expandf when calling expand causes compile-time cycle error.
var expandf func(Type) Type

func init() { expandf = expand }
