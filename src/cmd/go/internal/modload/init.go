// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modload

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"internal/lazyregexp"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/fsys"
	"cmd/go/internal/lockedfile"
	"cmd/go/internal/modconv"
	"cmd/go/internal/modfetch"
	"cmd/go/internal/search"
	"cmd/go/internal/web"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func TODOWorkspaces(s string) error {
	return fmt.Errorf("need to support this for workspaces: %s", s)
}

// Variables set in Init.
var (
	// These are primarily used to initialize the MainModules, and should be
	// eventually superceded by them but are still used in cases where the module
	// roots are required but MainModules hasn't been initialized yet. Set to
	// the modRoots of the main modules.
	// modRoots != nil implies len(modRoots) > 0
	modRoots          []string
	workFileGoVersion string
)

type MainModuleSet struct {
	// workspaceMode indicates whether the go command is running in a workspace,
	// which may contain multiple main modules. If workspaceMode is false, there
	// may be at most one main module.
	workspaceMode bool

	// versions are the module.Version values of each of the main modules.
	// For each of them, the Path fields are ordinary module paths and the Version
	// fields are empty strings.
	versions []module.Version

	// modRoot maps each module in versions to its absolute filesystem path.
	modRoot map[module.Version]string

	// pathPrefix is the path prefix for packages in the module, without a trailing
	// slash. For most modules, pathPrefix is just version.Path, but the
	// standard-library module "std" has an empty prefix.
	pathPrefix map[module.Version]string

	// inGorootSrc caches whether modRoot is within GOROOT/src.
	// The "std" module is special within GOROOT/src, but not otherwise.
	inGorootSrc map[module.Version]bool

	modFiles map[module.Version]*modfile.File

	modContainingCWD module.Version

	workFileGoVersion string

	indexMu sync.Mutex
	indices map[module.Version]*modFileIndex
}

func (mms *MainModuleSet) PathPrefix(m module.Version) string {
	return mms.pathPrefix[m]
}

// Versions returns the module.Version values of each of the main modules.
// For each of them, the Path fields are ordinary module paths and the Version
// fields are empty strings.
// Callers should not modify the returned slice.
func (mms *MainModuleSet) Versions() []module.Version {
	if mms == nil {
		return nil
	}
	return mms.versions
}

func (mms *MainModuleSet) Contains(path string) bool {
	if mms == nil {
		return false
	}
	for _, v := range mms.versions {
		if v.Path == path {
			return true
		}
	}
	return false
}

func (mms *MainModuleSet) ModRoot(m module.Version) string {
	if mms == nil {
		return ""
	}
	return mms.modRoot[m]
}

func (mms *MainModuleSet) InGorootSrc(m module.Version) bool {
	if mms == nil {
		return false
	}
	return mms.inGorootSrc[m]
}

func (mms *MainModuleSet) mustGetSingleMainModule() module.Version {
	if mms == nil || len(mms.versions) == 0 {
		panic("internal error: mustGetSingleMainModule called in context with no main modules")
	}
	if len(mms.versions) != 1 {
		if mms.workspaceMode {
			panic("internal error: mustGetSingleMainModule called in workspace mode")
		} else {
			panic("internal error: multiple main modules present outside of workspace mode")
		}
	}
	return mms.versions[0]
}

func (mms *MainModuleSet) GetSingleIndexOrNil() *modFileIndex {
	if mms == nil {
		return nil
	}
	if len(mms.versions) == 0 {
		return nil
	}
	return mms.indices[mms.mustGetSingleMainModule()]
}

func (mms *MainModuleSet) Index(m module.Version) *modFileIndex {
	mms.indexMu.Lock()
	defer mms.indexMu.Unlock()
	return mms.indices[m]
}

func (mms *MainModuleSet) SetIndex(m module.Version, index *modFileIndex) {
	mms.indexMu.Lock()
	defer mms.indexMu.Unlock()
	mms.indices[m] = index
}

func (mms *MainModuleSet) ModFile(m module.Version) *modfile.File {
	return mms.modFiles[m]
}

func (mms *MainModuleSet) Len() int {
	if mms == nil {
		return 0
	}
	return len(mms.versions)
}

// ModContainingCWD returns the main module containing the working directory,
// or module.Version{} if none of the main modules contain the working
// directory.
func (mms *MainModuleSet) ModContainingCWD() module.Version {
	return mms.modContainingCWD
}

// GoVersion returns the go version set on the single module, in module mode,
// or the go.work file in workspace mode.
func (mms *MainModuleSet) GoVersion() string {
	if !mms.workspaceMode {
		return modFileGoVersion(mms.ModFile(mms.mustGetSingleMainModule()))
	}
	v := mms.workFileGoVersion
	if v == "" {
		// Fall back to 1.18 for go.work files.
		v = "1.18"
	}
	return v
}

var MainModules *MainModuleSet

type Root int

const (
	// AutoRoot is the default for most commands. modload.Init will look for
	// a go.mod file in the current directory or any parent. If none is found,
	// modules may be disabled (GO111MODULE=auto) or commands may run in a
	// limited module mode.
	AutoRoot Root = iota

	// NoRoot is used for commands that run in module mode and ignore any go.mod
	// file the current directory or in parent directories.
	NoRoot

	// NeedRoot is used for commands that must run in module mode and don't
	// make sense without a main module.
	NeedRoot
)

// ModFile returns the parsed go.mod file.
//
// Note that after calling LoadPackages or LoadModGraph,
// the require statements in the modfile.File are no longer
// the source of truth and will be ignored: edits made directly
// will be lost at the next call to WriteGoMod.
// To make permanent changes to the require statements
// in go.mod, edit it before loading.
func ModFile() *modfile.File {
	modFile := MainModules.ModFile(MainModules.mustGetSingleMainModule())
	if modFile == nil {
		die()
	}
	return modFile
}

// Opts is a set of options commands can use when initializing modules with
// modload.Init.
type Opts struct {
	// ForceUseModules may be set to force modules to be enabled when
	// GO111MODULE=auto or to report an error when GO111MODULE=off.
	ForceUseModules bool

	// RootMode determines whether a module root is needed.
	RootMode Root

	// AllowWorkspace indicates whether a workspace may be defined in a go.work
	// file. cfg.WorkFile controls whether workspace mode is actually enabled.
	AllowWorkspace bool

	// ForceBuildMod forces a specific value for the -mod flag for commands that
	// don't support the -mod flag. It may be "", "mod", "readonly", or "vendor".
	// If "", the mode is derived from cfg.BuildMod (-mod). Use State.Mod
	// for the effective setting.
	ForceBuildMod string

	// DontAddGoDirective instructs LoadModFile not to add a "go" directive when
	// one absent. By default, if the main module lacks a "go" directive,
	// LoadModFile will add one with the latest Go version and will convert
	// requirements to be lazy. This is usually desirable, but it alters the
	// output of commands like 'go mod graph'.
	DontAddGoDirective bool
}

