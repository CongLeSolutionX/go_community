// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modload

import (
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/imports"
	"cmd/go/internal/modfetch"
	"cmd/go/internal/mvs"
	"cmd/go/internal/par"
	"cmd/go/internal/search"
	"cmd/go/internal/str"
	"errors"
	"fmt"
	"go/build"
	"os"
	"path"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"golang.org/x/mod/module"
)

// buildList is the list of modules to use for building packages.
// It is initialized by calling ImportPaths, ImportFromFiles,
// LoadALL, or LoadBuildList, each of which uses loaded.load.
//
// Ideally, exactly ONE of those functions would be called,
// and exactly once. Most of the time, that's true.
// During "go get" it may not be. TODO(rsc): Figure out if
// that restriction can be established, or else document why not.
//
var buildList []module.Version

// loaded is the most recently-used package loader.
// It holds details about individual packages.
var loaded *loader

// ImportPaths returns the set of packages matching the args (patterns),
// on the target platform. Modules may be added to the build list
// to satisfy new imports.
func ImportPaths(patterns []string) []*search.Match {
	matches := ImportPathsQuiet(patterns, imports.Tags())
	search.WarnUnmatched(matches)
	return matches
}

// ImportPathsQuiet is like ImportPaths but does not warn about patterns with
// no matches. It also lets the caller specify a set of build tags to match
// packages. The build tags should typically be imports.Tags() or
// imports.AnyTags(); a nil map has no special meaning.
func ImportPathsQuiet(patterns []string, tags map[string]bool) []*search.Match {
	InitMod()

	var (
		matches  []*search.Match
		allLevel allLevel
	)
	for _, pattern := range search.CleanPatterns(patterns) {
		matches = append(matches, search.NewMatch(pattern))
		if pattern == "all" {
			allLevel = importedByTransitiveTestFromTarget
		}
	}

	updateMatches := func(ld *loader) {
		for _, m := range matches {
			switch {
			case m.IsLocal():
				// Evaluate list of file system directories on first iteration.
				if m.Dirs == nil {
					matchLocalDirs(m)
				}

				// Make a copy of the directory list and translate to import paths.
				// Note that whether a directory corresponds to an import path
				// changes as the build list is updated, and a directory can change
				// from not being in the build list to being in it and back as
				// the exact version of a particular module increases during
				// the loader iterations.
				m.Pkgs = m.Pkgs[:0]
				for _, dir := range m.Dirs {
					pkg, err := resolveLocalPackage(dir)
					if err != nil {
						if !m.IsLiteral() && (err == errPkgIsBuiltin || err == errPkgIsGorootSrc) {
							continue // Don't include "builtin" or GOROOT/src in wildcard patterns.
						}

						// If we're outside of a module, ensure that the failure mode
						// indicates that.
						ModRoot()

						if ld != nil {
							m.AddError(err)
						}
						continue
					}
					m.Pkgs = append(m.Pkgs, pkg)
				}

			case m.IsLiteral():
				m.Pkgs = []string{m.Pattern()}

			case strings.Contains(m.Pattern(), "..."):
				m.Errs = m.Errs[:0]
				matchPackages(m, tags, includeStd, buildList)

			case m.Pattern() == "all":
				if ld == nil {
					// The initial roots are the packages in the main module.
					// loadFromRoots will expand that to "all".
					m.Errs = m.Errs[:0]
					matchPackages(m, tags, omitStd, []module.Version{Target})
				} else {
					// Starting with the packages in the main module,
					// enumerate the full list of "all".
					m.Pkgs = ld.computePatternAll()
				}

			case m.Pattern() == "std" || m.Pattern() == "cmd":
				if m.Pkgs == nil {
					m.MatchPackages() // Locate the packages within GOROOT/src.
				}

			default:
				panic(fmt.Sprintf("internal error: modload missing case for pattern %s", m.Pattern()))
			}
		}
	}

	loaded = loadFromRoots(tags, allLevel, func() []string {
		updateMatches(nil)
		var roots []string
		for _, m := range matches {
			roots = append(roots, m.Pkgs...)
		}
		return roots
	})

	// One last pass to finalize wildcards.
	updateMatches(loaded)
	checkMultiplePaths()
	WriteGoMod()

	return matches
}

// checkMultiplePaths verifies that a given module path is used as itself
// or as a replacement for another module, but not both at the same time.
//
// (See https://golang.org/issue/26607 and https://golang.org/issue/34650.)
func checkMultiplePaths() {
	firstPath := make(map[module.Version]string, len(buildList))
	for _, mod := range buildList {
		src := mod
		if rep := Replacement(mod); rep.Path != "" {
			src = rep
		}
		if prev, ok := firstPath[src]; !ok {
			firstPath[src] = mod.Path
		} else if prev != mod.Path {
			base.Errorf("go: %s@%s used for two different module paths (%s and %s)", src.Path, src.Version, prev, mod.Path)
		}
	}
	base.ExitIfErrors()
}

