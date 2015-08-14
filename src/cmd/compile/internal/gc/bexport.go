// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Binary package export.
// Based loosely on x/tools/go/importer.
// (see fmt.go, go.y as "documentation" for how to use/setup data structures)
//
// Use "-newexport" flag to enable.

// TODO(gri):
// - handling of escape information with functions
// - inlined functions

// Basic export format:
//
// The export data starts with a header string and a version number followed by
// two sections:
// - The first one is compiler-independent and contains all and only the information
// required to type-check against the imported package (e.g. inlined functions and
// escape information are not present). The intent is that the format of this section
// is the same across compilers.
// - The second section contains compiler-specific information in a compiler-specific
// format. It is typically only readable by the compiler that produced the section.
//
// The compiler-idependent section starts with the exported package object
// followed by the list of exported "objects": constants, variables, types, or functions.
//
// Each of the exported objects encodes its respective fields sequentially. If the
// field is a type, it encodes the respective type "inline", if it is a value (constant)
// it encodes the value inline. All other fields are reduced to strings or integers;
// also encoded inline.
//
// The encoding of each object starts with a "tag" - the tag defines what kind
// of object is encoded. It implicitly defines the order of the object's fields so
// that the importer can read the object back. In particular, the fields don't need
// to be tagged.
//
// The same packages and types may be referred to from multiple places or even
// recursively. Ordinary tags are (small) values < 0. A "tag" value >= 0 is used
// to mean the index (= number) of a package or tag that has been encoded before
// (the context always makes clear what kind of object is meant). When an index
// is decoded, the object (package or type) can simply be read out from the tables
// of previously encoded packages and types.
//
// Instead of using special tags for predeclared types (bool, int, string, etc.)
// the same mechanism is used: The map of already encoded types is prepopulated
// with the predeclared types. Whenever one of these types is seen, it is found in
// that map and the corresponding type index is used to encode it.
//
// With the exception of the top-level list of objects which is terminated with
// a special endTag, "lists of things" start with the number of list elements,
// followed by the elements. This permits the importer to allocate the right
// amount of memory for the list. This applies also to strings, which are encoded
// as their length followed by the individual bytes.
//
// All integer values use a variable-length encoding for compact representation.
//
// If debugFormat is set, each integer and string value is preceeded by a marker
// and position information in the encoding. This mechanism permits an importer
// to recognize immediately when it is out of sync. The importer recognizes this
// mode automatically (i.e., it can import export data produced with debugging
// support even if debugFormat is not set at the time of import). Using this mode
// will massively increase the size of the export data (by a factor of 2 to 3)
// and is only recommended for debugging.
//
// The exporter and importer are completely symmetric in implementation: For
// each encoding routine there is the matching and symmetric decoding routine.
// This symmtry makes it very easy to change or extend the format: If a new
// field needs to be encoded, a symmetric change can be made to exporter and
// importer.

package gc

import (
	"bytes"
	"cmd/compile/internal/big"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
)

// debugging support
const (
	debugFormat = false // use debugging format for export data (emits a lot of additional data)
	traceExport = false // enable printing of export data in human-readable form
)

const exportVersion = "v0"

