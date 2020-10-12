// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics_test

import (
	"regexp"
	"runtime/metrics"
	"testing"
)

func TestDescriptionNameFormat(t *testing.T) {
	r := regexp.MustCompile("^(?P<name>/[^:]+):(?P<unit>[^:*/]+(?:[*/][^:*/]+)*)$")
	descriptions := metrics.All()
	for _, desc := range descriptions {
		if !r.MatchString(desc.Name) {
			t.Errorf("metrics %q does not match regexp %s", desc.Name, r)
		}
	}
}
