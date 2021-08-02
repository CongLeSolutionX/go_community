// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package netip contains a IP address type that's a small value type.
// Building on that Addr type, the package also contains AddrPort (an
// IP address and a port), and Prefix (a IP address and a bit length
// prefix).
//
// Compared to the net.IP type, this package's Addr type takes less
// memory, is immutable, and is comparable (supports == and being a
// map key).
package netip

import (
	"errors"
	"math"
	"sort"
	"strconv"

	"internal/bytealg"
	"internal/intern"
	"internal/itoa"
)

// Sizes: (64-bit)
//   net.IP:     24 byte slice header + {4, 16} = 28 to 40 bytes
//   net.IPAddr: 40 byte slice header + {4, 16} = 44 to 56 bytes + zone length
//   netip.Addr: 24 bytes (zone is per-name singleton, shared across all users)

// Addr represents an IPv4 or IPv6 address (with or without a scoped
// addressing zone), similar to Go's net.IP or net.IPAddr.
//
// Unlike net.IP or net.IPAddr, the netip.Addr is a comparable value
// type (it supports == and can be a map key) and is immutable.
// Its memory representation is 24 bytes on 64-bit machines (the same
// size as a Go slice header) for both IPv4 and IPv6 address.
type Addr struct {
	// addr are the hi and lo bits of an IPv6 address. If z==z4,
	// hi and lo contain the IPv4-mapped IPv6 address.
	//
	// hi and lo are constructed by interpreting a 16-byte IPv6
	// address as a big-endian 128-bit number. The most significant
	// bits of that number go into hi, the rest into lo.
	//
	// For example, 0011:2233:4455:6677:8899:aabb:ccdd:eeff is stored as:
	//  addr.hi = 0x0011223344556677
	//  addr.lo = 0x8899aabbccddeeff
	//
	// We store IPs like this, rather than as [16]byte, because it
	// turns most operations on IPs into arithmetic and bit-twiddling
	// operations on 64-bit registers, which is much faster than
	// bytewise processing.
	addr uint128

	// z is a combination of the address family and the IPv6 zone.
	//
	// nil means invalid IP address (for the IP zero value).
	// z4 means an IPv4 address.
	// z6noz means an IPv6 address without a zone.
	//
	// Otherwise it's the interned zone name string.
	z *intern.Value
}

// z0, z4, and z6noz are sentinel IP.z values.
// See the IP type's field docs.
var (
	z0    = (*intern.Value)(nil)
	z4    = new(intern.Value)
	z6noz = new(intern.Value)
)

// IPv6LinkLocalAllNodes returns the IPv6 link-local all nodes multicast
// address ff02::1.
func IPv6LinkLocalAllNodes() Addr { return AddrFrom16([16]byte{0: 0xff, 1: 0x02, 15: 0x01}) }

// IPv6Unspecified returns the IPv6 unspecified address ::.
func IPv6Unspecified() Addr { return Addr{z: z6noz} }

// AddrFrom4 returns the address of the IPv4 address given by the bytes in addr.
func AddrFrom4(addr [4]byte) Addr {
	return Addr{
		addr: uint128{0, 0xffff00000000 | uint64(addr[0])<<24 | uint64(addr[1])<<16 | uint64(addr[2])<<8 | uint64(addr[3])},
		z:    z4,
	}
}

// AddrFrom16 returns the IPv6 address given by the bytes in addr,
// without an implicit Unmap call to unmap any v6-mapped IPv4
// address.
func AddrFrom16(addr [16]byte) Addr {
	return Addr{
		addr: uint128{
			beUint64(addr[:8]),
			beUint64(addr[8:]),
		},
		z: z6noz,
	}
}

// ipv6Slice is like IPv6Raw, but operates on a 16-byte slice. Assumes
// slice is 16 bytes, caller must enforce this.
func ipv6Slice(addr []byte) Addr {
	return Addr{
		addr: uint128{
			beUint64(addr[:8]),
			beUint64(addr[8:]),
		},
		z: z6noz,
	}
}

// ParseAddr parses s as an IP address, returning the result. The string
// s can be in dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"),
// or IPv6 with a scoped addressing zone ("fe80::1cc0:3e8c:119f:c2e1%ens18").
func ParseAddr(s string) (Addr, error) {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return parseIPv4(s)
		case ':':
			return parseIPv6(s)
		case '%':
			// Assume that this was trying to be an IPv6 address with
			// a zone specifier, but the address is missing.
			return Addr{}, parseAddrError{in: s, msg: "missing IPv6 address"}
		}
	}
	return Addr{}, parseAddrError{in: s, msg: "unable to parse IP"}
}

// MustParseAddr calls ParseAddr(s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParseAddr(s string) Addr {
	ip, err := ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return ip
}

type parseAddrError struct {
	in  string // the string given to ParseAddr
	msg string // an explanation of the parse failure
	at  string // optionally, the unparsed portion of in at which the error occurred.
}

func (err parseAddrError) Error() string {
	if err.at != "" {
		return sprintf("ParseAddr(%q): %s (at %q)", err.in, err.msg, err.at)
	}
	return sprintf("ParseAddr(%q): %s", err.in, err.msg)
}

// parseIPv4 parses s as an IPv4 address (in form "192.168.0.1").
func parseIPv4(s string) (ip Addr, err error) {
	var fields [4]uint8
	var val, pos int
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			val = val*10 + int(s[i]) - '0'
			if val > 255 {
				return Addr{}, parseAddrError{in: s, msg: "IPv4 field has value >255"}
			}
		} else if s[i] == '.' {
			// .1.2.3
			// 1.2.3.
			// 1..2.3
			if i == 0 || i == len(s)-1 || s[i-1] == '.' {
				return Addr{}, parseAddrError{in: s, msg: "IPv4 field must have at least one digit", at: s[i:]}
			}
			// 1.2.3.4.5
			if pos == 3 {
				return Addr{}, parseAddrError{in: s, msg: "IPv4 address too long"}
			}
			fields[pos] = uint8(val)
			pos++
			val = 0
		} else {
			return Addr{}, parseAddrError{in: s, msg: "unexpected character", at: s[i:]}
		}
	}
	if pos < 3 {
		return Addr{}, parseAddrError{in: s, msg: "IPv4 address too short"}
	}
	fields[3] = uint8(val)
	return AddrFrom4(fields), nil
}

