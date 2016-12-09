// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/proflabel"
	_ "unsafe"
)

//go:linkname runtime_setProfLabel runtime/internal/proflabel.runtime_setProfLabel
func runtime_setProfLabel(labels *proflabel.Labels) {
	getg().labels = labels
}

//go:linkname runtime_getProfLabel runtime/internal/proflabel.runtime_getProfLabel
func runtime_getProfLabel() *proflabel.Labels {
	return getg().labels
}
