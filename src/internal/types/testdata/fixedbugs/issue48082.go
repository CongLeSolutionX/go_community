// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue48082

import "init" /* ERR init must be a func */ /* ERR could not import init */
