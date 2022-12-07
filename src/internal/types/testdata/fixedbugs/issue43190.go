// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The errors below are produced by the parser, but we check
// them here for consistency with the types2 tests.

package p

import ; // ERR missing import path
import "" // ERR invalid import path (empty string)
import
var /* ERR missing import path */ _ int
import .; // ERR missing import path
import 'x' // ERR import path must be a string
var _ int
import /* ERR imports must appear before other declarations */ _ "math"

// Don't repeat previous error for each immediately following import ...
import ()
import (.) // ERR missing import path
import (
	"fmt"
	.
) // ERR missing import path

// ... but remind with error again if we start a new import section after
// other declarations
var _ = fmt.Println
import /* ERR imports must appear before other declarations */ _ "math"
import _ "math"
