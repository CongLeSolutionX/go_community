// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package doc

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"internal/diff"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"
)

var update = flag.Bool("update", false, "update golden (.out) files")
var files = flag.String("files", "", "consider only Go test files matching this regular expression")

const dataDir = "testdata"

var templateTxt = readTemplate("template.txt")

func readTemplate(filename string) *template.Template {
	t := template.New(filename)
	t.Funcs(template.FuncMap{
		"node":     nodeFmt,
		"synopsis": synopsisFmt,
		"indent":   indentFmt,
	})
	return template.Must(t.ParseFiles(filepath.Join(dataDir, filename)))
}

func nodeFmt(node any, fset *token.FileSet) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, node)
	return strings.ReplaceAll(strings.TrimSpace(buf.String()), "\n", "\n\t")
}

func synopsisFmt(s string) string {
	const n = 64
	if len(s) > n {
		// cut off excess text and go back to a word boundary
		s = s[0:n]
		if i := strings.LastIndexAny(s, "\t\n "); i >= 0 {
			s = s[0:i]
		}
		s = strings.TrimSpace(s) + " ..."
	}
	return "// " + strings.ReplaceAll(s, "\n", " ")
}

func indentFmt(indent, s string) string {
	end := ""
	if strings.HasSuffix(s, "\n") {
		end = "\n"
		s = s[:len(s)-1]
	}
	return indent + strings.ReplaceAll(s, "\n", "\n"+indent) + end
}

func isGoFile(fi fs.FileInfo) bool {
	name := fi.Name()
	return !fi.IsDir() &&
		len(name) > 0 && name[0] != '.' && // ignore .files
		filepath.Ext(name) == ".go"
}

type bundle struct {
	*Package
	FSet *token.FileSet
}

