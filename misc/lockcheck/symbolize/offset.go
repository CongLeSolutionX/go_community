// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package symbolize

import "debug/dwarf"

type Offset struct {
	// FieldName is the name of a struct field to follow to this
	// offset. If this is "", then this offset is an index offset.
	FieldName string

	// StructName gives the name of the structure containing this
	// field if this is a field offset.
	StructName string

	// Index is an array index to follow to this offset. This is 0
	// for field offsets.
	Index int
}

// OffsetPath returns the symbolic path to access byte offset off
// within a value of type typ. If the offset cannot be resolved, it
// returns nil.
func (d *DWARF) OffsetPath(typ dwarf.Type, off uint64) []Offset {
	var path []Offset
	for typ != nil {
		if off >= uint64(typ.Size()) {
			// Offset is larger than the type.
			return nil
		}
		switch typ1 := typ.(type) {
		default:
			// Assume this is a leaf.
			typ = nil

		case *dwarf.ArrayType:
			eltSize := typ1.Type.Size()
			index := off / uint64(eltSize)
			off = off - index*uint64(eltSize)
			path = append(path, Offset{Index: int(index)})
			typ = typ1.Type

		case *dwarf.StructType:
			// Find the field that contains off.
			bestI, bestOff := -1, uint64(0)
			for i, field := range typ1.Field {
				if off < uint64(field.ByteOffset) {
					continue
				}
				thisOff := off - uint64(field.ByteOffset)
				if bestI == -1 || thisOff < bestOff {
					bestI, bestOff = i, thisOff
				}
			}
			if bestI == -1 {
				return nil
			}
			path = append(path, Offset{
				FieldName:  typ1.Field[bestI].Name,
				StructName: typ1.StructName,
			})
			typ = typ1.Field[bestI].Type
			off = bestOff

		case *dwarf.QualType:
			typ = typ1.Type

		case *dwarf.TypedefType:
			typ = typ1.Type
		}
	}
	return path
}
