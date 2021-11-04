// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"io"
	"math"
	"math/rand"
	"testing"
	"time"
)

func TestContextHashCollisions(t *testing.T) {
	if debug {
		t.Skip("hash collisions are expected, and would fail debug assertions")
	}
	// Unit test the de-duplication fall-back logic in Context.
	//
	// We can't test this via Instantiate because this is only a fall-back in
	// case our hash is imperfect.
	//
	// These lookups and updates use reasonable looking types in an attempt to
	// make them robust to internal type assertions, but could equally well use
	// arbitrary types.

	// Create some distinct origin types. nullaryP and nullaryQ have no
	// parameters and are identical (but have different type parameter names).
	// unaryP has a parameter.
	var nullaryP, nullaryQ, unaryP Type
	{
		// type nullaryP = func[P any]()
		tparam := NewTypeParam(NewTypeName(token.NoPos, nil, "P", nil), &emptyInterface)
		nullaryP = NewSignatureType(nil, nil, []*TypeParam{tparam}, nil, nil, false)
	}
	{
		// type nullaryQ = func[Q any]()
		tparam := NewTypeParam(NewTypeName(token.NoPos, nil, "Q", nil), &emptyInterface)
		nullaryQ = NewSignatureType(nil, nil, []*TypeParam{tparam}, nil, nil, false)
	}
	{
		// type unaryP = func[P any](_ P)
		tparam := NewTypeParam(NewTypeName(token.NoPos, nil, "P", nil), &emptyInterface)
		params := NewTuple(NewVar(token.NoPos, nil, "_", tparam))
		unaryP = NewSignatureType(nil, nil, []*TypeParam{tparam}, params, nil, false)
	}

	ctxt := NewContext()

	// Update the context with an instantiation of nullaryP.
	inst := NewSignatureType(nil, nil, nil, nil, nil, false)
	if got := ctxt.update("", nullaryP, []Type{Typ[Int]}, inst); got != inst {
		t.Error("bad")
	}

	// unaryP is not identical to nullaryP, so we should not get inst when
	// instantiated with identical type arguments.
	if got := ctxt.lookup("", unaryP, []Type{Typ[Int]}); got != nil {
		t.Error("bad")
	}

	// nullaryQ is identical to nullaryP, so we *should* get inst when
	// instantiated with identical type arguments.
	if got := ctxt.lookup("", nullaryQ, []Type{Typ[Int]}); got != inst {
		t.Error("bad")
	}

	// ...but verify we don't get inst with different type arguments.
	if got := ctxt.lookup("", nullaryQ, []Type{Typ[String]}); got != nil {
		t.Error("bad")
	}
}

const (
	// Type kinds, based off iimport.go.
	definedType = iota
	pointerType
	sliceType
	arrayType
	chanType
	mapType
	signatureType
	structType
	interfaceType
	// Note: no typeParamType here. Type parameters are read as part of the
	// containing type.
	instanceType
	// unionType
)

func TestOneRandomType(t *testing.T) {
	t.Skip("experimental")
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, 1000)
	_, err := rand.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	r := &typeReader{
		input: bytes.NewReader(data),
		scope: NewScope(nil, token.NoPos, token.NoPos, ""),
	}
	r.definedType()
	// r.typ()
	fmt.Println(len(r.scope.Names()))
	fmt.Println(packageFromScope(r.scope))
}

func FuzzContext(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 40 {
			t.Skip("not enough data")
		}
		l := len(data) / 2
		typ1 := newType(data[:l])
		typ2 := newType(data[l:])
		if typ1 == nil || typ2 == nil {
			return
		}
		ctxt := NewContext()
		hash1 := typHash(ctxt, typ1)
		hash2 := typHash(ctxt, typ2)
		identical := Identical(typ1, typ2)
		if identical != (hash1 == hash2) {
			panic(fmt.Sprintf("bad types %s (%T) and %s (%T) (identical: %t, hash1: %q hash2: %q)", typ1, typ1, typ2, typ2, identical, hash1, hash2))
		}
	})
}

