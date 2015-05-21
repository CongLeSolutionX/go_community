// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cmd_go_bootstrap

// This is not built when bootstrapping to avoid having go_bootstrap depend on
// debug/elf.

package main

import (
	"bufio"
	"bytes"
	"debug/elf"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

func rnd(v int32, r int32) int32 {
	if r <= 0 {
		return v
	}
	v += r - 1
	c := v % r
	if c < 0 {
		c += r
	}
	v -= c
	return v
}

func readwithpad(r io.Reader, sz int32) ([]byte, error) {
	full := rnd(sz, 4)
	data := make([]byte, full)
	_, err := io.ReadFull(r, data)
	if err != nil {
		return nil, err
	}
	data = data[:sz]
	return data, nil
}

// readnote returns the ELF note with the given section and type from filename. The
// layout of note sections is explained on page 2-4 of "Tool Interface Standard (TIS)
// Executable and Linking Format (ELF) Specification Version 1.2" available at
// http://refspecs.linuxbase.org/elf/elf.pdf.
func readnote(filename, name string, typ int32) ([]byte, error) {
	f, err := elf.Open(filename)
	if err != nil {
		return nil, err
	}
	for _, sect := range f.Sections {
		if sect.Type != elf.SHT_NOTE {
			continue
		}
		r := sect.Open()
		for {
			var namesize, descsize, noteType int32
			err = binary.Read(r, f.ByteOrder, &namesize)
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, fmt.Errorf("read namesize failed:", err)
			}
			err = binary.Read(r, f.ByteOrder, &descsize)
			if err != nil {
				return nil, fmt.Errorf("read descsize failed:", err)
			}
			err = binary.Read(r, f.ByteOrder, &noteType)
			if err != nil {
				return nil, fmt.Errorf("read type failed:", err)
			}
			noteName, err := readwithpad(r, namesize)
			if err != nil {
				return nil, fmt.Errorf("read name failed:", err)
			}
			desc, err := readwithpad(r, descsize)
			if err != nil {
				return nil, fmt.Errorf("read desc failed:", err)
			}
			if name == string(noteName) && typ == noteType {
				return desc, nil
			}
		}
	}
	return nil, nil
}

// readpkglist returns the list of packages that were built into the shared library
// at shlibpath. For the native toolchain this list is stored, newline separated, in
// an ELF note with name "GO\x00\x00" and type 1. For GCCGO it is extracted from the
// .go_export section.
func readpkglist(shlibpath string) (pkgs []*Package) {
	var stk importStack
	if _, gccgo := buildToolchain.(gccgoToolchain); gccgo {
		f, _ := elf.Open(shlibpath)
		sect := f.Section(".go_export")
		data, _ := sect.Data()
		scanner := bufio.NewScanner(bytes.NewBuffer(data))
		for scanner.Scan() {
			t := scanner.Text()
			if strings.HasPrefix(t, "pkgpath ") {
				t = strings.TrimPrefix(t, "pkgpath ")
				t = strings.TrimSuffix(t, ";")
				pkgs = append(pkgs, loadPackage(t, &stk))
			}
		}
	} else {
		pkglistbytes, err := readnote(shlibpath, "GO\x00\x00", 1)
		if err != nil {
			fatalf("readnote failed: %v", err)
		}
		scanner := bufio.NewScanner(bytes.NewBuffer(pkglistbytes))
		for scanner.Scan() {
			t := scanner.Text()
			pkgs = append(pkgs, loadPackage(t, &stk))
		}
	}
	return
}
