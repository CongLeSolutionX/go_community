// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package symbolize

import (
	"debug/dwarf"
	"fmt"
	"sort"
)

// VarTab stores the locations of global variables for a DWARF object.
//
// It is safe to access concurrently.
type VarTab struct {
	dd   *dwarf.Data
	vars []Var
}

type Var struct {
	Addr   uint64
	Name   string
	typOff dwarf.Offset
}

func NewVarTab(dd *dwarf.Data) (*VarTab, error) {
	// Index the global variables.
	var vars []Var
	r := dd.Reader()
	for {
		ent, err := r.Next()
		if err != nil {
			return nil, err
		}
		if ent == nil {
			break
		}

		switch ent.Tag {
		default:
			r.SkipChildren()

		case dwarf.TagCompileUnit, dwarf.TagModule, dwarf.TagNamespace:
			// Descend into these.
			break

		case dwarf.TagVariable:
			name, ok := ent.Val(dwarf.AttrName).(string)
			if !ok {
				break
			}
			loc, ok := ent.Val(dwarf.AttrLocation).([]byte)
			if !ok {
				break
			}
			// Parse basic locations descriptions.
			const OpAddr = 0x03
			if len(loc) != 1+r.AddressSize() || loc[0] != OpAddr {
				// Can't handle complex location description.
				break
			}
			var addr uint64
			switch r.AddressSize() {
			case 4:
				addr = uint64(r.ByteOrder().Uint32(loc[1:]))
			case 8:
				addr = r.ByteOrder().Uint64(loc[1:])
			default:
				panic(fmt.Sprintf("unexpected address size %d", r.AddressSize()))
			}
			typOff, ok := ent.Val(dwarf.AttrType).(dwarf.Offset)
			if !ok {
				break
			}
			vars = append(vars, Var{addr, name, typOff})
		}
	}
	sort.Slice(vars, func(i, j int) bool {
		return vars[i].Addr < vars[j].Addr
	})
	return &VarTab{dd, vars}, nil
}

// Lookup returns the variable at addr, or false if there is no
// variable at this location. Lookup does not know the size of
// variables, so a lookup that does not fall in a variable may return
// the last variable before its address.
func (vt *VarTab) Lookup(addr uint64) (Var, bool) {
	i := sort.Search(len(vt.vars), func(i int) bool {
		return addr < vt.vars[i].Addr
	}) - 1
	if i < 0 {
		return Var{}, false
	}
	return vt.vars[i], true
}

// VarType returns the type of a Var as a dwarf.Type.
func (vt *VarTab) VarType(v Var) (dwarf.Type, error) {
	return vt.dd.Type(v.typOff)
}
