package modindex

import (
	"bytes"
	"cmd/go/internal/imports"
	"encoding/binary"
	"errors"
	"fmt"
	"go/build"
	"go/build/constraint"
	"go/token"
	"internal/buildinternal"
	"io"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

type ModuleIndex struct {
	md       mmapData
	moddir   string
	st       *stringTable
	packages map[string]pkgInfo
}

type pkgInfo struct {
	dir        string
	offset     uint32
	rawPkgData *RawPackage2
}

func Open(path string) (mi *ModuleIndex, err error) {
	md := mmap(path)

	moddir := filepath.Dir(path)

	mi = &ModuleIndex{md: md, moddir: moddir}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	// TODO(matloob): clean this up
	const indexv0 = "go index v0\n"
	if string(md.d[:len(indexv0)]) != indexv0 {
		return nil, fmt.Errorf("bad index version string: %q", string(md.d[:len(indexv0)]))
	}
	ud := newUintDecoder(md.d[len(indexv0):])
	stringTableOffset := ud.uint32()
	mi.st = newStringTable(md.d[stringTableOffset:])
	d := &decoder{ud, mi.st}
	numPackages := int(d.uint32())

	pkgInfos := make([]pkgInfo, numPackages)

	for i := 0; i < numPackages; i++ {
		pkgInfos[i].dir = d.string()
	}
	for i := 0; i < numPackages; i++ {
		pkgInfos[i].offset = d.uint32()
	}
	mi.packages = make(map[string]pkgInfo)
	for i := range pkgInfos {
		mi.packages[pkgInfos[i].dir] = pkgInfos[i]
	}

	return mi, nil
}
func (mi *ModuleIndex) ImportPackage(bctxt build.Context, reldir string, mode build.ImportMode) (p *build.Package, err error) {
	ctxt := (*buildinternal.Context)(&bctxt)
	// TODO(matloob): dir should be relative to mi. join dir with mi's dir for full directory
	dir := reldir

	p = &build.Package{}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	rp, ok := mi.RawPackage(dir)
	if !ok {
		return &build.Package{
			ImportPath: ".",
			Dir:        dir,
		}, fmt.Errorf("cannot find package . in:\n\t%s", dir)
	}

	p.ImportPath = rp.Path
	p.Dir = rp.SrcDir
	if rp.Error != "" {
		return p, errors.New(rp.Error)
	}

	const path = "." // TODO(matloob): clean this up; ImportDir calls ctxt.Import with Path == "."
	srcDir := rp.SrcDir

	var pkgtargetroot string
	var pkga string
	var pkgerr error
	suffix := ""
	if ctxt.InstallSuffix != "" {
		suffix = "_" + ctxt.InstallSuffix
	}
	switch ctxt.Compiler {
	case "gccgo":
		pkgtargetroot = "pkg/gccgo_" + ctxt.GOOS + "_" + ctxt.GOARCH + suffix
	case "gc":
		pkgtargetroot = "pkg/" + ctxt.GOOS + "_" + ctxt.GOARCH + suffix
	default:
		// Save error for end of function.
		pkgerr = fmt.Errorf("import %q: unknown compiler %q", path, ctxt.Compiler)
	}
	setPkga := func() {
		switch ctxt.Compiler {
		case "gccgo":
			dir, elem := pathpkg.Split(p.ImportPath)
			pkga = pkgtargetroot + "/" + dir + "lib" + elem + ".a"
		case "gc":
			pkga = pkgtargetroot + "/" + p.ImportPath + ".a"
		}
	}
	setPkga()

	pkga = "" // local Imports have no installed Path
	if srcDir == "" {
		return p, fmt.Errorf("import %q: import relative to unknown directory", path)
	}
	if !isAbsPath(path) {
		p.Dir = joinPath(srcDir, path)
	}
	// p.Dir directory may or may not exist. Gather partial information first, check if it exists later.
	// Determine canonical import Path, if any.
	// Exclude results where the import Path would include /testdata/.

	// Assumption: directory is in the module cache.

	// It's okay that we didn't find a root containing dir.
	// Keep going with the information we have.

	if p.Root != "" {
		p.SrcRoot = joinPath(p.Root, "src")
		p.PkgRoot = joinPath(p.Root, "pkg")
		p.BinDir = joinPath(p.Root, "bin")
		if pkga != "" {
			p.PkgTargetRoot = joinPath(p.Root, pkgtargetroot)
			p.PkgObj = joinPath(p.Root, pkga)
		}
	}

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
	var firstFile, firstCommentFile string
	embedPos := make(map[string][]token.Position)
	testEmbedPos := make(map[string][]token.Position)
	xTestEmbedPos := make(map[string][]token.Position)
	importPos := make(map[string][]token.Position)
	testImportPos := make(map[string][]token.Position)
	xTestImportPos := make(map[string][]token.Position)
	allTags := make(map[string]bool)
	for _, tf := range rp.SourceFiles {
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
		if !buildinternal.GoodOSArchFile(ctxt, name, allTags) && !ctxt.UseAllFiles {
			shouldBuild = false
		} else if goBuildConstraint := tf.goBuildConstraint(); goBuildConstraint != "" {
			x, err := constraint.Parse(goBuildConstraint)
			if err != nil {
				return p, fmt.Errorf("%s: parsing //go:build line: %v", name, err)
			}
			shouldBuild = buildinternal.Eval(ctxt, x, allTags)
		} else if plusBuildConstraints := tf.plusBuildConstraints(); len(plusBuildConstraints) > 0 {
			for _, text := range plusBuildConstraints {
				if x, err := constraint.Parse(text); err == nil {
					if !buildinternal.Eval(ctxt, x, allTags) {
						shouldBuild = false
					}
				}
			}
		}

		ext := nameExt(name)
		if !shouldBuild || tf.ignoreFile() {
			if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
				// not due to build constraints - don't report
			} else if ext == ".go" {
				p.IgnoredGoFiles = append(p.IgnoredGoFiles, name)
			} else if fileListForExt(p, ext) != nil {
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
			if list := fileListForExt(p, ext); list != nil {
				*list = append(*list, name)
			}
			continue
		}

		// TODO(matloob): determine pkg Name here? pkg variable

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

		if mode&build.ImportComment != 0 {
			com, err := strconv.Unquote(tf.quotedImportComment())
			if err != nil {
				badFile(name, fmt.Errorf("%s:%d: cannot parse import comment", name, tf.quotedImportCommentLine()))
			} else if p.ImportComment == "" {
				p.ImportComment = com
				firstCommentFile = name
			} else if p.ImportComment != com {
				badFile(name, fmt.Errorf("found import comments %q (%s) and %q (%s) in %s", p.ImportComment, firstCommentFile, com, name, p.Dir))
			}
		}

		// Record Imports and information about cgo.
		isCgo := false
		imports := tf.imports()
		for _, imp := range imports {
			if imp.Path == "C" {
				if isTest {
					badFile(name, fmt.Errorf("use of cgo in test %s not supported", name))
					continue
				}
				isCgo = true

				if imp.Doc != "" {
					if err := buildinternal.SaveCgo(ctxt, name, (*buildinternal.Package)(p), imp.Doc); err != nil {
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
				importMap[imp.Path] = append(importMap[imp.Path], imp.Position)
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
	// TODO Remove SFiles if we're not using cgo.

	if badGoError != nil {
		return p, badGoError
	}
	if len(p.GoFiles)+len(p.CgoFiles)+len(p.TestGoFiles)+len(p.XTestGoFiles) == 0 {
		return p, &NoGoError{p.Dir}
	}
	return p, pkgerr
}

func (mi *ModuleIndex) IsDirWithGoFiles(dir string) (_ bool, err error) {
	rp, ok := mi.RawPackage(dir)
	if !ok {
		return false, nil
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()
	for _, sf := range rp.SourceFiles {
		if strings.HasSuffix(sf.name(), ".go") {
			return true, nil
		}
	}
	return false, nil
}

// Implements imports.ScanDir in terms of module index.
func (mi *ModuleIndex) ScanDir(dir string, tags map[string]bool) (_ []string, _ []string, err error) {
	// TODO(matloob) dir should eventually be relative to indexed directory
	// TODO(matloob): skip reading raw package and jump straight to data we need?
	rp, ok := mi.RawPackage(dir)
	if !ok {
		panic("should this ever happen?")
	}

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	imports_ := make(map[string]bool)
	testImports := make(map[string]bool)
	numFiles := 0

Files:
	for _, sf := range rp.SourceFiles {
		name := sf.name()
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".go") || !imports.MatchFile(name, tags) {
			continue
		}

		imps := sf.imports() // TODO(matloob): directly read import paths to avoid the extra strings?
		for _, imp := range imps {
			if imp.Path == "C" && !tags["cgo"] && !tags["*"] {
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
			m[p.Path] = true
		}
	}
	if numFiles == 0 {
		return nil, nil, imports.ErrNoGo
	}
	return keys(imports_), keys(testImports), nil
}

func keys(m map[string]bool) []string {
	var list []string
	for k := range m {
		list = append(list, k)
	}
	sort.Strings(list)
	return list
}

// Implements imports.ShouldBuild in terms of an index sourcefile.
func shouldBuild(sf SourceFile, tags map[string]bool) bool {
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

///// TODO(matloob) delete all this stuff if we end up merging back into go/build

// joinPath calls joinPath (if not nil) or else filepath.Join.
func joinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// isAbsPath calls isAbsPath (if not nil) or else filepath.IsAbs.
func isAbsPath(path string) bool {
	return filepath.IsAbs(path)
}

func (mi *ModuleIndex) RawPackage(path string) (*RawPackage2, bool) {
	pkgData, ok := mi.packages[path]
	if !ok {
		return nil, false
	}
	rp := new(RawPackage2)
	// TODO(matloob): do we want to lock on the module index?
	d := mi.newDecoder(pkgData.offset)
	rp.Error = d.string()
	rp.Path = d.string()
	rp.SrcDir = d.string()
	rp.Dir = d.string()
	numSourceFiles := d.uint32()
	rp.SourceFiles = make([]SourceFile, numSourceFiles)
	for i := uint32(0); i < numSourceFiles; i++ {
		rp.SourceFiles[i].mi = mi
		rp.SourceFiles[i].offset = d.uint32()
	}
	return rp, true
}

type RawPackage2 struct {
	// TODO(matloob): Do we need AllTags in RawPackage?
	// We can produce it from contstraints when we evaluate them.

	Error string

	// Arguments to build.Import. Is Path always "."?
	Path   string
	SrcDir string

	Dir string // directory containing package sources

	// Source files
	SourceFiles []SourceFile

	// No ConflictDir-- only relevant togopath
}

type SourceFile struct {
	mi *ModuleIndex // index file. TODO(matloob): make a specific decoder type?

	offset uint32
}

const (
	sourceFileErrorOffset = 4 * iota
	sourceFileParseError
	sourceFileSynopsis
	sourceFileName
	sourceFilePkgName
	sourceFileIgnoreFile
	sourceFileBinaryOnly
	sourceFileQuotedImportComment
	sourceFileQuotedImportCommentLine
	sourceFileGoBuildConstraint
	sourceFileNumPlusBuildConstraints
)

func (sf *SourceFile) error() string {
	return sf.mi.stringAt(sf.offset + sourceFileErrorOffset)
}

func (sf *SourceFile) parseError() string {
	return sf.mi.stringAt(sf.offset + sourceFileParseError)
}

func (sf *SourceFile) name() string {
	return sf.mi.stringAt(sf.offset + sourceFileName)
}

func (sf *SourceFile) synopsis() string {
	return sf.mi.stringAt(sf.offset + sourceFileSynopsis)
}

func (sf *SourceFile) pkgName() string {
	return sf.mi.stringAt(sf.offset + sourceFilePkgName)
}

func (sf *SourceFile) ignoreFile() bool {
	return sf.mi.boolAt(sf.offset + sourceFileIgnoreFile)
}

func (sf *SourceFile) binaryOnly() bool {
	return sf.mi.boolAt(sf.offset + sourceFileBinaryOnly)
}

func (sf *SourceFile) quotedImportComment() string {
	return sf.mi.stringAt(sf.offset + sourceFileQuotedImportComment)
}

func (sf *SourceFile) quotedImportCommentLine() int {
	return int(sf.mi.uint32At(sf.offset + sourceFileQuotedImportCommentLine))
}

func (sf *SourceFile) goBuildConstraint() string {
	return sf.mi.stringAt(sf.offset + sourceFileGoBuildConstraint)
}

func (sf *SourceFile) plusBuildConstraints() []string {
	var ret []string

	d := sf.mi.newDecoder(sf.offset + sourceFileNumPlusBuildConstraints)
	n := int(d.uint32())
	for i := 0; i < n; i++ {
		ret = append(ret, d.string())
	}
	return ret
}

func importsOffset(sfOffset, numPlusBuildConstraints uint32) uint32 {
	// 4 bytes per uin32, add one to advance past numPlusBuildConstraints itself
	return sfOffset + sourceFileNumPlusBuildConstraints + 4*(numPlusBuildConstraints+1)
}

func (sf *SourceFile) importsOffset() uint32 {
	numPlusBuildConstraints := sf.mi.uint32At(sf.offset + sourceFileNumPlusBuildConstraints)
	return importsOffset(sf.offset, numPlusBuildConstraints)
}

func embedsOffset(importsOffset, numImports uint32) uint32 {
	// 4 bytes per uint32; 1 to advance past numImports itself, and 6 uint32s per import
	return importsOffset + 4*(1+(6*numImports))
}

func (sf *SourceFile) embedsOffset() uint32 {
	importsOffset := sf.importsOffset()
	numImports := sf.mi.newDecoder(importsOffset).uint32()
	return embedsOffset(importsOffset, numImports)
}

func (sf *SourceFile) imports() []TFImport {
	var ret []TFImport

	importsOffset := sf.importsOffset()
	d := sf.mi.newDecoder(importsOffset)
	numImports := int(d.uint32())
	for i := 0; i < numImports; i++ {
		path := d.string()
		doc := d.string()
		pos := d.tokpos()
		ret = append(ret, TFImport{
			Path:     path,
			Doc:      doc, // TODO(matloob): only save for cgo?
			Position: pos,
		})
	}
	return ret
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

func (sf *SourceFile) embeds() []embed {
	var ret []embed

	embedsOffset := sf.embedsOffset()
	d := sf.mi.newDecoder(embedsOffset)
	numEmbeds := int(d.uint32())
	for i := 0; i < numEmbeds; i++ {
		pattern := d.string()
		pos := d.tokpos()
		ret = append(ret, embed{pattern, pos})
	}
	return ret
}

func (mi *ModuleIndex) newDecoder(offset uint32) *decoder {
	return &decoder{newUintDecoder(mi.md.d[offset:]), mi.st}
}

func newUintDecoder(b []byte) uintDecoder {
	return uintDecoder{bytes.NewReader(b)}
}

type decoder struct {
	uintDecoder
	st *stringTable
}

type uintDecoder struct {
	r io.Reader
}

func (d uintDecoder) uint32() uint32 {
	var n uint32
	if err := binary.Read(d.r, binary.LittleEndian, &n); err != nil {
		panic(err)
	}
	return n
}

func (d *decoder) string() string {
	return d.st.String(d.uint32())
}

func (mi *ModuleIndex) uint32At(offset uint32) uint32 {
	return mi.newDecoder(offset).uint32()
}

func (mi *ModuleIndex) boolAt(offset uint32) bool {
	switch v := mi.uint32At(offset); v {
	case 0:
		return false
	case 1:
		return true
	default:
		panic(fmt.Errorf("invalid bool value for SourceFile.IgnoreFile:", v))
	}
}

type stringTable struct {
	b       []byte
	mu      sync.Mutex
	strings map[uint32]string
}

func (mi *ModuleIndex) stringAt(offset uint32) string {
	return mi.newDecoder(offset).string()
}

// TODO(matloob): is it ok to read the entire string table? Should we read strings directly
// from the file?

func newStringTable(b []byte) *stringTable {
	return &stringTable{b: b, strings: make(map[uint32]string)}
}

func (st *stringTable) String(pos uint32) string {
	//	st.mu.Lock()
	//	defer st.mu.Unlock()

	if pos == 0 {
		return ""
	}
	//	if s, ok := st.strings[pos]; ok {
	//		return s
	//	}
	i := int(pos)
	for ; i < len(st.b); i++ {
		c := st.b[i]
		if c == 0 {
			break
		}
	}
	s := AsString(st.b[int(pos):i])
	if i == len(st.b) {
		panic("reached end of string table trying to read string")
	}
	//	st.strings[pos] = s
	return s
}

func AsString(b []byte) string {
	p := unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(&b)).Data)

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(p)
	hdr.Len = len(b)

	return s
}