// parseIPv6 parses s as an IPv6 address (in form "2001:db8::68").
func parseIPv6(in string) (Addr, error) {
	s := in

	// Split off the zone right from the start. Yes it's a second scan
	// of the string, but trying to handle it inline makes a bunch of
	// other inner loop conditionals more expensive, and it ends up
	// being slower.
	zone := ""
	i := stringsIndexByte(s, '%')
	if i != -1 {
		s, zone = s[:i], s[i+1:]
		if zone == "" {
			// Not allowed to have an empty zone if explicitly specified.
			return Addr{}, parseAddrError{in: in, msg: "zone must be a non-empty string"}
		}
	}

	var ip [16]byte
	ellipsis := -1 // position of ellipsis in ip

	// Might have leading ellipsis
	if len(s) >= 2 && s[0] == ':' && s[1] == ':' {
		ellipsis = 0
		s = s[2:]
		// Might be only ellipsis
		if len(s) == 0 {
			return IPv6Unspecified().WithZone(zone), nil
		}
	}

	// Loop, parsing hex numbers followed by colon.
	i = 0
	for i < 16 {
		// Hex number. Similar to parseIPv4, inlining the hex number
		// parsing yields a significant performance increase.
		off := 0
		acc := uint32(0)
		for ; off < len(s); off++ {
			c := s[off]
			if c >= '0' && c <= '9' {
				acc = (acc << 4) + uint32(c-'0')
			} else if c >= 'a' && c <= 'f' {
				acc = (acc << 4) + uint32(c-'a'+10)
			} else if c >= 'A' && c <= 'F' {
				acc = (acc << 4) + uint32(c-'A'+10)
			} else {
				break
			}
			if acc > math.MaxUint16 {
				// Overflow, fail.
				return Addr{}, parseAddrError{in: in, msg: "IPv6 field has value >=2^16", at: s}
			}
		}
		if off == 0 {
			// No digits found, fail.
			return Addr{}, parseAddrError{in: in, msg: "each colon-separated field must have at least one digit", at: s}
		}

		// If followed by dot, might be in trailing IPv4.
		if off < len(s) && s[off] == '.' {
			if ellipsis < 0 && i != 12 {
				// Not the right place.
				return Addr{}, parseAddrError{in: in, msg: "embedded IPv4 address must replace the final 2 fields of the address", at: s}
			}
			if i+4 > 16 {
				// Not enough room.
				return Addr{}, parseAddrError{in: in, msg: "too many hex fields to fit an embedded IPv4 at the end of the address", at: s}
			}
			// TODO: could make this a bit faster by having a helper
			// that parses to a [4]byte, and have both parseIPv4 and
			// parseIPv6 use it.
			ip4, err := parseIPv4(s)
			if err != nil {
				return Addr{}, parseAddrError{in: in, msg: err.Error(), at: s}
			}
			ip[i] = ip4.v4(0)
			ip[i+1] = ip4.v4(1)
			ip[i+2] = ip4.v4(2)
			ip[i+3] = ip4.v4(3)
			s = ""
			i += 4
			break
		}

		// Save this 16-bit chunk.
		ip[i] = byte(acc >> 8)
		ip[i+1] = byte(acc)
		i += 2

		// Stop at end of string.
		s = s[off:]
		if len(s) == 0 {
			break
		}

		// Otherwise must be followed by colon and more.
		if s[0] != ':' {
			return Addr{}, parseAddrError{in: in, msg: "unexpected character, want colon", at: s}
		} else if len(s) == 1 {
			return Addr{}, parseAddrError{in: in, msg: "colon must be followed by more characters", at: s}
		}
		s = s[1:]

		// Look for ellipsis.
		if s[0] == ':' {
			if ellipsis >= 0 { // already have one
				return Addr{}, parseAddrError{in: in, msg: "multiple :: in address", at: s}
			}
			ellipsis = i
			s = s[1:]
			if len(s) == 0 { // can be at end
				break
			}
		}
	}

	// Must have used entire string.
	if len(s) != 0 {
		return Addr{}, parseAddrError{in: in, msg: "trailing garbage after address", at: s}
	}

	// If didn't parse enough, expand ellipsis.
	if i < 16 {
		if ellipsis < 0 {
			return Addr{}, parseAddrError{in: in, msg: "address string too short"}
		}
		n := 16 - i
		for j := i - 1; j >= ellipsis; j-- {
			ip[j+n] = ip[j]
		}
		for j := ellipsis + n - 1; j >= ellipsis; j-- {
			ip[j] = 0
		}
	} else if ellipsis >= 0 {
		// Ellipsis must represent at least one 0 group.
		return Addr{}, parseAddrError{in: in, msg: "the :: must expand to at least one field of zeros"}
	}
	return AddrFrom16(ip).WithZone(zone), nil
}

// FromStdIP returns an IP from the standard library's IP type.
//
// If std is invalid, ok is false.
//
// FromStdIP implicitly unmaps IPv6-mapped IPv4 addresses. That is, if
// len(std) == 16 and contains an IPv4 address, only the IPv4 part is
// returned, without the IPv6 wrapper. This is the common form returned by
// the standard library's ParseAddr: https://play.golang.org/p/qdjylUkKWxl.
// To convert a standard library IP without the implicit unmapping, use
// FromStdIPRaw.
func FromStdIP(std []byte) (ip Addr, ok bool) {
	ret, ok := FromStdIPRaw(std)
	if ret.Is4in6() {
		ret.z = z4
	}
	return ret, ok
}

// FromStdIPRaw returns an IP from the standard library's IP type.
// If std is invalid, ok is false.
// Unlike FromStdIP, FromStdIPRaw does not do an implicit Unmap if
// len(std) == 16 and contains an IPv6-mapped IPv4 address.
func FromStdIPRaw(std []byte) (ip Addr, ok bool) {
	switch len(std) {
	case 4:
		return AddrFrom4(*(*[4]byte)(std)), true
	case 16:
		return ipv6Slice(std), true
	}
	return Addr{}, false
}

// v4 returns the i'th byte of ip. If ip is not an IPv4, v4 returns
// unspecified garbage.
func (ip Addr) v4(i uint8) uint8 {
	return uint8(ip.addr.lo >> ((3 - i) * 8))
}

// v6 returns the i'th byte of ip. If ip is an IPv4 address, this
// accesses the IPv4-mapped IPv6 address form of the IP.
func (ip Addr) v6(i uint8) uint8 {
	return uint8(*(ip.addr.halves()[(i/8)%2]) >> ((7 - i%8) * 8))
}

// v6u16 returns the i'th 16-bit word of ip. If ip is an IPv4 address,
// this accesses the IPv4-mapped IPv6 address form of the IP.
func (ip Addr) v6u16(i uint8) uint16 {
	return uint16(*(ip.addr.halves()[(i/4)%2]) >> ((3 - i%4) * 16))
}

// IsZero reports whether ip is the zero value of the IP type.
// The zero value is not a valid IP address of any type.
//
// Note that "0.0.0.0" and "::" are not the zero value. Use IsUnspecified to
// check for these values instead.
func (ip Addr) IsZero() bool {
	// Faster than comparing ip == Addr{}, but effectively equivalent,
	// as there's no way to make an IP with a nil z from this package.
	return ip.z == z0
}

// IsValid whether the IP is an initialized value and not the IP
// type's zero value.
//
// Note that both "0.0.0.0" and "::" are valid, non-zero values.
func (ip Addr) IsValid() bool { return ip.z != z0 }

// BitLen returns the number of bits in the IP address:
// 32 for IPv4 or 128 for IPv6.
// For the zero value (see IP.IsZero), it returns 0.
// For IP4-mapped IPv6 addresses, it returns 128.
func (ip Addr) BitLen() uint8 {
	switch ip.z {
	case z0:
		return 0
	case z4:
		return 32
	}
	return 128
}

// Zone returns ip's IPv6 scoped addressing zone, if any.
func (ip Addr) Zone() string {
	if ip.z == nil {
		return ""
	}
	zone, _ := ip.z.Get().(string)
	return zone
}