// Export writes the export data for localpkg to out and returns the number of bytes written.
func Export(out io.Writer) int {
	var format byte = 'c' // compact
	if debugFormat {
		format = 'd'
	}
	p := exporter{
		out:      out,
		buf:      []byte{format},
		pkgIndex: make(map[*Pkg]int),
		symIndex: make(map[*Sym]int),
		typIndex: make(map[*Type]int),
		written:  1, // format byte
	}

	// --- generic export data ---

	if traceExport {
		p.tracef("\n--- generic export data ---\n")
		if p.indent != 0 {
			Fatalf("incorrect indentation %d", p.indent)
		}
	}

	p.string(exportVersion)
	if traceExport {
		p.tracef("\n")
	}

	// collect only inline-specific objects via exportlist
	// (elements are added via reexportdeplist in collectInlined)
	exportlist = nil

	// populate type map with predeclared "known" types
	var last int
	for index, typ := range predeclared() {
		p.typIndex[typ] = index
		last = index
	}
	if last+1 != len(p.typIndex) {
		Fatalf("duplicate entries in type map?")
	}

	// write package data
	if localpkg.Path != "" {
		Fatalf("local package path not empty: %q", localpkg.Path)
	}
	p.pkg(localpkg)

	// write compiler-specific flags
	{
		var flags string
		if safemode != 0 {
			flags = "safe"
		}
		p.string(flags)
	}

	if traceExport {
		p.tracef("\n")
	}

	// collect objects to export
	var list []*Sym
	for _, sym := range localpkg.Syms {
		if sym.Flags&SymExport != 0 {
			list = append(list, sym)
		}
	}
	sort.Sort(byName(list)) // for reproducible output

	// write objects
	// (We use an endTag instead of an object count because named types that
	// were exported already as part of another type don't need to be re-exported
	// and thus can be ignored. Thus, we don't know the exact number of objects
	// a priori.)
	for _, sym := range list {
		p.obj(sym)
	}
	p.tag(endTag)

	if traceExport {
		p.tracef("\n")
	}

	// --- compiler-specific export data ---

	// TODO(gri) complete this

	if traceExport && (len(exportlist) > 0 || len(p.symIndex) > 0) {
		p.tracef("\n--- compiler specific export data ---\n")
		if p.indent != 0 {
			Fatalf("incorrect indentation")
		}
	}

	// export object required for inlined functions/methods
	/*
		for len(exportlist) > 0 {
			n := exportlist[0]
			exportlist = exportlist[1:]
			p.obj(n.Sym)
		}

		// export inlined function bodies
		for sym, index := range p.symIndex {
			// TODO(gri) complete this - for now we just trace the names
			p.tracef("%04d  %s\n", index, sym.Name)
		}
	*/
	p.tag(endTag)

	if traceExport {
		p.tracef("\n")
	}

	// --- end of export data ---

	p.flush()
	return p.written
}

type byName []*Sym

func (a byName) Len() int           { return len(a) }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type exporter struct {
	out      io.Writer
	buf      []byte
	pkgIndex map[*Pkg]int
	symIndex map[*Sym]int
	typIndex map[*Type]int

	written int // bytes written
	indent  int // for traceExport
}

func (p *exporter) pkg(pkg *Pkg) {
	if pkg == nil {
		Fatalf("unexpected nil pkg")
	}

	// if we saw the package before, write its index (>= 0)
	if i, ok := p.pkgIndex[pkg]; ok {
		p.index('P', i)
		return
	}

	// otherwise, remember the package, write the package tag (< 0) and package data
	if traceExport {
		p.tracef("P%d = { ", len(p.pkgIndex))
		defer p.tracef("} ")
	}
	p.pkgIndex[pkg] = len(p.pkgIndex)

	p.tag(packageTag)
	p.string(pkg.Name)
	p.string(pkg.Path)
}

func (p *exporter) obj(sym *Sym) {
	if sym.Def == nil {
		Fatalf("unknown export symbol: %v", sym)
	}

	// Export format:
	// - tag
	// - name (except for types where it is written when emitting the type)
	// - type
	// - value (for constants only)

	switch sym.Def.Op {
	case OLITERAL:
		// constant
		n := sym.Def
		typecheck(&n, Erv)
		if n == nil || n.Op != OLITERAL {
			Fatalf("dumpexportconst: oconst nil: %v", sym)
		}

		typ := n.Type // may or may not be specified
		if typ == nil || isideal(typ) {
			typ = untype(n.Val().Ctype())
		}

		p.tag(constTag)
		p.string(sym.Name)
		p.typ(typ)
		p.value(n.Val())

	case OTYPE:
		// named type
		t := sym.Def.Type
		if t.Etype == TFORW {
			Fatalf("export of incomplete type %v", sym)
		}

		// no need to write type if it was encountered already as part of some other object
		if _, ok := p.typIndex[t]; ok {
			return
		}

		p.tag(typeTag)
		p.typ(t) // name is written by corresponding named type

	case ONAME:
		// variable or function
		n := sym.Def
		typecheck(&n, Erv|Ecall)
		if n == nil || n.Type == nil {
			Fatalf("variable/function exported but not defined: %v", sym)
		}

		tag := varTag
		if n.Type.Etype == TFUNC && n.Class == PFUNC {
			// function
			tag = funcTag
			p.collectInlined(sym, n)
		}

		p.tag(tag)
		p.string(sym.Name)
		// The type can only be a signature for functions. However, by always
		// writing the complete type specification (rather than just a signature)
		// we keep the option open of sharing common signatures across multiple
		// functions as a means to further compress the export data.
		p.typ(n.Type)

	default:
		Fatalf("unexpected export symbol: %v %v", Oconv(int(sym.Def.Op), 0), sym)
	}

	if traceExport {
		p.tracef("\n")
	}
}

