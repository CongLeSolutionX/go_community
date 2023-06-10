// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputil

import (
	// Import the package to for the side effect of configuring the fake network
	// stack on platforms that need it.
	_ "internal/testenv"
)