// matchLocalDirs is like m.MatchDirs, but tries to avoid scanning directories
// outside of the standard library and active modules.
func matchLocalDirs(m *search.Match) {
	if !m.IsLocal() {
		panic(fmt.Sprintf("internal error: resolveLocalDirs on non-local pattern %s", m.Pattern()))
	}

	if i := strings.Index(m.Pattern(), "..."); i >= 0 {
		// The pattern is local, but it is a wildcard. Its packages will
		// only resolve to paths if they are inside of the standard
		// library, the main module, or some dependency of the main
		// module. Verify that before we walk the filesystem: a filesystem
		// walk in a directory like /var or /etc can be very expensive!
		dir := filepath.Dir(filepath.Clean(m.Pattern()[:i+3]))
		absDir := dir
		if !filepath.IsAbs(dir) {
			absDir = filepath.Join(base.Cwd, dir)
		}
		if search.InDir(absDir, cfg.GOROOTsrc) == "" && search.InDir(absDir, ModRoot()) == "" && pathInModuleCache(absDir) == "" {
			m.Dirs = []string{}
			m.AddError(fmt.Errorf("directory prefix %s outside available modules", base.ShortPath(absDir)))
			return
		}
	}

	m.MatchDirs()
}

// resolveLocalPackage resolves a filesystem path to a package path.
func resolveLocalPackage(dir string) (string, error) {
	var absDir string
	if filepath.IsAbs(dir) {
		absDir = filepath.Clean(dir)
	} else {
		absDir = filepath.Join(base.Cwd, dir)
	}

	bp, err := cfg.BuildContext.ImportDir(absDir, 0)
	if err != nil && (bp == nil || len(bp.IgnoredGoFiles) == 0) {
		// golang.org/issue/32917: We should resolve a relative path to a
		// package path only if the relative path actually contains the code
		// for that package.
		//
		// If the named directory does not exist or contains no Go files,
		// the package does not exist.
		// Other errors may affect package loading, but not resolution.
		if _, err := os.Stat(absDir); err != nil {
			if os.IsNotExist(err) {
				// Canonicalize OS-specific errors to errDirectoryNotFound so that error
				// messages will be easier for users to search for.
				return "", &os.PathError{Op: "stat", Path: absDir, Err: errDirectoryNotFound}
			}
			return "", err
		}
		if _, noGo := err.(*build.NoGoError); noGo {
			// A directory that does not contain any Go source files — even ignored
			// ones! — is not a Go package, and we can't resolve it to a package
			// path because that path could plausibly be provided by some other
			// module.
			//
			// Any other error indicates that the package “exists” (at least in the
			// sense that it cannot exist in any other module), but has some other
			// problem (such as a syntax error).
			return "", err
		}
	}

	if modRoot != "" && absDir == modRoot {
		if absDir == cfg.GOROOTsrc {
			return "", errPkgIsGorootSrc
		}
		return targetPrefix, nil
	}

	// Note: The checks for @ here are just to avoid misinterpreting
	// the module cache directories (formerly GOPATH/src/mod/foo@v1.5.2/bar).
	// It's not strictly necessary but helpful to keep the checks.
	if modRoot != "" && strings.HasPrefix(absDir, modRoot+string(filepath.Separator)) && !strings.Contains(absDir[len(modRoot):], "@") {
		suffix := filepath.ToSlash(absDir[len(modRoot):])
		if strings.HasPrefix(suffix, "/vendor/") {
			if cfg.BuildMod != "vendor" {
				return "", fmt.Errorf("without -mod=vendor, directory %s has no package path", absDir)
			}

			readVendorList()
			pkg := strings.TrimPrefix(suffix, "/vendor/")
			if _, ok := vendorPkgModule[pkg]; !ok {
				return "", fmt.Errorf("directory %s is not a package listed in vendor/modules.txt", absDir)
			}
			return pkg, nil
		}

		if targetPrefix == "" {
			pkg := strings.TrimPrefix(suffix, "/")
			if pkg == "builtin" {
				// "builtin" is a pseudo-package with a real source file.
				// It's not included in "std", so it shouldn't resolve from "."
				// within module "std" either.
				return "", errPkgIsBuiltin
			}
			return pkg, nil
		}

		pkg := targetPrefix + suffix
		if _, ok, err := dirInModule(pkg, targetPrefix, modRoot, true); err != nil {
			return "", err
		} else if !ok {
			return "", &PackageNotInModuleError{Mod: Target, Pattern: pkg}
		}
		return pkg, nil
	}

	if sub := search.InDir(absDir, cfg.GOROOTsrc); sub != "" && sub != "." && !strings.Contains(sub, "@") {
		pkg := filepath.ToSlash(sub)
		if pkg == "builtin" {
			return "", errPkgIsBuiltin
		}
		return pkg, nil
	}

	pkg := pathInModuleCache(absDir)
	if pkg == "" {
		return "", fmt.Errorf("directory %s outside available modules", base.ShortPath(absDir))
	}
	return pkg, nil
}

