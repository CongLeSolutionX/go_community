package x509

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/crypto/cryptobyte"
	cbasn1 "golang.org/x/crypto/cryptobyte/asn1"
)

// parseName parses a DER encoded Name as defined in RFC 5280. We may
// want to export this function in the future for use in crypto/tls.
func parseName(raw cryptobyte.String) (*pkix.RDNSequence, error) {
	// input := cryptobyte.String(raw)
	var inner cryptobyte.String
	if !raw.ReadASN1(&inner, cbasn1.SEQUENCE) {
		return nil, errors.New("invalid RDNSequence")
	}

	var rdnSeq pkix.RDNSequence
	for !inner.Empty() {
		var rdnSet pkix.RelativeDistinguishedNameSET
		var set cryptobyte.String
		if !inner.ReadASN1(&set, cbasn1.SET) {
			return nil, errors.New("invalid RDNSequence")
		}
		for !set.Empty() {
			var atav cryptobyte.String
			if !set.ReadASN1(&atav, cbasn1.SEQUENCE) {
				return nil, errors.New("invalid RDNSequence")
			}
			var attr pkix.AttributeTypeAndValue
			if !atav.ReadASN1ObjectIdentifier(&attr.Type) {
				return nil, errors.New("invalid RDNSequence")
			}
			var rawValue cryptobyte.String
			var valueTag cbasn1.Tag
			if !atav.ReadAnyASN1(&rawValue, &valueTag) {
				return nil, errors.New("invalid RDNSequence")
			}
			var err error
			switch valueTag {
			case cbasn1.T61String:
				attr.Value = string(rawValue)
			case cbasn1.PrintableString:
				attr.Value, err = asn1.ParsePrintableString(rawValue)
				if err != nil {
					return nil, err
				}
			// case cbasn1.Tag(28): // UniversalString
			case cbasn1.UTF8String:
				attr.Value, err = asn1.ParseUTF8String(rawValue)
				if err != nil {
					return nil, err
				}
			case cbasn1.Tag(asn1.TagBMPString):
				attr.Value, err = asn1.ParseBMPString(rawValue)
				if err != nil {
					return nil, err
				}
			case cbasn1.IA5String:
				attr.Value, err = asn1.ParseIA5String(rawValue)
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unsupported AttributeTypeAndValue value type: %v", valueTag)
			}
			rdnSet = append(rdnSet, attr)
		}

		rdnSeq = append(rdnSeq, rdnSet)
	}

	return &rdnSeq, nil
}

func parseAI(der cryptobyte.String) (pkix.AlgorithmIdentifier, error) {
	ai := pkix.AlgorithmIdentifier{}
	if !der.ReadASN1ObjectIdentifier(&ai.Algorithm) {
		return ai, errors.New("malformed OID")
	}
	if der.Empty() {
		return ai, nil
	}
	var params cryptobyte.String
	var tag cbasn1.Tag
	if !der.ReadAnyASN1Element(&params, &tag) {
		return ai, errors.New("malformed parameters")
	}
	ai.Parameters.Tag = int(tag)
	ai.Parameters.FullBytes = params
	return ai, nil
}

func parseValidity(der cryptobyte.String) (time.Time, time.Time, error) {
	extract := func() (time.Time, error) {
		var t time.Time
		switch {
		case der.PeekASN1Tag(cbasn1.UTCTime):
			// This should be added to x/crypto/cryptobyte
			var utc cryptobyte.String
			if !der.ReadASN1(&utc, cbasn1.UTCTime) {
				return t, errors.New("malformed UTCTime")
			}
			s := string(utc)

			formatStr := "0601021504Z0700"
			var err error
			t, err = time.Parse(formatStr, s)
			if err != nil {
				formatStr = "060102150405Z0700"
				t, err = time.Parse(formatStr, s)
			}
			if err != nil {
				return t, err
			}

			if serialized := t.Format(formatStr); serialized != s {
				err = fmt.Errorf("asn1: time did not serialize back to the original value and may be invalid: given %q, but serialized as %q", s, serialized)
				return t, err
			}

			if t.Year() >= 2050 {
				// UTCTime only encodes times prior to 2050. See https://tools.ietf.org/html/rfc5280#section-4.1.2.5.1
				t = t.AddDate(-100, 0, 0)
			}
		case der.PeekASN1Tag(cbasn1.GeneralizedTime):
			if !der.ReadASN1GeneralizedTime(&t) {
				return t, errors.New("malformed GeneralizedTime")
			}
		default:
			return t, errors.New("unsupported time format")
		}
		return t, nil
	}

	notBefore, err := extract()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	notAfter, err := extract()
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return notBefore, notAfter, nil
}