// State holds information about modules loaded into memory. State is read
// and potentially written by most functions in this package.
type State struct {
	Opts

	// Mod is the effective setting for the -mod flag. It may be "mod",
	// "readonly", or "vendor". Mod may change after loading the main module's
	// go.mod file if automatic vendoring is enabled.
	Mod string

	// ModReason is the reason Mod is set. May be "" if Mod was set by -mod or
	// if ForceBuildMod was used.
	ModReason string

	// WorkFilePath is the path to the go.work file or "" if workspace mode
	// is disabled.
	WorkFilePath string
}

// Init determines whether module mode is enabled, locates the root of the
// current module (if any), sets environment variables for Git subprocesses, and
// configures the cfg, codehost, load, modfetch, and search packages for use
// with modules.
//
// If modules are enabled, Init returns a non-nil State holding module
// information. State may be passed to other modload functions. If modules
// are not enabled (for example, because GO111MODULE is "off"), Init returns
// nil.
func Init(opts Opts) (state *State, err error) {
	defer func() {
		if err == nil {
			err = checkModCommonFlags(state)
		}
	}()

	// Keep in sync with WillBeEnabled. We perform extra validation here, and
	// there are lots of diagnostics and side effects, so we can't use
	// WillBeEnabled directly.
	var mustUseModules bool
	env := cfg.Getenv("GO111MODULE")
	switch env {
	default:
		return nil, fmt.Errorf("unknown environment setting GO111MODULE=%s", env)
	case "auto":
		mustUseModules = opts.ForceUseModules
	case "on", "":
		mustUseModules = true
	case "off":
		if opts.ForceUseModules {
			return nil, fmt.Errorf("modules disabled by GO111MODULE=off; see 'go help modules'")
		}
		mustUseModules = false
		return nil, nil
	}

	if err := fsys.Init(base.Cwd()); err != nil {
		return nil, err
	}

	var workFilePath string
	if opts.AllowWorkspace {
		switch cfg.WorkFile {
		case "off":
			workFilePath = ""
		case "", "auto":
			workFilePath = findWorkspaceFile(base.Cwd())
		default:
			workFilePath = cfg.WorkFile
		}
	}

	if modRoots != nil {
		// modRoot set before Init was called ("go mod init" does this).
		// No need to search for go.mod.
	} else if opts.RootMode == NoRoot {
		if cfg.ModFile != "" && !base.InGOFLAGS("-modfile") {
			return nil, fmt.Errorf("-modfile cannot be used with commands that ignore the current module")
		}
		modRoots = nil
	} else if workFilePath != "" {
		// We're in workspace mode.
	} else {
		if modRoot := findModuleRoot(base.Cwd()); modRoot == "" {
			if cfg.ModFile != "" {
				return nil, fmt.Errorf("cannot find main module, but -modfile was set.\n\t-modfile cannot be used to set the module root directory.")
			}
			if opts.RootMode == NeedRoot {
				return nil, ErrNoModRoot
			}
			if !mustUseModules {
				// GO111MODULE is 'auto', and we can't find a module root.
				// Stay in GOPATH mode.
				return nil, nil
			}
		} else if search.InDir(modRoot, os.TempDir()) == "." {
			// If you create /tmp/go.mod for experimenting,
			// then any tests that create work directories under /tmp
			// will find it and get modules when they're not expecting them.
			// It's a bit of a peculiar thing to disallow but quite mysterious
			// when it happens. See golang.org/issue/26708.
			fmt.Fprintf(os.Stderr, "go: warning: ignoring go.mod in system temp root %v\n", os.TempDir())
			if !mustUseModules {
				return nil, nil
			}
		} else {
			modRoots = []string{modRoot}
		}
	}
	if cfg.ModFile != "" && !strings.HasSuffix(cfg.ModFile, ".mod") {
		return nil, fmt.Errorf("-modfile=%s: file does not have .mod extension", cfg.ModFile)
	}

	// We're in module mode. Set any global variables that need to be set.

	// Disable any prompting for passwords by Git.
	// Only has an effect for 2.3.0 or later, but avoiding
	// the prompt in earlier versions is just too hard.
	// If user has explicitly set GIT_TERMINAL_PROMPT=1, keep
	// prompting.
	// See golang.org/issue/9341 and golang.org/issue/12706.
	if os.Getenv("GIT_TERMINAL_PROMPT") == "" {
		os.Setenv("GIT_TERMINAL_PROMPT", "0")
	}

	// Disable any ssh connection pooling by Git.
	// If a Git subprocess forks a child into the background to cache a new connection,
	// that child keeps stdout/stderr open. After the Git subprocess exits,
	// os /exec expects to be able to read from the stdout/stderr pipe
	// until EOF to get all the data that the Git subprocess wrote before exiting.
	// The EOF doesn't come until the child exits too, because the child
	// is holding the write end of the pipe.
	// This is unfortunate, but it has come up at least twice
	// (see golang.org/issue/13453 and golang.org/issue/16104)
	// and confuses users when it does.
	// If the user has explicitly set GIT_SSH or GIT_SSH_COMMAND,
	// assume they know what they are doing and don't step on it.
	// But default to turning off ControlMaster.
	if os.Getenv("GIT_SSH") == "" && os.Getenv("GIT_SSH_COMMAND") == "" {
		os.Setenv("GIT_SSH_COMMAND", "ssh -o ControlMaster=no -o BatchMode=yes")
	}

	if os.Getenv("GCM_INTERACTIVE") == "" {
		os.Setenv("GCM_INTERACTIVE", "never")
	}
	gopath := filepath.SplitList(cfg.BuildContext.GOPATH)
	if len(gopath) == 0 || gopath[0] == "" {
		return nil, fmt.Errorf("missing $GOPATH")
	}
	if _, err := fsys.Stat(filepath.Join(gopath[0], "go.mod")); err == nil {
		return nil, fmt.Errorf("$GOPATH/go.mod exists but should not")
	}

	if workFilePath != "" {
		var err error
		workFileGoVersion, modRoots, err = loadWorkFile(workFilePath)
		if err != nil {
			return nil, fmt.Errorf("reading go.work: %v", err)
		}
		_ = TODOWorkspaces("Support falling back to individual module go.sum " +
			"files for sums not in the workspace sum file.")
		modfetch.GoSumFile = workFilePath + ".sum"
	} else if modRoots == nil {
		// We're in module mode, but not inside a module.
		//
		// Commands like 'go build', 'go run', 'go list' have no go.mod file to
		// read or write. They would need to find and download the latest versions
		// of a potentially large number of modules with no way to save version
		// information. We can succeed slowly (but not reproducibly), but that's
		// not usually a good experience.
		//
		// Instead, we forbid resolving import paths to modules other than std and
		// cmd. Users may still build packages specified with .go files on the
		// command line, but they'll see an error if those files import anything
		// outside std.
		//
		// This can be overridden by setting opts.ForceBuildMod = "mod".
		// For example, 'go get' does this, since it is expected to resolve paths.
		//
		// See golang.org/issue/32027.
	} else {
		modfetch.GoSumFile = strings.TrimSuffix(modFilePath(modRoots[0]), ".mod") + ".sum"
	}

	state = &State{
		Opts:         opts,
		WorkFilePath: workFilePath,
	}
	if opts.ForceBuildMod != "" {
		state.Mod = opts.ForceBuildMod
	} else if cfg.BuildMod != "" {
		state.Mod = cfg.BuildMod
	} else {
		state.Mod = "readonly"
	}
	if state.Mod == "vendor" {
		web.PanicIfNetworkUsed = true
	}
	return state, nil
}