// Compare returns an integer comparing two IPs.
// The result will be 0 if ip==ip2, -1 if ip < ip2, and +1 if ip > ip2.
// The definition of "less than" is the same as the IP.Less method.
func (ip Addr) Compare(ip2 Addr) int {
	f1, f2 := ip.BitLen(), ip2.BitLen()
	if f1 < f2 {
		return -1
	}
	if f1 > f2 {
		return 1
	}
	if hi1, hi2 := ip.addr.hi, ip2.addr.hi; hi1 < hi2 {
		return -1
	} else if hi1 > hi2 {
		return 1
	}
	if lo1, lo2 := ip.addr.lo, ip2.addr.lo; lo1 < lo2 {
		return -1
	} else if lo1 > lo2 {
		return 1
	}
	if ip.Is6() {
		za, zb := ip.Zone(), ip2.Zone()
		if za < zb {
			return -1
		} else if za > zb {
			return 1
		}
	}
	return 0
}

// Less reports whether ip sorts before ip2.
// IP addresses sort first by length, then their address.
// IPv6 addresses with zones sort just after the same address without a zone.
func (ip Addr) Less(ip2 Addr) bool { return ip.Compare(ip2) == -1 }

func (ip Addr) lessOrEq(ip2 Addr) bool { return ip.Compare(ip2) <= 0 }

// ipZone returns the standard library net.IP from ip, as well
// as the zone.
// The optional reuse IP provides memory to reuse.
func (ip Addr) ipZone(reuse []byte) (stdIP []byte, zone string) {
	base := reuse[:0]
	switch {
	case ip.z == z0:
		return nil, ""
	case ip.Is4():
		a4 := ip.As4()
		return append(base, a4[:]...), ""
	default:
		a16 := ip.As16()
		return append(base, a16[:]...), ip.Zone()
	}
}

// IPAddrParts returns the net.IPAddr representation of an IP. The returned value is
// always non-nil, but the IPAddr.IP will be nil if ip is the zero value.
// If ip contains a zone identifier, IPAddr.Zone is populated.
func (ip Addr) IPAddrParts() (stdIP []byte, zone string) {
	return ip.ipZone(nil)
}

// Is4 reports whether ip is an IPv4 address.
//
// It returns false for IP4-mapped IPv6 addresses. See IP.Unmap.
func (ip Addr) Is4() bool {
	return ip.z == z4
}

// Is4in6 reports whether ip is an IPv4-mapped IPv6 address.
func (ip Addr) Is4in6() bool {
	return ip.Is6() && ip.addr.hi == 0 && ip.addr.lo>>32 == 0xffff
}

// Is6 reports whether ip is an IPv6 address, including IPv4-mapped
// IPv6 addresses.
func (ip Addr) Is6() bool {
	return ip.z != z0 && ip.z != z4
}

// Unmap returns ip with any IPv4-mapped IPv6 address prefix removed.
//
// That is, if ip is an IPv6 address wrapping an IPv4 adddress, it
// returns the wrapped IPv4 address. Otherwise it returns ip, regardless
// of its type.
func (ip Addr) Unmap() Addr {
	if ip.Is4in6() {
		ip.z = z4
	}
	return ip
}

// WithZone returns an IP that's the same as ip but with the provided
// zone. If zone is empty, the zone is removed. If ip is an IPv4
// address it's returned unchanged.
func (ip Addr) WithZone(zone string) Addr {
	if !ip.Is6() {
		return ip
	}
	if zone == "" {
		ip.z = z6noz
		return ip
	}
	ip.z = intern.GetByString(zone)
	return ip
}

// noZone unconditionally strips the zone from IP.
// It's similar to WithZone, but small enough to be inlinable.
func (ip Addr) withoutZone() Addr {
	if !ip.Is6() {
		return ip
	}
	ip.z = z6noz
	return ip
}

// hasZone reports whether IP has an IPv6 zone.
func (ip Addr) hasZone() bool {
	return ip.z != z0 && ip.z != z4 && ip.z != z6noz
}

// IsLinkLocalUnicast reports whether ip is a link-local unicast address.
// If ip is the zero value, it will return false.
func (ip Addr) IsLinkLocalUnicast() bool {
	// Dynamic Configuration of IPv4 Link-Local Addresses
	// https://datatracker.ietf.org/doc/html/rfc3927#section-2.1
	if ip.Is4() {
		return ip.v4(0) == 169 && ip.v4(1) == 254
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.v6u16(0)&0xffc0 == 0xfe80
	}
	return false // zero value
}

// IsLoopback reports whether ip is a loopback address. If ip is the zero value,
// it will return false.
func (ip Addr) IsLoopback() bool {
	// Requirements for Internet Hosts -- Communication Layers (3.2.1.3 Addressing)
	// https://datatracker.ietf.org/doc/html/rfc1122#section-3.2.1.3
	if ip.Is4() {
		return ip.v4(0) == 127
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.addr.hi == 0 && ip.addr.lo == 1
	}
	return false // zero value
}

// IsMulticast reports whether ip is a multicast address. If ip is the zero
// value, it will return false.
func (ip Addr) IsMulticast() bool {
	// Host Extensions for IP Multicasting (4. HOST GROUP ADDRESSES)
	// https://datatracker.ietf.org/doc/html/rfc1112#section-4
	if ip.Is4() {
		return ip.v4(0)&0xf0 == 0xe0
	}
	// IP Version 6 Addressing Architecture (2.4 Address Type Identification)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.4
	if ip.Is6() {
		return ip.addr.hi>>(64-8) == 0xff // ip.v6(0) == 0xff
	}
	return false // zero value
}

// IsInterfaceLocalMulticast reports whether ip is an IPv6 interface-local
// multicast address. If ip is the zero value or an IPv4 address, it will return
// false.
func (ip Addr) IsInterfaceLocalMulticast() bool {
	// IPv6 Addressing Architecture (2.7.1. Pre-Defined Multicast Addresses)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.7.1
	if ip.Is6() {
		return ip.v6u16(0)&0xff0f == 0xff01
	}
	return false // zero value
}

// IsLinkLocalMulticast reports whether ip is a link-local multicast address.
// If ip is the zero value, it will return false.
func (ip Addr) IsLinkLocalMulticast() bool {
	// IPv4 Multicast Guidelines (4. Local Network Control Block (224.0.0/24))
	// https://datatracker.ietf.org/doc/html/rfc5771#section-4
	if ip.Is4() {
		return ip.v4(0) == 224 && ip.v4(1) == 0 && ip.v4(2) == 0
	}
	// IPv6 Addressing Architecture (2.7.1. Pre-Defined Multicast Addresses)
	// https://datatracker.ietf.org/doc/html/rfc4291#section-2.7.1
	if ip.Is6() {
		return ip.v6u16(0)&0xff0f == 0xff02
	}
	return false // zero value
}

