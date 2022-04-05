package modindex

import (
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/imports"
	"cmd/go/internal/par"
	"encoding/binary"
	"errors"
	"fmt"
	"go/build"
	"go/build/constraint"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unsafe"
)

// ModuleIndex represents and encoded module index file. It is used to
// do the equivalent of build.Import of packages in the module and answer other
// questions based on the index file's data.
type ModuleIndex struct {
	moddir   string
	od       offsetDecoder
	packages map[string]pkgInfo
}

var fcache par.Cache

func openIndex(indexPath string) *ModuleIndex {
	return fcache.Do(indexPath, func() any {
		mi, err := Open(indexPath) // TODO(matloob) is this the rightp ath
		if err != nil {
			base.Fatalf("index open %v", err)
		}
		return mi
	}).(*ModuleIndex)
}

// Get returns the ModuleIndex containing the directory dir in the module cache.
func Get(dir string) (*ModuleIndex, bool) {
	if !strings.HasPrefix(dir, cfg.GOMODCACHE+string(filepath.Separator)) {
		return nil, false
	}
	dir = dir[len(cfg.GOMODCACHE+string(filepath.Separator)):]
	at := strings.IndexRune(dir, '@')
	if at < 0 {
		return nil, false
	}
	modpathenc := dir[:at]
	rest := dir[at+1:]
	sepIndex := strings.IndexRune(rest, filepath.Separator)
	if sepIndex < 0 {
		// No separator means that the directory is at the top-level
		sepIndex = len(rest)
	}
	encVer := rest[:sepIndex]
	indexPath := filepath.Join(cfg.GOMODCACHE, "cache/download", modpathenc, "@v", encVer+".index")
	// TODO(matloob): ok to assume index exists at this point?
	return openIndex(indexPath), true
}

// pkgInfo holds a pointer into the module index for a given package.
type pkgInfo struct {
	dir        string
	offset     uint32
	rawPkgData *RawPackage
}

// useindex is used to flag off the behavior of the module index on tip.
// It will be removed before the release.
// TODO(matloob): Remove useindex once we have more confidence on the
// module index.
var useindex = os.Getenv("GOINDEX") == "true"

// Open opens a module index from disk.
func Open(path string) (mi *ModuleIndex, err error) {
	if !useindex {
		panic("use of index")
	}

	md := mmap(path)

	moddir := filepath.Dir(path)

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
	d := decoder{b: md.d[len(indexv0):]}
	stringTableOffset := d.uint32()
	st := newStringTable(md.d[stringTableOffset:])
	d.st = st
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
		offsetDecoder{md.d, st},
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

// ImportPackage is the equivalent of build.Import given the information in the RawPackage.
// TODO(matloob): dir should be relative to mi. join dir with mi's dir for full directory
func (rp *RawPackage) Import(bctxt build.Context, mode build.ImportMode) (p *build.Package, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error reading module index: %v", e)
		}
	}()

	ctxt := (*Context)(&bctxt)

	p = &build.Package{}

	p.ImportPath = rp.path
	p.Dir = rp.srcDir
	if rp.error != "" {
		return p, errors.New(rp.error)
	}

	const path = "." // TODO(matloob): clean this up; ImportDir calls ctxt.Import with Path == "."
	srcDir := rp.srcDir

	var pkgerr error
	switch ctxt.Compiler {
	case "gccgo":
	case "gc":
	default:
		// Save error for end of function.
		pkgerr = fmt.Errorf("import %q: unknown compiler %q", path, ctxt.Compiler)
	}

	if srcDir == "" {
		return p, fmt.Errorf("import %q: import relative to unknown directory", path)
	}
	if !filepath.IsAbs(path) {
		p.Dir = filepath.Join(srcDir, path)
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
	var firstFile, firstCommentFile string
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
			if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
				// not due to build constraints - don't report
			} else if ext == ".go" {
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
func (rp *RawPackage) IsDirWithGoFiles() (_ bool, err error) {
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
func (rp *RawPackage) ScanDir(tags map[string]bool) (_ []string, _ []string, err error) {
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
func shouldBuild(sf file, tags map[string]bool) bool {
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

// RawPackage returns a RawPackage constructed using the information in the ModuleIndex.
func (mi *ModuleIndex) RawPackage(path string) *RawPackage {
	defer func() {
		if e := recover(); e != nil {
			base.Fatalf("error reading module index: %v", e)
		}
	}()
	pkgData, ok := mi.packages[path]
	if !ok {
		return nil
	}
	rp := new(RawPackage)
	// TODO(matloob): do we want to lock on the module index?
	d := mi.od.decoderAt(pkgData.offset)
	rp.error = d.string()
	rp.path = d.string()
	rp.srcDir = d.string()
	rp.dir = d.string()
	numSourceFiles := d.uint32()
	rp.sourceFiles = make([]file, numSourceFiles)
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
	sourceFileQuotedImportComment
	sourceFileQuotedImportCommentLine
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
func (sf *sourceFile) quotedImportComment() string {
	return sf.od.stringAt(sourceFileQuotedImportComment)
}
func (sf *sourceFile) quotedImportCommentLine() int {
	return int(sf.od.uint32At(sourceFileQuotedImportCommentLine))
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