var willBeEnabled = struct {
	once    sync.Once
	enabled bool
}{}

// WillBeEnabled checks whether modules should be enabled but does not
// initialize modules by installing hooks. If Init has already been called,
// WillBeEnabled returns the same result as Enabled.
//
// This function is needed to break a cycle. The main package needs to know
// whether modules are enabled in order to install the module or GOPATH version
// of 'go get', but Init reads the -modfile flag in 'go get', so it shouldn't
// be called until the command is installed and flags are parsed. Instead of
// calling Init and Enabled, the main package can call this function.
func WillBeEnabled() bool {
	willBeEnabled.once.Do(func() {
		// Keep in sync with Init. Init does extra validation and prints warnings or
		// exits, so it can't call this function directly.
		env := cfg.Getenv("GO111MODULE")
		switch env {
		case "on", "":
			willBeEnabled.enabled = true
			return
		case "auto":
			break
		default:
			willBeEnabled.enabled = false
			return
		}

		if modRoot := findModuleRoot(base.Cwd()); modRoot == "" {
			// GO111MODULE is 'auto', and we can't find a module root.
			// Stay in GOPATH mode.
			willBeEnabled.enabled = false
			return
		} else if search.InDir(modRoot, os.TempDir()) == "." {
			// If you create /tmp/go.mod for experimenting,
			// then any tests that create work directories under /tmp
			// will find it and get modules when they're not expecting them.
			// It's a bit of a peculiar thing to disallow but quite mysterious
			// when it happens. See golang.org/issue/26708.
			willBeEnabled.enabled = false
			return
		}
		willBeEnabled.enabled = true
	})
	return willBeEnabled.enabled
}

func VendorDir() string {
	return filepath.Join(MainModules.ModRoot(MainModules.mustGetSingleMainModule()), "vendor")
}

// HasModRoot reports whether a main module is present.
// HasModRoot may return false even if Enabled returns true: for example, 'get'
// does not require a main module.
func HasModRoot() bool {
	return modRoots != nil
}

// MustHaveModRoot checks that a main module or main modules are present,
// and calls base.Fatalf if there are no main modules.
func MustHaveModRoot() {
	if !HasModRoot() {
		die()
	}
}

// ModFilePath returns the path that would be used for the go.mod
// file, if in module mode. ModFilePath calls base.Fatalf if there is no main
// module, even if -modfile is set.
func ModFilePath() string {
	MustHaveModRoot()
	return modFilePath(findModuleRoot(base.Cwd()))
}

func modFilePath(modRoot string) string {
	if cfg.ModFile != "" {
		return cfg.ModFile
	}
	return filepath.Join(modRoot, "go.mod")
}

func die() {
	if cfg.Getenv("GO111MODULE") == "off" {
		base.Fatalf("go: modules disabled by GO111MODULE=off; see 'go help modules'")
	}
	if dir, name := findAltConfig(base.Cwd()); dir != "" {
		rel, err := filepath.Rel(base.Cwd(), dir)
		if err != nil {
			rel = dir
		}
		cdCmd := ""
		if rel != "." {
			cdCmd = fmt.Sprintf("cd %s && ", rel)
		}
		base.Fatalf("go: cannot find main module, but found %s in %s\n\tto create a module there, run:\n\t%sgo mod init", name, dir, cdCmd)
	}
	base.Fatalf("go: %v", ErrNoModRoot)
}

var ErrNoModRoot = errors.New("go.mod file not found in current directory or any parent directory; see 'go help modules'")

type goModDirtyError struct {
	mod, modReason string
}

func newGoModDirtyError(state *State) error {
	return &goModDirtyError{mod: state.Mod, modReason: state.ModReason}
}

func (e *goModDirtyError) Error() string {
	if cfg.BuildMod != "" {
		return fmt.Sprintf("updates to go.mod needed, disabled by -mod=%v; to update it:\n\tgo mod tidy", cfg.BuildMod)
	}
	if e.modReason != "" {
		return fmt.Sprintf("updates to go.mod needed, disabled by -mod=%s\n\t(%s)\n\tto update it:\n\tgo mod tidy", e.mod, e.modReason)
	}
	return "updates to go.mod needed; to update it:\n\tgo mod tidy"
}

func loadWorkFile(path string) (goVersion string, modRoots []string, err error) {
	_ = TODOWorkspaces("Clean up and write back the go.work file: add module paths for workspace modules.")
	workDir := filepath.Dir(path)
	workData, err := lockedfile.Read(path)
	if err != nil {
		return "", nil, err
	}
	wf, err := modfile.ParseWork(path, workData, nil)
	if err != nil {
		return "", nil, err
	}
	if wf.Go != nil {
		goVersion = wf.Go.Version
	}
	seen := map[string]bool{}
	for _, d := range wf.Directory {
		modRoot := d.Path
		if !filepath.IsAbs(modRoot) {
			modRoot = filepath.Join(workDir, modRoot)
		}
		if seen[modRoot] {
			return "", nil, fmt.Errorf("path %s appears multiple times in workspace", modRoot)
		}
		seen[modRoot] = true
		modRoots = append(modRoots, modRoot)
	}
	return goVersion, modRoots, nil
}