// IsGlobalUnicast reports whether ip is a global unicast address.
//
// It returns true for IPv6 addresses which fall outside of the current
// IANA-allocated 2000::/3 global unicast space, with the exception of the
// link-local address space. It also returns true even if ip is in the IPv4
// private address space or IPv6 unique local address space. If ip is the zero
// value, it will return false.
//
// For reference, see RFC 1122, RFC 4291, and RFC 4632.
func (ip Addr) IsGlobalUnicast() bool {
	if ip.z == z0 {
		// Invalid or zero-value.
		return false
	}

	// Match the stdlib's IsGlobalUnicast logic. Notably private IPv4 addresses
	// and ULA IPv6 addresses are still considered "global unicast".
	if ip.Is4() && (ip == AddrFrom4([4]byte{}) || ip == AddrFrom4([4]byte{255, 255, 255, 255})) {
		return false
	}

	return ip != IPv6Unspecified() &&
		!ip.IsLoopback() &&
		!ip.IsMulticast() &&
		!ip.IsLinkLocalUnicast()
}

// IsPrivate reports whether ip is a private address, according to RFC 1918
// (IPv4 addresses) and RFC 4193 (IPv6 addresses). That is, it reports whether
// ip is in 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, or fc00::/7. This is the
// same as the standard library's net.IP.IsPrivate.
func (ip Addr) IsPrivate() bool {
	// Match the stdlib's IsPrivate logic.
	if ip.Is4() {
		// RFC 1918 allocates 10.0.0.0/8, 172.16.0.0/12, and 192.168.0.0/16 as
		// private IPv4 address subnets.
		return ip.v4(0) == 10 ||
			(ip.v4(0) == 172 && ip.v4(1)&0xf0 == 16) ||
			(ip.v4(0) == 192 && ip.v4(1) == 168)
	}

	if ip.Is6() {
		// RFC 4193 allocates fc00::/7 as the unique local unicast IPv6 address
		// subnet.
		return ip.v6(0)&0xfe == 0xfc
	}

	return false // zero value
}

// IsUnspecified reports whether ip is an unspecified address, either the IPv4
// address "0.0.0.0" or the IPv6 address "::".
//
// Note that the IP zero value is not an unspecified address. Use IsZero to
// check for the zero value instead.
func (ip Addr) IsUnspecified() bool {
	return ip == AddrFrom4([4]byte{}) || ip == IPv6Unspecified()
}

// Prefix applies a CIDR mask of leading bits to IP, producing an Prefix
// of the specified length. If IP is the zero value, a zero-value Prefix and
// a nil error are returned. If bits is larger than 32 for an IPv4 address or
// 128 for an IPv6 address, an error is returned.
func (ip Addr) Prefix(bits uint8) (Prefix, error) {
	effectiveBits := bits
	switch ip.z {
	case z0:
		return Prefix{}, nil
	case z4:
		if bits > 32 {
			return Prefix{}, errorf("prefix length %d too large for IPv4", bits)
		}
		effectiveBits += 96
	default:
		if bits > 128 {
			return Prefix{}, errorf("prefix length %d too large for IPv6", bits)
		}
	}
	ip.addr = ip.addr.and(mask6[effectiveBits])
	return PrefixFrom(ip, bits), nil
}

const (
	netIPv4len = 4
	netIPv6len = 16
)

// As16 returns the IP address in its 16 byte representation.
// IPv4 addresses are returned in their v6-mapped form.
// IPv6 addresses with zones are returned without their zone (use the
// Zone method to get it).
// The ip zero value returns all zeroes.
func (ip Addr) As16() [16]byte {
	var ret [16]byte
	bePutUint64(ret[:8], ip.addr.hi)
	bePutUint64(ret[8:], ip.addr.lo)
	return ret
}

// As4 returns an IPv4 or IPv4-in-IPv6 address in its 4 byte representation.
// If ip is the IP zero value or an IPv6 address, As4 panics.
// Note that 0.0.0.0 is not the zero value.
func (ip Addr) As4() [4]byte {
	if ip.z == z4 || ip.Is4in6() {
		var ret [4]byte
		bePutUint32(ret[:], uint32(ip.addr.lo))
		return ret
	}
	if ip.z == z0 {
		panic("As4 called on IP zero value")
	}
	panic("As4 called on IPv6 address")
}

// Next returns the IP following ip.
// If there is none, it returns the IP zero value.
func (ip Addr) Next() Addr {
	ip.addr = ip.addr.addOne()
	if ip.Is4() {
		if uint32(ip.addr.lo) == 0 {
			// Overflowed.
			return Addr{}
		}
	} else {
		if ip.addr.isZero() {
			// Overflowed
			return Addr{}
		}
	}
	return ip
}

// Prior returns the IP before ip.
// If there is none, it returns the IP zero value.
func (ip Addr) Prior() Addr {
	if ip.Is4() {
		if uint32(ip.addr.lo) == 0 {
			return Addr{}
		}
	} else if ip.addr.isZero() {
		return Addr{}
	}
	ip.addr = ip.addr.subOne()
	return ip
}

// String returns the string form of the IP address ip.
// It returns one of 4 forms:
//
//   - "invalid IP", if ip is the zero value
//   - IPv4 dotted decimal ("192.0.2.1")
//   - IPv6 ("2001:db8::1")
//   - IPv6 with zone ("fe80:db8::1%eth0")
//
// Note that unlike the Go standard library's IP.String method,
// IP4-mapped IPv6 addresses do not format as dotted decimals.
func (ip Addr) String() string {
	switch ip.z {
	case z0:
		return "zero IP"
	case z4:
		return ip.string4()
	default:
		return ip.string6()
	}
}

// AppendTo appends a text encoding of ip,
// as generated by MarshalText,
// to b and returns the extended buffer.
func (ip Addr) AppendTo(b []byte) []byte {
	switch ip.z {
	case z0:
		return b
	case z4:
		return ip.appendTo4(b)
	default:
		return ip.appendTo6(b)
	}
}

// digits is a string of the hex digits from 0 to f. It's used in
// appendDecimal and appendHex to format IP addresses.
const digits = "0123456789abcdef"

// appendDecimal appends the decimal string representation of x to b.
func appendDecimal(b []byte, x uint8) []byte {
	// Using this function rather than strconv.AppendUint makes IPv4
	// string building 2x faster.

	if x >= 100 {
		b = append(b, digits[x/100])
	}
	if x >= 10 {
		b = append(b, digits[x/10%10])
	}
	return append(b, digits[x%10])
}

// appendHex appends the hex string representation of x to b.
func appendHex(b []byte, x uint16) []byte {
	// Using this function rather than strconv.AppendUint makes IPv6
	// string building 2x faster.

	if x >= 0x1000 {
		b = append(b, digits[x>>12])
	}
	if x >= 0x100 {
		b = append(b, digits[x>>8&0xf])
	}
	if x >= 0x10 {
		b = append(b, digits[x>>4&0xf])
	}
	return append(b, digits[x&0xf])
}

// appendHexPad appends the fully padded hex string representation of x to b.
func appendHexPad(b []byte, x uint16) []byte {
	return append(b, digits[x>>12], digits[x>>8&0xf], digits[x>>4&0xf], digits[x&0xf])
}

func (ip Addr) string4() string {
	const max = len("255.255.255.255")
	ret := make([]byte, 0, max)
	ret = ip.appendTo4(ret)
	return string(ret)
}

func (ip Addr) appendTo4(ret []byte) []byte {
	ret = appendDecimal(ret, ip.v4(0))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(1))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(2))
	ret = append(ret, '.')
	ret = appendDecimal(ret, ip.v4(3))
	return ret
}

