// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package load

import (
	"cmd/go/internal/cfg"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// A Printer reports output about a Package.
type Printer interface {
	// Output reports output from building pkg. The arguments are of the form
	// expected by fmt.Print.
	Output(pkg *Package, args ...any)

	// Errorf prints output in the form of `log.Errorf` and reports that
	// building pkg failed. This ensures the output is terminated with a new
	// line if there's any output.
	Errorf(pkg *Package, format string, args ...any)
}

// DefaultPrinter returns the default Printer.
func DefaultPrinter() Printer {
	return defaultPrinter()
}

var defaultPrinter = sync.OnceValue(func() Printer {
	if cfg.BuildJSON {
		return NewJSONPrinter(os.Stdout)
	}
	return &TextPrinter{os.Stderr}
})

func ensureNewline(s string) string {
	if s == "" {
		return ""
	}
	if strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}

// A TextPrinter emits text format output to Writer.
type TextPrinter struct {
	Writer io.Writer
}

func (p *TextPrinter) Output(_ *Package, args ...any) {
	fmt.Fprint(p.Writer, args...)
}

func (p *TextPrinter) Errorf(_ *Package, format string, args ...any) {
	fmt.Fprint(p.Writer, ensureNewline(fmt.Sprintf(format, args...)))
}

// A JSONPrinter emits output about a build in JSON format.
type JSONPrinter struct {
	enc *json.Encoder
}

func NewJSONPrinter(w io.Writer) *JSONPrinter {
	return &JSONPrinter{json.NewEncoder(w)}
}

type jsonBuildEvent struct {
	ImportPath string
	Action     string
	Output     string `json:",omitempty"` // Non-empty if Action == “build-output”
}

func (p *JSONPrinter) Output(pkg *Package, args ...any) {
	ev := &jsonBuildEvent{
		Action: "build-output",
		Output: fmt.Sprint(args...),
	}
	if ev.Output == "" {
		// There's no point in emitting a completely empty output event.
		return
	}
	if pkg != nil {
		ev.ImportPath = pkg.Desc()
	}
	p.enc.Encode(ev)
}

func (p *JSONPrinter) Errorf(pkg *Package, format string, args ...any) {
	s := ensureNewline(fmt.Sprintf(format, args...))
	// For clarity, emit each line as a separate output event.
	for len(s) > 0 {
		i := strings.IndexByte(s, '\n')
		p.Output(pkg, s[:i+1])
		s = s[i+1:]
	}
	ev := &jsonBuildEvent{
		Action: "build-fail",
	}
	if pkg != nil {
		ev.ImportPath = pkg.Desc()
	}
	p.enc.Encode(ev)
}