var (
	errDirectoryNotFound = errors.New("directory not found")
	errPkgIsGorootSrc    = errors.New("GOROOT/src is not an importable package")
	errPkgIsBuiltin      = errors.New(`"builtin" is a pseudo-package, not an importable package`)
)

// pathInModuleCache returns the import path of the directory dir,
// if dir is in the module cache copy of a module in our build list.
func pathInModuleCache(dir string) string {
	for _, m := range buildList[1:] {
		var root string
		var err error
		if repl := Replacement(m); repl.Path != "" && repl.Version == "" {
			root = repl.Path
			if !filepath.IsAbs(root) {
				root = filepath.Join(ModRoot(), root)
			}
		} else if repl.Path != "" {
			root, err = modfetch.DownloadDir(repl)
		} else {
			root, err = modfetch.DownloadDir(m)
		}
		if err != nil {
			continue
		}
		if sub := search.InDir(dir, root); sub != "" {
			sub = filepath.ToSlash(sub)
			if !strings.Contains(sub, "/vendor/") && !strings.HasPrefix(sub, "vendor/") && !strings.Contains(sub, "@") {
				return path.Join(m.Path, filepath.ToSlash(sub))
			}
		}
	}
	return ""
}

// ImportFromFiles adds modules to the build list as needed
// to satisfy the imports in the named Go source files.
func ImportFromFiles(gofiles []string) {
	InitMod()

	tags := imports.Tags()
	imports, testImports, err := imports.ScanFiles(gofiles, tags)
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	loaded = loadFromRoots(tags, noAll, func() []string {
		var roots []string
		roots = append(roots, imports...)
		roots = append(roots, testImports...)
		return roots
	})
	WriteGoMod()
}

// DirImportPath returns the effective import path for dir,
// provided it is within the main module, or else returns ".".
func DirImportPath(dir string) string {
	if modRoot == "" {
		return "."
	}

	if !filepath.IsAbs(dir) {
		dir = filepath.Join(base.Cwd, dir)
	} else {
		dir = filepath.Clean(dir)
	}

	if dir == modRoot {
		return targetPrefix
	}
	if strings.HasPrefix(dir, modRoot+string(filepath.Separator)) {
		suffix := filepath.ToSlash(dir[len(modRoot):])
		if strings.HasPrefix(suffix, "/vendor/") {
			return strings.TrimPrefix(suffix, "/vendor/")
		}
		return targetPrefix + suffix
	}
	return "."
}

// LoadBuildList loads and returns the build list from go.mod.
// The loading of the build list happens automatically in ImportPaths:
// LoadBuildList need only be called if ImportPaths is not
// (typically in commands that care about the module but
// no particular package).
func LoadBuildList() []module.Version {
	InitMod()
	ReloadBuildList()
	WriteGoMod()
	return buildList
}

// ReloadBuildList resets the state of loaded packages, then loads and returns
// the build list set in SetBuildList.
func ReloadBuildList() []module.Version {
	loaded = loadFromRoots(imports.Tags(), noAll, func() []string { return nil })
	return buildList
}

// LoadALL returns the set of all packages in the current module
// and their dependencies in any other modules, without filtering
// due to build tags, except "+build ignore".
// It adds modules to the build list as needed to satisfy new imports.
// This set is useful for deciding whether a particular import is needed
// anywhere in a module.
func LoadALL() []string {
	InitMod()
	return loadAll(importedByTransitiveTestFromTarget)
}

// LoadVendor is like LoadALL but only follows test dependencies
// for tests in the main module. Tests in dependency modules are
// ignored completely.
// This set is useful for identifying the which packages to include in a vendor directory.
func LoadVendor() []string {
	InitMod()
	return loadAll(importedByTarget)
}

func loadAll(level allLevel) []string {
	inTarget := TargetPackages("...")
	loaded = loadFromRoots(imports.AnyTags(), level, func() []string { return inTarget.Pkgs })
	checkMultiplePaths()
	WriteGoMod()

	var paths []string
	for _, pkg := range loaded.pkgs {
		if pkg.err != nil {
			base.Errorf("%s: %v", pkg.stackText(), pkg.err)
			continue
		}
		paths = append(paths, pkg.path)
	}
	for _, err := range inTarget.Errs {
		base.Errorf("%v", err)
	}
	base.ExitIfErrors()
	return paths
}