// LoadModFile sets Target and, if there is a main module, parses the initial
// build list from its go.mod file.
//
// LoadModFile may make changes in memory, like adding a go directive and
// ensuring requirements are consistent. The caller is responsible for calling
// WriteModFile later to write those changes and any others to disk.
//
// As a side-effect, LoadModFile may change state.Mod to "vendor" if
// -mod wasn't set explicitly and automatic vendoring should be enabled.
//
// If LoadModFile or CreateModFile has already been called, LoadModFile returns
// the existing in-memory requirements (rather than re-reading them from disk).
//
// LoadModFile checks the roots of the module graph for consistency with each
// other, but unlike LoadModGraph does not load the full module graph or check
// it for global consistency. Most callers outside of the modload package should
// use LoadModGraph instead.
func LoadModFile(ctx context.Context, state *State) *Requirements {
	if requirements != nil {
		return requirements
	}

	if len(modRoots) == 0 {
		_ = TODOWorkspaces("Instead of creating a fake module with an empty modroot, make MainModules.Len() == 0 mean that we're in module mode but not inside any module.")
		mainModule := module.Version{Path: "command-line-arguments"}
		MainModules = makeMainModules(state.WorkFilePath != "", []module.Version{mainModule}, []string{""}, []*modfile.File{nil}, []*modFileIndex{nil}, "")
		goVersion := LatestGoVersion()
		rawGoVersion.Store(mainModule, goVersion)
		requirements = newRequirements(pruningForGoVersion(goVersion), nil, nil)
		return requirements
	}

	var modFiles []*modfile.File
	var mainModules []module.Version
	var indices []*modFileIndex
	for _, modroot := range modRoots {
		gomod := modFilePath(modroot)
		var fixed bool
		data, f, err := ReadModFile(gomod, fixVersion(ctx, state, &fixed))
		if err != nil {
			base.Fatalf("go: %v", err)
		}

		modFiles = append(modFiles, f)
		mainModule := f.Module.Mod
		mainModules = append(mainModules, mainModule)
		canWrite := state.Mod == "mod"
		indices = append(indices, indexModFile(data, f, mainModule, canWrite, fixed))

		if err := module.CheckImportPath(f.Module.Mod.Path); err != nil {
			if pathErr, ok := err.(*module.InvalidPathError); ok {
				pathErr.Kind = "module"
			}
			base.Fatalf("go: %v", err)
		}
	}

	MainModules = makeMainModules(state.WorkFilePath != "", mainModules, modRoots, modFiles, indices, workFileGoVersion)
	maybeEnableVendoring(state)
	rs := requirementsFromModFiles(ctx, state, modFiles)

	if state.WorkFilePath != "" {
		// We don't need to do anything for vendor or update the mod file so
		// return early.
		requirements = rs
		return rs
	}

	mainModule := MainModules.mustGetSingleMainModule()

	if state.Mod == "vendor" {
		readVendorList(mainModule)
		index := MainModules.Index(mainModule)
		modFile := MainModules.ModFile(mainModule)
		checkVendorConsistency(index, modFile)
		rs.initVendor(state, vendorList)
	}

	if rs.hasRedundantRoot() {
		// If any module path appears more than once in the roots, we know that the
		// go.mod file needs to be updated even though we have not yet loaded any
		// transitive dependencies.
		var err error
		rs, err = updateRoots(ctx, state, rs.direct, rs, nil, nil, false)
		if err != nil {
			base.Fatalf("go: %v", err)
		}
	}

	if modFile := MainModules.Index(mainModule); modFile.goVersionV == "" {
		if modFile.canWrite && !state.DontAddGoDirective {
			addGoStmt(MainModules.ModFile(mainModule), mainModule, LatestGoVersion())

			// We need to add a 'go' version to the go.mod file, but we must assume
			// that its existing contents match something between Go 1.11 and 1.16.
			// Go 1.11 through 1.16 do not support graph pruning, but the latest Go
			// version uses a pruned module graph — so we need to convert the
			// requirements to support pruning.
			var err error
			rs, err = convertPruning(ctx, state, rs, pruned)
			if err != nil {
				base.Fatalf("go: %v", err)
			}
		} else {
			rawGoVersion.Store(mainModule, modFileGoVersion(MainModules.ModFile(mainModule)))
		}
	}

	requirements = rs
	return requirements
}

// CreateModFile initializes a new module by creating a go.mod file.
//
// If modPath is empty, CreateModFile will attempt to infer the path from the
// directory location within GOPATH.
//
// If a vendoring configuration file is present, CreateModFile will attempt to
// translate it to go.mod directives. The resulting build list may not be
// exactly the same as in the legacy configuration (for example, we can't get
// packages at multiple versions from the same module).
func CreateModFile(ctx context.Context, state *State, modPath string) {
	modRoot := base.Cwd()
	modRoots = []string{modRoot}
	modFilePath := modFilePath(modRoot)
	if _, err := fsys.Stat(modFilePath); err == nil {
		base.Fatalf("go: %s already exists", modFilePath)
	}

	if modPath == "" {
		var err error
		modPath, err = findModulePath(modRoot)
		if err != nil {
			base.Fatalf("go: %v", err)
		}
	} else if err := module.CheckImportPath(modPath); err != nil {
		if pathErr, ok := err.(*module.InvalidPathError); ok {
			pathErr.Kind = "module"
			// Same as build.IsLocalPath()
			if pathErr.Path == "." || pathErr.Path == ".." ||
				strings.HasPrefix(pathErr.Path, "./") || strings.HasPrefix(pathErr.Path, "../") {
				pathErr.Err = errors.New("is a local import path")
			}
		}
		base.Fatalf("go: %v", err)
	} else if _, _, ok := module.SplitPathVersion(modPath); !ok {
		if strings.HasPrefix(modPath, "gopkg.in/") {
			invalidMajorVersionMsg := fmt.Errorf("module paths beginning with gopkg.in/ must always have a major version suffix in the form of .vN:\n\tgo mod init %s", suggestGopkgIn(modPath))
			base.Fatalf(`go: invalid module path "%v": %v`, modPath, invalidMajorVersionMsg)
		}
		invalidMajorVersionMsg := fmt.Errorf("major version suffixes must be in the form of /vN and are only allowed for v2 or later:\n\tgo mod init %s", suggestModulePath(modPath))
		base.Fatalf(`go: invalid module path "%v": %v`, modPath, invalidMajorVersionMsg)
	}

	fmt.Fprintf(os.Stderr, "go: creating new go.mod: module %s\n", modPath)
	modFile := new(modfile.File)
	modFile.AddModuleStmt(modPath)
	MainModules = makeMainModules(state.WorkFilePath != "", []module.Version{modFile.Module.Mod}, []string{modRoot}, []*modfile.File{modFile}, []*modFileIndex{nil}, "")
	addGoStmt(modFile, modFile.Module.Mod, LatestGoVersion()) // Add the go directive before converted module requirements.

	convertedFrom, err := convertLegacyConfig(state, modFile, modRoot)
	if convertedFrom != "" {
		fmt.Fprintf(os.Stderr, "go: copying requirements from %s\n", base.ShortPath(convertedFrom))
	}
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	rs := requirementsFromModFiles(ctx, state, []*modfile.File{modFile})
	rs, err = updateRoots(ctx, state, rs.direct, rs, nil, nil, false)
	if err != nil {
		base.Fatalf("go: %v", err)
	}
	requirements = rs
	if err := writeRequirements(ctx, state); err != nil {
		base.Fatalf("go: %v", err)
	}

	// Suggest running 'go mod tidy' unless the project is empty. Even if we
	// imported all the correct requirements above, we're probably missing
	// some sums, so the next build command in -mod=readonly will likely fail.
	//
	// We look for non-hidden .go files or subdirectories to determine whether
	// this is an existing project. Walking the tree for packages would be more
	// accurate, but could take much longer.
	empty := true
	files, _ := os.ReadDir(modRoot)
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
			continue
		}
		if strings.HasSuffix(name, ".go") || f.IsDir() {
			empty = false
			break
		}
	}
	if !empty {
		fmt.Fprintf(os.Stderr, "go: to add module requirements and sums:\n\tgo mod tidy\n")
	}
}

