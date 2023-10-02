// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tool implements the “go tool” command.
package tool

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"internal/platform"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"cmd/go/internal/base"
	"cmd/go/internal/cache"
	"cmd/go/internal/cfg"
	"cmd/go/internal/load"
	"cmd/go/internal/modload"
	"cmd/go/internal/str"
	"cmd/go/internal/work"

	"golang.org/x/mod/modfile"
)

var CmdTool = &base.Command{
	Run:       runTool,
	UsageLine: "go tool [-n] command [args...]",
	Short:     "run specified go tool",
	Long: `
Tool runs the go tool command identified by the arguments.

Go ships with a number of builtin tools, and additional tools
may be defined in the go.mod of the current module.

With no arguments it prints the list of known tools.

The -n flag causes tool to print the command that would be
executed but not execute it.

For more about each builtin tool command, see 'go doc cmd/<command>'.
`,
}

var toolN bool

// Return whether tool can be expected in the gccgo tool directory.
// Other binaries could be in the same directory so don't
// show those with the 'go tool' command.
func isGccgoTool(tool string) bool {
	switch tool {
	case "cgo", "fix", "cover", "godoc", "vet":
		return true
	}
	return false
}

func init() {
	base.AddChdirFlag(&CmdTool.Flag)
	CmdTool.Flag.BoolVar(&toolN, "n", false, "")
}

func runTool(ctx context.Context, cmd *base.Command, args []string) {
	if len(args) == 0 {
		listTools()
		return
	}
	toolName := args[0]

	toolPath, err := base.ToolPath(toolName)
	if err != nil {
		if toolName == "dist" && len(args) > 1 && args[1] == "list" {
			// cmd/distpack removes the 'dist' tool from the toolchain to save space,
			// since it is normally only used for building the toolchain in the first
			// place. However, 'go tool dist list' is useful for listing all supported
			// platforms.
			//
			// If the dist tool does not exist, impersonate this command.
			if impersonateDistList(args[2:]) {
				return
			}
		}

		if modFile := modFileForTools(); modFile != nil {
			for _, tool := range modFile.Tool {
				if path.Base(tool.Path) == toolName || tool.Path == toolName {
					buildAndRunModtool(ctx, modFile, tool.Path, args[1:])
					return
				}
			}
		}

		// Emit the usual error for the missing tool.
		_ = base.Tool(toolName)
	}

	if toolN {
		cmd := toolPath
		if len(args) > 1 {
			cmd += " " + strings.Join(args[1:], " ")
		}
		fmt.Printf("%s\n", cmd)
		return
	}
	args[0] = toolPath // in case the tool wants to re-exec itself, e.g. cmd/dist
	toolCmd := &exec.Cmd{
		Path:   toolPath,
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	err = toolCmd.Start()
	if err == nil {
		c := make(chan os.Signal, 100)
		signal.Notify(c)
		go func() {
			for sig := range c {
				toolCmd.Process.Signal(sig)
			}
		}()
		err = toolCmd.Wait()
		signal.Stop(c)
		close(c)
	}
	if err != nil {
		// Only print about the exit status if the command
		// didn't even run (not an ExitError) or it didn't exit cleanly
		// or we're printing command lines too (-x mode).
		// Assume if command exited cleanly (even with non-zero status)
		// it printed any messages it wanted to print.
		if e, ok := err.(*exec.ExitError); !ok || !e.Exited() || cfg.BuildX {
			fmt.Fprintf(os.Stderr, "go tool %s: %s\n", toolName, err)
		}
		base.SetExitStatus(1)
		return
	}
}

// listTools prints a list of the available tools in the tools directory.
func listTools() {
	f, err := os.Open(build.ToolDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go: no tool directory: %s\n", err)
		base.SetExitStatus(2)
		return
	}
	defer f.Close()
	names, err := f.Readdirnames(-1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "go: can't read tool directory: %s\n", err)
		base.SetExitStatus(2)
		return
	}

	modFile := modFileForTools()
	count := map[string]int{}
	if modFile != nil {
		fmt.Println("# builtin tools")
	}

	sort.Strings(names)
	for _, name := range names {
		// Unify presentation by going to lower case.
		// If it's windows, don't show the .exe suffix.
		name = strings.TrimSuffix(strings.ToLower(name), cfg.ToolExeSuffix())

		// The tool directory used by gccgo will have other binaries
		// in addition to go tools. Only display go tools here.
		if cfg.BuildToolchainName == "gccgo" && !isGccgoTool(name) {
			continue
		}
		count[name] += 1
		fmt.Println(name)
	}

	if modFile == nil {
		return
	}

	fmt.Println("# go.mod tools")
	for _, tool := range modFile.Tool {
		count[path.Base(tool.Path)] += 1
	}

	for _, tool := range modFile.Tool {
		if count[path.Base(tool.Path)] > 1 {
			fmt.Println(tool.Path)
		} else {
			fmt.Println(path.Base(tool.Path))
		}
	}
}

