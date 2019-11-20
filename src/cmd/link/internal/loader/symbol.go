// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

import (
	"cmd/link/internal/sym"
)

func (l *Loader) Name(i Sym) string {
	if l.Syms[i] == nil {
		panic("no name")
	}
	return l.Syms[i].Name
}

func (l *Loader) Type(i Sym) sym.SymKind {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Type
}

func (l *Loader) SetType(i Sym, t sym.SymKind) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].Type = t
}

func (l *Loader) P(i Sym) []byte {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].P
}

func (l *Loader) SetP(i Sym, p []byte) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].P = p
}

func (l *Loader) R(i Sym) []sym.Reloc {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].R
}

func (l *Loader) SetR(i Sym, r []sym.Reloc) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].R = r
}

func (l *Loader) Size(i Sym) int64 {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Size
}

func (l *Loader) SetSize(i Sym, s int64) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].Size = s
}

func (l *Loader) SetDynimplib(i Sym, lib string) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].SetDynimplib(lib)
}

func (l *Loader) SetAttr(i Sym, a sym.Attribute, v bool) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].Attr.Set(a, v)
}

func (l *Loader) DuplicateOK(i Sym) bool {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Attr.DuplicateOK()
}

func (l *Loader) External(i Sym) bool {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Attr.External()
}

func (l *Loader) OnList(i Sym) bool {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Attr.OnList()
}

func (l *Loader) CgoExportDynamic(i Sym) bool {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Attr.CgoExportDynamic()
}

func (l *Loader) Outer(i Sym) Sym {
	s := l.Syms[i]
	if s == nil || s.Outer == nil {
		return 0
	}
	return Sym(l.Syms[i].Outer.Index)
}

func (l *Loader) SetOuter(i, o Sym) {
	is, os := l.Syms[i], l.Syms[o]
	if is == nil {
		panic("no sym")
	}
	is.Outer = os
}

func (l *Loader) Sub(i Sym) Sym {
	if l.Syms[i] == nil {
		return 0
	}
	if ss := l.Syms[i].Sub; ss != nil {
		return Sym(ss.Index)
	}
	return 0
}

func (l *Loader) SetSub(o, s Sym) {
	os, ss := l.Syms[o], l.Syms[s]
	if os == nil {
		panic("no sym")
	}
	os.Sub = ss
}

func (l *Loader) Value(i Sym) int64 {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	return l.Syms[i].Value
}

func (l *Loader) SetValue(i Sym, v int64) {
	if l.Syms[i] == nil {
		panic("no sym")
	}
	l.Syms[i].Value = v
}

func (l *Loader) SetRelocSym(r *sym.Reloc, i Sym) {
	r.Sym = l.Syms[i]
}

// SortSub sorts a linked-list (by Sub) of *Symbol by Value.
// Used for sub-symbols when loading host objects (see e.g. ldelf.go).
func (ld *Loader) SortSub(l Sym) Sym {
	if l == 0 || ld.Sub(l) == 0 {
		return l
	}

	l1 := l
	l2 := l
	for {
		l2 = ld.Sub(l2)
		if l2 == 0 {
			break
		}
		l2 = ld.Sub(l2)
		if l2 == 0 {
			break
		}
		l1 = ld.Sub(l1)
	}

	l2 = ld.Sub(l1)
	ld.SetSub(l1, 0)
	l1 = ld.SortSub(l)
	l2 = ld.SortSub(l2)

	/* set up lead element */
	if ld.Value(l1) < ld.Value(l2) {
		l = l1
		l1 = ld.Sub(l1)
	} else {
		l = l2
		l2 = ld.Sub(l2)
	}

	le := l

	for {
		if l1 == 0 {
			for l2 != 0 {
				ld.SetSub(le, l2)
				le = l2
				l2 = ld.Sub(l2)
			}

			ld.SetSub(le, 0)
			break
		}

		if l2 == 0 {
			for l1 != 0 {
				ld.SetSub(le, l1)
				le = l1
				l1 = ld.Sub(l1)
			}

			break
		}

		if ld.Value(l1) < ld.Value(l2) {
			ld.SetSub(le, l1)
			le = l1
			l1 = ld.Sub(l1)
		} else {
			ld.SetSub(le, l2)
			le = l2
			l2 = ld.Sub(l2)
		}
	}

	ld.SetSub(le, 0)
	return l
}