// CreateWorkFile initializes a new workspace by creating a go.work file.
func CreateWorkFile(ctx context.Context, workFile string, modDirs []string) {
	if _, err := fsys.Stat(workFile); err == nil {
		base.Fatalf("go: %s already exists", workFile)
	}

	goV := LatestGoVersion() // Use current Go version by default
	workF := new(modfile.WorkFile)
	workF.Syntax = new(modfile.FileSyntax)
	workF.AddGoStmt(goV)

	for _, dir := range modDirs {
		_, f, err := ReadModFile(filepath.Join(dir, "go.mod"), nil)
		if err != nil {
			if os.IsNotExist(err) {
				base.Fatalf("go: creating workspace file: no go.mod file exists in directory %v", dir)
			}
			base.Fatalf("go: error parsing go.mod in directory %s: %v", dir, err)
		}
		workF.AddDirectory(ToDirectoryPath(dir), f.Module.Mod.Path)
	}

	data := modfile.Format(workF.Syntax)
	lockedfile.Write(workFile, bytes.NewReader(data), 0666)
}

// fixVersion returns a modfile.VersionFixer implemented using the Query function.
//
// It resolves commit hashes and branch names to versions,
// canonicalizes versions that appeared in early vgo drafts,
// and does nothing for versions that already appear to be canonical.
//
// The VersionFixer sets 'fixed' if it ever returns a non-canonical version.
func fixVersion(ctx context.Context, state *State, fixed *bool) modfile.VersionFixer {
	return func(path, vers string) (resolved string, err error) {
		defer func() {
			if err == nil && resolved != vers {
				*fixed = true
			}
		}()

		// Special case: remove the old -gopkgin- hack.
		if strings.HasPrefix(path, "gopkg.in/") && strings.Contains(vers, "-gopkgin-") {
			vers = vers[strings.Index(vers, "-gopkgin-")+len("-gopkgin-"):]
		}

		// fixVersion is called speculatively on every
		// module, version pair from every go.mod file.
		// Avoid the query if it looks OK.
		_, pathMajor, ok := module.SplitPathVersion(path)
		if !ok {
			return "", &module.ModuleError{
				Path: path,
				Err: &module.InvalidVersionError{
					Version: vers,
					Err:     fmt.Errorf("malformed module path %q", path),
				},
			}
		}
		if vers != "" && module.CanonicalVersion(vers) == vers {
			if err := module.CheckPathMajor(vers, pathMajor); err != nil {
				return "", module.VersionError(module.Version{Path: path, Version: vers}, err)
			}
			return vers, nil
		}

		if state.Mod != "mod" {
			// Don't perform a query, but do indicate go.mod needs to be fixed.
			*fixed = true
			return "", nil
		}
		info, err := Query(ctx, path, vers, "", nil)
		if err != nil {
			return "", err
		}
		return info.Version, nil
	}
}

// makeMainModules creates a MainModuleSet and associated variables according to
// the given main modules.
func makeMainModules(workspaceMode bool, ms []module.Version, rootDirs []string, modFiles []*modfile.File, indices []*modFileIndex, workFileGoVersion string) *MainModuleSet {
	for _, m := range ms {
		if m.Version != "" {
			panic("mainModulesCalled with module.Version with non empty Version field: " + fmt.Sprintf("%#v", m))
		}
	}
	modRootContainingCWD := findModuleRoot(base.Cwd())
	mainModules := &MainModuleSet{
		workspaceMode:     workspaceMode,
		versions:          ms[:len(ms):len(ms)],
		inGorootSrc:       map[module.Version]bool{},
		pathPrefix:        map[module.Version]string{},
		modRoot:           map[module.Version]string{},
		modFiles:          map[module.Version]*modfile.File{},
		indices:           map[module.Version]*modFileIndex{},
		workFileGoVersion: workFileGoVersion,
	}
	for i, m := range ms {
		mainModules.pathPrefix[m] = m.Path
		mainModules.modRoot[m] = rootDirs[i]
		mainModules.modFiles[m] = modFiles[i]
		mainModules.indices[m] = indices[i]

		if mainModules.modRoot[m] == modRootContainingCWD {
			mainModules.modContainingCWD = m
		}

		if rel := search.InDir(rootDirs[i], cfg.GOROOTsrc); rel != "" {
			mainModules.inGorootSrc[m] = true
			if m.Path == "std" {
				// The "std" module in GOROOT/src is the Go standard library. Unlike other
				// modules, the packages in the "std" module have no import-path prefix.
				//
				// Modules named "std" outside of GOROOT/src do not receive this special
				// treatment, so it is possible to run 'go test .' in other GOROOTs to
				// test individual packages using a combination of the modified package
				// and the ordinary standard library.
				// (See https://golang.org/issue/30756.)
				mainModules.pathPrefix[m] = ""
			}
		}
	}
	return mainModules
}

// requirementsFromModFiles returns the set of non-excluded requirements from
// the global modFile.
func requirementsFromModFiles(ctx context.Context, state *State, modFiles []*modfile.File) *Requirements {
	rootCap := 0
	for i := range modFiles {
		rootCap += len(modFiles[i].Require)
	}
	roots := make([]module.Version, 0, rootCap)
	mPathCount := make(map[string]int)
	for _, m := range MainModules.Versions() {
		mPathCount[m.Path] = 1
	}
	direct := map[string]bool{}
	for _, modFile := range modFiles {
	requirement:
		for _, r := range modFile.Require {
			// TODO(#45713): Maybe join
			for _, mainModule := range MainModules.Versions() {
				if index := MainModules.Index(mainModule); index != nil && index.exclude[r.Mod] {
					if state.Mod == "mod" {
						fmt.Fprintf(os.Stderr, "go: dropping requirement on excluded version %s %s\n", r.Mod.Path, r.Mod.Version)
					} else {
						fmt.Fprintf(os.Stderr, "go: ignoring requirement on excluded version %s %s\n", r.Mod.Path, r.Mod.Version)
					}
					continue requirement
				}
			}

			roots = append(roots, r.Mod)
			if !r.Indirect {
				direct[r.Mod.Path] = true
			}
		}
	}
	module.Sort(roots)
	rs := newRequirements(pruningForGoVersion(MainModules.GoVersion()), roots, direct)
	return rs
}