func test(t *testing.T, mode Mode) {
	// determine file filter
	filter := isGoFile
	if *files != "" {
		rx, err := regexp.Compile(*files)
		if err != nil {
			t.Fatal(err)
		}
		filter = func(fi fs.FileInfo) bool {
			return isGoFile(fi) && rx.MatchString(fi.Name())
		}
	}

	// get packages
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dataDir, filter, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	// test packages
	for _, pkg := range pkgs {
		t.Run(pkg.Name, func(t *testing.T) {
			importPath := dataDir + "/" + pkg.Name
			var files []*ast.File
			for _, f := range pkg.Files {
				files = append(files, f)
			}
			doc, err := NewFromFiles(fset, files, importPath, mode)
			if err != nil {
				t.Fatal(err)
			}

			// golden files always use / in filenames - canonicalize them
			for i, filename := range doc.Filenames {
				doc.Filenames[i] = filepath.ToSlash(filename)
			}

			// print documentation
			var buf bytes.Buffer
			if err := templateTxt.Execute(&buf, bundle{doc, fset}); err != nil {
				t.Fatal(err)
			}
			got := buf.Bytes()

			// update golden file if necessary
			golden := filepath.Join(dataDir, fmt.Sprintf("%s.%d.golden", pkg.Name, mode))
			if *update {
				err := os.WriteFile(golden, got, 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			// get golden file
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}

			// compare
			if d := diff.Diff(fmt.Sprintf("%s.%d", pkg.Name, mode), got, filepath.Base(golden), want); len(d) != 0 {
				t.Errorf("output differs:\n%s", d)
			}
		})
	}
}

func Test(t *testing.T) {
	t.Run("default", func(t *testing.T) { test(t, 0) })
	t.Run("AllDecls", func(t *testing.T) { test(t, AllDecls) })
	t.Run("AllMethods", func(t *testing.T) { test(t, AllMethods) })
}

func TestFuncs(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "funcs.go", strings.NewReader(funcsTestFile), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := NewFromFiles(fset, []*ast.File{file}, "importPath", Mode(0))
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range doc.Funcs {
		f.Decl = nil
	}
	for _, ty := range doc.Types {
		for _, f := range ty.Funcs {
			f.Decl = nil
		}
		for _, m := range ty.Methods {
			m.Decl = nil
		}
	}

	compareFuncs := func(t *testing.T, what string, got, want *Func) {
		// ignore Decl and Examples
		got.Decl = nil
		got.Examples = nil
		if !(got.Doc == want.Doc &&
			got.Name == want.Name &&
			got.Recv == want.Recv &&
			got.Orig == want.Orig &&
			got.Level == want.Level) {
			t.Errorf("%s:\ngot  %+v\nwant %+v", what, got, want)
		}
	}

	compareSlices(t, "Funcs", doc.Funcs, funcsPackage.Funcs, compareFuncs)
	compareSlices(t, "Types", doc.Types, funcsPackage.Types, func(t *testing.T, what string, got, want *Type) {
		compareSlices(t, got.Name+".Funcs", got.Funcs, want.Funcs, compareFuncs)
		compareSlices(t, got.Name+".Methods", got.Methods, want.Methods, compareFuncs)
	})
}

func compareSlices[E interface{ Type | Func }](t *testing.T, what string, got, want []*E, compareElem func(*testing.T, string, *E, *E)) {
	i, j := 0, 0
	for i < len(got) || j < len(want) {
		// Use reflect to obtain the Name fields because the typechecker can't figure it out.
		// (This works around https://go.dev/issue/48522.)
		var gotName, wantName string
		if j < len(want) {
			wantName = reflect.ValueOf(want[j]).Elem().FieldByName("Name").Interface().(string)
		}
		if i < len(got) {
			gotName = reflect.ValueOf(got[i]).Elem().FieldByName("Name").Interface().(string)
		}
		var ord int
		if i >= len(got) {
			ord = -1
		} else if j >= len(want) {
			ord = 1
		} else {
			ord = cmp.Compare(gotName, wantName)
		}
		switch ord {
		case -1:
			t.Errorf("%s: missing entry for %s", what, wantName)
			j++
		case 1:
			t.Errorf("%s: unexpected entry for %s", what, gotName)
			i++
		default:
			compareElem(t, fmt.Sprintf("%s[%s]", what, wantName), got[i], want[j])
			i++
			j++
		}
	}
}

const funcsTestFile = `
package funcs

func F() {}

type S1 struct {
	S2  // embedded, exported
	s3  // embedded, unexported
}

func NewS1()  S1 {return S1{} }
func NewS1p() *S1 { return &S1{} }

func (S1) M1() {}
func (r S1) M2() {}
func(S1) m3() {}		// unexported not shown
func (*S1) P1() {}		// pointer receiver

type S2 int
func (S2) M3() {}		// shown on S2

type s3 int
func (s3) M4() {}		// shown on S1

type G1[T any] struct {
	*s3
}

func NewG1[T any]() G1[T] { return G1[T]{} }

func (G1[T]) MG1() {}
func (*G1[U]) MG2() {}

type G2[T, U any] struct {}

func NewG2[T, U any]() G2[T, U] { return G2[T, U]{} }

func (G2[T, U]) MG3() {}
func (*G2[A, B]) MG4() {}


`

var funcsPackage = &Package{
	Funcs: []*Func{{Name: "F"}},
	Types: []*Type{
		{
			Name:  "G1",
			Funcs: []*Func{{Name: "NewG1"}},
			Methods: []*Func{
				{Name: "M4", Recv: "G1", // TODO: synthesize a param for G1?
					Orig: "s3", Level: 1},
				{Name: "MG1", Recv: "G1[T]", Orig: "G1[T]", Level: 0},
				{Name: "MG2", Recv: "*G1[U]", Orig: "*G1[U]", Level: 0},
			},
		},
		{
			Name:  "G2",
			Funcs: []*Func{{Name: "NewG2"}},
			Methods: []*Func{
				{Name: "MG3", Recv: "G2[T, U]", Orig: "G2[T, U]", Level: 0},
				{Name: "MG4", Recv: "*G2[A, B]", Orig: "*G2[A, B]", Level: 0},
			},
		},
		{
			Name:  "S1",
			Funcs: []*Func{{Name: "NewS1"}, {Name: "NewS1p"}},
			Methods: []*Func{
				{Name: "M1", Recv: "S1", Orig: "S1", Level: 0},
				{Name: "M2", Recv: "S1", Orig: "S1", Level: 0},
				{Name: "M4", Recv: "S1", Orig: "s3", Level: 1},
				{Name: "P1", Recv: "*S1", Orig: "*S1", Level: 0},
			},
		},
		{
			Name: "S2",
			Methods: []*Func{
				{Name: "M3", Recv: "S2", Orig: "S2", Level: 0},
			},
		},
	},
}
