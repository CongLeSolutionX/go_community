// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509

import (
	"crypto/x509/internal/macOS"
	"fmt"
	"os"
	"strings"
)

var debugDarwinRoots = strings.Contains(os.Getenv("GODEBUG"), "x509roots=1")

func (c *Certificate) systemVerify(opts *VerifyOptions) (chains [][]*Certificate, err error) {
	return nil, nil
}

func loadSystemRoots() (*CertPool, error) {
	certs, err := exportCertificates(macOS.SecTrustSettingsDomainSystem)
	if err != nil {
		return nil, err
	}
	println("system roots count:", len(certs))

	return NewCertPool(), nil
}

func exportCertificates(domain macOS.SecTrustSettingsDomain) ([]*Certificate, error) {
	certs, err := macOS.SecTrustSettingsCopyCertificates(domain)
	if err != nil {
		return nil, err
	}
	defer macOS.CFRelease(certs)

	var out []*Certificate
	for i := 0; i < macOS.CFArrayGetCount(certs); i++ {
		cert := macOS.CFArrayGetValueAtIndex(certs, i)
		data, err := macOS.SecItemExport(cert)
		if err != nil {
			if debugDarwinRoots {
				fmt.Fprintf(os.Stderr, "crypto/x509: domain %d, certificate #%d: %v\n", domain, i, err)
			}
			continue
		}
		der := macOS.CFDataCopyGoBytes(data)
		macOS.CFRelease(data)
		c, err := ParseCertificate(der)
		if err != nil {
			if debugDarwinRoots {
				fmt.Fprintf(os.Stderr, "crypto/x509: domain %d, certificate #%d: %v\n", domain, i, err)
			}
			continue
		}
		out = append(out, c)
	}
	return out, nil
}