// TargetPackages returns the list of packages in the target (top-level) module
// matching pattern, which may be relative to the working directory, under all
// build tag settings.
func TargetPackages(pattern string) *search.Match {
	// TargetPackages is relative to the main module, so ensure that the main
	// module is a thing that can contain packages.
	ModRoot()

	m := search.NewMatch(pattern)
	matchPackages(m, imports.AnyTags(), omitStd, []module.Version{Target})
	return m
}

// BuildList returns the module build list,
// typically constructed by a previous call to
// LoadBuildList or ImportPaths.
// The caller must not modify the returned list.
func BuildList() []module.Version {
	return buildList
}

// SetBuildList sets the module build list.
// The caller is responsible for ensuring that the list is valid.
// SetBuildList does not retain a reference to the original list.
func SetBuildList(list []module.Version) {
	buildList = append([]module.Version{}, list...)
}

// TidyBuildList trims the build list to the minimal requirements needed to
// retain the same versions of all packages from the preceding Load* or
// ImportPaths* call.
func TidyBuildList() {
	used := map[module.Version]bool{Target: true}
	for _, pkg := range loaded.pkgs {
		used[pkg.mod] = true
	}

	keep := []module.Version{Target}
	var direct []string
	for _, m := range buildList[1:] {
		if used[m] {
			keep = append(keep, m)
			if loaded.direct[m.Path] {
				direct = append(direct, m.Path)
			}
		} else if cfg.BuildV {
			if _, ok := index.require[m]; ok {
				fmt.Fprintf(os.Stderr, "unused %s\n", m.Path)
			}
		}
	}

	min, err := mvs.Req(Target, direct, &mvsReqs{buildList: keep})
	if err != nil {
		base.Fatalf("go: %v", err)
	}
	buildList = append([]module.Version{Target}, min...)
}

// ImportMap returns the actual package import path
// for an import path found in source code.
// If the given import path does not appear in the source code
// for the packages that have been loaded, ImportMap returns the empty string.
func ImportMap(path string) string {
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		return ""
	}
	return pkg.path
}

// PackageDir returns the directory containing the source code
// for the package named by the import path.
func PackageDir(path string) string {
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		return ""
	}
	return pkg.dir
}

// PackageModule returns the module providing the package named by the import path.
func PackageModule(path string) module.Version {
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		return module.Version{}
	}
	return pkg.mod
}

// PackageImports returns the imports for the package named by the import path.
// Test imports will be returned as well if tests were loaded for the package
// (i.e., if "all" was loaded or if LoadTests was set and the path was matched
// by a command line argument). PackageImports will return nil for
// unknown package paths.
func PackageImports(path string) (imports, testImports []string) {
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		return nil, nil
	}
	imports = make([]string, len(pkg.imports))
	for i, p := range pkg.imports {
		imports[i] = p.path
	}
	if pkg.test != nil {
		testImports = make([]string, len(pkg.test.imports))
		for i, p := range pkg.test.imports {
			testImports[i] = p.path
		}
	}
	return imports, testImports
}

// ModuleUsedDirectly reports whether the main module directly imports
// some package in the module with the given path.
func ModuleUsedDirectly(path string) bool {
	return loaded.direct[path]
}

// Lookup returns the source directory, import path, and any loading error for
// the package at path as imported from the package in parentDir.
// Lookup requires that one of the Load functions in this package has already
// been called.
func Lookup(parentPath string, parentIsStd bool, path string) (dir, realPath string, err error) {
	if path == "" {
		panic("Lookup called with empty package path")
	}

	if parentIsStd {
		path = loaded.stdVendor(parentPath, path)
	}
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		// The loader should have found all the relevant paths.
		// There are a few exceptions, though:
		//	- during go list without -test, the p.Resolve calls to process p.TestImports and p.XTestImports
		//	  end up here to canonicalize the import paths.
		//	- during any load, non-loaded packages like "unsafe" end up here.
		//	- during any load, build-injected dependencies like "runtime/cgo" end up here.
		//	- because we ignore appengine/* in the module loader,
		//	  the dependencies of any actual appengine/* library end up here.
		dir := findStandardImportPath(path)
		if dir != "" {
			return dir, path, nil
		}
		return "", "", errMissing
	}
	return pkg.dir, pkg.path, pkg.err
}

