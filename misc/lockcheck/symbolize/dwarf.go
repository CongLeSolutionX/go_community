// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package symbolize resolves various types of offsets in a binary to
// user-friendly symbols.
package symbolize

import (
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
)

// DWARF symbolizes PCs and addresses using DWARF data.
type DWARF struct {
	dd    *dwarf.Data
	Vars  *VarTab
	Lines *LineTab
}

func NewDWARF(path string) (*DWARF, error) {
	dd, err := openDWARF(path)
	if err != nil {
		return nil, err
	}

	vars, err := NewVarTab(dd)
	if err != nil {
		return nil, err
	}

	lines, err := NewLineTab(dd)
	if err != nil {
		return nil, err
	}

	return &DWARF{dd, vars, lines}, nil
}

func openDWARF(path string) (*dwarf.Data, error) {
	elfF, err := elf.Open(path)
	if err == nil {
		return elfF.DWARF()
	} else if _, ok := err.(*elf.FormatError); !ok {
		return nil, err
	}

	machoF, err := macho.Open(path)
	if err == nil {
		return machoF.DWARF()
	} else if _, ok := err.(*macho.FormatError); !ok {
		return nil, err
	}

	peF, err := pe.Open(path)
	if err == nil {
		return peF.DWARF()
	} else if _, ok := err.(*pe.FormatError); !ok {
		return nil, err
	}

	return nil, fmt.Errorf("object %s not a recognized format", path)
}