func parseExtension(der cryptobyte.String) (pkix.Extension, error) {
	var ext pkix.Extension
	if !der.ReadASN1ObjectIdentifier(&ext.Id) {
		return ext, errors.New("malformed OID") // extnID
	}
	if der.PeekASN1Tag(cbasn1.BOOLEAN) {
		if !der.ReadASN1Boolean(&ext.Critical) {
			return ext, errors.New("malformed critical boolean")
		}
	}
	var val cryptobyte.String
	if !der.ReadASN1(&val, cbasn1.OCTET_STRING) {
		return ext, errors.New("malformed value") // extnValue
	}
	ext.Value = val
	return ext, nil
}

func parsePublicKeyFast(algo PublicKeyAlgorithm, keyData *publicKeyInfo) (interface{}, error) {
	der := cryptobyte.String(keyData.PublicKey.RightAlign())
	switch algo {
	case RSA:
		// RSA public keys must have a NULL in the parameters.
		// See RFC 3279, Section 2.3.1.
		if !bytes.Equal(keyData.Algorithm.Parameters.FullBytes, asn1.NullBytes) {
			return nil, errors.New("x509: RSA key missing NULL parameters")
		}

		p := &pkcs1PublicKey{N: big.NewInt(0)}
		if !der.ReadASN1(&der, cbasn1.SEQUENCE) {
			return nil, errors.New("x509: invalid RSA public key")
		}
		if !der.ReadASN1Integer(p.N) {
			return nil, errors.New("x509: invalid RSA modulus")
		}
		if !der.ReadASN1Integer(&p.E) {
			return nil, errors.New("x509: invalid RSA public exponent")
		}

		if p.N.Sign() <= 0 {
			return nil, errors.New("x509: RSA modulus is not a positive number")
		}
		if p.E <= 0 {
			return nil, errors.New("x509: RSA public exponent is not a positive number")
		}

		pub := &rsa.PublicKey{
			E: p.E,
			N: p.N,
		}
		return pub, nil
	case ECDSA:
		paramsData := keyData.Algorithm.Parameters.FullBytes
		namedCurveOID := new(asn1.ObjectIdentifier)
		paramsDer := cryptobyte.String(paramsData)
		if !paramsDer.ReadASN1ObjectIdentifier(namedCurveOID) {
			return nil, errors.New("x509: invalid ECDSA parameters")
		}
		namedCurve := namedCurveFromOID(*namedCurveOID)
		if namedCurve == nil {
			return nil, errors.New("x509: unsupported elliptic curve")
		}
		x, y := elliptic.Unmarshal(namedCurve, der)
		if x == nil {
			return nil, errors.New("x509: failed to unmarshal elliptic curve point")
		}
		pub := &ecdsa.PublicKey{
			Curve: namedCurve,
			X:     x,
			Y:     y,
		}
		return pub, nil
	case Ed25519:
		// RFC 8410, Section 3
		// > For all of the OIDs, the parameters MUST be absent.
		if len(keyData.Algorithm.Parameters.FullBytes) != 0 {
			return nil, errors.New("x509: Ed25519 key encoded with illegal parameters")
		}
		if len(der) != ed25519.PublicKeySize {
			return nil, errors.New("x509: wrong Ed25519 public key size")
		}
		pub := make([]byte, ed25519.PublicKeySize)
		copy(pub, der)
		return ed25519.PublicKey(pub), nil
	default:
		return nil, nil
	}
}

func parseKeyUsageExtensionFast(der cryptobyte.String) (KeyUsage, error) {
	var usageBits asn1.BitString
	if !der.ReadASN1BitString(&usageBits) {
		return 0, errors.New("x509: invalid key usage")
	}

	var usage int
	for i := 0; i < 9; i++ {
		if usageBits.At(i) != 0 {
			usage |= 1 << uint(i)
		}
	}
	return KeyUsage(usage), nil
}

