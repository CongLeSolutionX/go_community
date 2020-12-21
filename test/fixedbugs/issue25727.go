// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "net/http"

var s = http.Server{}
<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = s.doneChan                  // ERROR "s.doneChan undefined .cannot refer to unexported field or method doneChan.$|s.doneChan undefined"
var _ = s.DoneChan                  // ERROR "s.DoneChan undefined .type http.Server has no field or method DoneChan.$"
var _ = http.Server{tlsConfig: nil} // ERROR "unknown field 'tlsConfig' in struct literal.+ .but does have TLSConfig.$|unknown field 'tlsConfig'"
var _ = http.Server{DoneChan: nil}  // ERROR "unknown field 'DoneChan' in struct literal of type http.Server$"
=======
var _ = s.doneChan                  // ERROR "s.doneChan undefined .cannot refer to unexported field or method doneChan.$|unexported field or method"
var _ = s.DoneChan                  // ERROR "s.DoneChan undefined .type http.Server has no field or method DoneChan.$|undefined field or method"
var _ = http.Server{tlsConfig: nil} // ERROR "unknown field 'tlsConfig' in struct literal.+ .but does have TLSConfig.$|unknown field .?tlsConfig.? in .?http.Server"
var _ = http.Server{DoneChan: nil}  // ERROR "unknown field 'DoneChan' in struct literal of type http.Server$|unknown field .?DoneChan.? in .?http.Server"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )

type foo struct {
	bar int
}

<<<<<<< HEAD   (060cdb [dev.typeparams] go/types: import object resolution from dev)
var _ = &foo{bAr: 10} // ERROR "unknown field 'bAr' in struct literal.+ .but does have bar.$|unknown field 'bAr'"
=======
var _ = &foo{bAr: 10} // ERROR "unknown field 'bAr' in struct literal.+ .but does have bar.$|unknown field .?bAr.? in .?foo"
>>>>>>> BRANCH (4e8f68 Merge "[dev.regabi] all: merge master into dev.regabi" into )
