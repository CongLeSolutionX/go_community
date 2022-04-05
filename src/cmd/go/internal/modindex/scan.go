package modindex

import (
	"cmd/go/internal/cfg"
	"cmd/go/internal/par"
	"fmt"
	"go/build"
	"go/doc"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
)

// Enabled is used to flag off the behavior of the module index on tip.
// It will be removed before the release.
// TODO(matloob): Remove Enabled once we have more confidence on the
// module index.
var Enabled = os.Getenv("GOINDEX") == "true"

// indexModule creates and writes the module index file for the module at the given directory
// into the file at the given path.
func indexModule(dir string) ([]byte, error) {
	var packages []*RawPackage
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		packages = append(packages, ImportRaw(path))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return encodeModule(packages, dir)
}

type RawModule struct {
	Dirs map[string]*RawPackage
}

// file abstracts over the information from each file that's needed
// to fill a build.Package once the context is available.
type file interface {
	error() string
	parseError() string
	name() string
	synopsis() string
	pkgName() string
	ignoreFile() bool
	binaryOnly() bool
	quotedImportComment() string
	quotedImportCommentLine() int
	goBuildConstraint() string
	plusBuildConstraints() []string

	imports() []rawImport
	embeds() []embed
}

// rawFile is the struct representation of the file holding all
// information in its fields.
type rawFile struct {
	Error      string
	ParseError string

	Name                    string
	Synopsis                string // doc.Synopsis of package comment... Compute synopsis on all of these?
	PkgName                 string
	IgnoreFile              bool // starts with _ or . or should otherwise always be ignored
	BinaryOnly              bool // cannot be rebuilt from source (has //go:binary-only-package comment)
	QuotedImportComment     string
	QuotedImportCommentLine int
	GoBuildConstraint       string
	PlusBuildConstraints    []string
	Imports                 []rawImport
	Embeds                  []embed
}

func (rf *rawFile) error() string                  { return rf.Error }
func (rf *rawFile) parseError() string             { return rf.ParseError }
func (rf *rawFile) name() string                   { return rf.Name }
func (rf *rawFile) synopsis() string               { return rf.Synopsis }
func (rf *rawFile) pkgName() string                { return rf.PkgName }
func (rf *rawFile) ignoreFile() bool               { return rf.IgnoreFile }
func (rf *rawFile) binaryOnly() bool               { return rf.BinaryOnly }
func (rf *rawFile) quotedImportComment() string    { return rf.QuotedImportComment }
func (rf *rawFile) quotedImportCommentLine() int   { return rf.QuotedImportCommentLine }
func (rf *rawFile) goBuildConstraint() string      { return rf.GoBuildConstraint }
func (rf *rawFile) plusBuildConstraints() []string { return rf.PlusBuildConstraints }
func (rf *rawFile) imports() []rawImport           { return rf.Imports }
func (rf *rawFile) embeds() []embed                { return rf.Embeds }

type rawImport struct {
	path     string
	doc      string // TODO(matloob): only save for cgo?
	position token.Position
}

type embed struct {
	pattern  string
	position token.Position
}

// RawPackage holds the information from each package that's needed to
// fill a build.Package once the context is available.
type RawPackage struct {
	error string

	path   string
	srcDir string // TODO(matloob): remove, replace with dir

	dir string // directory containing package sources

	// Source files
	sourceFiles []file
}

var pkgcache par.Cache // for packages not in modcache

// IndexedPackage returns the RawPackage for the directory, caching its work.
func IndexedPackage(dir string) *RawPackage {
	p := pkgcache.Do(dir, func() any {
		if cfg.BuildContext.GOROOT != "" {
			if _, ok := (*Context).hasSubdir((*Context)(&build.Default), cfg.BuildContext.GOROOT, dir); ok {
				return nil
			}
		} else if cfg.BuildContext.GOPATH != "" {
			return nil
		}
		// Assume package isn't in GOROOT or GOPATH. Should be filtered before here.
		// Change this?
		return ImportRaw(dir)
	})
	if p == nil {
		return nil
	}
	return p.(*RawPackage)
}

// ImportRaw fills the RawPackage from the package files in srcDir.
func ImportRaw(srcDir string) *RawPackage {
	path := "."
	p := &RawPackage{
		path:   path,
		srcDir: srcDir,
	}
	if path == "" {
		p.error = fmt.Errorf("import %q: invalid import Path", path).Error()
		return p
	}

	if !build.IsLocalImport(path) {
		panic(path)
	} else {
		if srcDir == "" {
			p.error = fmt.Errorf("import %q: import relative to unknown directory", path).Error()
			return p
		}
		if !filepath.IsAbs(path) {
			p.dir = filepath.Join(srcDir, path)
		}
	}

	// If it's a local import Path, by the time we get here, we still haven't checked
	// that p.dir directory exists. This is the right time to do that check.
	// We can't do it earlier, because we want to gather partial information for the
	// non-nil *Package returned when an error occurs.
	// We need to do this before we return early on FindOnly flag.
	if build.IsLocalImport(path) && !isDir(p.dir) {
		// package was not found
		p.error = fmt.Errorf("cannot find package %q in:\n\t%s", path, p.dir).Error()
		return p
	}

	dirs, err := os.ReadDir(p.dir)
	if err != nil {
		p.error = err.Error()
		return p
	}

	fset := token.NewFileSet()
	for _, d := range dirs {
		if d.IsDir() {
			continue
		}
		if d.Type()&fs.ModeSymlink != 0 {
			if isDir(filepath.Join(p.dir, d.Name())) {
				// Symlinks to directories are not source files.
				continue
			}
		}

		name := d.Name()
		ext := nameExt(name)

		info, err := getFileInfo(p.dir, name, fset)
		if err != nil {
			p.sourceFiles = append(p.sourceFiles, &rawFile{Name: name, Error: err.Error()})
			continue
		} else if info == nil {
			p.sourceFiles = append(p.sourceFiles, &rawFile{Name: name, IgnoreFile: true})
			continue
		}
		rf := &rawFile{
			Name:                 name,
			GoBuildConstraint:    info.goBuildConstraint,
			PlusBuildConstraints: info.plusBuildConstraints,
			BinaryOnly:           info.binaryOnly,
		}
		if info.parsed != nil {
			rf.PkgName = info.parsed.Name.Name
		}
		data := info.header

		// Going to save the file. For non-Go files, can stop here.
		p.sourceFiles = append(p.sourceFiles, rf)
		if ext != ".go" {
			continue
		}

		if info.parseErr != nil {
			rf.ParseError = info.parseErr.Error()
			// Fall through: we might still have a partial AST in info.Parsed,
			// and we want to list files with parse errors anyway.
		}

		if info.parsed != nil && info.parsed.Doc != nil {
			rf.Synopsis = doc.Synopsis(info.parsed.Doc.Text())
		}

		qcom, line := findImportComment(data)
		if line != 0 {
			rf.QuotedImportComment = qcom
			rf.QuotedImportCommentLine = line
		}

		for _, imp := range info.imports {
			rf.Imports = append(rf.Imports, rawImport{path: imp.path, doc: imp.doc.Text(), position: fset.Position(imp.pos)})
		}
		for _, emb := range info.embeds {
			rf.Embeds = append(rf.Embeds, embed{emb.pattern, emb.pos})
		}

	}
	return p
}