// string6 formats ip in IPv6 textual representation. It follows the
// guidelines in section 4 of RFC 5952
// (https://tools.ietf.org/html/rfc5952#section-4): no unnecessary
// zeros, use :: to elide the longest run of zeros, and don't use ::
// to compact a single zero field.
func (ip Addr) string6() string {
	// Use a zone with a "plausibly long" name, so that most zone-ful
	// IP addresses won't require additional allocation.
	//
	// The compiler does a cool optimization here, where ret ends up
	// stack-allocated and so the only allocation this function does
	// is to construct the returned string. As such, it's okay to be a
	// bit greedy here, size-wise.
	const max = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0")
	ret := make([]byte, 0, max)
	ret = ip.appendTo6(ret)
	return string(ret)
}

func (ip Addr) appendTo6(ret []byte) []byte {
	zeroStart, zeroEnd := uint8(255), uint8(255)
	for i := uint8(0); i < 8; i++ {
		j := i
		for j < 8 && ip.v6u16(j) == 0 {
			j++
		}
		if l := j - i; l >= 2 && l > zeroEnd-zeroStart {
			zeroStart, zeroEnd = i, j
		}
	}

	for i := uint8(0); i < 8; i++ {
		if i == zeroStart {
			ret = append(ret, ':', ':')
			i = zeroEnd
			if i >= 8 {
				break
			}
		} else if i > 0 {
			ret = append(ret, ':')
		}

		ret = appendHex(ret, ip.v6u16(i))
	}

	if ip.z != z6noz {
		ret = append(ret, '%')
		ret = append(ret, ip.Zone()...)
	}
	return ret
}

// StringExpanded is like String but IPv6 addresses are expanded with leading
// zeroes and no "::" compression. For example, "2001:db8::1" becomes
// "2001:0db8:0000:0000:0000:0000:0000:0001".
func (ip Addr) StringExpanded() string {
	switch ip.z {
	case z0, z4:
		return ip.String()
	}

	const size = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	ret := make([]byte, 0, size)
	for i := uint8(0); i < 8; i++ {
		if i > 0 {
			ret = append(ret, ':')
		}

		ret = appendHexPad(ret, ip.v6u16(i))
	}

	if ip.z != z6noz {
		// The addition of a zone will cause a second allocation, but when there
		// is no zone the ret slice will be stack allocated.
		ret = append(ret, '%')
		ret = append(ret, ip.Zone()...)
	}
	return string(ret)
}

// MarshalText implements the encoding.TextMarshaler interface,
// The encoding is the same as returned by String, with one exception:
// If ip is the zero value, the encoding is the empty string.
func (ip Addr) MarshalText() ([]byte, error) {
	switch ip.z {
	case z0:
		return []byte(""), nil
	case z4:
		max := len("255.255.255.255")
		b := make([]byte, 0, max)
		return ip.appendTo4(b), nil
	default:
		max := len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0")
		b := make([]byte, 0, max)
		return ip.appendTo6(b), nil
	}
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The IP address is expected in a form accepted by ParseAddr.
// It returns an error if *ip is not the IP zero value.
func (ip *Addr) UnmarshalText(text []byte) error {
	if ip.z != z0 {
		return errors.New("refusing to Unmarshal into non-zero IP")
	}
	if len(text) == 0 {
		return nil
	}
	var err error
	*ip, err = ParseAddr(string(text))
	return err
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (ip Addr) MarshalBinary() ([]byte, error) {
	switch ip.z {
	case z0:
		return nil, nil
	case z4:
		b := ip.As4()
		return b[:], nil
	default:
		b16 := ip.As16()
		b := b16[:]
		if z := ip.Zone(); z != "" {
			b = append(b, []byte(z)...)
		}
		return b, nil
	}
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (ip *Addr) UnmarshalBinary(b []byte) error {
	if ip.z != z0 {
		return errors.New("refusing to Unmarshal into non-zero IP")
	}
	n := len(b)
	switch {
	case n == 0:
		return nil
	case n == 4:
		*ip = AddrFrom4(*(*[4]byte)(b))
		return nil
	case n == 16:
		*ip = ipv6Slice(b)
		return nil
	case n > 16:
		*ip = ipv6Slice(b[:16]).WithZone(string(b[16:]))
		return nil
	}
	return errorf("unexpected ip size: %v", len(b))
}

// AddrPort is an IP and a port number.
type AddrPort struct {
	ip   Addr
	port uint16
}

// AddrPortFrom returns an AddrPort with IP ip and port port.
// It does not allocate.
func AddrPortFrom(ip Addr, port uint16) AddrPort { return AddrPort{ip: ip, port: port} }

// WithIP returns an AddrPort with IP ip and port p.Port().
func (p AddrPort) WithIP(ip Addr) AddrPort { return AddrPort{ip: ip, port: p.port} }

// WithIP returns an AddrPort with IP p.IP() and port port.
func (p AddrPort) WithPort(port uint16) AddrPort { return AddrPort{ip: p.ip, port: port} }

// IP returns p's IP.
func (p AddrPort) IP() Addr { return p.ip }

// Port returns p's port.
func (p AddrPort) Port() uint16 { return p.port }

// splitAddrPort splits s into an IP address string and a port
// string. It splits strings shaped like "foo:bar" or "[foo]:bar",
// without further validating the substrings. v6 indicates whether the
// ip string should parse as an IPv6 address or an IPv4 address, in
// order for s to be a valid ip:port string.
func splitAddrPort(s string) (ip, port string, v6 bool, err error) {
	i := stringsLastIndexByte(s, ':')
	if i == -1 {
		return "", "", false, errors.New("not an ip:port")
	}

	ip, port = s[:i], s[i+1:]
	if len(ip) == 0 {
		return "", "", false, errors.New("no IP")
	}
	if len(port) == 0 {
		return "", "", false, errors.New("no port")
	}
	if ip[0] == '[' {
		if len(ip) < 2 || ip[len(ip)-1] != ']' {
			return "", "", false, errors.New("missing ]")
		}
		ip = ip[1 : len(ip)-1]
		v6 = true
	}

	return ip, port, v6, nil
}

// ParseAddrPort parses s as an AddrPort.
//
// It doesn't do any name resolution, and ports must be numeric.
func ParseAddrPort(s string) (AddrPort, error) {
	var ipp AddrPort
	ip, port, v6, err := splitAddrPort(s)
	if err != nil {
		return ipp, err
	}
	port16, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return ipp, errorf("invalid port %q parsing %q", port, s)
	}
	ipp.port = uint16(port16)
	ipp.ip, err = ParseAddr(ip)
	if err != nil {
		return AddrPort{}, err
	}
	if v6 && ipp.ip.Is4() {
		return AddrPort{}, errorf("invalid ip:port %q, square brackets can only be used with IPv6 addresses", s)
	} else if !v6 && ipp.ip.Is6() {
		return AddrPort{}, errorf("invalid ip:port %q, IPv6 addresses must be surrounded by square brackets", s)
	}
	return ipp, nil
}

// MustParseAddrPort calls ParseAddrPort(s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParseAddrPort(s string) AddrPort {
	ip, err := ParseAddrPort(s)
	if err != nil {
		panic(err)
	}
	return ip
}

// IsZero reports whether p is its zero value.
func (p AddrPort) IsZero() bool { return p == AddrPort{} }

// IsValid reports whether p.IP() is valid.
// All ports are valid, including zero.
func (p AddrPort) IsValid() bool { return p.ip.IsValid() }

// Valid reports whether p.IP() is valid.
// All ports are valid, including zero.
//
// Deprecated: use the correctly named and identical IsValid method instead.
func (p AddrPort) Valid() bool { return p.IsValid() }

func (p AddrPort) String() string {
	switch p.ip.z {
	case z0:
		return "invalid AddrPort"
	case z4:
		a := p.ip.As4()
		return sprintf("%d.%d.%d.%d:%d", a[0], a[1], a[2], a[3], p.port)
	default:
		// TODO: this could be more efficient allocation-wise:
		return joinHostPort(p.ip.String(), itoa.Itoa(int(p.port)))
	}
}

func joinHostPort(host, port string) string {
	// We assume that host is a literal IPv6 address if host has
	// colons.
	if bytealg.IndexByteString(host, ':') >= 0 {
		return "[" + host + "]:" + port
	}
	return host + ":" + port
}

// AppendTo appends a text encoding of p,
// as generated by MarshalText,
// to b and returns the extended buffer.
func (p AddrPort) AppendTo(b []byte) []byte {
	switch p.ip.z {
	case z0:
		return b
	case z4:
		b = p.ip.appendTo4(b)
	default:
		b = append(b, '[')
		b = p.ip.appendTo6(b)
		b = append(b, ']')
	}
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(p.port), 10)
	return b
}