func impersonateDistList(args []string) (handled bool) {
	fs := flag.NewFlagSet("go tool dist list", flag.ContinueOnError)
	jsonFlag := fs.Bool("json", false, "produce JSON output")
	brokenFlag := fs.Bool("broken", false, "include broken ports")

	// The usage for 'go tool dist' claims that
	// “All commands take -v flags to emit extra information”,
	// but list -v appears not to have any effect.
	_ = fs.Bool("v", false, "emit extra information")

	if err := fs.Parse(args); err != nil || len(fs.Args()) > 0 {
		// Unrecognized flag or argument.
		// Force fallback to the real 'go tool dist'.
		return false
	}

	if !*jsonFlag {
		for _, p := range platform.List {
			if !*brokenFlag && platform.Broken(p.GOOS, p.GOARCH) {
				continue
			}
			fmt.Println(p)
		}
		return true
	}

	type jsonResult struct {
		GOOS         string
		GOARCH       string
		CgoSupported bool
		FirstClass   bool
		Broken       bool `json:",omitempty"`
	}

	var results []jsonResult
	for _, p := range platform.List {
		broken := platform.Broken(p.GOOS, p.GOARCH)
		if broken && !*brokenFlag {
			continue
		}
		if *jsonFlag {
			results = append(results, jsonResult{
				GOOS:         p.GOOS,
				GOARCH:       p.GOARCH,
				CgoSupported: platform.CgoSupported(p.GOOS, p.GOARCH),
				FirstClass:   platform.FirstClass(p.GOOS, p.GOARCH),
				Broken:       broken,
			})
		}
	}
	out, err := json.MarshalIndent(results, "", "\t")
	if err != nil {
		return false
	}

	os.Stdout.Write(out)
	return true
}

func modFileForTools() *modfile.File {
	if !modload.Enabled() || !modload.HasModRoot() {
		return nil
	}
	bytes, err := os.ReadFile(modload.ModFilePath())
	if err != nil {
		base.Fatal(err)
	}
	file, err := modfile.Parse(modload.ModFilePath(), bytes, nil)
	if err != nil {
		base.Fatal(err)
	}
	return file
}

func buildAndRunModtool(ctx context.Context, modFile *modfile.File, toolPath string, args []string) {
	work.BuildInit()
	b := work.NewBuilder("")
	defer func() {
		if err := b.Close(); err != nil {
			base.Fatal(err)
		}
	}()
	b.Print = printStderr

	pkgOpts := load.PackageOpts{MainOnly: true}
	p := load.PackagesAndErrors(ctx, pkgOpts, []string{toolPath})[0]
	p.Internal.OmitDebug = true

	cacheDir := filepath.Join(cache.Default().ToolDir(), modFileForTools().Module.Mod.Path)
	toolName := filepath.Base(toolPath)
	p.Target = filepath.Join(cacheDir, toolName)

	if err := os.MkdirAll(cacheDir, 0777); err != nil {
		base.Fatal(err)
	}
	a1 := b.LinkAction(work.ModeInstall, work.ModeBuild, p)
	a := &work.Action{Mode: "go tool", Actor: work.ActorFunc(runBuiltTool), Args: args, Deps: []*work.Action{a1}}
	b.Do(ctx, a)
}

func runBuiltTool(b *work.Builder, ctx context.Context, a *work.Action) error {
	cmdline := str.StringList(work.FindExecCmd(), a.Deps[0].Target, a.Args)

	if toolN {
		fmt.Println(strings.Join(cmdline, " "))
		return nil
	}

	toolCmd := &exec.Cmd{
		Path:   cmdline[0],
		Args:   cmdline[1:],
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	err := toolCmd.Start()
	if err == nil {
		c := make(chan os.Signal, 100)
		signal.Notify(c)
		go func() {
			for sig := range c {
				toolCmd.Process.Signal(sig)
			}
		}()
		err = toolCmd.Wait()
		signal.Stop(c)
		close(c)
	}
	if err != nil {
		// Only print about the exit status if the command
		// didn't even run (not an ExitError)
		// Assume if command exited cleanly (even with non-zero status)
		// it printed any messages it wanted to print.
		if e, ok := err.(*exec.ExitError); ok {
			base.SetExitStatus(e.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "go tool %s: %s\n", filepath.Base(a.Deps[0].Target), err)
			base.SetExitStatus(1)
		}
	}

	return nil
}

func printStderr(args ...any) (int, error) {
	return fmt.Fprint(os.Stderr, args...)
}