func parseBasicConstraintsExtensionFast(der cryptobyte.String) (bool, int, error) {
	// var constraints basicConstraints
	// if rest, err := asn1.Unmarshal(ext, &constraints); err != nil {
	// 	return false, 0, err
	// } else if len(rest) != 0 {
	// 	return false, 0, errors.New("x509: trailing data after X.509 BasicConstraints")
	// }

	var isCA bool
	if !der.ReadASN1(&der, cbasn1.SEQUENCE) {
		return false, 0, errors.New("x509: invalid basic constraints")
	}
	if der.PeekASN1Tag(cbasn1.BOOLEAN) {
		if !der.ReadASN1Boolean(&isCA) {
			return false, 0, errors.New("x509: invalid basic constraints")
		}
	}
	var maxPathLen int
	if !der.ReadOptionalASN1Integer(&maxPathLen, cbasn1.INTEGER, -1) {
		return false, 0, errors.New("x509: invalid basic constraints")
	}

	// TODO: map out.MaxPathLen to 0 if it has the -1 default value? (Issue 19285)
	return isCA, maxPathLen, nil
}

func forEachSANFast(der cryptobyte.String, callback func(tag int, data []byte) error) error {
	if !der.ReadASN1(&der, cbasn1.SEQUENCE) {
		return errors.New("x509: invalid subject alternative names")
	}
	for !der.Empty() {
		var san cryptobyte.String
		var tag cbasn1.Tag
		if !der.ReadAnyASN1(&san, &tag) {
			return errors.New("x509: invalid subject alternative name")
		}
		if err := callback(int(tag^0x80), san); err != nil {
			return err
		}
	}

	return nil
}

func parseSANExtensionFast(der cryptobyte.String) (dnsNames, emailAddresses []string, ipAddresses []net.IP, uris []*url.URL, err error) {
	err = forEachSANFast(der, func(tag int, data []byte) error {
		switch tag {
		case nameTypeEmail:
			email := string(data)
			if err := isIA5String(email); err != nil {
				return errors.New("x509: SAN rfc822Name is malformed")
			}
			emailAddresses = append(emailAddresses, email)
		case nameTypeDNS:
			name := string(data)
			if err := isIA5String(name); err != nil {
				return errors.New("x509: SAN dNSName is malformed")
			}
			dnsNames = append(dnsNames, string(name))
		case nameTypeURI:
			uriStr := string(data)
			if err := isIA5String(uriStr); err != nil {
				return errors.New("x509: SAN uniformResourceIdentifier is malformed")
			}
			uri, err := url.Parse(uriStr)
			if err != nil {
				return fmt.Errorf("x509: cannot parse URI %q: %s", uriStr, err)
			}
			if len(uri.Host) > 0 {
				if _, ok := domainToReverseLabels(uri.Host); !ok {
					return fmt.Errorf("x509: cannot parse URI %q: invalid domain", uriStr)
				}
			}
			uris = append(uris, uri)
		case nameTypeIP:
			switch len(data) {
			case net.IPv4len, net.IPv6len:
				ipAddresses = append(ipAddresses, data)
			default:
				return errors.New("x509: cannot parse IP address of length " + strconv.Itoa(len(data)))
			}
		}

		return nil
	})

	return
}

func parseExtKeyUsageExtensionFast(der cryptobyte.String) ([]ExtKeyUsage, []asn1.ObjectIdentifier, error) {
	var extKeyUsages []ExtKeyUsage
	var unknownUsages []asn1.ObjectIdentifier
	if !der.ReadASN1(&der, cbasn1.SEQUENCE) {
		return nil, nil, errors.New("x509: invalid extended key usages")
	}
	for !der.Empty() {
		var eku asn1.ObjectIdentifier
		if !der.ReadASN1ObjectIdentifier(&eku) {
			return nil, nil, errors.New("x509: invalid extended key usages")
		}
		if extKeyUsage, ok := extKeyUsageFromOID(eku); ok {
			extKeyUsages = append(extKeyUsages, extKeyUsage)
		} else {
			unknownUsages = append(unknownUsages, eku)
		}
	}
	return extKeyUsages, unknownUsages, nil
}

