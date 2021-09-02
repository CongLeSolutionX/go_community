// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fix implements the ``go fix'' command.
package fix

import (
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/load"
	"cmd/go/internal/modload"
	"cmd/internal/str"
	"context"
)

var CmdFix = &base.Command{
	Run:       runFix,
	UsageLine: "go fix [packages]",
	Short:     "update packages to use new APIs",
	Long: `
Fix runs the Go fix command on the packages named by the import paths.

For more about fix, see 'go doc cmd/fix'.
For more about specifying packages, see 'go help packages'.

To run fix with specific options, run 'go tool fix'.

See also: go fmt, go vet.
	`,
}

func runFix(ctx context.Context, cmd *base.Command, args []string) {
	modState, err := modload.Init(modload.Opts{})
	if err != nil {
		base.CmdFatalf("%v", err)
	}
	pkgs := load.PackagesAndErrors(ctx, load.PackageOpts{ModState: modState}, args)
	// TODO(#45551): report an error if go.mod or go.sum need to be updated.
	if modState != nil {
		if err := modload.WriteGoMod(ctx); err != nil {
			base.CmdFatalf("%v", err)
		}
	}
	w := 0
	for _, pkg := range pkgs {
		if pkg.Error != nil {
			base.Errorf("%v", pkg.Error)
			continue
		}
		pkgs[w] = pkg
		w++
	}
	pkgs = pkgs[:w]

	printed := false
	for _, pkg := range pkgs {
		if modState != nil && pkg.Module != nil && !pkg.Module.Main {
			if !printed {
				base.CmdLogf("not fixing packages in dependency modules\n")
				printed = true
			}
			continue
		}
		// Use pkg.gofiles instead of pkg.Dir so that
		// the command only applies to this package,
		// not to packages in subdirectories.
		files := base.RelPaths(pkg.InternalAllGoFiles())
		base.Run(str.StringList(cfg.BuildToolexec, base.Tool("fix"), files))
	}
}
