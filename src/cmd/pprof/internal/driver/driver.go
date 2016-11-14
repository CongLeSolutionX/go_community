// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package driver implements the core pprof functionality. It can be
// parameterized with a flag implementation, fetch and symbolize
// mechanisms.
package driver

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"cmd/pprof/internal/plugin"
	"cmd/pprof/internal/report"
	"internal/pprof/profile"
)

// PProf acquires a profile, and symbolizes it using a profile
// manager. Then it generates a report formatted according to the
// options selected through the flags package.
func PProf(eo *plugin.Options) error {
	// Remove any temporary files created during pprof processing.
	defer cleanupTempFiles()

	o := setDefaults(eo)

	src, cmd, err := parseFlags(o)
	if err != nil {
		return err
	}

	p, err := fetchProfiles(src, o)
	if err != nil {
		return err
	}

	if cmd != nil {
		return generateReport(p, cmd, pprofVariables, o)
	}

	return interactive(p, o)
}

func generateReport(p *profile.Profile, cmd []string, vars variables, o *plugin.Options) error {
	p = p.Copy() // Prevent modification to the incoming profile.

	var w io.Writer
	switch output := vars["output"].value; output {
	case "":
		w = os.Stdout
	default:
		o.UI.PrintErr("Generating report in ", output)
		outputFile, err := o.Writer.Open(output)
		if err != nil {
			return err
		}
		defer outputFile.Close()
		w = outputFile
	}

	vars = applyCommandOverrides(cmd, vars)

	// Delay focus after configuring report to get percentages on all samples.
	relative := vars["relative_percentages"].boolValue()
	if relative {
		if err := applyFocus(p, vars, o.UI); err != nil {
			return err
		}
	}
	ropt, err := reportOptions(p, vars)
	if err != nil {
		return err
	}
	c := pprofCommands[cmd[0]]
	if c == nil {
		panic("unexpected nil command")
	}
	ropt.OutputFormat = c.format
	post := c.postProcess
	if len(cmd) == 2 {
		s, err := regexp.Compile(cmd[1])
		if err != nil {
			return fmt.Errorf("parsing argument regexp %s: %v", cmd[1], err)
		}
		ropt.Symbol = s
	}

	rpt := report.New(p, ropt)
	if !relative {
		if err := applyFocus(p, vars, o.UI); err != nil {
			return err
		}
	}

	if err := aggregate(p, vars); err != nil {
		return err
	}

	if post == nil {
		return report.Generate(w, rpt, o.Obj)
	}

	// Capture output into buffer and send to postprocessing command.
	buf := &bytes.Buffer{}
	if err := report.Generate(buf, rpt, o.Obj); err != nil {
		return err
	}
	return post(buf.Bytes(), w, o.UI)
}

func applyCommandOverrides(cmd []string, v variables) variables {
	trim, focus, tagfocus, hide := v["trim"].boolValue(), true, true, true

	switch cmd[0] {
	case "proto", "raw":
		trim, focus, tagfocus, hide = false, false, false, false
		v.set("addresses", "t")
	case "callgrind", "kcachegrind":
		trim = false
		v.set("addresses", "t")
	case "disasm", "weblist":
		trim = false
		v.set("addressnoinlines", "t")
	case "peek":
		trim, focus, hide = false, false, false
	case "list":
		v.set("nodecount", "0")
		v.set("lines", "t")
	case "text", "top", "topproto":
		if v["nodecount"].intValue() == -1 {
			v.set("nodecount", "0")
		}
	default:
		if v["nodecount"].intValue() == -1 {
			v.set("nodecount", "80")
		}
	}
	if trim == false {
		v.set("nodecount", "0")
		v.set("nodefraction", "0")
		v.set("edgefraction", "0")
	}
	if focus == false {
		v.set("focus", "")
		v.set("ignore", "")
	}
	if tagfocus == false {
		v.set("tagfocus", "")
		v.set("tagignore", "")
	}
	if hide == false {
		v.set("hide", "")
		v.set("show", "")
	}
	return v
}

func aggregate(prof *profile.Profile, v variables) error {
	var inlines, function, filename, linenumber, address bool
	switch {
	case v["addresses"].boolValue():
		return nil
	case v["lines"].boolValue():
		inlines = true
		function = true
		filename = true
		linenumber = true
	case v["files"].boolValue():
		inlines = true
		filename = true
	case v["functions"].boolValue():
		inlines = true
		function = true
		filename = true
	case v["noinlines"].boolValue():
		function = true
		filename = true
	case v["addressnoinlines"].boolValue():
		function = true
		filename = true
		linenumber = true
		address = true
	case v["functionnameonly"].boolValue():
		inlines = true
		function = true
	default:
		return fmt.Errorf("unexpected granularity")
	}
	return prof.Aggregate(inlines, function, filename, linenumber, address)
}

func reportOptions(p *profile.Profile, vars variables) (*report.Options, error) {
	si, mean := vars["sample_index"].value, vars["mean"].boolValue()
	value, meanDiv, sample, err := sampleFormat(p, si, mean)
	if err != nil {
		return nil, err
	}

	stype := sample.Type
	if mean {
		stype = "mean_" + stype
	}

	if vars["divide_by"].floatValue() == 0 {
		return nil, fmt.Errorf("zero divisor specified")
	}

	ropt := &report.Options{
		CumSort:             vars["cum"].boolValue(),
		CallTree:            vars["call_tree"].boolValue(),
		DropNegative:        vars["drop_negative"].boolValue(),
		PositivePercentages: vars["positive_percentages"].boolValue(),

		CompactLabels: vars["compact_labels"].boolValue(),
		Ratio:         1 / vars["divide_by"].floatValue(),

		NodeCount:    vars["nodecount"].intValue(),
		NodeFraction: vars["nodefraction"].floatValue(),
		EdgeFraction: vars["edgefraction"].floatValue(),

		SampleValue:       value,
		SampleMeanDivisor: meanDiv,
		SampleType:        stype,
		SampleUnit:        sample.Unit,

		OutputUnit: vars["unit"].value,

		SourcePath: vars["source_path"].stringValue(),
	}

	if len(p.Mapping) > 0 && p.Mapping[0].File != "" {
		ropt.Title = filepath.Base(p.Mapping[0].File)
	}

	return ropt, nil
}

type sampleValueFunc func([]int64) int64

// sampleFormat returns a function to extract values out of a profile.Sample,
// and the type/units of those values.
func sampleFormat(p *profile.Profile, sampleIndex string, mean bool) (value, meanDiv sampleValueFunc, v *profile.ValueType, err error) {
	if len(p.SampleType) == 0 {
		return nil, nil, nil, fmt.Errorf("profile has no samples")
	}
	index, err := locateSampleIndex(p, sampleIndex)
	if err != nil {
		return nil, nil, nil, err
	}
	value = valueExtractor(index)
	if mean {
		meanDiv = valueExtractor(0)
	}
	v = p.SampleType[index]
	return
}

func valueExtractor(ix int) sampleValueFunc {
	return func(v []int64) int64 {
		return v[ix]
	}
}
