// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"sort"
	"unicode"
	"unicode/utf8"
)

type Package struct {
	name       string // Package name, json for encoding/json.
	importPath string
	pkg        *ast.Package   // Parsed package.
	file       *ast.File      // Merged from all files in the package
	fs         *token.FileSet // Needed for printing.
}

// parsePackage turns the build package we found into a parsed package
// we can then use to generate documentation.
func parsePackage(pkg *build.Package) *Package {
	fs := token.NewFileSet()
	// include tells parser.ParseDir which files to include.
	// That means the file must be in the build package's GoFiles or CgoFiles
	// list only (no tag-ignored files, tests, swig or other non-Go files).
	include := func(info os.FileInfo) bool {
		for _, name := range pkg.GoFiles {
			if name == info.Name() {
				return true
			}
		}
		for _, name := range pkg.CgoFiles {
			if name == info.Name() {
				return true
			}
		}
		return false
	}
	pkgs, err := parser.ParseDir(fs, pkg.Dir, include, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure they are all in one package.
	if len(pkgs) != 1 {
		log.Fatalf("multiple packages directory %s", pkg.Dir)
	}

	return &Package{
		name:       pkg.Name,
		importPath: pkg.ImportPath,
		pkg:        pkgs[pkg.Name],
		file:       ast.MergePackageFiles(pkgs[pkg.Name], 0),
		fs:         fs,
	}
}

var formatBuf bytes.Buffer // One instance to minimize allocation.

// emit prints the node.
func (pkg *Package) emit(node ast.Node) {
	if node != nil {
		formatBuf.Reset()
		err := format.Node(&formatBuf, pkg.fs, node)
		if err != nil {
			log.Fatal(err)
		}
		if formatBuf.Len() > 0 && formatBuf.Bytes()[formatBuf.Len()-1] != '\n' {
			formatBuf.WriteRune('\n')
		}
		os.Stdout.Write(formatBuf.Bytes())
	}
}

// formatNode is a helper function for printing.
func (pkg *Package) formatNode(node ast.Node) []byte {
	formatBuf.Reset()
	format.Node(&formatBuf, pkg.fs, node)
	return formatBuf.Bytes()
}

// oneLineFunc prints a function declaration as a single line.
func (pkg *Package) oneLineFunc(decl *ast.FuncDecl) {
	decl.Doc = nil
	decl.Body = nil
	pkg.emit(decl)
}

// oneLineValueGenDecl prints a var or const declaration as a single line.
func (pkg *Package) oneLineValueGenDecl(decl *ast.GenDecl, symbol string) {
	decl.Doc = nil
	dotDotDot := ""
	if len(decl.Specs) > 1 {
		dotDotDot = " ..."
	}
	// Find the first relevant spec.
	for i, spec := range decl.Specs {
		valueSpec := spec.(*ast.ValueSpec) // Must succeed; we can't mix types in one genDecl.
		if symbol != "" {
			if name, ok := valueSpec.Type.(*ast.Ident); !ok || !match(symbol, name.Name) {
				continue
			}
		}
		typ := ""
		if valueSpec.Type != nil {
			typ = fmt.Sprintf(" %s", pkg.formatNode(valueSpec.Type))
		}
		val := ""
		if i < len(valueSpec.Values) && valueSpec.Values[i] != nil {
			val = fmt.Sprintf(" = %s", pkg.formatNode(valueSpec.Values[i]))
		}
		fmt.Printf("%s %s%s%s%s\n", decl.Tok, valueSpec.Names[0], typ, val, dotDotDot)
		break
	}
}

// oneLineTypeDecl prints a type declaration as a single line.
func (pkg *Package) oneLineTypeDecl(spec *ast.TypeSpec) {
	spec.Doc = nil
	spec.Comment = nil
	switch spec.Type.(type) {
	case *ast.InterfaceType:
		fmt.Printf("type %s interface { ... }\n", spec.Name)
	case *ast.StructType:
		fmt.Printf("type %s struct { ... }\n", spec.Name)
	default:
		fmt.Printf("type %s %s\n", spec.Name, pkg.formatNode(spec.Type))
	}
}

// packageDoc prints the docs for the package (package doc plus one-liners of the rest).
// TODO: Sort the output.
func (pkg *Package) packageDoc() {
	for _, f := range pkg.pkg.Files {
		if f.Doc != nil {
			fmt.Printf("package %s // import %q\n\n", pkg.name, pkg.importPath)
			doc.ToText(os.Stdout, f.Doc.Text(), "", "\t", 80)
			fmt.Print("\n")
			break // First one counts.
		}
	}
	for _, d := range pkg.file.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			if spec, ok := decl.Specs[0].(*ast.ValueSpec); ok {
				exported := true
				for _, name := range spec.Names {
					if !isExported(name.Name) {
						exported = false
						break
					}
				}
				if exported {
					pkg.oneLineValueGenDecl(decl, "")
				}
				continue
			}
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					continue // Nothing to do.
				case *ast.TypeSpec:
					if isExported(spec.Name.Name) {
						pkg.oneLineTypeDecl(spec)
					}
				default:
					log.Fatalf("unrecognized type spec %T", spec)
				}
			}
		case *ast.FuncDecl:
			// Exported functions only, not methods.
			if isExported(decl.Name.Name) && decl.Recv == nil {
				pkg.oneLineFunc(decl)
			}
		default:
			log.Fatalf("can't handle node of type %T", decl)
		}
	}
}

