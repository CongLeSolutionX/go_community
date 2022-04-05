package modindex

import (
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/cache"
	"cmd/go/internal/cfg"
	"cmd/go/internal/fsys"
	"cmd/go/internal/imports"
	"cmd/go/internal/par"
	"encoding/binary"
	"errors"
	"fmt"
	"go/build"
	"go/build/constraint"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unsafe"
)

// ModuleIndex represents and encoded module index file. It is used to
// do the equivalent of build.Import of packages in the module and answer other
// questions based on the index file's data.
type ModuleIndex struct {
	modroot  string
	od       offsetDecoder
	packages map[string]pkgInfo
}

var fcache par.Cache

func moduleHash(modPath string, isReadOnly bool) cache.ActionID {
	h := cache.NewHash("moduleIndex")
	fmt.Fprintf(h, "module index %s %v", indexVersion, modPath)
	if isReadOnly {
		return h.Sum()
	}
	// walkdir happens in deterministic order.
	filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if fi, err := os.Stat(filepath.Join(path, "go.mod")); path == modPath && err == nil && !fi.IsDir() {
				return filepath.SkipDir
			}
		}
		// Avoid .foo, _foo, and testdata subdirectory trees.
		_, elem := filepath.Split(path)
		if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		n, err := d.Info()
		if err != nil {
			base.Fatalf("go: %v", err)
		}
		fmt.Fprintln(h, "file %v %v\n", d.Name(), n.ModTime())
		return nil
	})
	return h.Sum()
}

