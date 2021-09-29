// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x509_test

import (
	"crypto/tls"
	"crypto/x509"
	"internal/testenv"
	"testing"
	"time"
)

// var runRootTests = flag.Bool("run-darwin-x509-verification-tests", false, "run darwin x509 verifier tests (this requires installing test root certificates which may not be cleaned up)")

// func TestPlatformVerifier(t *testing.T) {
// 	if !*runRootTests && testenv.Builder() == "" {
// 		t.Skip("skipping darwin x509 verifier tests")
// 	}
// 	// if u, err := user.Current(); err != nil && u.Uid != "0" {
// 	// 	t.Skip("skipping darwin x509 verifier tests, must be root")
// 	// }
// 	u, _ := user.Current()
// 	fmt.Println("DEBUG", u.Uid)

// 	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		t.Fatalf("failed to generate test key: %s", err)
// 	}

// 	name := pkix.Name{
// 		CommonName: "golang darwin testing root",
// 	}
// 	now := time.Now()
// 	tmpl := &Certificate{
// 		Subject:               name,
// 		SerialNumber:          big.NewInt(5),
// 		NotBefore:             now.Add(-time.Hour),
// 		NotAfter:              now.Add(time.Hour),
// 		IsCA:                  true,
// 		BasicConstraintsValid: true,
// 		KeyUsage:              KeyUsageCertSign,
// 	}
// 	rootDER, err := CreateCertificate(rand.Reader, tmpl, tmpl, k.Public(), k)
// 	if err != nil {
// 		t.Fatalf("failed to create test root: %s", err)
// 	}
// 	root, err := ParseCertificate(rootDER)
// 	if err != nil {
// 		t.Fatalf("failed to parse test root: %s", err)
// 	}
// 	tmpDir := t.TempDir()
// 	rootFile := filepath.Join(tmpDir, "golang-test-root.crt")
// 	rootPEM := pem.EncodeToMemory(&pem.Block{Bytes: rootDER, Type: "CERTIFICATE"})
// 	// fmt.Println(string(rootPEM))
// 	if err = os.WriteFile(rootFile, rootPEM, 0600); err != nil {
// 		t.Fatalf("failed to write test root: %s", err)
// 	}

// 	out, err := exec.Command("security", "add-trusted-cert", "-k", "/Library/Keychains/System.keychain", "-p", "ssl", rootFile).CombinedOutput()
// 	if err != nil {
// 		t.Fatalf("security add-trusted-cert failed: %s\n[output] %s", err, out)
// 	}

// 	defer func() {
// 		out, err := exec.Command("security", "remove-trusted-cert", rootFile).CombinedOutput()
// 		if err != nil {
// 			t.Fatalf("security remove-trusted-cert failed: %s\n[output] %s", err, out)
// 		}
// 	}()

// 	leafTmpl := &Certificate{
// 		Subject:      pkix.Name{Organization: []string{"golang"}},
// 		DNSNames:     []string{"localhost"},
// 		SerialNumber: big.NewInt(1),
// 		KeyUsage:     KeyUsageKeyEncipherment | KeyUsageDigitalSignature,
// 		ExtKeyUsage:  []ExtKeyUsage{ExtKeyUsageServerAuth},
// 		NotBefore:    now.Add(-time.Hour),
// 		NotAfter:     now.Add(time.Hour),
// 	}
// 	leafDER, err := CreateCertificate(rand.Reader, leafTmpl, root, k.Public(), k)
// 	if err != nil {
// 		t.Fatalf("failed to create leaf certificate: %s", err)
// 	}
// 	leaf, err := ParseCertificate(leafDER)
// 	if err != nil {
// 		t.Fatalf("failed to parse leaf certificate: %s", err)
// 	}
// 	if _, err := leaf.Verify(VerifyOptions{}); err != nil {
// 		t.Fatalf("failed to verify leaf: %s", err)
// 	}
// }

func TestPlatformVerifier(t *testing.T) {
	if !testenv.HasExternalNetwork() {
		t.Skip()
	}

	getChain := func(host string) []*x509.Certificate {
		t.Helper()
		c, err := tls.Dial("tcp", host+":443", &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			t.Fatalf("tls connection failed: %s", err)
		}
		return c.ConnectionState().PeerCertificates
	}

	tests := []struct {
		name        string
		host        string
		verifyName  string
		verifyTime  time.Time
		expectedErr string
	}{
		{
			// whatever google.com serves should. hopefully, be trusted
			name: "valid chain",
			host: "google.com",
		},
		{
			name:        "expired leaf",
			host:        "expired.badssl.com",
			expectedErr: "x509: “*.badssl.com” certificate is expired",
		},
		{
			name:        "wrong host for leaf",
			host:        "wrong.host.badssl.com",
			verifyName:  "wrong.host.badssl.com",
			expectedErr: "x509: “*.badssl.com” certificate name does not match input",
		},
		{
			name:        "self-signed leaf",
			host:        "self-signed.badssl.com",
			expectedErr: "x509: “*.badssl.com” certificate is not trusted",
		},
		{
			name:        "untrusted root",
			host:        "untrusted-root.badssl.com",
			expectedErr: "x509: “BadSSL Untrusted Root Certificate Authority” certificate is not trusted",
		},
		{
			name:        "revoked leaf",
			host:        "revoked.badssl.com",
			expectedErr: "x509: “revoked.badssl.com” certificate is revoked",
		},
		{
			name:        "leaf missing SCTs",
			host:        "no-sct.badssl.com",
			expectedErr: "x509: “no-sct.badssl.com” certificate is not standards compliant",
		},
		{
			name:        "expired leaf (custom time)",
			host:        "google.com",
			verifyTime:  time.Time{}.Add(time.Hour),
			expectedErr: "x509: “*.google.com” certificate is expired",
		},
		{
			name:       "valid chain (custom time)",
			host:       "google.com",
			verifyTime: time.Now(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chain := getChain(tc.host)
			var opts x509.VerifyOptions
			if len(chain) > 1 {
				opts.Intermediates = x509.NewCertPool()
				for _, c := range chain[1:] {
					opts.Intermediates.AddCert(c)
				}
			}
			if tc.verifyName != "" {
				opts.DNSName = tc.verifyName
			}
			if !tc.verifyTime.IsZero() {
				opts.CurrentTime = tc.verifyTime
			}

			_, err := chain[0].Verify(opts)
			if err != nil && tc.expectedErr == "" {
				t.Errorf("unexpected verification error: %s", err)
			} else if err != nil && err.Error() != tc.expectedErr {
				t.Errorf("unexpected verification error: got %q, want %q", err.Error(), tc.expectedErr)
			} else if err == nil && tc.expectedErr != "" {
				t.Errorf("unexpected verification success: want %q", tc.expectedErr)
			}
		})
	}
}
