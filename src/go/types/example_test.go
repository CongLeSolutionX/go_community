// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

// This file shows examples of basic usage of the go/types API.
//
// To locate a Go package, use (*go/build.Context).Import.
// To load, parse, and type-check a complete Go program
// from source, use golang.org/x/tools/go/loader.

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"sort"
	"strings"
)

// ExampleScope prints the tree of Scopes of a package creates from a
// set of parsed files.  Type information for imported packages is
// provided by gcimporter, which loads export data from .a files
func ExampleScope() {
	// Parse the source files of a single package.
	fset := token.NewFileSet()
	var files []*ast.File
	for _, file := range []struct{ name, data string }{
		{"main.go", `
package main
import "fmt"
func main() {
	freezing := FToC(-18)
	fmt.Println(freezing, Boiling) }
`},
		{"celsius.go", `
package main
import "fmt"
type Celsius float64
func (c Celsius) String() string { return fmt.Sprintf("%g°C", c) }
func FToC(f float64) Celsius { return Celsius(f - 32 / 9 * 5) }
const Boiling Celsius = 100
`},
	} {
		f, err := parser.ParseFile(fset, file.name, file.data, 0)
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, f)
	}

	// Type-check the package.
	// Type information for the imported "fmt" package
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check("temperature", fset, files, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Package %s (%q):\n", pkg.Name(), pkg.Path())
	printScope(pkg.Scope(), "", pkg)

	// Output:
	// Package main ("temperature"):
	// scope package "temperature" {
	// 	const Boiling Celsius
	// 	type Celsius float64
	// 		method (temperature.Celsius) String() string
	// 	func FToC(f float64) Celsius
	// 	func main()
	// 	scope main.go {
	// 		package fmt
	// 		scope function {
	// 			var freezing Celsius
	// 		}
	// 	}
	// 	scope celsius.go {
	// 		package fmt
	// 		scope function {
	// 			var c Celsius
	// 		}
	// 		scope function {
	// 			var f float64
	// 		}
	// 	}
	// }
}

func printScope(scope *types.Scope, indent string, pkg *types.Package) {
	s := scope.String()
	fmt.Printf("%sscope %s {\n", indent, s[:strings.Index(s, " scope")])
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		fmt.Printf("%s\t%s\n", indent, types.ObjectString(pkg, obj))
		if named, ok := obj.(*types.TypeName); ok {
			mset := types.NewMethodSet(named.Type())
			for i := 0; i < mset.Len(); i++ {
				fmt.Printf("%s\t\t%s\n", indent, mset.At(i))
			}
		}
	}
	for i := 0; i < scope.NumChildren(); i++ {
		printScope(scope.Child(i), indent+"\t", pkg)
	}
	fmt.Printf("%s}\n", indent)
}

// ExampleInfo prints various facts recorded by the type checker in a
// types.Info: definitions and references of each named object, and
// the type, value, and mode of every expression in the package.
func ExampleInfo() {
	const data = `
package fib

type S string

var a, b, c = len(b), S(c), "hello"

func fib(x int) int {
	if x < 2 {
		return x
	}
	return fib(x-1) - fib(x-2)
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fib.go", data, 0)
	if err != nil {
		log.Fatal(err)
	}

	// Type-check the package.
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	var conf types.Config
	pkg, err := conf.Check("fib", fset, []*ast.File{f}, &info)
	if err != nil {
		log.Fatal(err)
	}

	// Print package-level variables in initialization order.
	fmt.Printf("InitOrder: %v\n\n", info.InitOrder)

	// For each named object, print the line and
	// column of its definition and each of its uses.
	fmt.Println("Defs and Uses of each named object:")
	usesByObj := make(map[types.Object][]string)
	for id, obj := range info.Uses {
		posn := fset.Position(id.Pos())
		lineCol := fmt.Sprintf("%d:%d", posn.Line, posn.Column)
		usesByObj[obj] = append(usesByObj[obj], lineCol)
	}
	var items []string
	for obj, uses := range usesByObj {
		sort.Strings(uses)
		item := fmt.Sprintf("%s:\n  defined at %s\n  used at %s",
			types.ObjectString(pkg, obj),
			fset.Position(obj.Pos()),
			strings.Join(uses, ", "))
		items = append(items, item)
	}
	sort.Strings(items)
	fmt.Println(strings.Join(items, "\n"))

	fmt.Println("\nTypes and Values of each expression:")
	items = nil
	for expr, tv := range info.Types {
		var buf bytes.Buffer
		posn := fset.Position(expr.Pos())
		tvstr := tv.Type.String()
		if tv.Value != nil {
			tvstr += " = " + tv.Value.String()
		}
		// line:col | expr | mode : type = value
		fmt.Fprintf(&buf, "%2d:%2d | %-19s | %-7s : %s",
			posn.Line, posn.Column, exprString(fset, expr),
			mode(tv), tvstr)
		items = append(items, buf.String())
	}
	sort.Strings(items)
	fmt.Println(strings.Join(items, "\n"))

	// Output:
	// InitOrder: [c = "hello" b = S(c) a = len(b)]
	//
	// Defs and Uses of each named object:
	// builtin len:
	//   defined at -
	//   used at 6:15
	// func fib(x int) int:
	//   defined at fib.go:8:6
	//   used at 12:20, 12:9
	// type S string:
	//   defined at fib.go:4:6
	//   used at 6:23
	// type int int:
	//   defined at -
	//   used at 8:12, 8:17
	// type string string:
	//   defined at -
	//   used at 4:8
	// var b S:
	//   defined at fib.go:6:8
	//   used at 6:19
	// var c string:
	//   defined at fib.go:6:11
	//   used at 6:25
	// var x int:
	//   defined at fib.go:8:10
	//   used at 10:10, 12:13, 12:24, 9:5
	//
	// Types and Values of each expression:
	//  4: 8 | string              | type    : string
	//  6:15 | len                 | builtin : func(string) int
	//  6:15 | len(b)              | value   : int
	//  6:19 | b                   | var     : fib.S
	//  6:23 | S                   | type    : fib.S
	//  6:23 | S(c)                | value   : fib.S
	//  6:25 | c                   | var     : string
	//  6:29 | "hello"             | value   : string = "hello"
	//  8:12 | int                 | type    : int
	//  8:17 | int                 | type    : int
	//  9: 5 | x                   | var     : int
	//  9: 5 | x < 2               | value   : untyped bool
	//  9: 9 | 2                   | value   : int = 2
	// 10:10 | x                   | var     : int
	// 12: 9 | fib                 | value   : func(x int) int
	// 12: 9 | fib(x - 1)          | value   : int
	// 12: 9 | fib(x-1) - fib(x-2) | value   : int
	// 12:13 | x                   | var     : int
	// 12:13 | x - 1               | value   : int
	// 12:15 | 1                   | value   : int = 1
	// 12:20 | fib                 | value   : func(x int) int
	// 12:20 | fib(x - 2)          | value   : int
	// 12:24 | x                   | var     : int
	// 12:24 | x - 2               | value   : int
	// 12:26 | 2                   | value   : int = 2
}

func mode(tv types.TypeAndValue) string {
	if tv.IsVoid() {
		return "void"
	} else if tv.IsType() {
		return "type"
	} else if tv.IsBuiltin() {
		return "builtin"
	} else if tv.IsNil() {
		return "nil"
	} else if tv.Assignable() {
		if tv.Addressable() {
			return "var"
		}
		return "mapindex"
	} else if tv.IsValue() {
		return "value"
	}
	return "unknown"
}

func exprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, fset, expr)
	return buf.String()
}
