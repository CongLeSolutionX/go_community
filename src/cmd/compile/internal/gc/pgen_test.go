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
	n.Name.SetUsed(true)
	return n
}

func markNeedZero(n *ir.Node) *ir.Node {
	n.Name.SetNeedzero(true)
	return n
}

func nodeWithClass(n ir.Node, c ir.Class) *ir.Node {
	n.SetClass(c)
	n.Name = new(ir.Name)
	return &n
}

// Test all code paths for cmpstackvarlt.
func TestCmpstackvar(t *testing.T) {
	testdata := []struct {
		a, b *ir.Node
		lt   bool
	}{
		{
			nodeWithClass(ir.Node{}, ir.PAUTO),
			nodeWithClass(ir.Node{}, ir.PFUNC),
			false,
		},
		{
			nodeWithClass(ir.Node{}, ir.PFUNC),
			nodeWithClass(ir.Node{}, ir.PAUTO),
			true,
		},
		{
			nodeWithClass(ir.Node{Xoffset: 0}, ir.PFUNC),
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PFUNC),
			true,
		},
		{
			nodeWithClass(ir.Node{Xoffset: 20}, ir.PFUNC),
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PFUNC),
			false,
		},
		{
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PFUNC),
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PFUNC),
			false,
		},
		{
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PPARAM),
			nodeWithClass(ir.Node{Xoffset: 20}, ir.PPARAMOUT),
			true,
		},
		{
			nodeWithClass(ir.Node{Xoffset: 10}, ir.PPARAMOUT),
			nodeWithClass(ir.Node{Xoffset: 20}, ir.PPARAM),
			true,
		},
		{
			markUsed(nodeWithClass(ir.Node{}, ir.PAUTO)),
			nodeWithClass(ir.Node{}, ir.PAUTO),
			true,
		},
		{
			nodeWithClass(ir.Node{}, ir.PAUTO),
			markUsed(nodeWithClass(ir.Node{}, ir.PAUTO)),
			false,
		},
		{
			nodeWithClass(ir.Node{Type: typeWithoutPointers()}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: typeWithPointers()}, ir.PAUTO),
			false,
		},
		{
			nodeWithClass(ir.Node{Type: typeWithPointers()}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: typeWithoutPointers()}, ir.PAUTO),
			true,
		},
		{
			markNeedZero(nodeWithClass(ir.Node{Type: &types.Type{}}, ir.PAUTO)),
			nodeWithClass(ir.Node{Type: &types.Type{}, Name: &ir.Name{}}, ir.PAUTO),
			true,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{}, Name: &ir.Name{}}, ir.PAUTO),
			markNeedZero(nodeWithClass(ir.Node{Type: &types.Type{}}, ir.PAUTO)),
			false,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{Width: 1}, Name: &ir.Name{}}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: &types.Type{Width: 2}, Name: &ir.Name{}}, ir.PAUTO),
			false,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{Width: 2}, Name: &ir.Name{}}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: &types.Type{Width: 1}, Name: &ir.Name{}}, ir.PAUTO),
			true,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "xyz"}}, ir.PAUTO),
			true,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
			false,
		},
		{
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "xyz"}}, ir.PAUTO),
			nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
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
	inp := []*ir.Node{
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Xoffset: 0, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Xoffset: 10, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Xoffset: 20, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		markUsed(nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO)),
		nodeWithClass(ir.Node{Type: typeWithoutPointers(), Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO),
		markNeedZero(nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO)),
		nodeWithClass(ir.Node{Type: &types.Type{Width: 1}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{Width: 2}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "xyz"}}, ir.PAUTO),
	}
	want := []*ir.Node{
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Xoffset: 0, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Xoffset: 10, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		nodeWithClass(ir.Node{Xoffset: 20, Type: &types.Type{}, Sym: &types.Sym{}}, ir.PFUNC),
		markUsed(nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO)),
		markNeedZero(nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO)),
		nodeWithClass(ir.Node{Type: &types.Type{Width: 2}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{Width: 1}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "abc"}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: &types.Type{}, Sym: &types.Sym{Name: "xyz"}}, ir.PAUTO),
		nodeWithClass(ir.Node{Type: typeWithoutPointers(), Sym: &types.Sym{}}, ir.PAUTO),
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