func parseCertificatePoliciesExtensionFast(der cryptobyte.String) ([]asn1.ObjectIdentifier, error) {
	var oids []asn1.ObjectIdentifier
	if !der.ReadASN1(&der, cbasn1.SEQUENCE) {
		return nil, errors.New("x509: invalid certificate policies")
	}
	for !der.Empty() {
		var cp cryptobyte.String
		if !der.ReadASN1(&cp, cbasn1.SEQUENCE) {
			return nil, errors.New("x509: invalid certificate policies")
		}
		var oid asn1.ObjectIdentifier
		if !cp.ReadASN1ObjectIdentifier(&oid) {
			return nil, errors.New("x509: invalid certificate policies")
		}
		oids = append(oids, oid)
	}

	return oids, nil
}

func processExtensionsFast(out *Certificate) error {
	var err error
	for _, e := range out.Extensions {
		unhandled := false

		if len(e.Id) == 4 && e.Id[0] == 2 && e.Id[1] == 5 && e.Id[2] == 29 {
			switch e.Id[3] {
			case 15:
				out.KeyUsage, err = parseKeyUsageExtensionFast(e.Value)
				if err != nil {
					return err
				}
			case 19:
				out.IsCA, out.MaxPathLen, err = parseBasicConstraintsExtensionFast(e.Value)
				if err != nil {
					return err
				}
				out.BasicConstraintsValid = true
				out.MaxPathLenZero = out.MaxPathLen == 0
			case 17:
				out.DNSNames, out.EmailAddresses, out.IPAddresses, out.URIs, err = parseSANExtensionFast(e.Value)
				if err != nil {
					return err
				}

				if len(out.DNSNames) == 0 && len(out.EmailAddresses) == 0 && len(out.IPAddresses) == 0 && len(out.URIs) == 0 {
					// If we didn't parse anything then we do the critical check, below.
					unhandled = true
				}

			case 30:
				unhandled, err = parseNameConstraintsExtension(out, e) // already cryptobyte
				if err != nil {
					return err
				}

			case 31:
				// RFC 5280, 4.2.1.13

				// CRLDistributionPoints ::= SEQUENCE SIZE (1..MAX) OF DistributionPoint
				//
				// DistributionPoint ::= SEQUENCE {
				//     distributionPoint       [0]     DistributionPointName OPTIONAL,
				//     reasons                 [1]     ReasonFlags OPTIONAL,
				//     cRLIssuer               [2]     GeneralNames OPTIONAL }
				//
				// DistributionPointName ::= CHOICE {
				//     fullName                [0]     GeneralNames,
				//     nameRelativeToCRLIssuer [1]     RelativeDistinguishedName }
				val := cryptobyte.String(e.Value)
				if !val.ReadASN1(&val, cbasn1.SEQUENCE) {
					return errors.New("x509: invalid CRL distribution points")
				}
				for !val.Empty() {
					var dpDER cryptobyte.String
					if !val.ReadASN1(&dpDER, cbasn1.SEQUENCE) {
						return errors.New("x509: invalid CRL distribution point")
					}
					var dpNameDER cryptobyte.String
					var dpNamePresent bool
					if !dpDER.ReadOptionalASN1(&dpNameDER, &dpNamePresent, cbasn1.Tag(0).Constructed().ContextSpecific()) {
						return errors.New("x509: invalid CRL distribution point")
					}
					if !dpNamePresent {
						continue
					}
					if !dpNameDER.ReadASN1(&dpNameDER, cbasn1.Tag(0).Constructed().ContextSpecific()) {
						return errors.New("x509: invalid CRL distribution point")
					}
					// if !dpNameDER.PeekASN1Tag(cbasn1.Tag(6).ContextSpecific()) {
					// 	continue
					// }
					for !dpNameDER.Empty() {
						if !dpNameDER.PeekASN1Tag(cbasn1.Tag(6).ContextSpecific()) {
							break
						}
						var uri cryptobyte.String
						if !dpNameDER.ReadASN1(&uri, cbasn1.Tag(6).ContextSpecific()) {
							return errors.New("x509: invalid CRL distribution point")
						}
						out.CRLDistributionPoints = append(out.CRLDistributionPoints, string(uri))
					}
				}

			case 35:
				// RFC 5280, 4.2.1.1
				val := cryptobyte.String(e.Value)
				var akid cryptobyte.String
				if !val.ReadASN1(&akid, cbasn1.SEQUENCE) {
					return errors.New("x509: invalid authority key identifier")
				}
				if !akid.ReadASN1(&akid, cbasn1.Tag(0).ContextSpecific()) {
					return errors.New("x509: invalid authority key identifier")
				}
				out.AuthorityKeyId = akid
			case 37:
				out.ExtKeyUsage, out.UnknownExtKeyUsage, err = parseExtKeyUsageExtensionFast(e.Value)
				if err != nil {
					return err
				}
			case 14:
				// RFC 5280, 4.2.1.2
				val := cryptobyte.String(e.Value)
				var skid cryptobyte.String
				if !val.ReadASN1(&skid, cbasn1.OCTET_STRING) {
					return errors.New("x509: invalid subject key identifier")
				}
				out.SubjectKeyId = skid
			case 32:
				out.PolicyIdentifiers, err = parseCertificatePoliciesExtensionFast(e.Value)
				if err != nil {
					return err
				}
			default:
				// Unknown extensions are recorded if critical.
				unhandled = true
			}
		} else if e.Id.Equal(oidExtensionAuthorityInfoAccess) {
			// RFC 5280 4.2.2.1: Authority Information Access
			val := cryptobyte.String(e.Value)
			if !val.ReadASN1(&val, cbasn1.SEQUENCE) {
				return errors.New("x509: invalid authority info access")
			}
			for !val.Empty() {
				var aiaDER cryptobyte.String
				if !val.ReadASN1(&aiaDER, cbasn1.SEQUENCE) {
					return errors.New("x509: invalid authority info access")
				}
				var method asn1.ObjectIdentifier
				if !aiaDER.ReadASN1ObjectIdentifier(&method) {
					return errors.New("x509: invalid authority info access")
				}
				if !aiaDER.PeekASN1Tag(cbasn1.Tag(6).ContextSpecific()) {
					continue
				}
				if !aiaDER.ReadASN1(&aiaDER, cbasn1.Tag(6).ContextSpecific()) {
					return errors.New("x509: invalid authority info access")
				}
				switch {
				case method.Equal(oidAuthorityInfoAccessOcsp):
					out.OCSPServer = append(out.OCSPServer, string(aiaDER))
				case method.Equal(oidAuthorityInfoAccessIssuers):
					out.IssuingCertificateURL = append(out.IssuingCertificateURL, string(aiaDER))
				}
			}
		} else {
			// Unknown extensions are recorded if critical.
			unhandled = true
		}

		if e.Critical && unhandled {
			out.UnhandledCriticalExtensions = append(out.UnhandledCriticalExtensions, e.Id)
		}
	}

	return nil
}