func (p *exporter) collectInlined(sym *Sym, n *Node) {
	if n != nil && n.Func != nil && n.Func.Inl != nil {
		// when lazily typechecking inlined bodies, some re-exported ones may not have been typechecked yet.
		// currently that can leave unresolved ONONAMEs in import-dot-ed packages in the wrong package
		if Debug['l'] < 2 {
			typecheckinl(n)
		}
		// TODO(gri) this is incorrect and won't work...
		p.symIndex[sym] = len(p.symIndex)
		// TODO(gri) collect n (can it be found from sym?)
		reexportdeplist(n.Func.Inl)
	}
}

func (p *exporter) typ(t *Type) {
	if t == nil {
		Fatalf("nil type")
	}

	// Possible optimization: Anonymous pointer types *T where
	// T is a named type are common. We could canonicalize all
	// such types *T to a single type PT = *T. This would lead
	// to at most one *T entry in typIndex, and all future *T's
	// would be encoded as the respective index directly. Would
	// save 1 byte (pointerTag) per *T and reduce the typIndex
	// size (at the cost of a canonicalization map). We can do
	// this later, without encoding format change.

	// if we saw the type before, write its index (>= 0)
	if i, ok := p.typIndex[t]; ok {
		p.index('T', i)
		return
	}

	// otherwise, remember the type, write the type tag (< 0) and type data
	if traceExport {
		p.tracef("T%d = {>\n", len(p.typIndex))
		defer p.tracef("<\n} ")
	}
	p.typIndex[t] = len(p.typIndex)

	// pick off named types
	if sym := t.Sym; sym != nil {
		// a few explicit assertions to be on the safe side for now:
		// fields have symbols too, but they are printed elsewhere
		if t.Etype == TFIELD {
			Fatalf("printing a field/parameter with wrong function")
		}
		// if it was a predeclared type, it was found in the type map
		if t.Orig == t {
			Fatalf("predeclared type missing from type map?")
		}
		// we expect the respective definition to point to us
		if sym.Def.Type != t {
			Fatalf("type definition doesn't point to us?")
		}

		p.tag(namedTag)
		p.qualifiedName(sym)

		// write underlying type
		p.typ(t.Orig)

		// interfaces don't have associated methods
		if t.Orig.Etype == TINTER {
			return
		}

		// write associated methods
		n := 0 // method count
		for m := t.Method; m != nil; m = m.Down {
			n++
		}
		p.int(n)

		if traceExport && t.Method != nil {
			p.tracef("associated methods {>\n")
		}

		for m := t.Method; m != nil; m = m.Down {
			p.string(m.Sym.Name)
			p.paramList(getthisx(m.Type))
			p.paramList(getinargx(m.Type))
			p.paramList(getoutargx(m.Type))

			p.collectInlined(m.Sym, m.Type.Nname)

			if traceExport && m.Down != nil {
				p.tracef("\n")
			}
		}

		if traceExport && t.Method != nil {
			p.tracef("<\n} ")
		}

		return
	}

	// otherwise we have a type literal
	switch t.Etype {
	case TARRAY:
		// TODO(gri) define named constant for the -100
		if t.Bound >= 0 || t.Bound == -100 {
			p.tag(arrayTag)
			p.int64(t.Bound)
		} else {
			p.tag(sliceTag)
		}
		p.typ(t.Type)

	case T_old_DARRAY:
		p.tag(dddTag)
		p.typ(t.Type)

	case TSTRUCT:
		p.tag(structTag)
		p.fieldList(t)

	case TPTR32, TPTR64: // could use Tptr but these are constants
		p.tag(pointerTag)
		p.typ(t.Type)

	case TFUNC:
		p.tag(signatureTag)
		p.paramList(getinargx(t))
		p.paramList(getoutargx(t))

	case TINTER:
		p.tag(interfaceTag)

		// gc doesn't separate between embedded interfaces
		// and methods declared explicitly with an interface
		p.int(0) // no embedded interfaces
		p.methodList(t)

	case TMAP:
		p.tag(mapTag)
		p.typ(t.Down) // key
		p.typ(t.Type) // val

	case TCHAN:
		p.tag(chanTag)
		p.int(int(t.Chan))
		p.typ(t.Type)

	default:
		Fatalf("unexpected type: %s (Etype = %d)", Tconv(t, 0), t.Etype)
	}
}