// A loader manages the process of loading information about
// the required packages for a particular build,
// checking that the packages are available in the module set,
// and updating the module set if needed.
// Loading is an iterative process: try to load all the needed packages,
// but if imports are missing, try to resolve those imports, and repeat.
//
// Although most of the loading state is maintained in the loader struct,
// one key piece - the build list - is a global, so that it can be modified
// separate from the loading operation, such as during "go get"
// upgrades/downgrades or in "go mod" operations.
// TODO(rsc): It might be nice to make the loader take and return
// a buildList rather than hard-coding use of the global.
type loader struct {
	tags           map[string]bool // tags for scanDir
	allLevel       allLevel
	forceStdVendor bool // if true, load standard-library dependencies from the vendor subtree

	work      chan *workQueue // 1-buffered; holds a non-nil *workQueue
	idle      chan token      // 1-buffered; holds a token when work.active == 0
	pkgLoaded chan *loadPkg   // unbuffered; receives each package as it is loaded

	// reset on each iteration
	roots    []*loadPkg
	pkgCache *par.Cache
	pkgs     []*loadPkg // populated in buildStacks

	// computed at end of iterations
	direct    map[string]bool   // module paths providing any package imported directly by main module
	goVersion map[string]string // go version recorded in each module
}

type workQueue struct {
	active, maxActive int
	queue             []*loadPkg
}

type token struct{}

// An allLevel is a possible meaning of the package pattern "all".
//
// There are two possible meanings in Go 1.11–1.15: the one used in
// 'go mod vendor' (which is, as a result, also the meaning of an explicit
// "all" pattern with other commands with -mod=vendor), and the one used in
// 'go mod tidy' (which is the same as "all" for other commands with -mod=mod).
//
// Lazy loading will add a third possible meaning in between the two, used in
// 'go mod tidy' to ensure that 'go test all' has all dependencies needed to
// build the tests.
type allLevel int8

const (
	noAll allLevel = iota

	// importedByTarget includes all packages transitively imported by packages
	// and tests in the main module. importedByTarget is the set of packages
	// included by "go mod vendor" in Go 1.11–1.15.
	importedByTarget

	// importedByTransitiveTestFromTarget includes the transitive closure of the
	// imports of all packages and tests of those packages starting with the set
	// of packages and tests in the main module. It is the root of both "go mod
	// tidy" (ignoring tags) and "all" in Go 1.11–1.15.
	importedByTransitiveTestFromTarget
)

// LoadTests controls whether the loaders load tests of the root packages.
var LoadTests bool

func (ld *loader) reset() {
	select {
	case <-ld.idle:
		ld.idle <- token{}
	default:
		panic("loader.reset when not idle")
	}

	ld.roots = nil
	ld.pkgCache = new(par.Cache)
	ld.pkgs = nil
}

// A loadPkg records information about a single loaded package.
type loadPkg struct {
	path string // import path

	// Populated by (*loader).load:
	mod         module.Version // module providing package
	dir         string         // directory containing source code
	err         error          // error loading package
	imports     []*loadPkg     // packages imported by this one
	testImports []string       // test-only imports, saved for use by pkg.test.
	inStd       bool

	// Populated by a single goroutine in loadFromRoots:
	inAll  bool
	loaded bool     // if true, imports and either testImports or test are populated
	test   *loadPkg // package with test imports, if we need test
	testOf *loadPkg

	// Populated in (*loader).buildStacks:
	stack *loadPkg // package importing this one in minimal import stack for this pkg
}

func (pkg *loadPkg) isTest() bool {
	return pkg.testOf != nil
}

var errMissing = errors.New("cannot find package")

