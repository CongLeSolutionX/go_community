// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/proflabel"
)

func runtime_setProfLabel(labels *proflabel.Labels) {
	getg().labels = labels
}

func runtime_getProfLabel() *proflabel.Labels {
	return getg().labels
}
