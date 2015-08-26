// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asm

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cmd/asm/internal/lex"
	"cmd/internal/obj"
)

// redirectStderr runs a function while stderr is redirected to a normal file
func redirectStderr(t *testing.T, output string, f func()) {
	var err error
	save := os.Stderr
	os.Stderr, err = os.Create(output)
	if err != nil {
		t.Fatalf("Cannot create output file")
	}
	defer func() {
		os.Stderr.Close()
		os.Stderr = save
	}()
	f()
}

func tryToParse(t *testing.T, goarch string, data []byte) ([]byte, error) {
	dir, err := ioutil.TempDir("", "asmtest")
	if err != nil {
		t.Fatalf("Cannot create temporary directory")
	}
	defer func() {
		os.RemoveAll(dir)
	}()

	input := filepath.Join(dir, goarch+".s")
	if err := ioutil.WriteFile(input, data, 0600); err != nil {
		t.Fatalf("Cannot create input file")
	}

	lex.InitHist()
	architecture, ctxt := setArch(goarch)
	lexer := lex.NewLexer(input, ctxt)
	parser := NewParser(ctxt, architecture, lexer)
	ctxt.Bso = obj.Binitw(ioutil.Discard)
	ctxt.Diag = log.Fatalf
	obj.Binitw(ioutil.Discard)

	output := filepath.Join(dir, goarch+".out")
	redirectStderr(t, output, func() {
		if _, ok := parser.Parse(); ok {
			t.Errorf("Expected some errors")
		}
	})
	return ioutil.ReadFile(output)
}

func TestErroneous(t *testing.T) {

	// Beware: asm only returns the 10 first errors
	input := []byte(`	TEXT
	TEXT%
	TEXT 1,1
	TEXT $"toto", 0, $1
	FUNCDATA
	DATA 0
	DATA(0),1
	FUNCDATA(SB
	GLOBL 0, 1
	PCDATA 1
`)

	expected := []string{
		"1: expect two or three operands for TEXT",
		"2: expect two or three operands for TEXT",
		"3: TEXT symbol \"<erroneous symbol>\" must be a symbol(SB)",
		"4: TEXT symbol \"<erroneous symbol>\" must be a symbol(SB)",
		"5: expect two operands for FUNCDATA",
		"6: expect two operands for DATA",
		"7: expect /size for DATA argument",
		"8: expect two operands for FUNCDATA",
		"9: GLOBL symbol \"<erroneous symbol>\" must be a symbol(SB)",
		"10: expect two operands for PCDATA",
	}

	// Note these errors should be independent of the architecture.
	// Just run the test with amd64.
	output, err := tryToParse(t, "amd64", input)
	if err != nil {
		t.Fatalf("Cannot read output file")
	}

	s := bufio.NewScanner(bytes.NewReader(output))
	for i := 0; s.Scan(); i++ {
		line := strings.SplitN(s.Text(), ":", 2)[1]
		if expected[i] != line {
			t.Errorf("Unexpected error %q; expected %q", line, expected[i])
		}
	}
	if err := s.Err(); err != nil {
		t.Errorf("Cannot parse output file")
	}
}
