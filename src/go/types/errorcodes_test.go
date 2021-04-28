// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	. "go/types"
)

var (
	ecodeOut = flag.String("ecode_out", "", "output error code data to this file in TestWriteErrorCodes")
)

func TestErrorCodeExamples(t *testing.T) {
	walkCodes(t, func(name string, value int, spec *ast.ValueSpec) {
		t.Run(name, func(t *testing.T) {
			doc := spec.Doc.Text()
			examples := strings.Split(doc, "Example:")
			for i := 1; i < len(examples); i++ {
				example := examples[i]
				err := checkExample(t, example)
				if err == nil {
					t.Fatalf("no error in example #%d", i)
				}
				typerr, ok := err.(Error)
				if !ok {
					t.Fatalf("not a types.Error: %v", err)
				}
				if got := readCode(typerr); got != value {
					t.Errorf("%s: example #%d returned code %d (%s), want %d", name, i, got, err, value)
				}
			}
		})
	})
}

func walkCodes(t *testing.T, f func(name string, value int, spec *ast.ValueSpec)) {
	t.Helper()
	fset := token.NewFileSet()
	files, err := pkgFiles(fset, ".", parser.ParseComments) // from self_test.go
	if err != nil {
		t.Fatal(err)
	}
	conf := Config{Importer: importer.Default()}
	info := &Info{
		Types: make(map[ast.Expr]TypeAndValue),
		Defs:  make(map[*ast.Ident]Object),
		Uses:  make(map[*ast.Ident]Object),
	}
	_, err = conf.Check("types", fset, files, info)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST {
				continue
			}
			for _, spec := range decl.Specs {
				spec, ok := spec.(*ast.ValueSpec)
				if !ok || len(spec.Names) == 0 {
					continue
				}
				obj := info.ObjectOf(spec.Names[0])
				if named, ok := obj.Type().(*Named); ok && named.Obj().Name() == "errorCode" {
					if len(spec.Names) != 1 {
						t.Fatalf("bad Code declaration for %q: got %d names, want exactly 1", spec.Names[0].Name, len(spec.Names))
					}
					codename := spec.Names[0].Name
					value := int(constant.Val(obj.(*Const).Val()).(int64))
					f(codename, value, spec)
				}
			}
		}
	}
}

func readCode(err Error) int {
	v := reflect.ValueOf(err)
	return int(v.FieldByName("go116code").Int())
}

func checkExample(t *testing.T, example string) error {
	t.Helper()
	fset := token.NewFileSet()
	src := fmt.Sprintf("package p\n\n%s", example)
	file, err := parser.ParseFile(fset, "example.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	conf := Config{
		FakeImportC: true,
		Importer:    importer.Default(),
	}
	_, err = conf.Check("example", fset, []*ast.File{file}, nil)
	return err
}

func TestErrorCodeStyle(t *testing.T) {
	// The set of error codes is large and intended to be self-documenting, so
	// this test enforces some style conventions.
	forbiddenInIdent := []string{
		// use invalid instead
		"illegal",
		// words with a common short-form
		"argument",
		"assertion",
		"assignment",
		"boolean",
		"channel",
		"condition",
		"declaration",
		"expression",
		"function",
		"initial", // use init for initializer, initialization, etc.
		"integer",
		"interface",
		"iterat", // use iter for iterator, iteration, etc.
		"literal",
		"operation",
		"package",
		"pointer",
		"receiver",
		"signature",
		"statement",
		"variable",
	}
	forbiddenInComment := []string{
		// lhs and rhs should be spelled-out.
		"lhs", "rhs",
		// builtin should be hyphenated.
		"builtin",
		// Use dot-dot-dot.
		"ellipsis",
	}
	nameHist := make(map[int]int)
	longestName := ""
	maxValue := 0

	walkCodes(t, func(name string, value int, spec *ast.ValueSpec) {
		if name == "_" {
			return
		}
		nameHist[len(name)]++
		if value > maxValue {
			maxValue = value
		}
		if len(name) > len(longestName) {
			longestName = name
		}
		if token.IsExported(name) {
			// This is an experimental API, and errorCode values should not be
			// exported.
			t.Errorf("%q is exported", name)
		}
		if name[0] != '_' || !token.IsExported(name[1:]) {
			t.Errorf("%q should start with _, followed by an exported identifier", name)
		}
		lower := strings.ToLower(name)
		for _, bad := range forbiddenInIdent {
			if strings.Contains(lower, bad) {
				t.Errorf("%q contains forbidden word %q", name, bad)
			}
		}
		doc := spec.Doc.Text()
		if !strings.HasPrefix(doc, name) {
			t.Errorf("doc for %q does not start with identifier", name)
		}
		lowerComment := strings.ToLower(strings.TrimPrefix(doc, name))
		for _, bad := range forbiddenInComment {
			if strings.Contains(lowerComment, bad) {
				t.Errorf("doc for %q contains forbidden word %q", name, bad)
			}
		}
	})

	if testing.Verbose() {
		var totChars, totCount int
		for chars, count := range nameHist {
			totChars += chars * count
			totCount += count
		}
		avg := float64(totChars) / float64(totCount)
		fmt.Println()
		fmt.Printf("%d error codes\n", totCount)
		fmt.Printf("average length: %.2f chars\n", avg)
		fmt.Printf("max length: %d (%s)\n", len(longestName), longestName)
	}
}

