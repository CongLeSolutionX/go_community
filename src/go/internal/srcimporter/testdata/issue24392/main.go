package main

import (
	"fmt"
	"go/importer"
)

func main() {
	imp := importer.For("source", nil)
	pkg, err := imp.Import("example.com/testdata/testpkg")
	if err != nil {
		panic(err)
	}
	if got, want := pkg.Path(), "example.com/testdata/testpkg"; got != want {
		panic(fmt.Errorf("got %q; want %q", got, want))
	}
}
