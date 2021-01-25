<<<<<<< HEAD   (79f796 [dev.go2go] go/format: parse type parameters)
=======
// UNREVIEWED
>>>>>>> BRANCH (945680 [dev.typeparams] test: fix excluded files lookup so it works)
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Input for TestIssue13566

package b

import "./a"

type A a.A
