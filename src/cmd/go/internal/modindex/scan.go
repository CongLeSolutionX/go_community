package modindex

import (
	"cmd/go/internal/fsys"
	"cmd/go/internal/par"
	"fmt"
	"go/doc"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Enabled is used to flag off the behavior of the module index on tip.
// It will be removed before the release.
// TODO(matloob): Remove Enabled once we have more confidence on the
// module index.
var Enabled = os.Getenv("GOINDEX") == "true"

// indexModule creates and writes the module index file for the module at the given directory
// into the file at the given path.
func indexModule(modroot string) ([]byte, error) {
	var packages []*rawPackage
	err := filepath.WalkDir(modroot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		// Don't enter directories starting with "_" or "."
		_, elem := filepath.Split(path)
		if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") {
			return filepath.SkipDir
		}
		// stop at module boundaries
		if modroot != path {
			if fi, err := os.Stat(filepath.Join(path, "go.mod")); err == nil && !fi.IsDir() {
				return filepath.SkipDir
			}
		}
		// TODO(matloob): what do we do about symlinks
		rel, err := filepath.Rel(modroot, path)
		if err != nil {
			panic(err)
		}
		packages = append(packages, importRaw(modroot, rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return encodeModule(packages)
}

// rawPackage holds the information from each package that's needed to
// fill a build.Package once the context is available.
type rawPackage struct {
	error string
	dir   string // directory containing package sources

	// Source files
	sourceFiles []*rawFile
}

// rawFile is the struct representation of the file holding all
// information in its fields.
type rawFile struct {
	error      string
	parseError string

	name                    string
	synopsis                string // doc.Synopsis of package comment... Compute synopsis on all of these?
	pkgName                 string
	ignoreFile              bool // starts with _ or . or should otherwise always be ignored
	binaryOnly              bool // cannot be rebuilt from source (has //go:binary-only-package comment)
	quotedImportComment     string
	quotedImportCommentLine int
	goBuildConstraint       string
	plusBuildConstraints    []string
	imports                 []rawImport
	embeds                  []embed
}

type rawImport struct {
	path     string
	doc      string // The comment on import "C" when using cgo. Only set if path == "C".
	position token.Position
}

type embed struct {
	pattern  string
	position token.Position
}

var pkgcache par.Cache // for packages not in modcache

// importRaw fills the rawPackage from the package files in srcDir.
// dir is the package's path relative to the modroot.
func importRaw(modroot, reldir string) *rawPackage {
	p := &rawPackage{
		dir: reldir,
	}

	absdir := filepath.Join(modroot, reldir)

	// We still haven't checked
	// that p.dir directory exists. This is the right time to do that check.
	// We can't do it earlier, because we want to gather partial information for the
	// non-nil *Package returned when an error occurs.
	// We need to do this before we return early on FindOnly flag.
	if !isDir(absdir) {
		// package was not found
		p.error = fmt.Errorf("cannot find package in:\n\t%s", p.dir).Error()
		return p
	}

	entries, err := fsys.ReadDir(absdir)
	if err != nil {
		p.error = err.Error()
		return p
	}

	fset := token.NewFileSet()
	for _, d := range entries {
		if d.IsDir() {
			continue
		}
		if d.Mode()&fs.ModeSymlink != 0 {
			if isDir(filepath.Join(absdir, d.Name())) {
				// Symlinks to directories are not source files.
				continue
			}
		}

		name := d.Name()
		ext := nameExt(name)

		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		info, err := getFileInfo(absdir, name, fset)
		if err != nil {
			p.sourceFiles = append(p.sourceFiles, &rawFile{name: name, error: err.Error()})
			continue
		} else if info == nil {
			p.sourceFiles = append(p.sourceFiles, &rawFile{name: name, ignoreFile: true})
			continue
		}
		rf := &rawFile{
			name:                 name,
			goBuildConstraint:    info.goBuildConstraint,
			plusBuildConstraints: info.plusBuildConstraints,
			binaryOnly:           info.binaryOnly,
		}
		if info.parsed != nil {
			rf.pkgName = info.parsed.Name.Name
		}
		data := info.header

		// Going to save the file. For non-Go files, can stop here.
		p.sourceFiles = append(p.sourceFiles, rf)
		if ext != ".go" {
			continue
		}

		if info.parseErr != nil {
			rf.parseError = info.parseErr.Error()
			// Fall through: we might still have a partial AST in info.Parsed,
			// and we want to list files with parse errors anyway.
		}

		if info.parsed != nil && info.parsed.Doc != nil {
			rf.synopsis = doc.Synopsis(info.parsed.Doc.Text())
		}

		qcom, line := findImportComment(data)
		if line != 0 {
			rf.quotedImportComment = qcom
			rf.quotedImportCommentLine = line
		}

		for _, imp := range info.imports {
			var doc string
			if imp.path == "C" {
				doc = imp.doc.Text()
			}
			rf.imports = append(rf.imports, rawImport{path: imp.path, doc: doc, position: fset.Position(imp.pos)})
		}
		for _, emb := range info.embeds {
			rf.embeds = append(rf.embeds, embed{emb.pattern, emb.pos})
		}

	}
	return p
}
