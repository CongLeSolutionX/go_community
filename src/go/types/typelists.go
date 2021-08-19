// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// TParamList holds a list of type parameters.
type TParamList struct{ tparams []*TypeParam }

// Len returns the number of type parameters in the list.
// It is safe to call on a nil receiver.
func (tps *TParamList) Len() int { return len(tps.list()) }

// At returns the i'th type parameter in the list.
func (tps *TParamList) At(i int) *TypeParam { return tps.tparams[i] }

// list is for internal use where we expect a []*TypeParam.
// TODO(rfindley): list should probably be eliminated: we can pass around a
// TypeList instead.
func (tps *TParamList) list() []*TypeParam {
	if tps == nil {
		return nil
	}
	return tps.tparams
}

// TypeList holds a list of types.
type TypeList struct{ types []Type }

// Len returns the number of types in the list.
// It is safe to call on a nil receiver.
func (tl *TypeList) Len() int { return len(tl.list()) }

// At returns the i'th type in the list.
func (tl *TypeList) At(i int) Type { return tl.types[i] }

// list is for internal use where we expect a []Type.
// TODO(rfindley): list should probably be eliminated: we can pass around a
// TypeList instead.
func (tl *TypeList) list() []Type {
	if tl == nil {
		return nil
	}
	return tl.types
}

// ----------------------------------------------------------------------------
// Implementation

func bindTParams(list []*TypeParam) *TParamList {
	if len(list) == 0 {
		return nil
	}
	for i, typ := range list {
		if typ.index >= 0 {
			panic("type parameter bound more than once")
		}
		typ.index = i
	}
	return &TParamList{tparams: list}
}