// loadFromRoots attempts to load the build graph needed to process a set of
// root packages and their dependencies.
//
// The set of root packages is returned by the roots function,
// and expanded to the full set of packages by tracing imports and tests.
//
// The allLevel determines which tests are traced for imports.
func loadFromRoots(tags map[string]bool, allLevel allLevel, roots func() []string) *loader {
	ld := &loader{
		tags:      tags,
		allLevel:  allLevel,
		work:      make(chan *workQueue, 1),
		idle:      make(chan token, 1),
		pkgLoaded: make(chan *loadPkg),
	}
	ld.work <- &workQueue{maxActive: runtime.GOMAXPROCS(0)}
	ld.idle <- token{}

	// Inside the "std" and "cmd" modules, we prefer to use the vendor directory
	// unless the command explicitly changes the module graph.
	// TODO(bcmills): Is this still needed now that we have automatic vendoring?
	if !targetInGorootSrc || (cfg.CmdName != "get" && !strings.HasPrefix(cfg.CmdName, "mod ")) {
		ld.forceStdVendor = true
	}

	var err error
	reqs := Reqs()
	buildList, err = mvs.BuildList(Target, reqs)
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	addedModuleFor := make(map[string]bool)
	for {
		ld.reset()

		// Load the root packages and their imports.
		// Note: the returned roots can change on each iteration,
		// since the expansion of package patterns depends on the
		// build list we're using.
		ld.loadAll(roots())
		ld.buildStacks()

		modAddedBy := resolveMissingImports(addedModuleFor, ld.pkgs)
		if len(modAddedBy) == 0 {
			break
		}

		// Recompute buildList with all our additions.
		reqs = Reqs()
		buildList, err = mvs.BuildList(Target, reqs)
		if err != nil {
			// If an error was found in a newly added module, report the package
			// import stack instead of the module requirement stack. Packages
			// are more descriptive.
			if err, ok := err.(*mvs.BuildListError); ok {
				if pkg := modAddedBy[err.Module()]; pkg != nil {
					base.Fatalf("go: %s: %v", pkg.stackText(), err.Err)
				}
			}
			base.Fatalf("go: %v", err)
		}
	}
	base.ExitIfErrors()

	// Compute directly referenced dependency modules.
	ld.direct = make(map[string]bool)
	for _, pkg := range ld.pkgs {
		if pkg.mod == Target {
			for _, dep := range pkg.imports {
				if dep.mod.Path != "" {
					ld.direct[dep.mod.Path] = true
				}
			}
		}
	}

	// Add Go versions, computed during walk.
	ld.goVersion = make(map[string]string)
	for _, m := range buildList {
		v, _ := reqs.(*mvsReqs).versions.Load(m)
		ld.goVersion[m.Path], _ = v.(string)
	}

	// If we didn't scan all of the imports from the main module, or didn't use
	// imports.AnyTags, then we didn't necessarily load every package that
	// contributes “direct” imports — so we can't safely mark existing
	// dependencies as indirect-only.
	// Conservatively mark those dependencies as direct.
	if modFile != nil && (ld.allLevel == noAll || !reflect.DeepEqual(tags, imports.AnyTags())) {
		for _, r := range modFile.Require {
			if !r.Indirect {
				ld.direct[r.Mod.Path] = true
			}
		}
	}

	return ld
}

// resolveMissingImports adds module dependencies to the global build list
// in order to resolve missing packages from pkgs.
//
// The newly-resolved packages are added to the addedModuleFor map, and
// resolveMissingImports returns a map from each newly-added module version to
// the first package for which that module was added.
func resolveMissingImports(addedModuleFor map[string]bool, pkgs []*loadPkg) (modAddedBy map[module.Version]*loadPkg) {
	haveMod := make(map[module.Version]bool)
	for _, m := range buildList {
		haveMod[m] = true
	}

	modAddedBy = make(map[module.Version]*loadPkg)
	for _, pkg := range pkgs {
		if pkg.isTest() {
			// If we are missing a test, we are also missing its non-test version, and
			// we should only add the missing import once.
			continue
		}
		if err, ok := pkg.err.(*ImportMissingError); ok && err.Module.Path != "" {
			if err.newMissingVersion != "" {
				base.Fatalf("go: %s: package provided by %s at latest version %s but not at required version %s", pkg.stackText(), err.Module.Path, err.Module.Version, err.newMissingVersion)
			}
			fmt.Fprintf(os.Stderr, "go: found %s in %s %s\n", pkg.path, err.Module.Path, err.Module.Version)
			if addedModuleFor[pkg.path] {
				base.Fatalf("go: %s: looping trying to add package", pkg.stackText())
			}
			addedModuleFor[pkg.path] = true
			if !haveMod[err.Module] {
				haveMod[err.Module] = true
				modAddedBy[err.Module] = pkg
				buildList = append(buildList, err.Module)
			}
			continue
		}
		// Leave other errors for Import or load.Packages to report.
	}
	base.ExitIfErrors()

	return modAddedBy
}

// pkg returns the *loadPkg for path, creating and queuing it if needed.
// If the package should be tested, its test is created but not queued
// (the test is queued after processing pkg).
// If isRoot is true, the pkg is being queued as one of the roots of the work graph.
func (ld *loader) pkg(path string) *loadPkg {
	return ld.pkgCache.Do(path, func() interface{} {
		pkg := &loadPkg{
			path: path,
		}
		ld.startLoad(pkg)
		return pkg
	}).(*loadPkg)
}

// startLoad either adds pkg to ld's work queue, or spawns a new goroutine to
// begin loading pkg immediately.
func (ld *loader) startLoad(pkg *loadPkg) {
	work := <-ld.work
	if work.active >= work.maxActive {
		work.queue = append(work.queue, pkg)
		ld.work <- work
		return
	}
	if work.active == 0 {
		<-ld.idle // Mark ld as non-idle.
	}
	work.active++
	ld.work <- work

	go func() {
		for {
			ld.load(pkg)
			ld.pkgLoaded <- pkg

			work := <-ld.work
			if len(work.queue) == 0 {
				if work.active--; work.active == 0 {
					ld.idle <- token{}
				}
				ld.work <- work
				return
			}
			pkg, work.queue = work.queue[0], work.queue[1:]
			ld.work <- work
		}
	}()
}

