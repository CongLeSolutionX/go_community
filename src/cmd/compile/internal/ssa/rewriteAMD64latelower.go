// Code generated from gen/AMD64latelower.rules; DO NOT EDIT.
// generated with: cd gen; go run *.go

package ssa

func rewriteValueAMD64latelower(v *Value) bool {
	switch v.Op {
	case OpAMD64LEAL1:
		return rewriteValueAMD64latelower_OpAMD64LEAL1(v)
	case OpAMD64LEAL2:
		return rewriteValueAMD64latelower_OpAMD64LEAL2(v)
	case OpAMD64LEAL4:
		return rewriteValueAMD64latelower_OpAMD64LEAL4(v)
	case OpAMD64LEAL8:
		return rewriteValueAMD64latelower_OpAMD64LEAL8(v)
	case OpAMD64LEAQ1:
		return rewriteValueAMD64latelower_OpAMD64LEAQ1(v)
	case OpAMD64LEAQ2:
		return rewriteValueAMD64latelower_OpAMD64LEAQ2(v)
	case OpAMD64LEAQ4:
		return rewriteValueAMD64latelower_OpAMD64LEAQ4(v)
	case OpAMD64LEAQ8:
		return rewriteValueAMD64latelower_OpAMD64LEAQ8(v)
	case OpAMD64LEAW1:
		return rewriteValueAMD64latelower_OpAMD64LEAW1(v)
	case OpAMD64LEAW2:
		return rewriteValueAMD64latelower_OpAMD64LEAW2(v)
	case OpAMD64LEAW4:
		return rewriteValueAMD64latelower_OpAMD64LEAW4(v)
	case OpAMD64LEAW8:
		return rewriteValueAMD64latelower_OpAMD64LEAW8(v)
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL1(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL1 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAL [c] {s} (LEAL1 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAL)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL1, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL2(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL2 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAL [c] {s} (LEAL2 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAL)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL2, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL4(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL4 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAL [c] {s} (LEAL4 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAL)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL4, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL8 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAL [c] {s} (LEAL8 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAL)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL8, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ1(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ1 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAQ [c] {s} (LEAQ1 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAQ)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ1, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ2(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ2 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAQ [c] {s} (LEAQ2 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAQ)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ2, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ4(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ4 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAQ [c] {s} (LEAQ4 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAQ)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ4, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ8 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAQ [c] {s} (LEAQ8 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAQ)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ8, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAW1(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAW1 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAW [c] {s} (LEAW1 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAW)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAW1, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAW2(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAW2 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAW [c] {s} (LEAW2 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAW)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAW2, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAW4(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAW4 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAW [c] {s} (LEAW4 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAW)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAW4, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAW8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAW8 <t> [c] {s} x y)
	// cond: !t.IsPtrShaped() && c != 0 && s == nil
	// result: (LEAW [c] {s} (LEAW8 <x.Type> x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(!t.IsPtrShaped() && c != 0 && s == nil) {
			break
		}
		v.reset(OpAMD64LEAW)
		v.AuxInt = int32ToAuxInt(c)
		v.Aux = symToAux(s)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAW8, x.Type)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteBlockAMD64latelower(b *Block) bool {
	return false
}
