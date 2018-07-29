// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"cmd/go/internal/cfg"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	goflags   []string
	knownFlag = make(map[string]bool)
)

// AddKnownFlag adds name to the list of known flags for use in GOFLAGS.
func AddKnownFlag(name string) {
	knownFlag[name] = true
}

// GOFLAGS returns the flags from GOFLAGS.
// The list can be assumed to contain one string per flag,
// with each string either beginning with -name or --name.
func GOFLAGS() []string {
	InitGOFLAGS()
	return goflags
}

// InitGOFLAGS initializes the goflags list from from the GOFLAGS environment variable.
// It is idempotent.
func InitGOFLAGS() {
	if goflags != nil { // already initialized
		return
	}

	goflags = strings.Fields(os.Getenv("GOFLAGS"))
	if goflags == nil {
		goflags = []string{}
	}

	// Build list of all flags for all commands.
	// If no command has that flag, then we report the problem.
	// This catches typos while still letting users record flags in GOFLAGS
	// that only apply to a subset of go commands.
	// Commands using CustomFlags can report their flag names
	// by calling AddKnownFlag instead.
	var walkFlags func(*Command)
	walkFlags = func(cmd *Command) {
		for _, sub := range cmd.Commands {
			walkFlags(sub)
		}
		cmd.Flag.VisitAll(func(f *flag.Flag) {
			knownFlag[f.Name] = true
		})
	}
	walkFlags(Go)

	// Ignore bad flag in go env and go bug, because
	// they are what people reach for when debugging
	// a problem, and maybe they're debugging GOFLAGS.
	// (Both will show the GOFLAGS setting if let succeed.)
	hideErrors := cfg.CmdName == "env" || cfg.CmdName == "bug"

	// Each of the words returned by strings.Fields must be its own flag.
	// (To set flag arguments use -x=value instead of -x value.)
	for _, f := range goflags {
		if !strings.HasPrefix(f, "-") || f == "-" || f == "--" || strings.HasPrefix(f, "---") {
			if hideErrors {
				continue
			}
			Fatalf("go: parsing $GOFLAGS: non-flag %q", f)
		}

		name := f[1:]
		if i := strings.Index(name, "="); i >= 0 {
			name = name[:i]
		}
		if !knownFlag[name] {
			if hideErrors {
				continue
			}
			Fatalf("go: parsing $GOFLAGS: unknown flag -%s", name)
		}
	}
}

type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// SetFromGOFLAGS sets the flags in the given flag set using settings in $GOFLAGS.
func SetFromGOFLAGS(flags flag.FlagSet) {
	InitGOFLAGS()
	for _, goflag := range goflags {
		name, value, hasValue := goflag, "", false
		if i := strings.Index(goflag, "="); i >= 0 {
			name, value, hasValue = goflag[:i], goflag[i+1:], true
		}
		if strings.HasPrefix(name, "--") {
			name = name[1:]
		}
		f := flags.Lookup(name[1:])
		if f == nil {
			continue
		}
		if fb, ok := f.Value.(boolFlag); ok && fb.IsBoolFlag() {
			if hasValue {
				if err := fb.Set(value); err != nil {
					fmt.Fprintf(flags.Output(), "invalid boolean value %q for flag -%s (from $GOFLAGS): %v", value, name, err)
					flags.Usage()
				}
			} else {
				if err := fb.Set("true"); err != nil {
					fmt.Fprintf(flags.Output(), "invalid boolean flag -%s (from $GOFLAGS): %v", name, err)
					flags.Usage()
				}
			}
		} else {
			if !hasValue {
				fmt.Fprintf(flags.Output(), "flag needs an argument: -%s (from $GOFLAGS)", name)
				flags.Usage()
			}
			if err := f.Value.Set(value); err != nil {
				fmt.Fprintf(flags.Output(), "invalid value %q for flag -%s (from $GOFLAGS): %v", value, name, err)
				flags.Usage()
			}
		}
	}
}
