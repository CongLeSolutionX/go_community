// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate ./mkalldocs.sh

package main

import (
	"context"
	"flag"
	"fmt"
	"internal/buildcfg"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"cmd/go/internal/base"
	"cmd/go/internal/bug"
	"cmd/go/internal/cfg"
	"cmd/go/internal/clean"
	"cmd/go/internal/doc"
	"cmd/go/internal/envcmd"
	"cmd/go/internal/fix"
	"cmd/go/internal/fmtcmd"
	"cmd/go/internal/generate"
	"cmd/go/internal/get"
	"cmd/go/internal/gotoolchain"
	"cmd/go/internal/help"
	"cmd/go/internal/list"
	"cmd/go/internal/modcmd"
	"cmd/go/internal/modfetch"
	"cmd/go/internal/modget"
	"cmd/go/internal/modload"
	"cmd/go/internal/run"
	"cmd/go/internal/test"
	"cmd/go/internal/tool"
	"cmd/go/internal/trace"
	"cmd/go/internal/version"
	"cmd/go/internal/vet"
	"cmd/go/internal/work"
	"cmd/go/internal/workcmd"
)

func init() {
	base.Go.Commands = []*base.Command{
		bug.CmdBug,
		work.CmdBuild,
		clean.CmdClean,
		doc.CmdDoc,
		envcmd.CmdEnv,
		fix.CmdFix,
		fmtcmd.CmdFmt,
		generate.CmdGenerate,
		modget.CmdGet,
		work.CmdInstall,
		list.CmdList,
		modcmd.CmdMod,
		workcmd.CmdWork,
		run.CmdRun,
		test.CmdTest,
		tool.CmdTool,
		version.CmdVersion,
		vet.CmdVet,

		help.HelpBuildConstraint,
		help.HelpBuildmode,
		help.HelpC,
		help.HelpCache,
		help.HelpEnvironment,
		help.HelpFileType,
		modload.HelpGoMod,
		help.HelpGopath,
		get.HelpGopathGet,
		modfetch.HelpGoproxy,
		help.HelpImportPath,
		modload.HelpModules,
		modget.HelpModuleGet,
		modfetch.HelpModuleAuth,
		help.HelpPackages,
		modfetch.HelpPrivate,
		test.HelpTestflag,
		test.HelpTestfunc,
		modget.HelpVCS,
	}
}

func main() {
	_ = go11tag
	flag.Usage = base.Usage
	flag.Parse()
	log.SetFlags(0)

	setupGoToolchain()

	args := flag.Args()
	if len(args) < 1 {
		base.Usage()
	}

	if args[0] == "get" || args[0] == "help" {
		if !modload.WillBeEnabled() {
			// Replace module-aware get with GOPATH get if appropriate.
			*modget.CmdGet = *get.CmdGet
		}
	}

	cfg.CmdName = args[0] // for error messages
	if args[0] == "help" {
		help.Help(os.Stdout, args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := cfg.BuildContext.GOPATH; filepath.Clean(gopath) == filepath.Clean(runtime.GOROOT()) {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Some GOPATHs have empty directory elements - ignore them.
			// See issue 21928 for details.
			if p == "" {
				continue
			}
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(os.Stderr, "go: GOPATH entry cannot start with shell metacharacter '~': %q\n", p)
				os.Exit(2)
			}
			if !filepath.IsAbs(p) {
				if cfg.Getenv("GOPATH") == "" {
					// We inferred $GOPATH from $HOME and did a bad job at it.
					// Instead of dying, uninfer it.
					cfg.BuildContext.GOPATH = ""
				} else {
					fmt.Fprintf(os.Stderr, "go: GOPATH entry is relative; must be absolute path: %q.\nFor more details see: 'go help gopath'\n", p)
					os.Exit(2)
				}
			}
		}
	}

	if cfg.GOROOT == "" {
		fmt.Fprintf(os.Stderr, "go: cannot find GOROOT directory: 'go' binary is trimmed and GOROOT is not set\n")
		os.Exit(2)
	}
	if fi, err := os.Stat(cfg.GOROOT); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "go: cannot find GOROOT directory: %v\n", cfg.GOROOT)
		os.Exit(2)
	}

BigCmdLoop:
	for bigCmd := base.Go; ; {
		for _, cmd := range bigCmd.Commands {
			if cmd.Name() != args[0] {
				continue
			}
			if len(cmd.Commands) > 0 {
				bigCmd = cmd
				args = args[1:]
				if len(args) == 0 {
					help.PrintUsage(os.Stderr, bigCmd)
					base.SetExitStatus(2)
					base.Exit()
				}
				if args[0] == "help" {
					// Accept 'go mod help' and 'go mod help foo' for 'go help mod' and 'go help mod foo'.
					help.Help(os.Stdout, append(strings.Split(cfg.CmdName, " "), args[1:]...))
					return
				}
				cfg.CmdName += " " + args[0]
				continue BigCmdLoop
			}
			if !cmd.Runnable() {
				continue
			}
			invoke(cmd, args)
			base.Exit()
			return
		}
		helpArg := ""
		if i := strings.LastIndex(cfg.CmdName, " "); i >= 0 {
			helpArg = " " + cfg.CmdName[:i]
		}
		fmt.Fprintf(os.Stderr, "go %s: unknown command\nRun 'go help%s' for usage.\n", cfg.CmdName, helpArg)
		base.SetExitStatus(2)
		base.Exit()
	}
}