func (p *exporter) fieldList(t *Type) {
	if traceExport && t.Type != nil {
		p.tracef("fields {>\n")
		defer p.tracef("<\n} ")
	}

	p.int(countfield(t))
	for f := t.Type; f != nil; f = f.Down {
		p.field(f)
		if traceExport && f.Down != nil {
			p.tracef("\n")
		}
	}
}

func (p *exporter) field(f *Type) {
	if f.Etype != TFIELD {
		Fatalf("field expected")
	}

	note := ""
	if f.Note != nil {
		note = *f.Note
	}

	p.fieldName(f.Sym, f.Embedded != 0)
	p.typ(f.Type)
	p.string(note)
}

func (p *exporter) methodList(t *Type) {
	if traceExport && t.Type != nil {
		p.tracef("methods {>\n")
		defer p.tracef("<\n} ")
	}

	p.int(countfield(t))
	for m := t.Type; m != nil; m = m.Down {
		p.method(m)
		if traceExport && m.Down != nil {
			p.tracef("\n")
		}
	}
}

func (p *exporter) method(m *Type) {
	if m.Etype != TFIELD {
		Fatalf("method expected")
	}

	p.fieldName(m.Sym, false)
	p.paramList(getinargx(m.Type))
	p.paramList(getoutargx(m.Type))
}

func (p *exporter) qualifiedName(sym *Sym) {
	p.string(sym.Name)
	p.pkg(sym.Pkg)
}

// fieldName is like qualifiedName but it doesn't record
// the package for anonyous or exported names.
func (p *exporter) fieldName(sym *Sym, anonymous bool) {
	name := sym.Name
	if anonymous {
		name = ""
	}
	p.string(name)
	if name != "" && !exportname(name) {
		p.pkg(sym.Pkg)
	}
}

func (p *exporter) paramList(params *Type) {
	if params.Etype != TSTRUCT || params.Funarg == 0 {
		Fatalf("parameter list expected")
	}

	// use negative length to indicate unnamed parameters
	// (look at the first parameter only since either all
	// names are present or all are absent)
	n := countfield(params)
	if n > 0 && parName(params.Type) == "" {
		n = -n
	}
	p.int(n)
	for q := params.Type; q != nil; q = q.Down {
		p.param(q, n)
	}
}

func (p *exporter) param(q *Type, n int) {
	if q.Etype != TFIELD {
		Fatalf("parameter expected")
	}
	t := q.Type
	if q.Isddd {
		// create a fake type to encode ... just for the p.typ call
		t = &Type{Etype: T_old_DARRAY, Type: t.Type}
	}
	p.typ(t)
	if n > 0 {
		p.string(parName(q))
	}
}

func parName(q *Type) string {
	if q.Sym == nil {
		return ""
	}
	name := q.Sym.Name
	if len(name) > 0 && name[0] == '~' {
		// name is ~b%d or ~r%d
		switch name[1] {
		case 'b':
			return "_"
		case 'r':
			return ""
		default:
			Fatalf("unexpected parameter name: %s", name)
		}
	}
	if i := strings.Index(name, "·"); i > 0 {
		name = name[:i] // cut off numbering
	}
	return name
}

