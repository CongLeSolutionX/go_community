package x509

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
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

func parseCertificateFast(der []byte) (*Certificate, error) {
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
	cert.PublicKey, err = parsePublicKey(cert.PublicKeyAlgorithm, &publicKeyInfo{
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
				err = processExtensions(cert)
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
