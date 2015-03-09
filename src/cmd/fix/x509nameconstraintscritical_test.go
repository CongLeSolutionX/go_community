// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func init() {
	addTestCases(x509nameconstraintscriticalTests, x509nameconstraintscritical)
}

var x509nameconstraintscriticalTests = []testCase{
	{
		Name: "x509nameconstraintscritical.0",
		In: `package main

import "crypto/x509"

func f() x509.Certificate {
	a := &x509.Certificate{PermittedDNSDomainsCritical: true}
	b := &x509.Certificate{PermittedDNSDomainsCritical: false}
	return &x509.Certificate{PermittedDNSDomainsCritical: true}, nil
}
`,
		Out: `package main

import "crypto/x509"

func f() x509.Certificate {
	a := &x509.Certificate{NameConstraintsCritical: true}
	b := &x509.Certificate{NameConstraintsCritical: false}
	return &x509.Certificate{NameConstraintsCritical: true}, nil
}
`,
	},
}
