// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package toolchain

import (
	"cmd/go/internal/base"
	"cmd/go/internal/cfg"
	"cmd/go/internal/gover"
	"cmd/go/internal/modload"
	"context"
	"os"
	"strings"
)

// TryVersion tries to reinvoke an appropriate Go toolchain for the given Go version.
// If it cannot locate an appropriate Go toolchain, it prints an error and returns.
// (The assumption is the caller has other errors to print about being too old.)
//
// If TryVersion does find an appropriate toolchain, it downloads and runs it.
// If any part of that fails, TryVersion exits with a fatal error.
//
// TryVersion adds the selected Go toolchain version to the command line before
// running it, in the form go@version. The assumption is that the program being
// invoked is go get or some similar program that can take an extra go@version.
func TryVersion(version string) {
	// Only bother at all if the GOTOOLCHAIN setting allows switching.
	env := cfg.Getenv("GOTOOLCHAIN")
	if env != "auto" && env != "path" && !strings.HasSuffix(env, "+auto") && !strings.HasSuffix(env, "+path") {
		return
	}

	if gover.TestVersion != "" && os.Getenv("TESTGO_VERSION_SWITCH") == "1" {
		if gover.IsLang(version) {
			version += ".99"
		}
		os.Args = append(os.Args, "go@"+version)
		base.Errorf("go: switching to go%v", version)
		SwitchTo("go" + version)
	}

	// Look up version as query. This will find the latest point release for the
	// language version indicated by version. So if version is 1.21 or 1.21rc1 or 1.21.4,
	// those all resolve to 1.21.9 if that's the latest point release.
	ctx := context.Background()
	allowed := modload.CheckAllowed
	noneSelected := func(path string) (version string) { return "none" }
	_, m, err := modload.QueryPattern(ctx, "go", gover.Lang(version), noneSelected, allowed)
	if err != nil {
		base.Errorf("go: looking up newest %v: %v", gover.Lang(version), err)
		return
	}

	// Assuming
	v := m.Mod.Version
	os.Args = append(os.Args, "go@"+v)
	base.Errorf("go: switching to go%v for %v", v, gover.Lang(version))
	SwitchTo("go" + v)
}
