// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package objabi

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"internal/goexperiment"
)

// Experiment contains the toolchain experiments enabled for the
// current build.
//
// (This is not necessarily the set of experiments the compiler itself
// was built with.)
var Experiment = goexperiment.DefaultFlags

// EnabledExperiments is a list of enabled experiments, as lower-cased
// experiment names.
var EnabledExperiments []string

// GOEXPERIMENT is a comma-separated list of enabled or disabled
// experiments that differ from the default experiment settings.
var GOEXPERIMENT string

// FramePointerEnabled enables the use of platform conventions for
// saving frame pointers.
//
// This used to be an experiment, but now it's always enabled on
// platforms that support it.
//
// Note: must agree with runtime.framepointer_enabled.
var FramePointerEnabled = GOARCH == "amd64" || GOARCH == "arm64"

func init() {
	env := envOr("GOEXPERIMENT", defaultGOEXPERIMENT)

	// GOEXPERIMENT=none disables all experiments.
	if env == "none" {
		Experiment = goexperiment.Flags{}
	} else {
		// Create a map of known experiment names.
		names := make(map[string]reflect.Value)
		rv := reflect.ValueOf(&Experiment).Elem()
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rv.Field(i)
			names[strings.ToLower(rt.Field(i).Name)] = field
		}

		// Parse names.
		for _, f := range strings.Split(env, ",") {
			if f == "" {
				continue
			}
			val := true
			if strings.HasPrefix(f, "no") {
				f, val = f[2:], false
			}
			field, ok := names[f]
			if !ok {
				fmt.Printf("unknown experiment %s\n", f)
				os.Exit(2)
			}
			field.SetBool(val)
		}
	}

	// regabi is only supported on amd64.
	if GOARCH != "amd64" {
		Experiment.Regabi = false
		Experiment.RegabiWrappers = false
		Experiment.RegabiG = false
		Experiment.RegabiReflect = false
		Experiment.RegabiDefer = false
		Experiment.RegabiArgs = false
	}
	// Setting regabi sets working sub-experiments.
	if Experiment.Regabi {
		Experiment.RegabiWrappers = true
		Experiment.RegabiG = true
		Experiment.RegabiReflect = true
		Experiment.RegabiDefer = true
		// Not ready yet:
		//Experiment.RegabiArgs = true
	}
	// Check regabi dependencies.
	if Experiment.RegabiG && !Experiment.RegabiWrappers {
		panic("GOEXPERIMENT regabig requires regabiwrappers")
	}
	if Experiment.RegabiArgs && !(Experiment.RegabiWrappers && Experiment.RegabiG && Experiment.RegabiReflect && Experiment.RegabiDefer) {
		panic("GOEXPERIMENT regabiargs requires regabiwrappers,regabig,regabireflect,regabidefer")
	}

	// Now that all experiment flags are set, get the list of
	// enabled experiments, and the different-from-default list.
	var diff []string
	rv := reflect.ValueOf(&Experiment).Elem()
	rDef := reflect.ValueOf(&goexperiment.DefaultFlags).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		name := strings.ToLower(rt.Field(i).Name)
		val := rv.Field(i).Bool()
		if val {
			EnabledExperiments = append(EnabledExperiments, name)
		}
		if val != rDef.Field(i).Bool() {
			if val {
				diff = append(diff, name)
			} else {
				diff = append(diff, "no"+name)
			}
		}

	}

	// Set GOEXPERIMENT to the delta from the default experiments.
	// This way, GOEXPERIMENT is exactly what a user would set on
	// the command line to get the set of enabled experiments.
	GOEXPERIMENT = strings.Join(diff, ",")
}
