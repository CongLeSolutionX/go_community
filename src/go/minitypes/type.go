// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minitypes

import (
	itypes "internal/types"
)

type Type interface {
	Underlying() Type

	internal() itypes.Type
}

type Basic struct {
	*itypes.Basic
}

func (b *Basic) internal() itypes.Type { return b.Basic }
func (b *Basic) Underlying() Type      { return b }

type Struct struct {
	inner *itypes.Struct
}

func NewStruct(fields []*Var, tags []string) *Struct {
	var innerFields []*itypes.Var
	for _, f := range fields {
		innerFields = append(innerFields, f.inner)
	}
	inner := itypes.NewStruct(innerFields, tags)
	return &Struct{inner}
}

func (s *Struct) internal() itypes.Type { return s.inner }
func (s *Struct) Underlying() Type      { return s }

func (s *Struct) Field(i int) *Var {
	mapped := mapObj(s.inner.Field(i))
	if mapped == nil {
		return nil
	}
	return mapped.(*Var)
}

func (s *Struct) NumFields() int {
	return s.inner.NumFields()
}

func (s *Struct) Tag(i int) string {
	return s.inner.Tag(i)
}

type Named struct {
	inner *itypes.Named
}

func NewNamed(obj *TypeName, underlying Type) *Named {
	inner := itypes.NewNamed(obj.inner, underlying.internal())
	return &Named{inner}
}

func (t *Named) internal() itypes.Type {
	return t.inner
}

func (t *Named) Obj() *TypeName {
	mapped := mapObj(t.inner.Obj())
	if mapped == nil {
		return nil
	}
	return mapped.(*TypeName)
}

func (t *Named) SetUnderlying(underlying Type) {
	if underlying == nil {
		panic("types.Named.SetUnderlying: underlying type must not be nil")
	}
	if _, ok := underlying.(*Named); ok {
		panic("types.Named.SetUnderlying: underlying type must not be *Named")
	}
	t.inner.SetUnderlying(underlying.internal())
}

func (t *Named) Underlying() Type { return mapType(t.inner.Underlying()) }
