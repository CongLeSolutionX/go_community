// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run root_darwin_arm_gen.go -output root_darwin_armx.go

package x509

import (
	"bytes"
	"encoding/pem"
	"os/exec"
	"sync"
)

func (c *Certificate) systemVerify(opts *VerifyOptions) (chains [][]*Certificate, err error) {
	return nil, nil
}

func execSecurityRoots() (*CertPool, error) {
	cmd := exec.Command("/usr/bin/security", "find-certificate", "-a", "-p", "/System/Library/Keychains/SystemRootCertificates.keychain")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	roots := NewCertPool()
	var mtx sync.Mutex
	add := func(cert *Certificate) {
		mtx.Lock()
		defer mtx.Unlock()
		roots.AddCert(cert)
	}
	blockCh := make(chan *pem.Block)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for block := range blockCh {
				cmd := exec.Command("/usr/bin/security", "verify-cert", "-c", "/dev/stdin", "-l", "-q")
				cmd.Stdin = bytes.NewReader(pem.EncodeToMemory(block))
				if err := cmd.Run(); err == nil {
					// Non-zero exit means untrusted
					cert, err := ParseCertificate(block.Bytes)
					if err != nil {
						continue
					}

					add(cert)
				}
			}
		}()
	}
	for len(data) > 0 {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		blockCh <- block
	}
	close(blockCh)
	wg.Wait()
	return roots, nil
}
