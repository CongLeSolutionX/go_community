// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const (
	// Data directory, also the package directory for the test.
	testdata = "testdata"

	// Binaries we compile.
	testcover = "./testcover.exe"
)

var (
	// Files we use.
	testMain     = filepath.Join(testdata, "main.go")
	testTest     = filepath.Join(testdata, "test.go")
	coverInput   = filepath.Join(testdata, "test_line.go")
	coverOutput  = filepath.Join(testdata, "test_cover.go")
	coverProfile = filepath.Join(testdata, "profile.cov")
)

var debug = false // Keeps the rewritten files around if set.

// Run this shell script, but do it in Go so it can be run by "go test".
//
//	replace the word LINE with the line number < testdata/test.go > testdata/test_line.go
// 	go build -o ./testcover
// 	./testcover -mode=count -var=CoverTest -o ./testdata/test_cover.go testdata/test_line.go
//	go run ./testdata/main.go ./testdata/test.go
//
func TestCover(t *testing.T) {
	testenv.MustHaveGoBuild(t)

	// Read in the test file (testTest) and write it, with LINEs specified, to coverInput.
	file, err := ioutil.ReadFile(testTest)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(file, []byte("\n"))
	for i, line := range lines {
		lines[i] = bytes.Replace(line, []byte("LINE"), []byte(fmt.Sprint(i+1)), -1)
	}
	if err := ioutil.WriteFile(coverInput, bytes.Join(lines, []byte("\n")), 0666); err != nil {
		t.Fatal(err)
	}

	// defer removal of test_line.go
	if !debug {
		defer os.Remove(coverInput)
	}

	// go build -o testcover
	cmd := exec.Command(testenv.GoToolPath(t), "build", "-o", testcover)
	run(cmd, t)

	// defer removal of testcover
	defer os.Remove(testcover)

	// ./testcover -mode=count -var=thisNameMustBeVeryLongToCauseOverflowOfCounterIncrementStatementOntoNextLineForTest -o ./testdata/test_cover.go testdata/test_line.go
	cmd = exec.Command(testcover, "-mode=count", "-var=thisNameMustBeVeryLongToCauseOverflowOfCounterIncrementStatementOntoNextLineForTest", "-o", coverOutput, coverInput)
	run(cmd, t)

	// defer removal of ./testdata/test_cover.go
	if !debug {
		defer os.Remove(coverOutput)
	}

	// go run ./testdata/main.go ./testdata/test.go
	cmd = exec.Command(testenv.GoToolPath(t), "run", testMain, coverOutput)
	run(cmd, t)

	file, err = ioutil.ReadFile(coverOutput)
	if err != nil {
		t.Fatal(err)
	}
	// pragmas must appear right next to function declaration.
	if got, err := regexp.MatchString(".*\n//go:nosplit\nfunc someFunction().*", string(file)); err != nil || !got {
		t.Errorf("misplaced pragma: got=(%v, %v); want=(true; nil)", got, err)
	}
	// "go:linkname" pragma should be present.
	if got, err := regexp.MatchString(`.*go\:linkname some\_name some\_name.*`, string(file)); err != nil || !got {
		t.Errorf("'go:linkname' pragma not found: got=(%v, %v); want=(true; nil)", got, err)
	}

	// No other comments should be present in generated code.
	c := ".*// This comment shouldn't appear in generated go code.*"
	if got, err := regexp.MatchString(c, string(file)); err != nil || got {
		t.Errorf("non pragma comment %q found. got=(%v, %v); want=(false; nil)", c, got, err)
	}
}

var testPragmas = filepath.Join(testdata, "pragmas.go")

// TestPragmas checks that pragma comments are preserved and positioned
// correctly. Pragmas that occur before top-level declarations should remain
// above those declarations, even if they are not part of the block of
// documentation comments.
func TestPragmas(t *testing.T) {
	// Read the source file and find all the pragmas. We'll keep track of whether
	// each one has been seen in the output.
	source, err := ioutil.ReadFile(testPragmas)
	if err != nil {
		t.Fatal(err)
	}
	sourcePragmas := findPragmas(source)

	// go tool cover -mode=set ./testdata/pragmas.go
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "cover", "-mode=set", testPragmas)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	// Check that all pragmas are present in the output.
	outputPragmas := findPragmas(output)
	foundPragma := make(map[string]bool)
	for _, p := range sourcePragmas {
		foundPragma[p.name] = false
	}
	for _, p := range outputPragmas {
		if found, ok := foundPragma[p.name]; !ok {
			t.Errorf("unexpected pragma in output: %s", p.text)
		} else if found {
			t.Errorf("pragma found multiple times in output: %s", p.text)
		}
		foundPragma[p.name] = true
	}
	var missing []string
	for name, found := range foundPragma {
		if !found {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Errorf("the following pragmas were missing: %s", strings.Join(missing, ", "))
	}

	// Check that pragmas that start with the name of top-level declarations
	// come before the beginning of the named declaration and after the end
	// of the previous declaration.
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, testPragmas, output, 0)
	if err != nil {
		t.Fatal(err)
	}

	prevEnd := 0
	for _, decl := range astFile.Decls {
		var name string
		switch d := decl.(type) {
		case *ast.FuncDecl:
			name = d.Name.Name
		case *ast.GenDecl:
			if len(d.Specs) != 1 {
				continue
			}
			if spec, ok := d.Specs[0].(*ast.TypeSpec); ok {
				name = spec.Name.Name
			}
		}
		pos := fset.Position(decl.Pos()).Offset
		end := fset.Position(decl.End()).Offset
		if name == "" {
			prevEnd = end
			continue
		}
		for _, p := range outputPragmas {
			if !strings.HasPrefix(p.name, name) {
				continue
			}
			if p.offset < prevEnd || pos < p.offset {
				t.Errorf("pragma %s does not appear before definition %s", p.text, name)
			}
		}
		prevEnd = end
	}
}

type pragmaInfo struct {
	text   string // full text of the comment, not including newline
	name   string // text after //go:
	offset int    // byte offset of first slash in comment
}

func findPragmas(source []byte) []pragmaInfo {
	var pragmas []pragmaInfo
	pragmaPrefix := []byte("\n//go:")
	offset := 0
	for {
		i := bytes.Index(source[offset:], pragmaPrefix)
		if i == -1 {
			break
		}
		p := source[offset+i:]
		j := bytes.IndexByte(p[1:], '\n') + 1
		if j == -1 {
			j = len(p)
		}
		pragma := pragmaInfo{
			text:   string(p[1:j]),
			name:   string(p[len(pragmaPrefix):j]),
			offset: offset + i,
		}
		pragmas = append(pragmas, pragma)
		offset += i + j
	}
	return pragmas
}

// Makes sure that `cover -func=profile.cov` reports accurate coverage.
// Issue #20515.
func TestCoverFunc(t *testing.T) {
	// go tool cover -func ./testdata/profile.cov
	cmd := exec.Command(testenv.GoToolPath(t), "tool", "cover", "-func", coverProfile)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Logf("%s", ee.Stderr)
		}
		t.Fatal(err)
	}

	if got, err := regexp.Match(".*total:.*100.0.*", out); err != nil || !got {
		t.Logf("%s", out)
		t.Errorf("invalid coverage counts. got=(%v, %v); want=(true; nil)", got, err)
	}
}

func run(c *exec.Cmd, t *testing.T) {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		t.Fatal(err)
	}
}