// findDecl returns the declaration that holds the declaration for
// the symbol, and returns it. It will be either be a function
// declaration or a "general" declaration. If the latter, findDecl
// also returns the "spec" that declares the actual symbol.
func (pkg *Package) findDecl(symbol string) (*ast.FuncDecl, *ast.GenDecl, ast.Spec) {
	for _, d := range pkg.file.Decls {
		switch decl := d.(type) {
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ImportSpec:
					// Nothing to do.
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						if match(symbol, name.Name) {
							return nil, decl, spec
						}
					}
				case *ast.TypeSpec:
					if match(symbol, spec.Name.Name) {
						return nil, decl, spec
					}
				default:
					log.Fatalf("unrecognized type spec %T", spec)
				}
			}
		case *ast.FuncDecl:
			if decl.Recv == nil && match(symbol, decl.Name.Name) {
				return decl, nil, nil
			}
		default:
			log.Fatalf("can't handle node of type %T", decl)
		}
	}
	return nil, nil, nil
}

// findAllMethods returns the list of (exported) methods for the type symbol or *symbol.
func (pkg *Package) findAllMethods(symbol string) (methods []*ast.FuncDecl) {
	for _, d := range pkg.file.Decls {
		switch fun := d.(type) {
		case *ast.FuncDecl:
			if pkg.matchMethod(fun, symbol, "") {
				methods = append(methods, fun)
			}
		}
	}
	return methods
}

// findAllConstants returns the list of declarations of (exported) constants for the type symbol or *symbol.
// The return value is a slice of genDecls whose specs are ValueSpecs.
func (pkg *Package) findAllConstants(symbol string) (constants []*ast.GenDecl) {
OuterLoop:
	for _, d := range pkg.file.Decls {
		switch genDecl := d.(type) {
		case *ast.GenDecl:
			// Is this a list of constants?
			if genDecl.Tok != token.CONST {
				continue
			}
			// Is the type of one of these the identifier "symbol"?
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if name, ok := valueSpec.Type.(*ast.Ident); ok && match(symbol, name.Name) {
					constants = append(constants, genDecl)
					continue OuterLoop // We're good; go to next declaration.
				}
			}
		}
	}
	return constants
}

// symbolDoc prints the doc for symbol. If it is a type, this includes its methods,
// factories (TODO) and associated constants.
func (pkg *Package) symbolDoc(symbol string) {
	funcDecl, genDecl, node := pkg.findDecl(symbol)
	if funcDecl != nil {
		// Symbol is a function.
		funcDecl.Body = nil
		pkg.emit(funcDecl)
		return
	}
	if node == nil {
		log.Fatalf("symbol %s not present in package %q", symbol, pkg.name)
	}
	switch node := node.(type) {
	case *ast.TypeSpec:
		trimUnexportedFields(node)
		// If there are multiple types defined, reduce to just this one.
		if len(genDecl.Specs) > 1 {
			genDecl.Specs = []ast.Spec{node}
		}
		pkg.emit(genDecl)
		// NO TOO MANY pkg.emit(node)
	case *ast.ValueSpec:
		pkg.emit(genDecl)
	default:
		log.Fatalf("can't doc node of type %T", node)
	}
	// Show associated methods.
	methods := pkg.findAllMethods(symbol)
	sort.Sort(methodList(methods))
	for _, m := range methods {
		m.Doc = nil
		m.Body = nil
		pkg.emit(m)
	}
	// Show associated constants. (Actually just the first one in any const block.)
	constants := pkg.findAllConstants(symbol)
	for _, genDecl := range constants {
		pkg.oneLineValueGenDecl(genDecl, symbol)
	}
}