// MarshalText implements the encoding.TextMarshaler interface. The
// encoding is the same as returned by String, with one exception: if
// p.IP() is the zero value, the encoding is the empty string.
func (p AddrPort) MarshalText() ([]byte, error) {
	var max int
	switch p.ip.z {
	case z0:
	case z4:
		max = len("255.255.255.255:65535")
	default:
		max = len("[ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0]:65535")
	}
	b := make([]byte, 0, max)
	b = p.AppendTo(b)
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler
// interface. The AddrPort is expected in a form accepted by
// ParseAddrPort. It returns an error if *p is not the AddrPort zero
// value.
func (p *AddrPort) UnmarshalText(text []byte) error {
	if p.ip.z != z0 || p.port != 0 {
		return errors.New("refusing to Unmarshal into non-zero AddrPort")
	}
	if len(text) == 0 {
		return nil
	}
	var err error
	*p, err = ParseAddrPort(string(text))
	return err
}

// FromStdAddr maps the components of a standard library TCPAddr or
// UDPAddr into an AddrPort.
func FromStdAddr(stdIP []byte, port int, zone string) (_ AddrPort, ok bool) {
	ip, ok := FromStdIP(stdIP)
	if !ok || port < 0 || port > math.MaxUint16 {
		return
	}
	ip = ip.Unmap()
	if zone != "" {
		if ip.Is4() {
			ok = false
			return
		}
		ip = ip.WithZone(zone)
	}
	return AddrPort{ip: ip, port: uint16(port)}, true
}

// Prefix is an IP address prefix (CIDR) representing an IP network.
//
// The first Bits() of IP() are specified. The remaining bits match any address.
// The range of Bits() is [0,32] for IPv4 or [0,128] for IPv6.
type Prefix struct {
	ip   Addr
	bits uint8
}

// PrefixFrom returns an Prefix with IP ip and port port.
// It does not allocate.
func PrefixFrom(ip Addr, bits uint8) Prefix {
	return Prefix{
		ip:   ip.withoutZone(),
		bits: bits,
	}
}

// IP returns p's IP.
func (p Prefix) IP() Addr { return p.ip }

// Bits returns p's prefix length.
func (p Prefix) Bits() uint8 { return p.bits }

// IsValid reports whether whether p.Bits() has a valid range for p.IP().
// If p.IP() is zero, Valid returns false.
func (p Prefix) IsValid() bool { return !p.ip.IsZero() && p.bits <= p.ip.BitLen() }

// Valid reports whether whether p.Bits() has a valid range for p.IP().
// If p.IP() is zero, Valid returns false.
//
// Deprecated: use the correctly named and identical IsValid method instead.
func (p Prefix) Valid() bool { return p.IsValid() }

// IsZero reports whether p is its zero value.
func (p Prefix) IsZero() bool { return p == Prefix{} }

// IsSingleIP reports whether p contains exactly one IP.
func (p Prefix) IsSingleIP() bool { return p.bits != 0 && p.bits == p.ip.BitLen() }

// ParsePrefix parses s as an IP address prefix.
// The string can be in the form "192.168.1.0/24" or "2001::db8::/32",
// the CIDR notation defined in RFC 4632 and RFC 4291.
//
// Note that masked address bits are not zeroed. Use Masked for that.
func ParsePrefix(s string) (Prefix, error) {
	i := stringsLastIndexByte(s, '/')
	if i < 0 {
		return Prefix{}, errorf("netip.ParsePrefix(%q): no '/'", s)
	}
	ip, err := ParseAddr(s[:i])
	if err != nil {
		return Prefix{}, errorf("netip.ParsePrefix(%q): %v", s, err)
	}
	s = s[i+1:]
	bits, err := strconv.Atoi(s)
	if err != nil {
		return Prefix{}, errorf("netip.ParsePrefix(%q): bad prefix: %v", s, err)
	}
	maxBits := 32
	if ip.Is6() {
		maxBits = 128
	}
	if bits < 0 || bits > maxBits {
		return Prefix{}, errorf("netip.ParsePrefix(%q): prefix length out of range", s)
	}
	return PrefixFrom(ip, uint8(bits)), nil
}

// MustParsePrefix calls ParsePrefix(s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParsePrefix(s string) Prefix {
	ip, err := ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return ip
}

// Masked returns p in its canonical form, with bits of p.IP() not in p.Bits() masked off.
// If p is zero or otherwise invalid, Masked returns the zero value.
func (p Prefix) Masked() Prefix {
	if m, err := p.ip.Prefix(p.bits); err == nil {
		return m
	}
	return Prefix{}
}

// Range returns the inclusive range of IPs that p covers.
//
// If p is zero or otherwise invalid, Range returns the zero value.
func (p Prefix) Range() Range {
	p = p.Masked()
	if p.IsZero() {
		return Range{}
	}
	return RangeFrom(p.ip, p.lastIP())
}

// Contains reports whether the network p includes ip.
//
// An IPv4 address will not match an IPv6 prefix.
// A v6-mapped IPv6 address will not match an IPv4 prefix.
// A zero-value IP will not match any prefix.
// If ip has an IPv6 zone, Contains returns false,
// because Prefixes strip zones.
func (p Prefix) Contains(ip Addr) bool {
	if !p.IsValid() || ip.hasZone() {
		return false
	}
	if f1, f2 := p.ip.BitLen(), ip.BitLen(); f1 == 0 || f2 == 0 || f1 != f2 {
		return false
	}
	if ip.Is4() {
		// xor the IP addresses together; mismatched bits are now ones.
		// Shift away the number of bits we don't care about.
		// Shifts in Go are more efficient if the compiler can prove
		// that the shift amount is smaller than the width of the shifted type (64 here).
		// We know that p.bits is in the range 0..32 because p is Valid;
		// the compiler doesn't know that, so mask with 63 to help it.
		// Now truncate to 32 bits, because this is IPv4.
		// If all the bits we care about are equal, the result will be zero.
		return uint32((ip.addr.lo^p.ip.addr.lo)>>((32-p.bits)&63)) == 0
	} else {
		// xor the IP addresses together.
		// Mask away the bits we don't care about.
		// If all the bits we care about are equal, the result will be zero.
		return ip.addr.xor(p.ip.addr).and(mask6[p.bits]).isZero()
	}
}