// checkModCommonFlags checks the values of several flags common to module-aware
// commands like -mod. If modules are not enabled (state is nil), these flags
// must not be set explicitly.
//
// checkModCommonFlags sets cfg.BuildMod and BuildModReason in module-aware mode
// if -mod was not set explicitly.
// TODO(#40775): set a field in State instead.
func checkModCommonFlags(state *State) error {
	if state == nil {
		if cfg.BuildMod != "" && !base.InGOFLAGS("-mod") {
			return fmt.Errorf("flag -mod=%s only valid when using modules", cfg.BuildMod)
		}
		if cfg.ModCacheRW && !base.InGOFLAGS("-modcacherw") {
			return fmt.Errorf("flag -modcacherw only valid when using modules")
		}
		if cfg.ModFile != "" && !base.InGOFLAGS("-mod") {
			return fmt.Errorf("flag -modfile only valid when using modules")
		}
		return nil
	}

	if cfg.BuildMod != "" {
		switch cfg.BuildMod {
		case "mod", "readonly", "vendor":
		default:
			return fmt.Errorf("-mod=%s not supported (can be '', 'mod', 'readonly', or 'vendor')", cfg.BuildMod)
		}
		if state.WorkFilePath != "" && cfg.BuildMod != "readonly" {
			return fmt.Errorf("-mod may only be set to readonly when in workspace mode, but it is set to %q"+
				"\n\tRemove the -mod flag to use the default readonly value,"+
				"\n\tor set -workfile=off to disable workspace mode.", cfg.BuildMod)
		}
		return nil
	}

	return nil
}

// maybeEnableVendoring checks whether automatic vendoring should be enabled.
// This may change state.Mod.
func maybeEnableVendoring(state *State) {
	if cfg.BuildMod != "" || state.Opts.ForceBuildMod != "" || len(modRoots) != 1 {
		return
	}
	if fi, err := fsys.Stat(filepath.Join(modRoots[0], "vendor")); err != nil || !fi.IsDir() {
		return
	}
	index := MainModules.GetSingleIndexOrNil()
	if index == nil {
		panic("main module's go.mod not indexed")
	}
	if semver.Compare(index.goVersionV, "v1.14") < 0 {
		return
	}
	state.Mod = "vendor"
	state.ModReason = "Go version in go.mod is at least 1.14 and vendor directory exists."
	web.PanicIfNetworkUsed = true
}

var _ = TODOWorkspaces("In workspace mode, mod will not be readonly for go mod download," +
	"verify, graph, and why. Implement support for go mod download and add test cases" +
	"to ensure verify, graph, and why work properly.")

func mustHaveCompleteRequirements(state *State) bool {
	return state.Mod != "mod" && state.WorkFilePath == ""
}

// convertLegacyConfig imports module requirements from a legacy vendoring
// configuration file, if one is present.
func convertLegacyConfig(state *State, modFile *modfile.File, modRoot string) (from string, err error) {
	noneSelected := func(path string) (version string) { return "none" }
	queryPackage := func(path, rev string) (module.Version, error) {
		pkgMods, modOnly, err := QueryPattern(context.TODO(), state, path, rev, noneSelected, nil)
		if err != nil {
			return module.Version{}, err
		}
		if len(pkgMods) > 0 {
			return pkgMods[0].Mod, nil
		}
		return modOnly.Mod, nil
	}
	for _, name := range altConfigs {
		cfg := filepath.Join(modRoot, name)
		data, err := os.ReadFile(cfg)
		if err == nil {
			convert := modconv.Converters[name]
			if convert == nil {
				return "", nil
			}
			cfg = filepath.ToSlash(cfg)
			err := modconv.ConvertLegacyConfig(modFile, cfg, data, queryPackage)
			return name, err
		}
	}
	return "", nil
}

// addGoStmt adds a go directive to the go.mod file if it does not already
// include one. The 'go' version added, if any, is the latest version supported
// by this toolchain.
func addGoStmt(modFile *modfile.File, mod module.Version, v string) {
	if modFile.Go != nil && modFile.Go.Version != "" {
		return
	}
	if err := modFile.AddGoStmt(v); err != nil {
		base.Fatalf("go: internal error: %v", err)
	}
	rawGoVersion.Store(mod, v)
}

// LatestGoVersion returns the latest version of the Go language supported by
// this toolchain, like "1.17".
func LatestGoVersion() string {
	tags := build.Default.ReleaseTags
	version := tags[len(tags)-1]
	if !strings.HasPrefix(version, "go") || !modfile.GoVersionRE.MatchString(version[2:]) {
		base.Fatalf("go: internal error: unrecognized default version %q", version)
	}
	return version[2:]
}

// priorGoVersion returns the Go major release immediately preceding v,
// or v itself if v is the first Go major release (1.0) or not a supported
// Go version.
func priorGoVersion(v string) string {
	vTag := "go" + v
	tags := build.Default.ReleaseTags
	for i, tag := range tags {
		if tag == vTag {
			if i == 0 {
				return v
			}

			version := tags[i-1]
			if !strings.HasPrefix(version, "go") || !modfile.GoVersionRE.MatchString(version[2:]) {
				base.Fatalf("go: internal error: unrecognized version %q", version)
			}
			return version[2:]
		}
	}
	return v
}

var altConfigs = []string{
	"Gopkg.lock",

	"GLOCKFILE",
	"Godeps/Godeps.json",
	"dependencies.tsv",
	"glide.lock",
	"vendor.conf",
	"vendor.yml",
	"vendor/manifest",
	"vendor/vendor.json",

	".git/config",
}

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

func findWorkspaceFile(dir string) (root string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		f := filepath.Join(dir, "go.work")
		if fi, err := fsys.Stat(f); err == nil && !fi.IsDir() {
			return f
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		if d == cfg.GOROOT {
			_ = TODOWorkspaces("If we end up checking in a go.work file to GOROOT/src," +
				"remove this case.")
			return "" // As a special case, don't cross GOROOT to find a go.work file.
		}
		dir = d
	}
	return ""
}