// trimUnexportedFields modifies spec in place to elide unexported fields (unless
// the unexported flag is set). If spec is not a structure declartion, nothing happens.
func trimUnexportedFields(spec *ast.TypeSpec) {
	if unexported {
		// We're printing all fields.
		return
	}
	// It must be a struct for us to care. (We show unexported methods in interfaces.)
	structType, ok := spec.Type.(*ast.StructType)
	if !ok {
		return
	}
	trimmed := false
	list := make([]*ast.Field, 0, len(structType.Fields.List))
	for _, field := range structType.Fields.List {
		// Trims if any is unexported. Fine in practice.
		ok := true
		for _, name := range field.Names {
			if !isExported(name.Name) {
				trimmed = true
				ok = false
				break
			}
		}
		if ok {
			list = append(list, field)
		}
	}
	if trimmed {
		unexportedField := &ast.Field{
			Type: ast.NewIdent(""), // Hack: printer will treat this as a field with a named type.
			Comment: &ast.CommentGroup{
				List: []*ast.Comment{
					&ast.Comment{
						Text: "// Has unexported fields.\n",
					},
				},
			},
		}
		list = append(list, unexportedField)
		structType.Fields.List = list
	}
}

// findMethod returns the declaration for symbol.method.
func (pkg *Package) findMethod(symbol, method string) *ast.FuncDecl {
	for _, d := range pkg.file.Decls {
		switch fun := d.(type) {
		case *ast.FuncDecl:
			if pkg.matchMethod(fun, symbol, method) {
				return fun
			}
		}
	}
	log.Fatalf("no such method %s.%s", symbol, method)
	return nil
}

// matchMethod reports whether the function declaration is one for the specified
// symbol and method name; the method name can be empty, meaning always match.
func (pkg *Package) matchMethod(fun *ast.FuncDecl, symbol, method string) bool {
	// Is it a method?
	recv := fun.Recv
	if recv == nil || len(recv.List) != 1 {
		return false // Not a method (or strange somehow).
	}
	// Does the method name match?
	if method != "" && !match(method, fun.Name.Name) {
		return false
	}
	if !isExported(fun.Name.Name) {
		return false
	}
	recvField := recv.List[0]
	// Does the type match? Must be T or *T.
	var name string
	switch typ := recvField.Type.(type) {
	case *ast.Ident:
		name = typ.Name
	case *ast.StarExpr:
		name = typ.X.(*ast.Ident).Name
	default:
		log.Fatalf("unexpected receiver type %s", typ)
	}
	if !isExported(name) {
		return false
	}
	if !match(symbol, name) {
		return false
	}
	return true
}

// methodDoc prints the doc for symbol.method.
func (pkg *Package) methodDoc(symbol, method string) {
	fnDecl, _, node := pkg.findDecl(symbol)
	if fnDecl != nil {
		log.Fatalf("%s is a function; it has no methods", symbol)
	}
	if node == nil {
		log.Fatalf("symbol %s not present in package %q", symbol, pkg.name)
	}
	var ok bool
	if _, ok = node.(*ast.TypeSpec); !ok {
		log.Fatalf("%s is not a type, cannot have method", symbol)
	}
	funcDecl := pkg.findMethod(symbol, method)
	funcDecl.Body = nil
	pkg.emit(funcDecl)
}

// match reports whether the user's symbol matches the program's.
// A lower-case character in the user's string matches either case in the program's.
// The program string must be exported.
func match(user, program string) bool {
	if !isExported(program) {
		return false
	}
	for _, u := range user {
		p, w := utf8.DecodeRuneInString(program)
		program = program[w:]
		if u == p {
			continue
		}
		if unicode.IsLower(u) && simpleFold(u) == simpleFold(p) {
			continue
		}
		return false
	}
	return program == ""
}

// simpleFold returns the minimum rune equivalent to r
// under Unicode-defined simple case folding.
func simpleFold(r rune) rune {
	for {
		r1 := unicode.SimpleFold(r)
		if r1 <= r {
			return r1 // wrapped around, found min
		}
		r = r1
	}
}

// methodList lets us sort list of methods by name alphabetically.
type methodList []*ast.FuncDecl

func (m methodList) Len() int           { return len(m) }
func (m methodList) Less(i, j int) bool { return m[i].Name.Name < m[j].Name.Name }
func (m methodList) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
