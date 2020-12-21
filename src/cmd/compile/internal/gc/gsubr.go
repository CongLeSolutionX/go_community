// Derived from Inferno utils/6c/txt.c
// https://bitbucket.org/inferno-os/inferno-os/src/master/utils/6c/txt.c
//
//	Copyright © 1994-1999 Lucent Technologies Inc.  All rights reserved.
//	Portions Copyright © 1995-1997 C H Forsyth (forsyth@terzarima.net)
//	Portions Copyright © 1997-1999 Vita Nuova Limited
//	Portions Copyright © 2000-2007 Vita Nuova Holdings Limited (www.vitanuova.com)
//	Portions Copyright © 2004,2006 Bruce Ellis
//	Portions Copyright © 2005-2007 C H Forsyth (forsyth@terzarima.net)
//	Revisions Copyright © 2000-2007 Lucent Technologies Inc. and others
//	Portions Copyright © 2009 The Go Authors. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package gc

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"

	"cmd/internal/obj"
	"cmd/internal/objabi"
)

// initLSym defines f's obj.LSym and initializes it based on the
// properties of f. This includes setting the symbol flags and ABI and
// creating and initializing related DWARF symbols.
//
// initLSym must be called exactly once per function and must be
// called for both functions with bodies and functions without bodies.
func initLSym(f *ir.Func, hasBody bool) {
	if f.LSym != nil {
		base.Fatalf("Func.initLSym called twice")
	}

	if nam := f.Nname; !ir.IsBlank(nam) {
		f.LSym = nam.Sym().Linksym()
		if f.Pragma&ir.Systemstack != 0 {
			f.LSym.Set(obj.AttrCFunc, true)
		}

		var aliasABI obj.ABI
		needABIAlias := false
		defABI, hasDefABI := symabiDefs[f.LSym.Name]
		if hasDefABI && defABI == obj.ABI0 {
			// Symbol is defined as ABI0. Create an
			// Internal -> ABI0 wrapper.
			f.LSym.SetABI(obj.ABI0)
			needABIAlias, aliasABI = true, obj.ABIInternal
		} else {
			// No ABI override. Check that the symbol is
			// using the expected ABI.
			want := obj.ABIInternal
			if f.LSym.ABI() != want {
				base.Fatalf("function symbol %s has the wrong ABI %v, expected %v", f.LSym.Name, f.LSym.ABI(), want)
			}
		}

		isLinknameExported := nam.Sym().Linkname != "" && (hasBody || hasDefABI)
		if abi, ok := symabiRefs[f.LSym.Name]; (ok && abi == obj.ABI0) || isLinknameExported {
			// Either 1) this symbol is definitely
			// referenced as ABI0 from this package; or 2)
			// this symbol is defined in this package but
			// given a linkname, indicating that it may be
			// referenced from another package. Create an
			// ABI0 -> Internal wrapper so it can be
			// called as ABI0. In case 2, it's important
			// that we know it's defined in this package
			// since other packages may "pull" symbols
			// using linkname and we don't want to create
			// duplicate ABI wrappers.
			if f.LSym.ABI() != obj.ABI0 {
				needABIAlias, aliasABI = true, obj.ABI0
			}
		}

		if needABIAlias {
			// These LSyms have the same name as the
			// native function, so we create them directly
			// rather than looking them up. The uniqueness
			// of f.lsym ensures uniqueness of asym.
			asym := &obj.LSym{
				Name: f.LSym.Name,
				Type: objabi.SABIALIAS,
				R:    []obj.Reloc{{Sym: f.LSym}}, // 0 size, so "informational"
			}
			asym.SetABI(aliasABI)
			asym.Set(obj.AttrDuplicateOK, true)
			base.Ctxt.ABIAliases = append(base.Ctxt.ABIAliases, asym)
		}
	}

	if !hasBody {
		// For body-less functions, we only create the LSym.
		return
	}

	var flag int
	if f.Dupok() {
		flag |= obj.DUPOK
	}
	if f.Wrapper() {
		flag |= obj.WRAPPER
	}
	if f.Needctxt() {
		flag |= obj.NEEDCTXT
	}
	if f.Pragma&ir.Nosplit != 0 {
		flag |= obj.NOSPLIT
	}
	if f.ReflectMethod() {
		flag |= obj.REFLECTMETHOD
	}

	// Clumsy but important.
	// See test/recover.go for test cases and src/reflect/value.go
	// for the actual functions being considered.
	if base.Ctxt.Pkgpath == "reflect" {
		switch f.Sym().Name {
		case "callReflect", "callMethod":
			flag |= obj.WRAPPER
		}
	}

	base.Ctxt.InitTextSym(f.LSym, flag)
}