// Overlaps reports whether p and o overlap at all.
//
// If p and o are of different address families or either have a zero
// IP, it reports false. Like the Contains method, a prefix with a
// v6-mapped IPv4 IP is still treated as an IPv6 mask.
//
// If either has a Bits of zero, it returns true.
func (p Prefix) Overlaps(o Prefix) bool {
	if !p.IsValid() || !o.IsValid() {
		return false
	}
	if p == o {
		return true
	}
	if p.ip.Is4() != o.ip.Is4() {
		return false
	}
	var minBits uint8
	if p.bits < o.bits {
		minBits = p.bits
	} else {
		minBits = o.bits
	}
	if minBits == 0 {
		return true
	}
	// One of these Prefix calls might look redundant, but we don't require
	// that p and o values are normalized (via Prefix.Masked) first,
	// so the Prefix call on the one that's already minBits serves to zero
	// out any remaining bits in IP.
	var err error
	if p, err = p.ip.Prefix(minBits); err != nil {
		return false
	}
	if o, err = o.ip.Prefix(minBits); err != nil {
		return false
	}
	return p.ip == o.ip
}

// AppendTo appends a text encoding of p,
// as generated by MarshalText,
// to b and returns the extended buffer.
func (p Prefix) AppendTo(b []byte) []byte {
	if p.IsZero() {
		return b
	}
	if !p.IsValid() {
		return append(b, "invalid Prefix"...)
	}

	// p.IP is non-zero, because p is valid.
	if p.ip.z == z4 {
		b = p.ip.appendTo4(b)
	} else {
		b = p.ip.appendTo6(b)
	}

	b = append(b, '/')
	b = appendDecimal(b, p.bits)
	return b
}

// MarshalText implements the encoding.TextMarshaler interface,
// The encoding is the same as returned by String, with one exception:
// If p is the zero value, the encoding is the empty string.
func (p Prefix) MarshalText() ([]byte, error) {
	var max int
	switch p.ip.z {
	case z0:
	case z4:
		max = len("255.255.255.255/32")
	default:
		max = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff%enp5s0/128")
	}
	b := make([]byte, 0, max)
	b = p.AppendTo(b)
	return b, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The IP address is expected in a form accepted by ParsePrefix.
// It returns an error if *p is not the Prefix zero value.
func (p *Prefix) UnmarshalText(text []byte) error {
	if *p != (Prefix{}) {
		return errors.New("refusing to Unmarshal into non-zero Prefix")
	}
	if len(text) == 0 {
		return nil
	}
	var err error
	*p, err = ParsePrefix(string(text))
	return err
}

// String returns the CIDR notation of p: "<ip>/<bits>".
func (p Prefix) String() string {
	if p.IsZero() {
		return "zero Prefix"
	}
	if !p.IsValid() {
		return "invalid Prefix"
	}
	return sprintf("%s/%d", p.ip, p.bits)
}

// lastIP returns the last IP in the prefix.
func (p Prefix) lastIP() Addr {
	if !p.IsValid() {
		return Addr{}
	}
	a16 := p.ip.As16()
	var off uint8
	var bits uint8 = 128
	if p.ip.Is4() {
		off = 12
		bits = 32
	}
	for b := p.bits; b < bits; b++ {
		byteNum, bitInByte := b/8, 7-(b%8)
		a16[off+byteNum] |= 1 << uint(bitInByte)
	}
	if p.ip.Is4() {
		return AddrFrom16(a16).Unmap()
	} else {
		return AddrFrom16(a16) // doesn't unmap
	}
}

// Range represents an inclusive range of IP addresses
// from the same address family.
//
// The From() and To() IPs are inclusive bounds, both included in the
// range.
//
// To be valid, the From() and To() values must be non-zero, have matching
// address families (IPv4 vs IPv6), and From() must be less than or equal to To().
// IPv6 zones are stripped out and ignored.
// An invalid range may be ignored.
type Range struct {
	// from is the initial IP address in the range.
	from Addr

	// to is the final IP address in the range.
	to Addr
}

// RangeFrom returns an Range from from to to.
// It does not allocate.
func RangeFrom(from, to Addr) Range {
	return Range{
		from: from.withoutZone(),
		to:   to.withoutZone(),
	}
}

// From returns the lower bound of r.
func (r Range) From() Addr { return r.from }

// To returns the upper bound of r.
func (r Range) To() Addr { return r.to }

// ParseRange parses a range out of two IPs separated by a hyphen.
//
// It returns an error if the range is not valid.
func ParseRange(s string) (Range, error) {
	var r Range
	h := stringsIndexByte(s, '-')
	if h == -1 {
		return r, errorf("no hyphen in range %q", s)
	}
	from, to := s[:h], s[h+1:]
	var err error
	r.from, err = ParseAddr(from)
	if err != nil {
		return r, errorf("invalid From IP %q in range %q", from, s)
	}
	r.from = r.from.withoutZone()
	r.to, err = ParseAddr(to)
	if err != nil {
		return r, errorf("invalid To IP %q in range %q", to, s)
	}
	r.to = r.to.withoutZone()
	if !r.IsValid() {
		return r, errorf("range %v to %v not valid", r.from, r.to)
	}
	return r, nil
}

// MustParseRange calls ParseRange(s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParseRange(s string) Range {
	r, err := ParseRange(s)
	if err != nil {
		panic(err)
	}
	return r
}

// String returns a string representation of the range.
//
// For a valid range, the form is "From-To" with a single hyphen
// separating the IPs, the same format recognized by
// ParseRange.
func (r Range) String() string {
	if r.IsValid() {
		return sprintf("%s-%s", r.from, r.to)
	}
	if r.from.IsZero() || r.to.IsZero() {
		return "zero Range"
	}
	return "invalid Range"
}

// AppendTo appends a text encoding of r,
// as generated by MarshalText,
// to b and returns the extended buffer.
func (r Range) AppendTo(b []byte) []byte {
	if r.IsZero() {
		return b
	}
	b = r.from.AppendTo(b)
	b = append(b, '-')
	b = r.to.AppendTo(b)
	return b
}

