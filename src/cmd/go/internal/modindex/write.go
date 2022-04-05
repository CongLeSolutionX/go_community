package modindex

import (
	"encoding/binary"
	"go/token"
	"path/filepath"
	"sort"
)

const indexVersion = "go index v0\n"

// encodeModule produces the encoded representation of the module index.
func encodeModule(packages []*RawPackage, moddir string) ([]byte, error) {
	// fix up dir
	for i := range packages {
		rel, err := filepath.Rel(moddir, packages[i].dir)
		if err != nil {
			return nil, err
		}
		packages[i].dir = rel
	}

	e := newEncoder()
	e.Bytes([]byte(indexVersion))
	stringTableOffsetPos := e.Pos() // fill this at the end
	e.Uint32(0)                     // string table offset
	e.Uint32(uint32(len(packages)))
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].dir < packages[j].dir
	})
	for _, p := range packages {
		e.String(p.srcDir)
	}
	packagesOffsetPos := make([]uint32, len(packages))
	for i := range packages {
		packagesOffsetPos[i] = e.Pos()
		e.Uint32(0)
	}
	for i, p := range packages {
		e.Uint32At(e.Pos(), packagesOffsetPos[i])
		encodePackage(e, p)
	}
	e.Uint32At(e.Pos(), stringTableOffsetPos)
	e.Bytes(e.stringTable)
	return e.b, nil
}

func encodePackage(e *encoder, p *RawPackage) {
	e.String(p.error)
	e.String(p.path)
	e.String(p.srcDir)
	e.String(p.dir)
	e.Uint32(uint32(len(p.sourceFiles)))                      // number of source files
	sourceFileOffsetPos := make([]uint32, len(p.sourceFiles)) // where to place the ith source file's offset
	for i := range p.sourceFiles {
		sourceFileOffsetPos[i] = e.Pos()
		e.Uint32(0)
	}
	for i, f := range p.sourceFiles {
		e.Uint32At(e.Pos(), sourceFileOffsetPos[i])
		encodeFile(e, f)
	}
}

func encodeFile(e *encoder, f file) {
	e.String(f.error())
	e.String(f.parseError())
	e.String(f.synopsis())
	e.String(f.name())
	e.String(f.pkgName())
	e.Bool(f.ignoreFile())
	e.Bool(f.binaryOnly())
	e.String(f.quotedImportComment())
	e.Uint32(uint32(f.quotedImportCommentLine()))
	e.String(f.goBuildConstraint())

	e.Uint32(uint32(len(f.plusBuildConstraints())))
	for _, s := range f.plusBuildConstraints() {
		e.String(s)
	}

	e.Uint32(uint32(len(f.imports())))
	for _, m := range f.imports() {
		e.String(m.path)
		e.String(m.doc) // TODO(matloob): only save for cgo?
		e.Position(m.position)
	}
	// TODO(matloob) produce the slice earlier

	e.Uint32(uint32(len(f.embeds())))
	for _, embed := range f.embeds() {
		e.String(embed.pattern)
		e.Position(embed.position)

	}
}

func newEncoder() *encoder {
	e := &encoder{strings: make(map[string]uint32)}

	// place the empty string at position 0 in the string table
	e.stringTable = append(e.stringTable, 0)
	e.strings[""] = 0

	return e
}

func (e *encoder) Position(position token.Position) {
	e.String(position.Filename)
	e.Uint32(uint32(position.Offset))
	e.Uint32(uint32(position.Line))
	e.Uint32(uint32(position.Column))
}

type encoder struct {
	b           []byte
	stringTable []byte
	strings     map[string]uint32
}

func (e *encoder) Pos() uint32 {
	return uint32(len(e.b))
}

func (e *encoder) Bytes(b []byte) {
	e.b = append(e.b, b...)
}

func (e *encoder) String(s string) {
	if n, ok := e.strings[s]; ok {
		e.Uint32(n)
		return
	}
	pos := uint32(len(e.stringTable))
	e.strings[s] = pos
	e.Uint32(pos)
	e.stringTable = append(e.stringTable, []byte(s)...)
	e.stringTable = append(e.stringTable, 0)
}

func (e *encoder) Bool(b bool) {
	if b {
		e.Uint32(1)
	} else {
		e.Uint32(0)
	}
}

func (e *encoder) Uint32(n uint32) {
	e.b = binary.LittleEndian.AppendUint32(e.b, n)
}

// There's got to be a better way to do this, right?
func (e *encoder) Uint32At(n uint32, at uint32) {
	binary.LittleEndian.PutUint32(e.b[at:], n)
}
