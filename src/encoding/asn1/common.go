// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asn1

import (
	"reflect"
	"strconv"
	"strings"
)

// ASN.1 objects have metadata preceding them:
//   the tag: the type of the object
//   a flag denoting if this object is compound or not
//   the class type: the namespace of the tag
//   the length of the object, in bytes

// Here are some standard tags and classes

// ASN.1 tags represent the type of the following object.
const (
	TagBoolean         = 1
	TagInteger         = 2
	TagBitString       = 3
	TagOctetString     = 4
	TagOID             = 6
	TagEnum            = 10
	TagUTF8String      = 12
	TagSequence        = 16
	TagSet             = 17
	TagPrintableString = 19
	TagT61String       = 20
	TagIA5String       = 22
	TagUTCTime         = 23
	TagGeneralizedTime = 24
	TagGeneralString   = 27
)

// ASN.1 class types represent the namespace of the tag.
const (
	ClassUniversal       = 0
	ClassApplication     = 1
	ClassContextSpecific = 2
	ClassPrivate         = 3
)

type tagAndLength struct {
	class, tag, length int
	isCompound         bool
}

// ASN.1 has IMPLICIT and EXPLICIT tags, which can be translated as "instead
// of" and "in addition to". When not specified, every primitive type has a
// default tag in the UNIVERSAL class.
//
// For example: a BIT STRING is tagged [UNIVERSAL 3] by default (although ASN.1
// doesn't actually have a UNIVERSAL keyword). However, by saying [IMPLICIT
// CONTEXT-SPECIFIC 42], that means that the tag is replaced by another.
//
// On the other hand, if it said [EXPLICIT CONTEXT-SPECIFIC 10], then an
// /additional/ tag would wrap the default tag. This explicit tag will have the
// compound flag set.
//
// (This is used in order to remove ambiguity with optional elements.)
//
// You can layer EXPLICIT and IMPLICIT tags to an arbitrary depth, however we
// don't support that here. We support a single layer of EXPLICIT or IMPLICIT
// tagging with tag strings on the fields of a structure.

// fieldParameters is the parsed representation of tag string from a structure field.
type fieldParameters struct {
	optional     bool   // true iff the field is OPTIONAL
	explicit     bool   // true iff an EXPLICIT tag is in use.
	application  bool   // true iff an APPLICATION tag is in use.
	defaultValue *int64 // a default value for INTEGER typed fields (maybe nil).
	tag          *int   // the EXPLICIT or IMPLICIT tag (maybe nil).
	stringType   int    // the string tag to use when marshaling.
	timeType     int    // the time tag to use when marshaling.
	set          bool   // true iff this should be encoded as a SET
	omitEmpty    bool   // true iff this should be omitted if empty when marshaling.

	// Invariants:
	//   if explicit is set, tag is non-nil.
}

// Given a tag string with the format specified in the package comment,
// parseFieldParameters will parse it into a fieldParameters structure,
// ignoring unknown parts of the string.
func parseFieldParameters(str string) (ret fieldParameters) {
	for _, part := range strings.Split(str, ",") {
		switch part {
		case "optional":
			ret.optional = true
		case "explicit":
			ret.explicit = true
			if ret.tag == nil {
				ret.tag = new(int)
			}
		case "generalized":
			ret.timeType = TagGeneralizedTime
		case "utc":
			ret.timeType = TagUTCTime
		case "ia5":
			ret.stringType = TagIA5String
		case "printable":
			ret.stringType = TagPrintableString
		case "utf8":
			ret.stringType = TagUTF8String
		case "set":
			ret.set = true
		case "application":
			ret.application = true
			if ret.tag == nil {
				ret.tag = new(int)
			}
		case "omitempty":
			ret.omitEmpty = true
		default:
			switch {
			case strings.HasPrefix(part, "default:"):
				i, err := strconv.ParseInt(part[8:], 10, 64)
				if err == nil {
					ret.defaultValue = new(int64)
					*ret.defaultValue = i
				}
			case strings.HasPrefix(part, "tag:"):
				i, err := strconv.Atoi(part[4:])
				if err == nil {
					ret.tag = new(int)
					*ret.tag = i
				}
			}
		}
	}
	return
}

// Given a reflected Go type, getUniversalType returns the default tag number
// and expected compound flag.
func getUniversalType(t reflect.Type) (tagNumber int, isCompound, ok bool) {
	switch t {
	case objectIdentifierType:
		return TagOID, false, true
	case bitStringType:
		return TagBitString, false, true
	case timeType:
		return TagUTCTime, false, true
	case enumeratedType:
		return TagEnum, false, true
	case bigIntType:
		return TagInteger, false, true
	}
	switch t.Kind() {
	case reflect.Bool:
		return TagBoolean, false, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return TagInteger, false, true
	case reflect.Struct:
		return TagSequence, true, true
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return TagOctetString, false, true
		}
		if strings.HasSuffix(t.Name(), "SET") {
			return TagSet, true, true
		}
		return TagSequence, true, true
	case reflect.String:
		return TagPrintableString, false, true
	}
	return 0, false, false
}

// isPrintable reports whether the given b is in the ASN.1 PrintableString set.
func isPrintable(b byte) bool {
	return 'a' <= b && b <= 'z' ||
		'A' <= b && b <= 'Z' ||
		'0' <= b && b <= '9' ||
		'\'' <= b && b <= ')' ||
		'+' <= b && b <= '/' ||
		b == ' ' ||
		b == ':' ||
		b == '=' ||
		b == '?' ||
		// This is technically not allowed in a PrintableString.
		// However, x509 certificates with wildcard strings don't
		// always use the correct string type so we permit it.
		b == '*'
}