// MarshalText implements the encoding.TextMarshaler interface,
// The encoding is the same as returned by String, with one exception:
// If ip is the zero value, the encoding is the empty string.
func (r Range) MarshalText() ([]byte, error) {
	if r.IsZero() {
		return []byte(""), nil
	}
	var max int
	if r.from.z == z4 {
		max = len("255.255.255.255-255.255.255.255")
	} else {
		max = len("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	}
	b := make([]byte, 0, max)
	return r.AppendTo(b), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The IP range is expected in a form accepted by ParseRange.
// It returns an error if *r is not the Range zero value.
func (r *Range) UnmarshalText(text []byte) error {
	if *r != (Range{}) {
		return errors.New("refusing to Unmarshal into non-zero Range")
	}
	if len(text) == 0 {
		return nil
	}
	var err error
	*r, err = ParseRange(string(text))
	return err
}

// IsZero reports whether r is the zero value of the Range type.
func (r Range) IsZero() bool {
	return r == Range{}
}

// IsValid reports whether r.From() and r.To() are both non-zero and
// obey the documented requirements: address families match, and From
// is less than or equal to To.
func (r Range) IsValid() bool {
	return !r.from.IsZero() &&
		r.from.z == r.to.z &&
		!r.to.Less(r.from)
}

// Valid reports whether r.From() and r.To() are both non-zero and
// obey the documented requirements: address families match, and From
// is less than or equal to To.
//
// Deprecated: use the correctly named and identical IsValid method instead.
func (r Range) Valid() bool { return r.IsValid() }

// Contains reports whether the range r includes addr.
//
// An invalid range always reports false.
//
// If ip has an IPv6 zone, Contains returns false,
// because Prefixes strip zones.
func (r Range) Contains(addr Addr) bool {
	return r.IsValid() && !addr.hasZone() && r.contains(addr)
}

// contains is like Contains, but without the validity check.
// addr must not have a zone.
func (r Range) contains(addr Addr) bool {
	return r.from.Compare(addr) <= 0 && r.to.Compare(addr) >= 0
}

// less reports whether r is "before" other. It is before if r.From()
// is before other.From(). If they're equal, then the larger range
// (higher To()) comes first.
func (r Range) less(other Range) bool {
	if cmp := r.from.Compare(other.from); cmp != 0 {
		return cmp < 0
	}
	return other.to.Less(r.to)
}

// entirelyBefore returns whether r lies entirely before other in IP
// space.
func (r Range) entirelyBefore(other Range) bool {
	return r.to.Less(other.from)
}

// entirelyWithin returns whether r is entirely contained within
// other.
func (r Range) coveredBy(other Range) bool {
	return other.from.lessOrEq(r.from) && r.to.lessOrEq(other.to)
}

// inMiddleOf returns whether r is inside other, but not touching the
// edges of other.
func (r Range) inMiddleOf(other Range) bool {
	return other.from.Less(r.from) && r.to.Less(other.to)
}

// overlapsStartOf returns whether r entirely overlaps the start of
// other, but not all of other.
func (r Range) overlapsStartOf(other Range) bool {
	return r.from.lessOrEq(other.from) && r.to.Less(other.to)
}

// overlapsEndOf returns whether r entirely overlaps the end of
// other, but not all of other.
func (r Range) overlapsEndOf(other Range) bool {
	return other.from.Less(r.from) && other.to.lessOrEq(r.to)
}

// mergeRanges returns the minimum and sorted set of IP ranges that
// cover r.
func mergeRanges(rr []Range) (out []Range, valid bool) {
	// Always return a copy of r, to avoid aliasing slice memory in
	// the caller.
	switch len(rr) {
	case 0:
		return nil, true
	case 1:
		return []Range{rr[0]}, true
	}

	sort.Slice(rr, func(i, j int) bool { return rr[i].less(rr[j]) })
	out = make([]Range, 1, len(rr))
	out[0] = rr[0]
	for _, r := range rr[1:] {
		prev := &out[len(out)-1]
		switch {
		case !r.IsValid():
			// Invalid ranges make no sense to merge, refuse to
			// perform.
			return nil, false
		case prev.to.Next() == r.from:
			// prev and r touch, merge them.
			//
			//   prev     r
			// f------tf-----t
			prev.to = r.to
		case prev.to.Less(r.from):
			// No overlap and not adjacent (per previous case), no
			// merging possible.
			//
			//   prev       r
			// f------t  f-----t
			out = append(out, r)
		case prev.to.Less(r.to):
			// Partial overlap, update prev
			//
			//   prev
			// f------t
			//     f-----t
			//        r
			prev.to = r.to
		default:
			// r entirely contained in prev, nothing to do.
			//
			//    prev
			// f--------t
			//  f-----t
			//     r
		}
	}
	return out, true
}

// Overlaps reports whether p and o overlap at all.
//
// If p and o are of different address families or either are invalid,
// it reports false.
func (r Range) Overlaps(o Range) bool {
	return r.IsValid() &&
		o.IsValid() &&
		r.from.Compare(o.to) <= 0 &&
		o.from.Compare(r.to) <= 0
}

// prefixMaker returns a address-family-corrected Prefix from a and bits,
// where the input bits is always in the IPv6-mapped form for IPv4 addresses.
type prefixMaker func(a uint128, bits uint8) Prefix

// Prefixes returns the set of Prefix entries that covers r.
//
// If either of r's bounds are invalid, in the wrong order, or if
// they're of different address families, then Prefixes returns nil.
//
// Prefixes necessarily allocates. See AppendPrefixes for a version that uses
// memory you provide.
func (r Range) Prefixes() []Prefix {
	return r.AppendPrefixes(nil)
}

// AppendPrefixes is an append version of Range.Prefixes. It appends
// the Prefix entries that cover r to dst.
func (r Range) AppendPrefixes(dst []Prefix) []Prefix {
	if !r.IsValid() {
		return nil
	}
	return appendRangePrefixes(dst, r.prefixFrom128AndBits, r.from.addr, r.to.addr)
}

func (r Range) prefixFrom128AndBits(a uint128, bits uint8) Prefix {
	ip := Addr{addr: a, z: r.from.z}
	if r.from.Is4() {
		bits -= 12 * 8
	}
	return Prefix{ip, bits}
}

// aZeroBSet is whether, after the common bits, a is all zero bits and
// b is all set (one) bits.
func comparePrefixes(a, b uint128) (common uint8, aZeroBSet bool) {
	common = a.commonPrefixLen(b)

	// See whether a and b, after their common shared bits, end
	// in all zero bits or all one bits, respectively.
	if common == 128 {
		return common, true
	}

	m := mask6[common]
	return common, (a.xor(a.and(m)).isZero() &&
		b.or(m) == uint128{^uint64(0), ^uint64(0)})
}

// Prefix returns r as an Prefix, if it can be presented exactly as such.
// If r is not valid or is not exactly equal to one prefix, ok is false.
func (r Range) Prefix() (p Prefix, ok bool) {
	if !r.IsValid() {
		return
	}
	if common, ok := comparePrefixes(r.from.addr, r.to.addr); ok {
		return r.prefixFrom128AndBits(r.from.addr, common), true
	}
	return
}

func appendRangePrefixes(dst []Prefix, makePrefix prefixMaker, a, b uint128) []Prefix {
	common, ok := comparePrefixes(a, b)
	if ok {
		// a to b represents a whole range, like 10.50.0.0/16.
		// (a being 10.50.0.0 and b being 10.50.255.255)
		return append(dst, makePrefix(a, common))
	}
	// Otherwise recursively do both halves.
	dst = appendRangePrefixes(dst, makePrefix, a, a.bitsSetFrom(common+1))
	dst = appendRangePrefixes(dst, makePrefix, b.bitsClearedFrom(common+1), b)
	return dst
}