// findModuleRoot finds the module root for dir by climbing up parent directories.
func findModuleRoot(dir string) (roots string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := fsys.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

// modroot returns the modroot for dir, by climbing up parent directories for most modules,
// or path manipulation in the module cache.
func modroot(dir string) (modroot string, isModCache, ok bool) {
	if !strings.HasPrefix(dir, cfg.GOMODCACHE+string(filepath.Separator)) {
		if !strings.HasPrefix(dir, cfg.GOROOT) {
			modroot = findModuleRoot(dir)

			return modroot, false, true
		}
		return "", false, false
		// TODO(matloob): add a case for stdlib here.
	}
	dir = dir[len(cfg.GOMODCACHE+string(filepath.Separator)):]
	at := strings.IndexRune(dir, '@')
	if at < 0 {
		return "", false, false
	}
	rest := dir[at+1:]
	if sepIndex := strings.IndexRune(rest, filepath.Separator); sepIndex >= 0 {
		rest = rest[:sepIndex]
	}

	modroot = cfg.GOMODCACHE + string(filepath.Separator) + dir[:at+1] + rest
	return modroot, true, true
}

// Get returns the ModuleIndex containing the directory dir in the module cache.
func Get(dir string) (*ModuleIndex, bool) {
	modroot, isModCache, ok := modroot(dir)
	if !ok {
		return nil, false
	}
	return openIndex(modroot, isModCache), true
}

func openIndex(modPath string, isReadOnly bool) *ModuleIndex {
	return fcache.Do(modPath, func() any {
		id := moduleHash(modPath, isReadOnly)
		data, _, err := cache.Default().GetMmap(id)
		if err != nil {
			// Couldn't read from modindex. cache.entryNotFoundError is unexported,
			// so just assume we couldn't read from the index because we haven't
			// indexed the module yet.
			data, err = indexModule(modPath)
			if err != nil {
				base.Fatalf("indexing module: %v", err)
			}
			if err = cache.Default().PutBytes(id, data); err != nil {
				base.Fatalf("writing module index to cache: %v", err)
			}
		}
		mi, err := fromBytes(modPath, data)
		if err != nil {
			base.Fatalf("index open %v", err)
		}
		return mi
	}).(*ModuleIndex)
}

// pkgInfo holds a pointer into the module index for a given package.
type pkgInfo struct {
	dir    string
	offset uint32
}

// fromBytes returns a *ModuleIndex given the encoded representation.
func fromBytes(moddir string, data []byte) (mi *ModuleIndex, err error) {
	if !Enabled {
		panic("use of index")
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	if gotVersion := string(data[:len(indexVersion)]); gotVersion != indexVersion {
		return nil, fmt.Errorf("bad index version string: %q", gotVersion)
	}
	stringTableOffset := binary.LittleEndian.Uint32(data[len(indexVersion):][:4])
	st := newStringTable(data[stringTableOffset:])
	d := decoder{data[len(indexVersion)+4:], st}
	numPackages := int(d.uint32())

	pkgInfos := make([]pkgInfo, numPackages)
	for i := 0; i < numPackages; i++ {
		pkgInfos[i].dir = d.string()
	}
	for i := 0; i < numPackages; i++ {
		pkgInfos[i].offset = d.uint32()
	}
	packages := make(map[string]pkgInfo)
	for i := range pkgInfos {
		packages[pkgInfos[i].dir] = pkgInfos[i]
	}

	return &ModuleIndex{
		moddir,
		offsetDecoder{data, st},
		packages,
	}, nil
}

func (mi *ModuleIndex) Packages() []string {
	var pkgs []string
	for p := range mi.packages {
		pkgs = append(pkgs, p)
	}
	return pkgs
}

// RelPath returns the path relative to the module's root.
func (mi *ModuleIndex) RelPath(path string) string {
	return filepath.Clean(strings.TrimPrefix(strings.TrimPrefix(path, mi.modroot), string(filepath.Separator)))
}

// ImportPackage is the equivalent of build.Import given the information in ModuleIndex.
// TODO(matloob): dir should be relative to mi. join dir with mi's dir for full directory
func (mi *ModuleIndex) Import(bctxt build.Context, relpath string, mode build.ImportMode) (p *build.Package, err error) {
	rp := mi.indexPackage(relpath)

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	ctxt := (*Context)(&bctxt)

	p = &build.Package{}

	p.ImportPath = "."
	p.Dir = filepath.Join(mi.modroot, rp.dir)
	if rp.error != "" {
		return p, errors.New(rp.error)
	}

	var pkgerr error
	switch ctxt.Compiler {
	case "gccgo":
	case "gc":
	default:
		// Save error for end of function.
		pkgerr = fmt.Errorf("import %q: unknown compiler %q", p.Dir, ctxt.Compiler)
	}

	if p.Dir == "" {
		return p, fmt.Errorf("import %q: import of unknown directory", p.Dir)
	}

	// Assumption: directory is in the module cache.

	if mode&build.FindOnly != 0 {
		return p, pkgerr
	}

	// We need to do a second round of bad file processing.
	var badGoError error
	badFiles := make(map[string]bool)
	badFile := func(name string, err error) {
		if badGoError == nil {
			badGoError = err
		}
		if !badFiles[name] {
			p.InvalidGoFiles = append(p.InvalidGoFiles, name)
			badFiles[name] = true
		}
	}

	var Sfiles []string // files with ".S"(capital S)/.sx(capital s equivalent for case insensitive filesystems)
	var firstFile string
	embedPos := make(map[string][]token.Position)
	testEmbedPos := make(map[string][]token.Position)
	xTestEmbedPos := make(map[string][]token.Position)
	importPos := make(map[string][]token.Position)
	testImportPos := make(map[string][]token.Position)
	xTestImportPos := make(map[string][]token.Position)
	allTags := make(map[string]bool)
	for _, tf := range rp.sourceFiles {
		name := tf.name()
		if error := tf.error(); error != "" {
			badFile(name, errors.New(tf.error()))
			continue
		} else if parseError := tf.parseError(); parseError != "" {
			badFile(name, errors.New(tf.parseError()))
			// Fall through: we might still have a partial AST in info.Parsed,
			// and we want to list files with parse errors anyway.
		}

		var shouldBuild = true
		if !ctxt.goodOSArchFile(name, allTags) && !ctxt.UseAllFiles {
			shouldBuild = false
		} else if goBuildConstraint := tf.goBuildConstraint(); goBuildConstraint != "" {
			x, err := constraint.Parse(goBuildConstraint)
			if err != nil {
				return p, fmt.Errorf("%s: parsing //go:build line: %v", name, err)
			}
			shouldBuild = ctxt.eval(x, allTags)
		} else if plusBuildConstraints := tf.plusBuildConstraints(); len(plusBuildConstraints) > 0 {
			for _, text := range plusBuildConstraints {
				if x, err := constraint.Parse(text); err == nil {
					if !ctxt.eval(x, allTags) {
						shouldBuild = false
					}
				}
			}
		}

		ext := nameExt(name)
		if !shouldBuild || tf.ignoreFile() {
			if ext == ".go" {
				p.IgnoredGoFiles = append(p.IgnoredGoFiles, name)
			} else if fileListForExt((*Package)(p), ext) != nil {
				p.IgnoredOtherFiles = append(p.IgnoredOtherFiles, name)
			}
			continue
		}

		// Going to save the file. For non-Go files, can stop here.
		switch ext {
		case ".go":
			// keep going
		case ".S", ".sx":
			// special case for cgo, handled at end
			Sfiles = append(Sfiles, name)
			continue
		default:
			if list := fileListForExt((*Package)(p), ext); list != nil {
				*list = append(*list, name)
			}
			continue
		}

		pkg := tf.pkgName()
		if pkg == "documentation" {
			p.IgnoredGoFiles = append(p.IgnoredGoFiles, name)
			continue
		}
		isTest := strings.HasSuffix(name, "_test.go")
		isXTest := false
		if isTest && strings.HasSuffix(tf.pkgName(), "_test") && p.Name != tf.pkgName() {
			isXTest = true
			pkg = pkg[:len(pkg)-len("_test")]
		}

		if !isTest && tf.binaryOnly() {
			p.BinaryOnly = true
		}

		// Grab the first package comment as docs, provided it is not from a test file.
		if p.Doc == "" && !isTest && !isXTest {
			if synopsis := tf.synopsis(); synopsis != "" {
				p.Doc = synopsis
			}
		}

		if p.Name == "" {
			p.Name = pkg
			firstFile = name
		} else if pkg != p.Name {
			// TODO(#45999): The choice of p.Name is arbitrary based on file iteration
			// order. Instead of resolving p.Name arbitrarily, we should clear out the
			// existing Name and mark the existing files as also invalid.
			badFile(name, &MultiplePackageError{
				Dir:      p.Dir,
				Packages: []string{p.Name, pkg},
				Files:    []string{firstFile, name},
			})
		}

		// Record Imports and information about cgo.
		isCgo := false
		imports := tf.imports()
		for _, imp := range imports {
			if imp.path == "C" {
				if isTest {
					badFile(name, fmt.Errorf("use of cgo in test %s not supported", name))
					continue
				}
				isCgo = true

				if imp.doc != "" {
					if err := ctxt.saveCgo(name, (*Package)(p), imp.doc); err != nil {
						badFile(name, err)
					}
				}

			}
		}

		var fileList *[]string
		var importMap, embedMap map[string][]token.Position
		switch {
		case isCgo:
			allTags["cgo"] = true
			if ctxt.CgoEnabled {
				fileList = &p.CgoFiles
				importMap = importPos
				embedMap = embedPos
			} else {
				// Ignore Imports and Embeds from cgo files if cgo is disabled.
				fileList = &p.IgnoredGoFiles
			}
		case isXTest:
			fileList = &p.XTestGoFiles
			importMap = xTestImportPos
			embedMap = xTestEmbedPos
		case isTest:
			fileList = &p.TestGoFiles
			importMap = testImportPos
			embedMap = testEmbedPos
		default:
			fileList = &p.GoFiles
			importMap = importPos
			embedMap = embedPos
		}
		*fileList = append(*fileList, name)
		if importMap != nil {
			for _, imp := range imports {
				importMap[imp.path] = append(importMap[imp.path], imp.position)
			}
		}
		if embedMap != nil {
			for _, e := range tf.embeds() {
				embedMap[e.pattern] = append(embedMap[e.pattern], e.position)
			}
		}
	}

	p.EmbedPatterns, p.EmbedPatternPos = cleanDecls(embedPos)
	p.TestEmbedPatterns, p.TestEmbedPatternPos = cleanDecls(testEmbedPos)
	p.XTestEmbedPatterns, p.XTestEmbedPatternPos = cleanDecls(xTestEmbedPos)

	p.Imports, p.ImportPos = cleanDecls(importPos)
	p.TestImports, p.TestImportPos = cleanDecls(testImportPos)
	p.XTestImports, p.XTestImportPos = cleanDecls(xTestImportPos)

	for tag := range allTags {
		p.AllTags = append(p.AllTags, tag)
	}
	sort.Strings(p.AllTags)

	if len(p.CgoFiles) > 0 {
		p.SFiles = append(p.SFiles, Sfiles...)
		sort.Strings(p.SFiles)
	} else {
		p.IgnoredOtherFiles = append(p.IgnoredOtherFiles, Sfiles...)
		sort.Strings(p.IgnoredOtherFiles)
	}

	if badGoError != nil {
		return p, badGoError
	}
	if len(p.GoFiles)+len(p.CgoFiles)+len(p.TestGoFiles)+len(p.XTestGoFiles) == 0 {
		return p, &NoGoError{p.Dir}
	}
	return p, pkgerr
}

// IsDirWithGoFiles is the equivalent of fsys.IsDirWithGoFiles using the information in the
// RawPackage.
func (mi *ModuleIndex) IsDirWithGoFiles(relpath string) (_ bool, err error) {
	rp := mi.indexPackage(relpath)

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()
	for _, sf := range rp.sourceFiles {
		if strings.HasSuffix(sf.name(), ".go") {
			return true, nil
		}
	}
	return false, nil
}

// ScanDir implements imports.ScanDir using the information in the RawPackage.
func (mi *ModuleIndex) ScanDir(path string, tags map[string]bool) (_ []string, _ []string, err error) {
	rp := mi.indexPackage(path)

	// TODO(matloob) dir should eventually be relative to indexed directory
	// TODO(matloob): skip reading raw package and jump straight to data we need?

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	imports_ := make(map[string]bool)
	testImports := make(map[string]bool)
	numFiles := 0

Files:
	for _, sf := range rp.sourceFiles {
		name := sf.name()
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".go") || !imports.MatchFile(name, tags) {
			continue
		}

		imps := sf.imports() // TODO(matloob): directly read import paths to avoid the extra strings?
		for _, imp := range imps {
			if imp.path == "C" && !tags["cgo"] && !tags["*"] {
				continue Files
			}
		}

		if !shouldBuild(sf, tags) {
			continue
		}
		numFiles++
		m := imports_
		if strings.HasSuffix(name, "_test.go") {
			m = testImports
		}
		for _, p := range imps {
			m[p.path] = true
		}
	}
	if numFiles == 0 {
		return nil, nil, imports.ErrNoGo
	}
	return keys(imports_), keys(testImports), nil
}

func keys(m map[string]bool) []string {
	list := make([]string, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

// implements imports.ShouldBuild in terms of an index sourcefile.
func shouldBuild(sf *sourceFile, tags map[string]bool) bool {
	if goBuildConstraint := sf.goBuildConstraint(); goBuildConstraint != "" {
		x, err := constraint.Parse(goBuildConstraint)
		if err != nil {
			return false
		}
		return imports.Eval(x, tags, true)
	}

	plusBuildConstraints := sf.plusBuildConstraints()
	for _, text := range plusBuildConstraints {
		if x, err := constraint.Parse(text); err == nil {
			if imports.Eval(x, tags, true) == false {
				return false
			}
		}
	}

	return true
}

// index package holds the information needed to access information in the
// index about a package.
type indexPackage struct {
	error string
	dir   string // directory of the package relative to the modroot

	// Source files
	sourceFiles []*sourceFile
}

// indexPackage returns an indexPackage constructed using the information in the ModuleIndex.
func (mi *ModuleIndex) indexPackage(path string) *indexPackage {
	defer func() {
		if e := recover(); e != nil {
			base.Fatalf("error reading module index: %v", e)
		}
	}()
	pkgData, ok := mi.packages[path]
	if !ok {
		return &indexPackage{error: "directory does not exist"}
	}

	// TODO(matloob): do we want to lock on the module index?
	d := mi.od.decoderAt(pkgData.offset)
	rp := new(indexPackage)
	rp.error = d.string()
	rp.dir = d.string()
	numSourceFiles := d.uint32()
	rp.sourceFiles = make([]*sourceFile, numSourceFiles)
	for i := uint32(0); i < numSourceFiles; i++ {
		offset := d.uint32()
		rp.sourceFiles[i] = &sourceFile{
			od: mi.od.offsetDecoderAt(offset),
		}
	}
	return rp
}

// sourceFile represents the information of a given source file in the module index.
type sourceFile struct {
	od offsetDecoder // od interprets all offsets relative to the start of the source file's data
}

// Offsets for fields in the sourceFile.
const (
	sourceFileError = 4 * iota
	sourceFileParseError
	sourceFileSynopsis
	sourceFileName
	sourceFilePkgName
	sourceFileIgnoreFile
	sourceFileBinaryOnly
	sourceFileGoBuildConstraint
	sourceFileNumPlusBuildConstraints
)

func (sf *sourceFile) error() string {
	return sf.od.stringAt(sourceFileError)
}
func (sf *sourceFile) parseError() string {
	return sf.od.stringAt(sourceFileParseError)
}
func (sf *sourceFile) name() string {
	return sf.od.stringAt(sourceFileName)
}
func (sf *sourceFile) synopsis() string {
	return sf.od.stringAt(sourceFileSynopsis)
}
func (sf *sourceFile) pkgName() string {
	return sf.od.stringAt(sourceFilePkgName)
}
func (sf *sourceFile) ignoreFile() bool {
	return sf.od.boolAt(sourceFileIgnoreFile)
}
func (sf *sourceFile) binaryOnly() bool {
	return sf.od.boolAt(sourceFileBinaryOnly)
}
func (sf *sourceFile) goBuildConstraint() string {
	return sf.od.stringAt(sourceFileGoBuildConstraint)
}

func (sf *sourceFile) plusBuildConstraints() []string {
	d := sf.od.decoderAt(sourceFileNumPlusBuildConstraints)
	n := int(d.uint32())
	ret := make([]string, n)
	for i := 0; i < n; i++ {
		ret[i] = d.string()
	}
	return ret
}

func importsOffset(numPlusBuildConstraints uint32) uint32 {
	// 4 bytes per uin32, add one to advance past numPlusBuildConstraints itself
	return sourceFileNumPlusBuildConstraints + 4*(numPlusBuildConstraints+1)
}

func (sf *sourceFile) importsOffset() uint32 {
	numPlusBuildConstraints := sf.od.uint32At(sourceFileNumPlusBuildConstraints)
	return importsOffset(numPlusBuildConstraints)
}

func embedsOffset(importsOffset, numImports uint32) uint32 {
	// 4 bytes per uint32; 1 to advance past numImports itself, and 6 uint32s per import
	return importsOffset + 4*(1+(6*numImports))
}

func (sf *sourceFile) embedsOffset() uint32 {
	importsOffset := sf.importsOffset()
	numImports := sf.od.uint32At(importsOffset)
	return embedsOffset(importsOffset, numImports)
}

func (sf *sourceFile) imports() []rawImport {
	importsOffset := sf.importsOffset()
	d := sf.od.decoderAt(importsOffset)
	numImports := int(d.uint32())
	ret := make([]rawImport, numImports)
	for i := 0; i < numImports; i++ {
		ret[i].path = d.string()
		ret[i].doc = d.string()
		ret[i].position = d.tokpos()
	}
	return ret
}

func (sf *sourceFile) embeds() []embed {
	embedsOffset := sf.embedsOffset()
	d := sf.od.decoderAt(embedsOffset)
	numEmbeds := int(d.uint32())
	ret := make([]embed, numEmbeds)
	for i := 0; i < numEmbeds; i++ {
		pattern := d.string()
		pos := d.tokpos()
		ret[i] = embed{pattern, pos}
	}
	return ret
}

// A decoder reads from the current position of the file and advances its position as it
// reads.
type decoder struct {
	b  []byte
	st *stringTable
}

func (d *decoder) uint32() uint32 {
	n := binary.LittleEndian.Uint32(d.b[:4])
	d.b = d.b[4:]
	return n
}

func (d *decoder) tokpos() token.Position {
	file := d.string()
	offset := int(d.uint32())
	line := int(d.uint32())
	column := int(d.uint32())
	return token.Position{
		Filename: file,
		Offset:   offset,
		Line:     line,
		Column:   column,
	}
}

func (d *decoder) string() string {
	return d.st.string(d.uint32())
}

// And offset decoder reads information offset from its position in the file.
// It's either offset from the beginning of the index, or the beginning of a sourceFile's data.
type offsetDecoder struct {
	b  []byte
	st *stringTable
}

func (od *offsetDecoder) uint32At(offset uint32) uint32 {
	return binary.LittleEndian.Uint32(od.b[offset:])
}

func (od *offsetDecoder) boolAt(offset uint32) bool {
	switch v := od.uint32At(offset); v {
	case 0:
		return false
	case 1:
		return true
	default:
		panic(fmt.Errorf("invalid bool value for SourceFile.IgnoreFile: %v", v))
	}
}

func (od *offsetDecoder) stringAt(offset uint32) string {
	return od.st.string(od.uint32At(offset))
}

func (od *offsetDecoder) decoderAt(offset uint32) *decoder {
	return &decoder{od.b[offset:], od.st}
}

func (od *offsetDecoder) offsetDecoderAt(offset uint32) offsetDecoder {
	return offsetDecoder{od.b[offset:], od.st}
}

type stringTable struct {
	b []byte
}

func newStringTable(b []byte) *stringTable {
	return &stringTable{b: b}
}

func (st *stringTable) string(pos uint32) string {
	if pos == 0 {
		return ""
	}

	bb := st.b[int(pos):]
	i := bytes.IndexByte(bb, 0)

	if i == -1 {
		panic("reached end of string table trying to read string")
	}
	s := asString(bb[:i])

	return s
}

func asString(b []byte) string {
	p := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(p)
	hdr.Len = len(b)

	return s
}
