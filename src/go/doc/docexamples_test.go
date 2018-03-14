package doc_test

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
)

func ExampleExample() {
	// build.Import collects all the information we need to parse a package.
	// It handles build tags and finds the correct path to the source.
	buildInfo, err := build.Import("go/doc", "", 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Note that, while we're collecting the test files together,
	// the buildInfo.XTestGoFiles is a separate package named
	// buildInfo.Name + "_test".
	var testFilenames []string
	testFilenames = append(testFilenames, buildInfo.TestGoFiles...)
	testFilenames = append(testFilenames, buildInfo.XTestGoFiles...)

	fset := token.NewFileSet()

	var examples []*doc.Example

	for _, filename := range testFilenames {
		// Compute the full file path.
		path := filepath.Join(buildInfo.Dir, filename)

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		examples = append(examples, doc.Examples(file)...)
	}

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Name < examples[j].Name
	})

	fmt.Println("Examples:")
	for _, example := range examples {
		fmt.Printf("- Example%s\n", example.Name)
	}

	// Output:
	//
	// Examples:
	// - ExampleExample
	// - ExamplePackage
}

func ExamplePackage() {
	// build.Import collects all the information we need to parse a package.
	// It handles build tags and finds the correct path to the source.
	buildInfo, err := build.Import("go/doc", "", 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Collect the names of the files of this package.
	// While go/doc does not have any cgo files, an arbitrary package may.
	// To parse a package and its tests, append buildInfo.TestGoFiles,
	// but note that buildInfo.XTestGoFiles is separate package
	// named buildInfo.Name + "_test".
	var filenames []string
	filenames = append(filenames, buildInfo.GoFiles...)
	filenames = append(filenames, buildInfo.CgoFiles...)

	fset := token.NewFileSet()

	// To construct an *ast.Package, we need a map of *ast.File,
	// keyed by the file's name.
	files := map[string]*ast.File{}

	for _, filename := range filenames {
		// Compute the full file path.
		path := filepath.Join(buildInfo.Dir, filename)

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}

		files[filename] = file
	}

	astPkg := &ast.Package{
		Name:  buildInfo.Name,
		Files: files,
	}

	pkg := doc.New(astPkg, buildInfo.ImportPath, 0)

	// Remove unexported types/funcs/etc.
	pkg.Filter(ast.IsExported)

	fmt.Println("Function names:")
	for _, Func := range pkg.Funcs {
		fmt.Println("-", Func.Name)
	}

	// Output:
	//
	// Function names:
	// - IsPredeclared
	// - Synopsis
	// - ToHTML
	// - ToText
}