func (p *exporter) value(x Val) {
	if traceExport {
		p.tracef("= ")
	}

	switch x := x.U.(type) {
	case bool:
		tag := falseTag
		if x {
			tag = trueTag
		}
		p.tag(tag)

	case *Mpint:
		if Mpcmpfixfix(Minintval[TINT64], x) <= 0 && Mpcmpfixfix(x, Maxintval[TINT64]) <= 0 {
			// common case: x fits into an int64 - use compact encoding
			p.tag(int64Tag)
			p.int64(Mpgetfix(x))
			return
		}
		// uncommon case: large x - use float encoding
		// (powers of 2 will be encoded efficiently with exponent)
		p.tag(floatTag)
		f := newMpflt()
		Mpmovefixflt(f, x)
		p.float(f)

	case *Mpflt:
		p.tag(floatTag)
		p.float(x)

	case *Mpcplx:
		p.tag(complexTag)
		p.float(&x.Real)
		p.float(&x.Imag)

	case string:
		p.tag(stringTag)
		p.string(x)

	default:
		Fatalf("unexpected value %v (%T)", x, x)
	}
}

func (p *exporter) float(x *Mpflt) {
	// extract sign, treat -0 as < 0
	f := &x.Val
	sign := f.Sign()
	if sign == 0 {
		// ±0
		if f.Signbit() {
			// -0: uncommon
			// represented with sign, empty (== 0) mantissa, and 0 exponent
			p.int(-1)
			p.string("")
		}
		p.int(0)
		return
	}
	// x != 0

	// extract exponent (for mantissa such that 0.5 <= mant < 1.0)
	var m big.Float
	exp := f.MantExp(&m)

	// extract mantissa as *big.Int
	// - set exponent large enough so mant satisfies mant.IsInt()
	// - get *big.Int from mant
	m.SetMantExp(&m, int(m.MinPrec()))
	mant, acc := m.Int(nil)
	if acc != big.Exact {
		Fatalf("internal error")
	}

	p.int(sign)
	p.string(string(mant.Bytes()))
	p.int(exp)
}

// ----------------------------------------------------------------------------
// Low-level encoders

func (p *exporter) index(marker byte, index int) {
	if index < 0 {
		Fatalf("invalid index < 0")
	}
	if debugFormat {
		p.marker('t')
	}
	if traceExport {
		p.tracef("%c%d ", marker, index)
	}
	p.rawInt64(int64(index))
}

func (p *exporter) tag(tag int) {
	if tag >= 0 {
		Fatalf("invalid tag >= 0")
	}
	if debugFormat {
		p.marker('t')
	}
	if traceExport {
		p.tracef("%s ", tagString[-tag])
	}
	p.rawInt64(int64(tag))

	// don't let the buffer get too large
	if len(p.buf) > 4<<10 {
		p.flush()
	}
}

func (p *exporter) int(x int) {
	p.int64(int64(x))
}

func (p *exporter) int64(x int64) {
	if debugFormat {
		p.marker('i')
	}
	if traceExport {
		p.tracef("%d ", x)
	}
	p.rawInt64(x)
}

func (p *exporter) string(s string) {
	if debugFormat {
		p.marker('s')
	}
	if traceExport {
		p.tracef("%q ", s)
	}
	p.rawInt64(int64(len(s)))
	p.buf = append(p.buf, s...)
	p.written += len(s)
}

// marker emits a marker byte and position information which makes
// it easy for a reader to detect if it is "out of sync". Used for
// debugFormat format only.
func (p *exporter) marker(m byte) {
	p.buf = append(p.buf, m)
	p.written++
	p.rawInt64(int64(p.written))
}

// rawInt64 should only be used by low-level encoders
func (p *exporter) rawInt64(x int64) {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutVarint(tmp[:], x)
	p.buf = append(p.buf, tmp[:n]...)
	p.written += n
}

func (p *exporter) flush() {
	w, err := p.out.Write(p.buf)
	if err != nil || w != len(p.buf) {
		Fatalf("error while writing export data")
	}
	p.buf = p.buf[:0]
}

