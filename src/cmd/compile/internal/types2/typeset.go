// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"bytes"
)

// topTypeSet may be used as type set for the empty interface.
var topTypeSet TypeSet

// A TypeSet represents the type set of an interface.
type TypeSet struct {
	// TODO(gri) consider using a set for the methods for faster lookup
	methods []*Func // all methods of the interface; sorted by unique ID
	types   Type    // typically a *Union; nil means no type restrictions
}

func (s *TypeSet) String() string {
	if s.IsTop() {
		return "⊤"
	}

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, m := range s.methods {
		if i > 0 {
			buf.WriteByte(';')
		}
		buf.WriteByte(' ')
		buf.WriteString(m.String())
	}
	if len(s.methods) > 0 && s.types != nil {
		buf.WriteByte(';')
	}
	if s.types != nil {
		buf.WriteByte(' ')
		writeType(&buf, s.types, nil, nil)
	}

	buf.WriteString(" }") // there was a least one method or type
	return buf.String()
}

// IsTop reports whether type set s is the top type set (corresponding to the empty interface).
func (s *TypeSet) IsTop() bool { return len(s.methods) == 0 && s.types == nil }

// NumMethods returns the number of methods available.
func (s *TypeSet) NumMethods() int { return len(s.methods) }

// Method returns the i'th method of type set s for 0 <= i < s.NumMethods().
// The methods are ordered by their unique Id.
func (s *TypeSet) Method(i int) *Func { return s.methods[i] }

// lookupMethod returns the index of and method with matching package and name, or (-1, nil).
func (s *TypeSet) LookupMethod(pkg *Package, name string) (int, *Func) {
	// TODO(gri) s.methods is sorted - consider binary search
	return lookupMethod(s.methods, pkg, name)
}
