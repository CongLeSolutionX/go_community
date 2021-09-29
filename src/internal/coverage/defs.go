// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

// Types and constants related to the output files files written
// by code coverage tooling. When an coverage-instrumented binary
// is run, it emits two output files: a meta-data output file, and
// a counter data output file.

//.....................................................................
//
// Meta-data definitions:
//
// The meta-data file is composed of a file header, a series of
// meta-data sections (one per instrumented package), and an offsets
// area storing the offsets of each section. Format of the meta-data
// file looks like:
//
// --header----------
//  | magic: [4]byte magic string
//  | module name [string table offset]
//  | size: size of file in bytes
//  | numPkgs: number of package entries in file
//  | hash: [16]byte hash for entire file
//  | offset to string table section
//  | length of string table
//  | number of entries in string table
//  --package offsets table------
//  <offset to pkg 0>
//  <offset to pkg 1>
//  ...
//  --package lengths table------
//  <length of pkg 0>
//  <length of pkg 1>
//  ...
//  --string table------
//  <uleb128 len> 8
//  <data> "mymodule"
//  ...
//  --package payloads------
//  <meta-symbol for pkg 0>
//  <meta-symbol for pkg 1>
//  ...
//
// Each package payload is a stand-alone blob emitted by the compiler.
// Note that the file-level string table is expected to be very
// short (most strings will be in the meta-data blobs themselves).

// CovMetaMagic holds the magic string for a meta-data file.
var CovMetaMagic = [4]byte{'\x00', '\x63', '\x76', '\x6d'}

// MetaFileHeader stores file header information for a meta-data file.
type MetaFileHeader struct {
	Magic        [4]byte
	_            [4]byte // padding
	TotalLength  uint64
	Entries      uint64
	MetaHash     [16]byte
	StrTabOffset uint32
	StrTabLength uint32
	StrTabEnts   uint32
	ModuleName   uint32 // string table index
}

const MetaFilePref = "covmeta"

// A counter data file is composed of a file header, offsets section,
// string section, then a series of counter entries each containing
// the counter data for a specific function.

// CovCounterMagic holds the magic string for a coverage counter-data file.
var CovCounterMagic = [4]byte{'\x00', '\x63', '\x77', '\x6d'}

// CounterFileHeader stores files header information for a counter-data file.
type CounterFileHeader struct {
	Magic          [4]byte
	_              [4]byte // padding
	TotalLength    uint64  // size in bytes
	MetaHash       [16]byte
	FcnEntries     uint64
	StrTabOff      uint32
	StrTabLen      uint32
	StrTabNentries uint32
	ArgsOff        uint32
	ArgsLen        uint32
	ArgsNentries   uint32
	BigEndian      bool
	_              [7]byte // padding
}

const CounterFilePref = "covcounters"

// The meta-data for a single package looks like the following:
//
// --header----------
//  | size: size of this blob in bytes
//  | packagepath: <path to p>
//  | classification: ...
//  | nfiles: 1
//  | nfunctions: 2
//  --func offsets table------
//  <offset to func 0>
//  <offset to func 1>
//  --file + function table------
//  | <uleb128 len> 4
//  | <data> "p.go"
//  | <uleb128 len> 5
//  | <data> "small"
//  | <uleb128 len> 6
//  | <data> "Medium"
//  --func 1------
//  | <uleb128> num units: 3
//  | <uleb128> func name: 1 (index into string table)
//  | <uleb128> file: 0 (index into string table)
//  | <unit 0>:  F0   L6     L8    2
//  | <unit 1>:  F0   L9     L9    1
//  | <unit 2>:  F0   L11    L11   1
//  --func 2------
//  | <uleb128> num units: 1
//  | <uleb128> func name: 2 (index into string table)
//  | <uleb128> file: 0 (index into string table)
//  | <unit 0>:  F0   L15    L19   5
//  ---end-----------

type MetaSymbolHeader struct {
	Length   uint32 // size of meta-symbol in bytes
	PkgPath  uint32 // string table index
	NumFiles uint32
	NumFuncs uint32
}

const CovMetaHeaderSize = 4 + 4 + 4 + 4 // keep in sync with above

// The following types and constants used by the meta-data encoder/decoder.

// FuncDesc encapsulates the meta-data definitions for a single Go function.
// This version assumes that we're looking at a function before inlining;
// if we want to capture a post-inlining view of the world, the
// representations of source positions would need to be a good deal more
// complicated.
type FuncDesc struct {
	Funcname string
	Srcfile  string
	Units    []CoverableUnit
}

// CoverableUnit describes the source characteristics of a single
// basic block (region of straight-line code with no jumps or control
// transfers) in a function being instrumented.
type CoverableUnit struct {
	StLine, StCol uint32
	EnLine, EnCol uint32
	NxStmts       uint32
}

func Round4(x int) int {
	return (x + 3) &^ 3
}
