// run -race

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 39807: data race in html/template & text/template

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	jsTempl := `
{{- define "jstempl" -}}
var foo = "bar";
{{- end -}}
<script type="application/javascript">
{{ template "jstempl" $ }}
</script>
`

	tpl := template.New("")
	_, err := tpl.New("templ.html").Parse(jsTempl)
	if err != nil {
		log.Fatal(err)
	}

	const numTemplates = 20

	for i := 0; i < numTemplates; i++ {
		_, err = tpl.New(fmt.Sprintf("main%d.html", i)).Parse(`{{ template "templ.html" . }}`)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numTemplates; j++ {
				templ := tpl.Lookup(fmt.Sprintf("main%d.html", j))
				if err := templ.Execute(ioutil.Discard, nil); err != nil {
					log.Fatal(err)
				}

			}
		}()
	}

	wg.Wait()
}