func typHash(ctxt *Context, typ Type) string {
	var buf bytes.Buffer
	hasher := newTypeHasher(&buf, ctxt)
	hasher.typ(typ)
	return buf.String()
}

func newType(data []byte) (res Type) {
	defer func() {
		if x := recover(); x == noType {
			res = nil
		}
	}()
	r := &typeReader{
		input: bytes.NewReader(data),
		scope: NewScope(nil, token.NoPos, token.NoPos, ""),
	}
	return r.typ()
}

type typeReader struct {
	input io.ByteReader
	scope *Scope
	depth int
}

func packageFromScope(scope *Scope) string {
	var buf bytes.Buffer
	buf.WriteString("package p\n\n")
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		buf.WriteString(ObjectString(obj, func(*Package) string { return "" }))
		buf.WriteString("\n\n")
	}
	return buf.String()
}

func (r *typeReader) pushScope() *Scope {
	r.scope = NewScope(r.scope, token.NoPos, token.NoPos, "")
	return r.scope
}

func (r *typeReader) popScope() {
	r.scope = r.scope.Parent()
}

// string interprets the next byte as a character, mixing in upper and lower
// case to trigger exported/non-exported behavior. Narrowing to 52 possible
// names increases the likelihood of interesting behavior due to referencing.
func (r *typeReader) string() string {
	v := r.uint8()
	char := rune(v % ('Z' - 'A'))
	ucase := (v>>4)%2 != 0 // interpret the second nibble as the case.
	if ucase {
		return string('A' + char)
	} else {
		return string('a' + char)
	}
}

// noType signals that the byte stream contains no more types.
var noType = errors.New("all done")

// uint8 reads a byte from the input, as a uint8.
func (r *typeReader) uint8() uint8 {
	b, err := r.input.ReadByte()
	if err != nil {
		panic(noType) // assume we're out of bytes
	}
	return b
}

// bool reads a bool from the input.
func (r *typeReader) bool() bool {
	return r.uint8()%2 != 0
}

// probability returns true with probability n (out of 100).
func (r *typeReader) probability(n int) bool {
	v := r.uint8()
	return (int(v) * 100 / math.MaxUint8) < n
}

func (r *typeReader) typ() Type {
	r.depth++
	defer func() {
		r.depth--
	}()
	// Choose a kind of type based on the next byte. We allocate low-range byte
	// values to basic types, then partition some amount of the remaining range
	// for complex types, depending on the complexity value.  Increasing the
	// complexity coefficient coefficient increases the likelihood that the
	// resulting type is not a basic type (approximately; there's some
	// inexactness due to modular arithmetic, since we choose basic types for low
	// remainders).
	//
	// Caution: at some level of complexity types definitions will expand
	// forever, depending on the distribution of field/method arity below.
	const complexity = 3

	rem := (1 << 8) - len(Typ) // remaining values to allocate to complex types
	nComplex := int(instanceType)
	if complexity > rem/(nComplex+1) { // modulus will exceed uint8
		panic(fmt.Sprintf("invalid complexity factor %d", complexity))
	}

	modulus := len(Typ) + complexity*int(instanceType+1)
	v := r.uint8() % uint8(modulus)
	if v < uint8(len(Typ)) {
		return Typ[v]
	}

	tag := (int(v) - len(Typ)) / complexity
	switch tag {
	case definedType:
		return r.definedType()
	case pointerType:
		return NewPointer(r.typ())
	case sliceType:
		return NewSlice(r.typ())
	case arrayType:
		return NewArray(r.typ(), int64(r.uint8()))
	case chanType:
		dir := r.choice(SendRecv, SendOnly, RecvOnly).(ChanDir)
		return NewChan(dir, r.typ())
	case mapType:
		return NewMap(r.typ(), r.typ())
	case signatureType:
		return r.signature()
	case structType:
		fields := some(r.inRange(0, 2), r.var_, (*Var).Name)
		return NewStruct(fields, nil)
	case interfaceType:
		methods := some(r.inRange(0, 2), r.func_, (*Func).Name)
		// embeddeds := some(r.inRange(0, 2), r.typ, nil)
		embeddeds := some(r.inRange(0, 2), r.union_, nil)
		t := NewInterfaceType(methods, embeddeds)
		t.Complete()
		return t
	case instanceType:
		var orig *Named
		for {
			orig = r.definedType()
			if orig.TypeParams().Len() > 0 {
				break
			}
		}
		var targs []Type
		for i := 0; i < orig.TypeParams().Len(); i++ {
			targs = append(targs, r.typ())
		}
		inst, err := Instantiate(nil, orig, targs, false)
		if err != nil {
			panic(err)
		}
		return inst

	default:
		panic(fmt.Sprintf("unknown tag %d", tag))
	}
}

