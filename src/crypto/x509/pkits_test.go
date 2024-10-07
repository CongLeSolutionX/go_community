package x509

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNISTPKITS(t *testing.T) {
	certDir := "testdata/nist-pkits/certs"
	currentDate, err := time.Parse("January 2 2006", "April 14 2011")
	if err != nil {
		t.Fatal(err)
	}

	var testcases []struct {
		Name           string
		CertPath       []string
		ShouldValidate bool
		Skipped        bool
	}
	b, err := os.ReadFile("testdata/nist-pkits/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &testcases); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Skipped {
				t.Skip()
			}
			if len(tc.CertPath) == 0 {
				t.Skip()
			}
			roots := NewCertPool()
			rootDER, err := os.ReadFile(filepath.Join(certDir, tc.CertPath[0]))
			if err != nil {
				t.Fatal(err)
			}
			root, err := ParseCertificate(rootDER)
			if err != nil {
				t.Fatal(err)
			}
			roots.AddCert(root)
			intermediates := NewCertPool()
			if len(tc.CertPath) > 2 {
				for _, interCert := range tc.CertPath[1 : len(tc.CertPath)-1] {
					interDER, err := os.ReadFile(filepath.Join(certDir, interCert))
					if err != nil {
						t.Fatal(err)
					}
					inter, err := ParseCertificate(interDER)
					if err != nil {
						t.Fatal(err)
					}
					intermediates.AddCert(inter)
				}
			}

			leafDER, err := os.ReadFile(filepath.Join(certDir, tc.CertPath[len(tc.CertPath)-1]))
			if err != nil {
				t.Fatal(err)
			}
			leaf, err := ParseCertificate(leafDER)
			if err != nil {
				t.Fatal(err)
			}

			path, err := leaf.Verify(VerifyOptions{
				Roots:         roots,
				Intermediates: intermediates,
				CurrentTime:   currentDate,
			})
			if err != nil {
				if !tc.ShouldValidate {
					return
				}
				t.Fatalf("Failed to validate: %s", err)
			}
			if !tc.ShouldValidate {
				t.Fatal("Expected path validation to fail")
			}
			if len(path) != len(tc.CertPath) {
				t.Logf("unexpected path length: got %d, want %d", len(path), len(tc.CertPath))
			}
		})
	}
}