func ParseCertificateFast(der []byte) (*Certificate, error) {
	cert := &Certificate{}
	cert.Raw = make([]byte, len(der))
	copy(cert.Raw, der)

	input := cryptobyte.String(der)
	var inner cryptobyte.String
	if !input.ReadASN1(&inner, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed certificate")
	}
	var tbs cryptobyte.String
	if !inner.ReadASN1Element(&tbs, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed tbs certificate")
	}
	cert.RawTBSCertificate = make([]byte, len(tbs))
	copy(cert.RawTBSCertificate, tbs)
	if !tbs.ReadASN1(&tbs, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed tbs certificate")
	}

	if !tbs.ReadOptionalASN1Integer(&cert.Version, cbasn1.Tag(0).Constructed().ContextSpecific(), 0) {
		return nil, errors.New("malformed version") // bad version
	}
	cert.Version++ // this is dumb, but matches current behavior
	if cert.Version > 3 {
		return nil, errors.New("invalid version") // bad version
	}
	serial := big.NewInt(0)
	if !tbs.ReadASN1Integer(serial) {
		return nil, errors.New("malformed serial number") // bad version
	}
	cert.SerialNumber = serial
	var sigAISeq cryptobyte.String
	if !tbs.ReadASN1(&sigAISeq, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed signature algorithm identifier")
	}
	sigAI, err := parseAI(sigAISeq)
	if err != nil {
		return nil, err
	}
	cert.SignatureAlgorithm = getSignatureAlgorithmFromAI(sigAI)
	var issuerSeq cryptobyte.String
	if !tbs.ReadASN1Element(&issuerSeq, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed issuer")
	}
	cert.RawIssuer = make([]byte, len(issuerSeq))
	copy(cert.RawIssuer, issuerSeq)
	issuerRDNs, err := parseName(issuerSeq)
	if err != nil {
		return nil, err
	}
	cert.Issuer.FillFromRDNSequence(issuerRDNs)
	var validity cryptobyte.String
	if !tbs.ReadASN1(&validity, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed certificate") // bad validity
	}
	cert.NotBefore, cert.NotAfter, err = parseValidity(validity)
	if err != nil {
		return nil, err
	}
	var subjectSeq cryptobyte.String
	if !tbs.ReadASN1Element(&subjectSeq, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed issuer")
	}
	cert.RawSubject = make([]byte, len(subjectSeq))
	copy(cert.RawSubject, subjectSeq)
	subjectRDNs, err := parseName(subjectSeq)
	if err != nil {
		return nil, err
	}
	cert.Subject.FillFromRDNSequence(subjectRDNs)
	var spki cryptobyte.String
	if !tbs.ReadASN1Element(&spki, cbasn1.SEQUENCE) {
		return nil, err // bad spki
	}
	cert.RawSubjectPublicKeyInfo = make([]byte, len(spki))
	copy(cert.RawSubjectPublicKeyInfo, spki)
	if !spki.ReadASN1(&spki, cbasn1.SEQUENCE) {
		return nil, err // bad spki
	}
	var pkAISeq cryptobyte.String
	if !spki.ReadASN1(&pkAISeq, cbasn1.SEQUENCE) {
		return nil, errors.New("malformed public key algorithm identifier")
	}
	pkAI, err := parseAI(pkAISeq)
	if err != nil {
		return nil, err
	}
	cert.PublicKeyAlgorithm = getPublicKeyAlgorithmFromOID(pkAI.Algorithm)
	var spk asn1.BitString
	if !spki.ReadASN1BitString(&spk) {
		return nil, errors.New("malformed certificate") // bad subjectPublicKey
	}
	cert.PublicKey, err = parsePublicKeyFast(cert.PublicKeyAlgorithm, &publicKeyInfo{
		Algorithm: pkAI,
		PublicKey: spk,
	})
	if err != nil {
		return nil, err
	}
	if cert.Version > 1 {
		// issuerUniqueID
		if !tbs.SkipOptionalASN1(cbasn1.Tag(1).Constructed().ContextSpecific()) {
			return nil, errors.New("malformed certificate")
		}
		// subjectUniqueID
		if !tbs.SkipOptionalASN1(cbasn1.Tag(2).Constructed().ContextSpecific()) {
			return nil, errors.New("malformed certificate")
		}
		if cert.Version == 3 {
			var extensions cryptobyte.String
			var present bool
			if !tbs.ReadOptionalASN1(&extensions, &present, cbasn1.Tag(3).Constructed().ContextSpecific()) {
				return nil, errors.New("malformed certificate extensions") // bad extensions
			}
			if present {
				if !extensions.ReadASN1(&extensions, cbasn1.SEQUENCE) {
					return nil, errors.New("malformed certificate extensions") // bad extensions
				}
				for !extensions.Empty() {
					var extension cryptobyte.String
					if !extensions.ReadASN1(&extension, cbasn1.SEQUENCE) {
						return nil, errors.New("malformed certificate extension")
					}
					ext, err := parseExtension(extension)
					if err != nil {
						return nil, err // bad extensions
					}
					cert.Extensions = append(cert.Extensions, ext)
				}
				// parse the extensions
				err = processExtensionsFast(cert)
				if err != nil {
					return nil, errors.New("malformed certificate") // bad extensions
				}
			}

		}
	}
	if !inner.SkipASN1(cbasn1.SEQUENCE) {
		return nil, errors.New("malformed certificate") // bad ai
	}
	var signature asn1.BitString
	if !inner.ReadASN1BitString(&signature) {
		return nil, errors.New("malformed certificate") // bad signature
	}
	cert.Signature = signature.RightAlign()

	return cert, nil
}