// tracef is like fmt.Printf but it rewrites the format string
// to take care of indentation.
func (p *exporter) tracef(format string, args ...interface{}) {
	if strings.IndexAny(format, "<>\n") >= 0 {
		var buf bytes.Buffer
		for i := 0; i < len(format); i++ {
			// no need to deal with runes
			ch := format[i]
			switch ch {
			case '>':
				p.indent++
				continue
			case '<':
				p.indent--
				continue
			}
			buf.WriteByte(ch)
			if ch == '\n' {
				for j := p.indent; j > 0; j-- {
					buf.WriteString(".  ")
				}
			}
		}
		format = buf.String()
	}
	fmt.Printf(format, args...)
}

// ----------------------------------------------------------------------------
// Export format

// Tags. Must be < 0.
const (
	endTag = -(iota + 1)

	// Packages
	packageTag

	// Objects
	constTag
	typeTag
	varTag
	funcTag

	// Types
	namedTag
	arrayTag
	sliceTag
	dddTag
	structTag
	pointerTag
	signatureTag
	interfaceTag
	mapTag
	chanTag

	// Values
	falseTag
	trueTag
	int64Tag
	floatTag
	fractionTag // not used by gc
	complexTag
	stringTag
)

// Debugging support.
// (tagString is only used when tracing is enabled)
var tagString = [...]string{
	-endTag: "end",

	// Packages:
	-packageTag: "package",

	// Objects:
	-constTag: "const",
	-typeTag:  "type",
	-varTag:   "var",
	-funcTag:  "func",

	// Types:
	-namedTag:     "named type",
	-arrayTag:     "array",
	-sliceTag:     "slice",
	-dddTag:       "ddd",
	-structTag:    "struct",
	-pointerTag:   "pointer",
	-signatureTag: "signature",
	-interfaceTag: "interface",
	-mapTag:       "map",
	-chanTag:      "chan",

	// Values:
	-falseTag:    "false",
	-trueTag:     "true",
	-int64Tag:    "int64",
	-floatTag:    "float",
	-fractionTag: "fraction",
	-complexTag:  "complex",
	-stringTag:   "string",
}

// untype returns the "pseudo" untyped type for a Ctype (import/export use only).
// (we can't use an pre-initialized array because we must be sure all types are
// set up)
func untype(ctype int) *Type {
	switch ctype {
	case CTINT:
		return idealint
	case CTRUNE:
		return idealrune
	case CTFLT:
		return idealfloat
	case CTCPLX:
		return idealcomplex
	case CTSTR:
		return idealstring
	case CTBOOL:
		return idealbool
	case CTNIL:
		return Types[TNIL]
	}
	Fatalf("unknown Ctype")
	panic("unreachable")
}

var (
	idealint     = typ(TIDEAL)
	idealrune    = typ(TIDEAL)
	idealfloat   = typ(TIDEAL)
	idealcomplex = typ(TIDEAL)
)

var predecl []*Type // initialized lazily

func predeclared() []*Type {
	if predecl == nil {
		// initialize lazily to be sure that all
		// elements have been initialized before
		predecl = []*Type{
			// basic types
			Types[TBOOL],
			Types[TINT],
			Types[TINT8],
			Types[TINT16],
			Types[TINT32],
			Types[TINT64],
			Types[TUINT],
			Types[TUINT8],
			Types[TUINT16],
			Types[TUINT32],
			Types[TUINT64],
			Types[TUINTPTR],
			Types[TFLOAT32],
			Types[TFLOAT64],
			Types[TCOMPLEX64],
			Types[TCOMPLEX128],
			Types[TSTRING],

			// aliases
			bytetype,
			runetype,

			// error
			errortype,

			// untyped types
			untype(CTBOOL),
			untype(CTINT),
			untype(CTRUNE),
			untype(CTFLT),
			untype(CTCPLX),
			untype(CTSTR),
			untype(CTNIL),

			// package unsafe
			Types[TUNSAFEPTR],
		}
	}
	return predecl
}