// min <= result < max
func (r *typeReader) inRange(min, max uint8) uint8 {
	remainder := r.uint8() % (max - min)
	return min + remainder
}

func (r *typeReader) choice(values ...interface{}) interface{} {
	return values[r.inRange(0, uint8(len(values)))]
}

// Returns a non-instance named type.
func (r *typeReader) definedType() *Named {
	var name string
	for {
		name = r.string()
		fmt.Println("name:", name)
		obj := r.scope.Lookup(name)
		if obj == nil {
			break // we found a new name
		}
		if existing, ok := obj.Type().(*Named); ok {
			return existing
		}
		// Existing name, but not a defined type. Keep looking.
	}
	obj := NewTypeName(token.NoPos, nil, name, nil)
	named := NewNamed(obj, nil, nil)
	r.declare(obj)
	r.pushScope()
	defer r.popScope()
	tparams := r.tparamList()
	// TODO: need a declare scope
	under := r.typ().Underlying()
	named.SetTypeParams(tparams)
	named.SetUnderlying(under)
	return named
}

func (r *typeReader) signature() *Signature {
	// tparams := r.tparamList()
	params := r.tuple()
	results := r.tuple()
	variadic := false
	if params.Len() > 0 {
		if _, ok := params.At(params.Len() - 1).Type().(*Slice); ok {
			variadic = r.bool()
		}
	}
	// return NewSignatureType(nil, nil, tparams, params, results, variadic)
	return NewSignatureType(nil, nil, nil, params, results, variadic)
}

func (r *typeReader) tuple() *Tuple {
	vars := some(r.inRange(0, 3), r.var_, (*Var).Name)
	return NewTuple(vars...)
}

func (r *typeReader) union_() Type {
	typs := some(r.inRange(1, 3), r.typ, nil)
	var terms []*Term
	for i := 0; i < len(typs); i++ {
		tilde := r.bool()
		terms = append(terms, NewTerm(tilde, typs[i]))
	}
	return NewUnion(terms)
}

func (r *typeReader) tparamList() []*TypeParam {
	if r.depth != 1 {
		// TODO: document
		return nil
	}
	seen := make(map[string]bool)
	want := r.inRange(0, 3)
	var res []*TypeParam
	for len(res) < int(want) {
		name := r.string()
		if !seen[name] {
			seen[name] = true
			tn := NewTypeName(token.NoPos, nil, name, nil)
			tparam := NewTypeParam(tn, nil)
			r.declare(tn)
			res = append(res, tparam)
		}
	}
	for _, tparam := range res {
		tparam.SetConstraint(r.typ())
	}
	return res
}

func (r *typeReader) declare(obj Object) {
	fmt.Println("declaring", obj.Name())
	r.scope.Insert(obj)
}

func (r *typeReader) var_() *Var {
	// TODO: this code is imprecise about scopes and duplicates
	return NewVar(token.NoPos, nil, r.string(), r.typ())
}

func (r *typeReader) func_() *Func {
	return NewFunc(token.NoPos, nil, r.string(), r.signature())
}

func some[T any](n uint8, next func() T, uniq func(T) string) []T {
	seen := make(map[string]bool)
	var res []T
	for len(res) < int(n) {
		x := next()
		if uniq != nil {
			name := uniq(x)
			if seen[name] {
				continue
			}
			seen[name] = true
		}
		res = append(res, x)
	}
	return res
}
