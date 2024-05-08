// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package telemetrycmd implements the "go telemetry" command.
package telemetrycmd

import (
	"cmd/go/internal/base"
	"cmd/internal/telemetry"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var CmdTelemetry = &base.Command{
	UsageLine: "go telemetry [off|local|on]",
	Short:     "manage telemetry data and settings",
	Long: `Telemetry is used to manage Go telemetry data and settings.

Telemetry can be in one of three modes: off, local, or on.

When telemetry is in local mode, counter data is written to the local file
system, but will not be uploaded to remote servers.

When telemetry is disabled, local counter data is neither collected nor uploaded.

When telemetry is enabled, telemetry data is written to the local file system
and periodically sent to https://telemetry.go.dev/. Uploaded data is used to
help improve the Go toolchain and related tools, and it will be published as
part of a public dataset.

For more details, see https://telemetry.go.dev/privacy.
This data is collected in accordance with the Google Privacy Policy
(https://policies.google.com/privacy).

To view the current telemetry mode, run "go telemetry".
To disable telemetry uploading, but keep local data collection, run
"go telemetry local".
To enable both collection and uploading, run “go telemetry on”.
To disable both collection and uploading, run "go telemetry off".

See https://go.dev/doc/telemetry for more information on telemetry.
`,
	Run: runTelemetry,
}

func runTelemetry(ctx context.Context, cmd *base.Command, args []string) {
	if len(args) == 0 {
		fmt.Println(telemetry.Mode())
		return
	}

	if len(args) != 1 {
		cmd.Usage()
	}

	mode := args[0]
	if old := telemetry.Mode(); old == mode {
		return
	}
	switch mode {
	case "local":
		if err := telemetry.SetMode("local"); err != nil {
			base.Fatalf("go: failed to set the telemetry mode to local: %v", err)
		}
	case "off":
		if err := telemetry.SetMode("off"); err != nil {
			base.Fatalf("Failed to disable telemetry: %v", err)
		}
	case "on":
		if err := telemetry.SetMode("on"); err != nil {
			base.Fatalf("go: failed to enable telemetry: %v", err)
		}
		fmt.Fprintln(os.Stderr, telemetryOnMessage())
	default:
		cmd.Usage()
	}
}

func telemetryOnMessage() string {
	return `Telemetry uploading is now enabled and data will be periodically sent to
https://telemetry.go.dev/. Uploaded data is used to help improve the Go
toolchain and related tools, and it will be published as part of a public
dataset.

For more details, see https://telemetry.go.dev/privacy.
This data is collected in accordance with the Google Privacy Policy
(https://policies.google.com/privacy).

To disable telemetry uploading, but keep local data collection, run
“gotelemetry local”.
To disable both collection and uploading, run “gotelemetry off“.`
}

func runClean(ctx context.Context, cmd *base.Command, args []string) {
	// For now, be careful to only remove counter files and reports.
	// It would probably be OK to just remove everything, but it may
	// be useful to preserve the weekends file.
	telemetryDir := telemetry.Dir()
	for dir, suffixes := range map[string][]string{
		filepath.Join(telemetryDir, "local"):  {".v1.count", ".json"},
		filepath.Join(telemetryDir, "upload"): {".json"},
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				base.Errorf("go: failed to read telemetry dir: %v", err)
			}
			continue
		}
		for _, entry := range entries {
			// TODO: use slices.ContainsFunc once it is available in all supported Go
			// versions.
			remove := false
			for _, suffix := range suffixes {
				if strings.HasSuffix(entry.Name(), suffix) {
					remove = true
					break
				}
			}
			if remove {
				path := filepath.Join(dir, entry.Name())
				if err := os.Remove(path); err != nil {
					base.Errorf("go: failed to remove %s: %v", path, err)
				}
			}
		}
	}
}
