// skip

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package foo

import (
	"debug/dwarf"
	"debug/pe"
	"fmt"
	"os"
)

var Dwarf *dwarf.Data
var TextStart uint64
var TextData []byte

func must(err error) {
	if err != nil {
		panic(err)
	}
}

/*


  Gratuitous white space to make the test lines match
  the bug. openPE should be on line 39.






*/

func openPE(path string) (*dwarf.Data, uint64, []byte) {
	file, _ := pe.Open(path)
	if file == nil {
		return nil, 0, nil
	}
	fmt.Fprintf(os.Stderr, "Found PE executable\n")
	dwarf, err := file.DWARF()
	must(err)

	var imageBase uint64
	switch oh := file.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		imageBase = uint64(oh.ImageBase)
	case *pe.OptionalHeader64:
		imageBase = oh.ImageBase
	default:
		panic(fmt.Errorf("pe file format not recognized"))
	}
	sect := file.Section(".text")
	if sect == nil {
		panic(fmt.Errorf("text section not found"))
	}
	textStart := imageBase + uint64(sect.VirtualAddress)
	textData, err := sect.Data()
	must(err)
	return dwarf, textStart, textData
}
