// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build generate

package main

//go:generate go run $GOFILE analyzers.go

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

var docTmpl = template.Must(template.New("goVetDoc").Parse(`// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Vet examines Go source code and reports suspicious constructs, such as Printf
calls whose arguments do not align with the format string. Vet uses heuristics
that do not guarantee all reports are genuine problems, but it can find errors
not caught by the compilers.

Vet is normally invoked through the go command.
This command vets the package in the current directory:

	go vet

whereas this one vets the packages whose path is provided:

	go vet my/project/...

Use "go help packages" to see other ways of specifying which packages to vet.

Vet's exit code is non-zero for erroneous invocation of the tool or if a
problem was reported, and 0 otherwise. Note that the tool does not
check every possible problem and depends on unreliable heuristics,
so it should be used as guidance only, not as a firm indicator of
program correctness.

To list the available checks, run "go tool vet help":

{{ .ChecksDescription }}

For details and flags of a particular check, such as printf, run "go tool vet help printf".

By default, all checks are performed.
If any flags are explicitly set to true, only those tests are run.
Conversely, if any flag is explicitly set to false, only those tests are disabled.
Thus -printf=true runs the printf check,
and -printf=false runs all checks except the printf check.

For information on writing a new check, see golang.org/x/tools/go/analysis.

Core flags:

	-c=N
	  	display offending line plus N lines of surrounding context
	-json
	  	emit analysis diagnostics (and errors) in JSON format
*/
package main
`))

func main() {
	if err := mainImpl(); err != nil {
		log.Panic(err)
	}
}

type check struct {
	name        string
	description string
}

func mainImpl() error {
	checks := make([]check, 0, len(analyzers))
	for _, anl := range analyzers {
		doc := anl.Doc
		if summaryEnd := strings.Index(anl.Doc, "\n"); summaryEnd > 0 {
			doc = doc[:summaryEnd]
		}

		checks = append(checks, check{
			name:        anl.Name,
			description: doc,
		})
	}

	var maxLength int
	for _, ch := range checks {
		if l := len(ch.name); l > maxLength {
			maxLength = l
		}
	}
	format := fmt.Sprintf("\t%%-%ds %%s", maxLength) // Example: `%-16s %s`.

	var description strings.Builder
	for i, ch := range checks {
		description.WriteString(fmt.Sprintf(format, ch.name, ch.description))
		if i != len(checks)-1 {
			description.WriteRune('\n')
		}
	}

	f, err := os.Create("doc.go")
	if err != nil {
		return err
	}
	defer f.Close()

	if err := docTmpl.Execute(f, struct {
		ChecksDescription string
	}{
		ChecksDescription: description.String(),
	}); err != nil {
		return err
	}

	return f.Sync()
}