func findAltConfig(dir string) (root, name string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)
	if rel := search.InDir(dir, cfg.BuildContext.GOROOT); rel != "" {
		// Don't suggest creating a module from $GOROOT/.git/config
		// or a config file found in any parent of $GOROOT (see #34191).
		return "", ""
	}
	for {
		for _, name := range altConfigs {
			if fi, err := fsys.Stat(filepath.Join(dir, name)); err == nil && !fi.IsDir() {
				return dir, name
			}
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return "", ""
}

func findModulePath(dir string) (string, error) {
	// TODO(bcmills): once we have located a plausible module path, we should
	// query version control (if available) to verify that it matches the major
	// version of the most recent tag.
	// See https://golang.org/issue/29433, https://golang.org/issue/27009, and
	// https://golang.org/issue/31549.

	// Cast about for import comments,
	// first in top-level directory, then in subdirectories.
	list, _ := os.ReadDir(dir)
	for _, info := range list {
		if info.Type().IsRegular() && strings.HasSuffix(info.Name(), ".go") {
			if com := findImportComment(filepath.Join(dir, info.Name())); com != "" {
				return com, nil
			}
		}
	}
	for _, info1 := range list {
		if info1.IsDir() {
			files, _ := os.ReadDir(filepath.Join(dir, info1.Name()))
			for _, info2 := range files {
				if info2.Type().IsRegular() && strings.HasSuffix(info2.Name(), ".go") {
					if com := findImportComment(filepath.Join(dir, info1.Name(), info2.Name())); com != "" {
						return path.Dir(com), nil
					}
				}
			}
		}
	}

	// Look for Godeps.json declaring import path.
	data, _ := os.ReadFile(filepath.Join(dir, "Godeps/Godeps.json"))
	var cfg1 struct{ ImportPath string }
	json.Unmarshal(data, &cfg1)
	if cfg1.ImportPath != "" {
		return cfg1.ImportPath, nil
	}

	// Look for vendor.json declaring import path.
	data, _ = os.ReadFile(filepath.Join(dir, "vendor/vendor.json"))
	var cfg2 struct{ RootPath string }
	json.Unmarshal(data, &cfg2)
	if cfg2.RootPath != "" {
		return cfg2.RootPath, nil
	}

	// Look for path in GOPATH.
	var badPathErr error
	for _, gpdir := range filepath.SplitList(cfg.BuildContext.GOPATH) {
		if gpdir == "" {
			continue
		}
		if rel := search.InDir(dir, filepath.Join(gpdir, "src")); rel != "" && rel != "." {
			path := filepath.ToSlash(rel)
			// gorelease will alert users publishing their modules to fix their paths.
			if err := module.CheckImportPath(path); err != nil {
				badPathErr = err
				break
			}
			return path, nil
		}
	}

	reason := "outside GOPATH, module path must be specified"
	if badPathErr != nil {
		// return a different error message if the module was in GOPATH, but
		// the module path determined above would be an invalid path.
		reason = fmt.Sprintf("bad module path inferred from directory in GOPATH: %v", badPathErr)
	}
	msg := `cannot determine module path for source directory %s (%s)

Example usage:
	'go mod init example.com/m' to initialize a v0 or v1 module
	'go mod init example.com/m/v2' to initialize a v2 module

Run 'go help mod init' for more information.
`
	return "", fmt.Errorf(msg, dir, reason)
}

var (
	importCommentRE = lazyregexp.New(`(?m)^package[ \t]+[^ \t\r\n/]+[ \t]+//[ \t]+import[ \t]+(\"[^"]+\")[ \t]*\r?\n`)
)

func findImportComment(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	m := importCommentRE.FindSubmatch(data)
	if m == nil {
		return ""
	}
	path, err := strconv.Unquote(string(m[1]))
	if err != nil {
		return ""
	}
	return path
}

// WriteGoMod writes the current build list back to go.mod.
func WriteGoMod(ctx context.Context, state *State) error {
	requirements = LoadModFile(ctx, state)
	return writeRequirements(ctx, state)
}

// writeRequirements rewrites the go.mod file (if there is one) using the
// global requirements variable.
func writeRequirements(ctx context.Context, state *State) error {
	if state.WorkFilePath != "" {
		// go.mod files aren't updated in workspace mode, but we still want to
		// update the go.work.sum file.
		if err := modfetch.WriteGoSum(keepSums(ctx, state, loaded, requirements, addBuildListZipSums), mustHaveCompleteRequirements(state)); err != nil {
			base.Fatalf("go: %v", err)
		}
		return nil
	}
	mainModule := MainModules.mustGetSingleMainModule()
	modFile := MainModules.ModFile(mainModule)
	if modFile == nil {
		// command-line-arguments has no .mod file to write.
		return nil
	}
	modFilePath := modFilePath(MainModules.ModRoot(mainModule))

	var list []*modfile.Require
	for _, m := range requirements.rootModules {
		list = append(list, &modfile.Require{
			Mod:      m,
			Indirect: !requirements.direct[m.Path],
		})
	}
	if modFile.Go == nil || modFile.Go.Version == "" {
		modFile.AddGoStmt(modFileGoVersion(modFile))
	}
	if semver.Compare("v"+modFileGoVersion(modFile), separateIndirectVersionV) < 0 {
		modFile.SetRequire(list)
	} else {
		modFile.SetRequireSeparateIndirect(list)
	}
	modFile.Cleanup()

	index := MainModules.GetSingleIndexOrNil()
	dirty := index.modFileIsDirty(modFile)
	if dirty && state.Mod != "mod" {
		// If we're about to fail due to -mod=readonly,
		// prefer to report a dirty go.mod over a dirty go.sum
		return newGoModDirtyError(state)
	}

	if !dirty && cfg.CmdName != "mod tidy" {
		// The go.mod file has the same semantic content that it had before
		// (but not necessarily the same exact bytes).
		// Don't write go.mod, but write go.sum in case we added or trimmed sums.
		// 'go mod init' shouldn't write go.sum, since it will be incomplete.
		if cfg.CmdName != "mod init" {
			if err := modfetch.WriteGoSum(keepSums(ctx, state, loaded, requirements, addBuildListZipSums), mustHaveCompleteRequirements(state)); err != nil {
				base.Fatalf("go: %v", err)
			}
		}
		return nil
	}
	if _, ok := fsys.OverlayPath(modFilePath); ok {
		if dirty {
			base.Fatalf("go: updates to go.mod needed, but go.mod is part of the overlay specified with -overlay")
		}
		return nil
	}

	new, err := modFile.Format()
	if err != nil {
		base.Fatalf("go: %v", err)
	}
	defer func() {
		// At this point we have determined to make the go.mod file on disk equal to new.
		MainModules.SetIndex(mainModule, indexModFile(new, modFile, mainModule, true, false))

		// Update go.sum after releasing the side lock and refreshing the index.
		// 'go mod init' shouldn't write go.sum, since it will be incomplete.
		// TODO(#45551): remove this special case or express it without the
		// command name.
		if cfg.CmdName != "mod init" {
			if err := modfetch.WriteGoSum(keepSums(ctx, state, loaded, requirements, addBuildListZipSums), mustHaveCompleteRequirements(state)); err != nil {
				base.Fatalf("go: %v", err)
			}
		}
	}()

	// Make a best-effort attempt to acquire the side lock, only to exclude
	// previous versions of the 'go' command from making simultaneous edits.
	if unlock, err := modfetch.SideLock(); err == nil {
		defer unlock()
	}

	errNoChange := errors.New("no update needed")

	err = lockedfile.Transform(modFilePath, func(old []byte) ([]byte, error) {
		if bytes.Equal(old, new) {
			// The go.mod file is already equal to new, possibly as the result of some
			// other process.
			return nil, errNoChange
		}

		if index != nil && !bytes.Equal(old, index.data) {
			// The contents of the go.mod file have changed. In theory we could add all
			// of the new modules to the build list, recompute, and check whether any
			// module in *our* build list got bumped to a different version, but that's
			// a lot of work for marginal benefit. Instead, fail the command: if users
			// want to run concurrent commands, they need to start with a complete,
			// consistent module definition.
			return nil, fmt.Errorf("existing contents have changed since last read")
		}

		return new, nil
	})

	if err != nil && err != errNoChange {
		return fmt.Errorf("updating go.mod: %w", err)
	}
	return nil
}

// keepSums returns the set of modules (and go.mod file entries) for which
// checksums would be needed in order to reload the same set of packages
// loaded by the most recent call to LoadPackages or ImportFromFiles,
// including any go.mod files needed to reconstruct the MVS result,
// in addition to the checksums for every module in keepMods.
func keepSums(ctx context.Context, state *State, ld *loader, rs *Requirements, which whichSums) map[module.Version]bool {
	// Every module in the full module graph contributes its requirements,
	// so in order to ensure that the build list itself is reproducible,
	// we need sums for every go.mod in the graph (regardless of whether
	// that version is selected).
	keep := make(map[module.Version]bool)

	// Add entries for modules in the build list with paths that are prefixes of
	// paths of loaded packages. We need to retain sums for all of these modules —
	// not just the modules containing the actual packages — in order to rule out
	// ambiguous import errors the next time we load the package.
	if ld != nil {
		for _, pkg := range ld.pkgs {
			// We check pkg.mod.Path here instead of pkg.inStd because the
			// pseudo-package "C" is not in std, but not provided by any module (and
			// shouldn't force loading the whole module graph).
			if pkg.testOf != nil || (pkg.mod.Path == "" && pkg.err == nil) || module.CheckImportPath(pkg.path) != nil {
				continue
			}

			if rs.pruning == pruned && pkg.mod.Path != "" {
				if v, ok := rs.rootSelected(pkg.mod.Path); ok && v == pkg.mod.Version {
					// pkg was loaded from a root module, and because the main module has
					// a pruned module graph we do not check non-root modules for
					// conflicts for packages that can be found in roots. So we only need
					// the checksums for the root modules that may contain pkg, not all
					// possible modules.
					for prefix := pkg.path; prefix != "."; prefix = path.Dir(prefix) {
						if v, ok := rs.rootSelected(prefix); ok && v != "none" {
							m := module.Version{Path: prefix, Version: v}
							r, _ := resolveReplacement(m)
							keep[r] = true
						}
					}
					continue
				}
			}

			mg, _ := rs.Graph(ctx, state)
			for prefix := pkg.path; prefix != "."; prefix = path.Dir(prefix) {
				if v := mg.Selected(prefix); v != "none" {
					m := module.Version{Path: prefix, Version: v}
					r, _ := resolveReplacement(m)
					keep[r] = true
				}
			}
		}
	}

	if rs.graph.Load() == nil {
		// We haven't needed to load the module graph so far.
		// Save sums for the root modules (or their replacements), but don't
		// incur the cost of loading the graph just to find and retain the sums.
		for _, m := range rs.rootModules {
			r, _ := resolveReplacement(m)
			keep[modkey(r)] = true
			if which == addBuildListZipSums {
				keep[r] = true
			}
		}
	} else {
		mg, _ := rs.Graph(ctx, state)
		mg.WalkBreadthFirst(func(m module.Version) {
			if _, ok := mg.RequiredBy(m); ok {
				// The requirements from m's go.mod file are present in the module graph,
				// so they are relevant to the MVS result regardless of whether m was
				// actually selected.
				r, _ := resolveReplacement(m)
				keep[modkey(r)] = true
			}
		})

		if which == addBuildListZipSums {
			for _, m := range mg.BuildList() {
				r, _ := resolveReplacement(m)
				keep[r] = true
			}
		}
	}

	return keep
}

type whichSums int8

const (
	loadedZipSumsOnly = whichSums(iota)
	addBuildListZipSums
)

// modKey returns the module.Version under which the checksum for m's go.mod
// file is stored in the go.sum file.
func modkey(m module.Version) module.Version {
	return module.Version{Path: m.Path, Version: m.Version + "/go.mod"}
}

func suggestModulePath(path string) string {
	var m string

	i := len(path)
	for i > 0 && ('0' <= path[i-1] && path[i-1] <= '9' || path[i-1] == '.') {
		i--
	}
	url := path[:i]
	url = strings.TrimSuffix(url, "/v")
	url = strings.TrimSuffix(url, "/")

	f := func(c rune) bool {
		return c > '9' || c < '0'
	}
	s := strings.FieldsFunc(path[i:], f)
	if len(s) > 0 {
		m = s[0]
	}
	m = strings.TrimLeft(m, "0")
	if m == "" || m == "1" {
		return url + "/v2"
	}

	return url + "/v" + m
}

func suggestGopkgIn(path string) string {
	var m string
	i := len(path)
	for i > 0 && (('0' <= path[i-1] && path[i-1] <= '9') || (path[i-1] == '.')) {
		i--
	}
	url := path[:i]
	url = strings.TrimSuffix(url, ".v")
	url = strings.TrimSuffix(url, "/v")
	url = strings.TrimSuffix(url, "/")

	f := func(c rune) bool {
		return c > '9' || c < '0'
	}
	s := strings.FieldsFunc(path, f)
	if len(s) > 0 {
		m = s[0]
	}

	m = strings.TrimLeft(m, "0")

	if m == "" {
		return url + ".v1"
	}
	return url + ".v" + m
}