func invoke(cmd *base.Command, args []string) {
	// 'go env' handles checking the build config
	if cmd != envcmd.CmdEnv {
		buildcfg.Check()
		if cfg.ExperimentErr != nil {
			base.Fatalf("go: %v", cfg.ExperimentErr)
		}
	}

	// Set environment (GOOS, GOARCH, etc) explicitly.
	// In theory all the commands we invoke should have
	// the same default computation of these as we do,
	// but in practice there might be skew
	// This makes sure we all agree.
	cfg.OrigEnv = os.Environ()
	cfg.CmdEnv = envcmd.MkEnv()
	for _, env := range cfg.CmdEnv {
		if os.Getenv(env.Name) != env.Value {
			os.Setenv(env.Name, env.Value)
		}
	}

	cmd.Flag.Usage = func() { cmd.Usage() }
	if cmd.CustomFlags {
		args = args[1:]
	} else {
		base.SetFromGOFLAGS(&cmd.Flag)
		cmd.Flag.Parse(args[1:])
		args = cmd.Flag.Args()
	}
	ctx := maybeStartTrace(context.Background())
	ctx, span := trace.StartSpan(ctx, fmt.Sprint("Running ", cmd.Name(), " command"))
	cmd.Run(ctx, cmd, args)
	span.Done()
}

func init() {
	base.Usage = mainUsage
}

func mainUsage() {
	help.PrintUsage(os.Stderr, base.Go)
	os.Exit(2)
}

func maybeStartTrace(pctx context.Context) context.Context {
	if cfg.DebugTrace == "" {
		return pctx
	}

	ctx, close, err := trace.Start(pctx, cfg.DebugTrace)
	if err != nil {
		base.Fatalf("failed to start trace: %v", err)
	}
	base.AtExit(func() {
		if err := close(); err != nil {
			base.Fatalf("failed to stop trace: %v", err)
		}
	})

	return ctx
}

var goLineRE = regexp.MustCompile(`^\s*go\s+([^/#\s]+)\s*$`)

/*
TODO:
 - all the panics
 - GOTOOLCHAIN=auto needs to allow older go lines to be handled by newer local toolchain
 - need to read the toolchain line from go.mod
    - if toolchain line says 'toolchain local', then we use the local toolchain.
 - if we get the go version from the go line, need to read the module version list to decide which to use.
    version list will have two versions in it - current release most recent patch, and previous release most recent patch.
    we want to use the minimum one that is acceptable.
    so if it says 1.19.3 and 1.18.8 and we need go 1.16, we'd use 1.18.8.
    but if we need go 1.19, we use 1.19.3.
 - when we make a choice based on the go line and it's different from the go line itself, we need to add a toolchain line immediately after the go line.
    - for reproducibility and also not re-fetching the module list every single time

 - go install path@version needs to run a lot of this code after downloading the go.mod for path@version,
    and it should ignore the toolchain line.

 - go get needs to watch the go version in dependencies and update the toolchain line in the work module go.mod
    to be at least as big as the minimum go version needed by dependencies

 - devel toolchain has to default to GOTOOLCHAIN=local, releases default to =auto

 - need to fetch GOTOOLCHAIN from go env -w file if not set in environment.

 - probably check that the go line only has 1.X, 1.X.Y, disallow beta1, rc1 etc.

 - go get go@1.20.1

 - go get toolchain@go1.20.1

 - go get go@latest (or just go get go), go get -p go, and go get toolchain@latest

*/

func setupGoToolchain() {
	if !modload.WillBeEnabled() {
		return
	}

	rel := os.Getenv("GOTOOLCHAIN")
	if rel == "" {
		if strings.HasPrefix(runtime.Version(), "go") {
			rel = "auto"
		} else {
			rel = "local"
		}
	}
	if rel == "auto" {
		// find go.mod or go.work
		modFile := modload.EarlyGoMod()
		if modFile == "" {
			return
		}
		data, err := ioutil.ReadFile(modFile)
		if err != nil {
			panic(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if m := goLineRE.FindStringSubmatch(line); m != nil {
				rel = "go" + m[1]
				break
			}
		}
		if rel == "auto" {
			panic("did not find go line")
		}
	}
	if rel == "local" || rel == runtime.Version() {
		// TODO: rel == runtime.Version() is wrong; if it's newer and rel was auto, we can keep using local too.
		// Let the current binary handle the command.
		return
	}

	modload.ForceUseModules = true
	modload.RootMode = modload.NoRoot
	modload.Init()

	path, vers, ok := gotoolchain.Module(rel)
	if !ok {
		panic("bad go version: " + rel)
	}

	m := &modcmd.ModuleJSON{Path: path, Version: vers + "." + runtime.GOOS + "-" + runtime.GOARCH}
	modcmd.DownloadModule(context.Background(), m)
	if m.Error != "" {
		panic(m.Error)
	}

	// look at m.Dir for execute bits
	dir := m.Dir
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(dir, "bin/go"))
		if err != nil {
			panic(err)
		}
		if info.Mode()&0111 == 0 {
			// Set execute bits on bin and pkg/tool files.
			setExecBits := func(dir string) {
				err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if !d.IsDir() {
						info, err := os.Stat(path)
						if err != nil {
							return err
						}
						if err := os.Chmod(path, info.Mode()&0777|0111); err != nil {
							return err
						}
					}
					return nil
				})
				if err != nil {
					panic(err)
				}
			}

			setExecBits(filepath.Join(dir, "bin"))
			setExecBits(filepath.Join(dir, "pkg/tool"))
		}
		os.Setenv("GOROOT", dir)
		err = syscall.Exec(filepath.Join(dir, "bin/go"), os.Args, os.Environ())
		panic(err)
	}
}
