package x509

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

var nistTestPolicies = map[string]OID{
	"anyPolicy":          anyPolicyOID,
	"NIST-test-policy-1": mustNewOIDFromInts([]uint64{2, 16, 840, 1, 101, 3, 2, 1, 48, 1}),
	"NIST-test-policy-2": mustNewOIDFromInts([]uint64{2, 16, 840, 1, 101, 3, 2, 1, 48, 2}),
	"NIST-test-policy-3": mustNewOIDFromInts([]uint64{2, 16, 840, 1, 101, 3, 2, 1, 48, 3}),
	"NIST-test-policy-6": mustNewOIDFromInts([]uint64{2, 16, 840, 1, 101, 3, 2, 1, 48, 6}),
}

func TestNISTPKITS(t *testing.T) {
	certDir := "testdata/nist-pkits/certs"
	currentDate, err := time.Parse("January 2 2006", "April 14 2011")
	if err != nil {
		t.Fatal(err)
	}

	var testcases []struct {
		Name                        string
		CertPath                    []string
		InitialPolicySet            []string
		InitialPolicyMappingInhibit bool
		InitialExplicitPolicy       bool
		InitialAnyPolicyInhibit     bool
		ShouldValidate              bool
		Skipped                     bool
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

			var initialPolicies []OID
			for _, pstr := range tc.InitialPolicySet {
				policy, ok := nistTestPolicies[pstr]
				if !ok {
					t.Fatalf("unknown test policy: %s", pstr)
				}
				initialPolicies = append(initialPolicies, policy)
			}

			path, err := leaf.Verify(VerifyOptions{
				Roots:                       roots,
				Intermediates:               intermediates,
				CurrentTime:                 currentDate,
				CertificatePolicies:         initialPolicies,
				InitialPolicyMappingInhibit: tc.InitialPolicyMappingInhibit,
				InitialExplicitPolicy:       tc.InitialExplicitPolicy,
				InitialInhibitAnyPolicy:     tc.InitialAnyPolicyInhibit,
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
				t.Errorf("unexpected path length: got %d, want %d", len(path), len(tc.CertPath))
			}
		})
	}
}

func TestNISTPKITSPolicy(t *testing.T) {
	certDir := "testdata/nist-pkits/certs"
	currentDate, err := time.Parse("January 2 2006", "April 14 2011")
	if err != nil {
		t.Fatal(err)
	}

	var testcases []struct {
		Name                        string
		CertPath                    []string
		InitialPolicySet            []string
		InitialPolicyMappingInhibit bool
		InitialExplicitPolicy       bool
		InitialAnyPolicyInhibit     bool
		ShouldValidate              bool
		Skipped                     bool
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
			var chain []*Certificate
			for _, c := range tc.CertPath {
				certDER, err := os.ReadFile(filepath.Join(certDir, c))
				if err != nil {
					t.Fatal(err)
				}
				cert, err := ParseCertificate(certDER)
				if err != nil {
					t.Fatal(err)
				}
				chain = append(chain, cert)
			}
			slices.Reverse(chain)

			var initialPolicies []OID
			for _, pstr := range tc.InitialPolicySet {
				policy, ok := nistTestPolicies[pstr]
				if !ok {
					t.Fatalf("unknown test policy: %s", pstr)
				}
				initialPolicies = append(initialPolicies, policy)
			}

			valid := policiesValid(chain, VerifyOptions{
				CurrentTime:                 currentDate,
				CertificatePolicies:         initialPolicies,
				InitialPolicyMappingInhibit: tc.InitialPolicyMappingInhibit,
				InitialExplicitPolicy:       tc.InitialExplicitPolicy,
				InitialInhibitAnyPolicy:     tc.InitialAnyPolicyInhibit,
			})
			if !valid {
				if !tc.ShouldValidate {
					return
				}
				t.Fatalf("Failed to validate: %s", err)
			}
			if !tc.ShouldValidate {
				t.Fatal("Expected path validation to fail")
			}
		})
	}
}