// load loads an individual package.
func (ld *loader) load(pkg *loadPkg) {
	if strings.Contains(pkg.path, "@") {
		// Leave for error during load.
		return
	}
	if build.IsLocalImport(pkg.path) || filepath.IsAbs(pkg.path) {
		// Leave for error during load.
		// (Module mode does not allow local imports.)
		return
	}

	pkg.mod, pkg.dir, pkg.err = Import(pkg.path)
	if pkg.dir == "" {
		return
	}

	imports, testImports, err := scanDir(pkg.dir, ld.tags)
	if err != nil {
		pkg.err = err
		return
	}

	pkg.inStd = (search.IsStandardImportPath(pkg.path) && search.InDir(pkg.dir, cfg.GOROOTsrc) != "")

	pkg.imports = make([]*loadPkg, 0, len(imports))
	for _, path := range imports {
		if pkg.inStd {
			path = ld.stdVendor(pkg.path, path)
		}
		pkg.imports = append(pkg.imports, ld.pkg(path))
	}

	pkg.testImports = testImports
}

// loadTestOf loads the test of pkg, if it has not already been loaded.
// loadTestOf requires that load(pkg) has already completed.
func (ld *loader) loadTestOf(pkg *loadPkg) *loadPkg {
	if pkg.test != nil {
		return pkg.test
	}
	if pkg.isTest() {
		panic("testOf called on a test package")
	}

	test := &loadPkg{
		path:   pkg.path,
		mod:    pkg.mod,
		dir:    pkg.dir,
		err:    pkg.err,
		testOf: pkg,
	}

	test.imports = make([]*loadPkg, 0, len(pkg.testImports))
	for _, path := range pkg.testImports {
		if pkg.inStd {
			path = ld.stdVendor(test.path, path)
		}
		test.imports = append(test.imports, ld.pkg(path))
	}
	test.loaded = true
	pkg.testImports = nil
	pkg.test = test
	return pkg.test
}

// stdVendor returns the canonical import path for the package with the given
// path when imported from the standard-library package at parentPath.
func (ld *loader) stdVendor(parentPath, path string) string {
	if search.IsStandardImportPath(path) {
		return path
	}

	if str.HasPathPrefix(parentPath, "cmd") {
		if ld.forceStdVendor || Target.Path != "cmd" {
			vendorPath := pathpkg.Join("cmd", "vendor", path)
			if _, err := os.Stat(filepath.Join(cfg.GOROOTsrc, filepath.FromSlash(vendorPath))); err == nil {
				return vendorPath
			}
		}
	} else if ld.forceStdVendor || Target.Path != "std" {
		vendorPath := pathpkg.Join("vendor", path)
		if _, err := os.Stat(filepath.Join(cfg.GOROOTsrc, filepath.FromSlash(vendorPath))); err == nil {
			return vendorPath
		}
	}

	// Not vendored: resolve from modules.
	return path
}

// loadAll imports all packages needed to build the given roots,
// then expands that set (if needed) to match ld.allLevel.
func (ld *loader) loadAll(roots []string) {
	isRoot := map[*loadPkg]bool{}
	for _, path := range roots {
		root := ld.pkg(path)
		if !isRoot[root] {
			isRoot[root] = true
			ld.roots = append(ld.roots, root)
		}
	}

	for {
		var pkg *loadPkg
		select {
		case <-ld.idle:
			ld.idle <- token{}
			return
		case pkg = <-ld.pkgLoaded:
		}
		pkg.loaded = true

		if ld.allLevel >= importedByTarget && pkg.mod == Target {
			pkg.inAll = true
		}
		if LoadTests && (pkg.inAll || isRoot[pkg]) {
			ld.loadTestOf(pkg)
		}
		if pkg.inAll {
			ld.propagateInAll(pkg)
		}
	}
}

// propagateInAll sets inAll for packages that are in "all" by virtue of their
// connection to pkg.
func (ld *loader) propagateInAll(pkg *loadPkg) {
	if !pkg.loaded {
		return
	}

	for _, dep := range pkg.imports {
		if !dep.inAll {
			dep.inAll = true
			ld.propagateInAll(dep)
		}
	}

	if !pkg.isTest() && (ld.allLevel >= importedByTransitiveTestFromTarget ||
		(ld.allLevel >= importedByTarget && pkg.mod == Target)) {
		if test := ld.loadTestOf(pkg); !test.inAll {
			test.inAll = true
			ld.propagateInAll(test)
		}
	}
}

