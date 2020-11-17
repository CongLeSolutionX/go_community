// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/compile/internal/ir"
	"cmd/compile/internal/types"
	"reflect"
	"sort"
	"testing"
)

func typeWithoutPointers() *types.Type {
	t := types.New(types.TSTRUCT)
	f := &types.Field{Type: types.New(types.TINT)}
	t.SetFields([]*types.Field{f})
	return t
}

func typeWithPointers() *types.Type {
	t := types.New(types.TSTRUCT)
	f := &types.Field{Type: types.NewPtr(types.New(types.TINT))}
	t.SetFields([]*types.Field{f})
	return t
}

func markUsed(n *ir.Node) *ir.Node {
	n.Name().SetUsed(true)
	return n
}

func markNeedZero(n *ir.Node) *ir.Node {
	n.Name().SetNeedzero(true)
	return n
}

func nodeWithClass(n ir.Node, c ir.Class) *ir.Node {
	n.SetClass(c)
	n.SetName(new(ir.Name))
	return &n
}

// Test all code paths for cmpstackvarlt.
func TestCmpstackvar(t *testing.T) {
	nod := func(xoffset int64, t *types.Type, s *types.Sym, nam *ir.Name) ir.Node {
		n := new(ir.Node)
		n.SetType(t)
		n.Sym = s
		n.Xoffset = xoffset
		n.SetName(nam)
		return *n
	}
	testdata := []struct {
		a, b *ir.Node
		lt   bool
	}{
		{
			nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO),
			nodeWithClass(nod(0, nil, nil, nil), ir.PFUNC),
			false,
		},
		{
			nodeWithClass(nod(0, nil, nil, nil), ir.PFUNC),
			nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO),
			true,
		},
		{
			nodeWithClass(nod(0, nil, nil, nil), ir.PFUNC),
			nodeWithClass(nod(10, nil, nil, nil), ir.PFUNC),
			true,
		},
		{
			nodeWithClass(nod(20, nil, nil, nil), ir.PFUNC),
			nodeWithClass(nod(10, nil, nil, nil), ir.PFUNC),
			false,
		},
		{
			nodeWithClass(nod(10, nil, nil, nil), ir.PFUNC),
			nodeWithClass(nod(10, nil, nil, nil), ir.PFUNC),
			false,
		},
		{
			nodeWithClass(nod(10, nil, nil, nil), ir.PPARAM),
			nodeWithClass(nod(20, nil, nil, nil), ir.PPARAMOUT),
			true,
		},
		{
			nodeWithClass(nod(10, nil, nil, nil), ir.PPARAMOUT),
			nodeWithClass(nod(20, nil, nil, nil), ir.PPARAM),
			true,
		},
		{
			markUsed(nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO)),
			nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO),
			true,
		},
		{
			nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO),
			markUsed(nodeWithClass(nod(0, nil, nil, nil), ir.PAUTO)),
			false,
		},
		{
			nodeWithClass(nod(0, typeWithoutPointers(), nil, nil), ir.PAUTO),
			nodeWithClass(nod(0, typeWithPointers(), nil, nil), ir.PAUTO),
			false,
		},
		{
			nodeWithClass(nod(0, typeWithPointers(), nil, nil), ir.PAUTO),
			nodeWithClass(nod(0, typeWithoutPointers(), nil, nil), ir.PAUTO),
			true,
		},
		{
			markNeedZero(nodeWithClass(nod(0, &types.Type{}, nil, nil), ir.PAUTO)),
			nodeWithClass(nod(0, &types.Type{}, nil, &ir.Name{}), ir.PAUTO),
			true,
		},
		{
			nodeWithClass(nod(0, &types.Type{}, nil, &ir.Name{}), ir.PAUTO),
			markNeedZero(nodeWithClass(nod(0, &types.Type{}, nil, nil), ir.PAUTO)),
			false,
		},
		{
			nodeWithClass(nod(0, &types.Type{Width: 1}, nil, &ir.Name{}), ir.PAUTO),
			nodeWithClass(nod(0, &types.Type{Width: 2}, nil, &ir.Name{}), ir.PAUTO),
			false,
		},
		{
			nodeWithClass(nod(0, &types.Type{Width: 2}, nil, &ir.Name{}), ir.PAUTO),
			nodeWithClass(nod(0, &types.Type{Width: 1}, nil, &ir.Name{}), ir.PAUTO),
			true,
		},
		{
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}, nil), ir.PAUTO),
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "xyz"}, nil), ir.PAUTO),
			true,
		},
		{
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}, nil), ir.PAUTO),
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}, nil), ir.PAUTO),
			false,
		},
		{
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "xyz"}, nil), ir.PAUTO),
			nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}, nil), ir.PAUTO),
			false,
		},
	}
	for _, d := range testdata {
		got := cmpstackvarlt(d.a, d.b)
		if got != d.lt {
			t.Errorf("want %#v < %#v", d.a, d.b)
		}
		// If we expect a < b to be true, check that b < a is false.
		if d.lt && cmpstackvarlt(d.b, d.a) {
			t.Errorf("unexpected %#v < %#v", d.b, d.a)
		}
	}
}

func TestStackvarSort(t *testing.T) {
	nod := func(xoffset int64, t *types.Type, s *types.Sym) ir.Node {
		n := new(ir.Node)
		n.SetType(t)
		n.Sym = s
		n.Xoffset = xoffset
		return *n
	}
	inp := []*ir.Node{
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(10, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(20, &types.Type{}, &types.Sym{}), ir.PFUNC),
		markUsed(nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO)),
		nodeWithClass(nod(0, typeWithoutPointers(), &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO),
		markNeedZero(nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO)),
		nodeWithClass(nod(0, &types.Type{Width: 1}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{Width: 2}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "xyz"}), ir.PAUTO),
	}
	want := []*ir.Node{
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(10, &types.Type{}, &types.Sym{}), ir.PFUNC),
		nodeWithClass(nod(20, &types.Type{}, &types.Sym{}), ir.PFUNC),
		markUsed(nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO)),
		markNeedZero(nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO)),
		nodeWithClass(nod(0, &types.Type{Width: 2}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{Width: 1}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "abc"}), ir.PAUTO),
		nodeWithClass(nod(0, &types.Type{}, &types.Sym{Name: "xyz"}), ir.PAUTO),
		nodeWithClass(nod(0, typeWithoutPointers(), &types.Sym{}), ir.PAUTO),
	}
	sort.Sort(byStackVar(inp))
	if !reflect.DeepEqual(want, inp) {
		t.Error("sort failed")
		for i := range inp {
			g := inp[i]
			w := want[i]
			eq := reflect.DeepEqual(w, g)
			if !eq {
				t.Log(i, w, g)
			}
		}
	}
}