type codeData struct {
	name  string
	value int
}

func TestWriteErrorCodes(t *testing.T) {
	if *ecodeOut == "" {
		t.Skip("no output specified")
	}

	var codes []codeData
	walkCodes(t, func(name string, value int, _ *ast.ValueSpec) {
		if name == "_" {
			return
		}
		codes = append(codes, codeData{name, value})
	})
	sort.Slice(codes, func(i, j int) bool {
		return codes[i].value < codes[j].value
	})

	f, err := os.Create(*ecodeOut)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	for _, code := range codes {
		fmt.Fprintf(f, "%s %d\n", code.name, code.value)
	}
}

func TestErrorCodeChanges(t *testing.T) {
	// Build maps to associate error code name<->value, and use these to try to
	// guess what happened with a given code from older versions of Go.
	//
	// There are three possibilities that we try to differentiate, in order to
	// make it clear what the corrective action should be.
	//  1. The code value changed.
	//  2. The code name changed.
	//  3. The code was renamed.
	//
	// Of these, 1 is almost certainly an error: if the same constant name means
	// something different, in a later version of go/types, it is likely that
	// code values were accidentally altered.
	//
	// 2 and 3 are OK, but if too many codes fall into this category it might be
	// that the entire naming schema has changed, and this test is of little
	// value. Try to detect if this has happened by looking for a minimum number
	// of matches.
	//
	// TODO(rFindley) replace this quick and dirty check with a more formal
	// generation of code values, once we decide where codes should live.
	current := make(map[string]int)
	currentInv := make(map[int]string)
	walkCodes(t, func(name string, value int, _ *ast.ValueSpec) {
		if name == "_" {
			return
		}
		current[name] = value
		if _, ok := currentInv[value]; ok {
			t.Fatalf("duplicate error code value %d", value)
		}
		currentInv[value] = name
	})
	dir := filepath.Join("testdata", "codes")
	fis, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Sanity check: we should at least match _some_ error code names, else the
	// entire naming schema may have changed, and test data should be updated
	// accordingly.
	matched := 0
	const wantMatched = 100

	for _, fi := range fis {
		t.Run(fi.Name(), func(t *testing.T) {
			path := filepath.Join(dir, fi.Name())
			found := readCodeData(t, path)
			for _, old := range found {
				if curval, ok := current[old.name]; ok {
					if old.value == curval {
						matched++
					} else {
						t.Errorf("%s: value changed from %d to %d", old.name, old.value, curval)
					}
					continue
				}
				if curname, ok := currentInv[old.value]; ok {
					t.Logf("%s (#%d): name changed to %s", old.name, old.value, curname)
					continue
				}
				t.Logf("%s (#%d) was removed", old.name, old.value)
			}
		})
	}
	if matched < wantMatched {
		t.Errorf("only matched %d error codes, want at least %d. Did the naming convention change?", matched, wantMatched)
	}
}

// readCodeData loads code names and values that were written to a data file
// using TestWriteErrorCodes.
func readCodeData(t *testing.T, path string) []codeData {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var codes []codeData
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			t.Fatalf("malformed error code data: %q", scanner.Text())
		}
		val, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatalf("bad error code value: %q", fields[1])
		}
		codes = append(codes, codeData{fields[0], val})
	}
	return codes
}