// computePatternAll returns the list of packages matching pattern "all",
// starting with a list of the import paths for the packages in the main module.
func (ld *loader) computePatternAll() (all []string) {
	for _, pkg := range ld.pkgs {
		if pkg.inAll && !pkg.isTest() {
			all = append(all, pkg.path)
		}
	}
	sort.Strings(all)
	return all
}

// scanDir is like imports.ScanDir but elides known magic imports from the list,
// so that we do not go looking for packages that don't really exist.
//
// The standard magic import is "C", for cgo.
//
// The only other known magic imports are appengine and appengine/*.
// These are so old that they predate "go get" and did not use URL-like paths.
// Most code today now uses google.golang.org/appengine instead,
// but not all code has been so updated. When we mostly ignore build tags
// during "go vendor", we look into "// +build appengine" files and
// may see these legacy imports. We drop them so that the module
// search does not look for modules to try to satisfy them.
func scanDir(dir string, tags map[string]bool) (imports_, testImports []string, err error) {
	imports_, testImports, err = imports.ScanDir(dir, tags)

	filter := func(x []string) []string {
		w := 0
		for _, pkg := range x {
			if pkg != "C" && pkg != "appengine" && !strings.HasPrefix(pkg, "appengine/") &&
				pkg != "appengine_internal" && !strings.HasPrefix(pkg, "appengine_internal/") {
				x[w] = pkg
				w++
			}
		}
		return x[:w]
	}

	return filter(imports_), filter(testImports), err
}

// buildStacks computes minimal import stacks for each package,
// for use in error messages. When it completes, packages that
// are part of the original root set have pkg.stack == nil,
// and other packages have pkg.stack pointing at the next
// package up the import stack in their minimal chain.
// As a side effect, buildStacks also constructs ld.pkgs,
// the list of all packages loaded.
func (ld *loader) buildStacks() {
	if len(ld.pkgs) > 0 {
		panic("buildStacks")
	}
	for _, pkg := range ld.roots {
		pkg.stack = pkg // sentinel to avoid processing in next loop
		ld.pkgs = append(ld.pkgs, pkg)
	}
	for i := 0; i < len(ld.pkgs); i++ { // not range: appending to ld.pkgs in loop
		pkg := ld.pkgs[i]
		for _, next := range pkg.imports {
			if next.stack == nil {
				next.stack = pkg
				ld.pkgs = append(ld.pkgs, next)
			}
		}
		if next := pkg.test; next != nil && next.stack == nil {
			next.stack = pkg
			ld.pkgs = append(ld.pkgs, next)
		}
	}
	for _, pkg := range ld.roots {
		pkg.stack = nil
	}
}

// stackText builds the import stack text to use when
// reporting an error in pkg. It has the general form
//
//	root imports
//		other imports
//		other2 tested by
//		other2.test imports
//		pkg
//
func (pkg *loadPkg) stackText() string {
	var stack []*loadPkg
	for p := pkg; p != nil; p = p.stack {
		stack = append(stack, p)
	}

	var buf bytes.Buffer
	for i := len(stack) - 1; i >= 0; i-- {
		p := stack[i]
		fmt.Fprint(&buf, p.path)
		if p.testOf != nil {
			fmt.Fprint(&buf, ".test")
		}
		if i > 0 {
			if stack[i-1].testOf == p {
				fmt.Fprint(&buf, " tested by\n\t")
			} else {
				fmt.Fprint(&buf, " imports\n\t")
			}
		}
	}
	return buf.String()
}

// why returns the text to use in "go mod why" output about the given package.
// It is less ornate than the stackText but contains the same information.
func (pkg *loadPkg) why() string {
	var buf strings.Builder
	var stack []*loadPkg
	for p := pkg; p != nil; p = p.stack {
		stack = append(stack, p)
	}

	for i := len(stack) - 1; i >= 0; i-- {
		p := stack[i]
		if p.testOf != nil {
			fmt.Fprintf(&buf, "%s.test\n", p.testOf.path)
		} else {
			fmt.Fprintf(&buf, "%s\n", p.path)
		}
	}
	return buf.String()
}

// Why returns the "go mod why" output stanza for the given package,
// without the leading # comment.
// The package graph must have been loaded already, usually by LoadALL.
// If there is no reason for the package to be in the current build,
// Why returns an empty string.
func Why(path string) string {
	pkg, ok := loaded.pkgCache.Get(path).(*loadPkg)
	if !ok {
		return ""
	}
	return pkg.why()
}

// WhyDepth returns the number of steps in the Why listing.
// If there is no reason for the package to be in the current build,
// WhyDepth returns 0.
func WhyDepth(path string) int {
	n := 0
	pkg, _ := loaded.pkgCache.Get(path).(*loadPkg)
	for p := pkg; p != nil; p = p.stack {
		n++
	}
	return n
}
