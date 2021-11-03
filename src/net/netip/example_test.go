// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip_test

import (
	"fmt"
	"log"
	"net/netip"
)

func ExampleAddrFrom16() {
	b := [16]byte{0xfc, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	fmt.Println(netip.AddrFrom16(b))
	// Output:
	// fc00::
}

func ExampleAddrFrom4() {
	fmt.Println(netip.AddrFrom4([4]byte{8, 8, 8, 8}))
	// Output:
	// 8.8.8.8
}

func ExampleAddrFromSlice() {
	b16 := []byte{0xfc, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	ip, ok := netip.AddrFromSlice(b16)
	fmt.Printf("%q - %t\n", ip, ok)

	b4 := []byte{8, 8, 8, 8}
	ip, ok = netip.AddrFromSlice(b4)
	fmt.Printf("%q - %t\n", ip, ok)

	invalidIP := []byte{8, 8, 8, 8, 8}
	ip, ok = netip.AddrFromSlice(invalidIP)
	fmt.Printf("%q - %t\n", ip, ok)
	// Output:
	// "fc00::" - true
	// "8.8.8.8" - true
	// "invalid IP" - false
}

func ExampleIPv6LinkLocalAllNodes() {
	fmt.Println(netip.IPv6LinkLocalAllNodes())
	// Output:
	// ff02::1
}

func ExampleMustParseAddr() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("2001:0000:0000:0000:0000:0000:0000:1234")

	fmt.Println(ipv4)
	fmt.Println(ipv6)
	// Output:
	// 8.8.8.8
	// 2001::1234
}

func ExampleParseAddr() {
	ip, err := netip.ParseAddr("8.8.8.8")
	fmt.Printf("IP: %q, Error: %v\n", ip, err)

	ip, err = netip.ParseAddr("2001:0000:0000:0000:0000:0000:0000:1234")
	fmt.Printf("IP: %q, Error: %v\n", ip, err)

	ip, err = netip.ParseAddr("fe80::1cc0:3e8c:119f:c2e1%ens18")
	fmt.Printf("IP: %q, Error: %v\n", ip, err)

	ip, err = netip.ParseAddr("8.8.meow.8")
	fmt.Printf("IP: %q, Error: %v\n", ip, err)
	// Output:
	// IP: "8.8.8.8", Error: <nil>
	// IP: "2001::1234", Error: <nil>
	// IP: "fe80::1cc0:3e8c:119f:c2e1%ens18", Error: <nil>
	// IP: "invalid IP", Error: ParseAddr("8.8.meow.8"): unexpected character (at "meow.8")
}

func ExampleAddr_AppendTo() {
	b := []byte("http://")
	ip := netip.MustParseAddr("8.8.8.8")
	fmt.Println(string(ip.AppendTo(b)))
	// Output:
	// http://8.8.8.8
}

func ExampleAddr_As16() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4.As16())
	fmt.Println(ipv6.As16())
	// Output:
	// [0 0 0 0 0 0 0 0 0 0 255 255 8 8 8 8]
	// [0 6 0 0 0 0 0 0 0 0 0 0 0 0 0 4]
}

func ExampleAddr_As4() {
	ip := netip.MustParseAddr("8.8.8.8")
	fmt.Println(ip.As4())
	// Output:
	// [8 8 8 8]
}

func ExampleAddr_BitLen() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4.BitLen())
	fmt.Println(ipv6.BitLen())
	// Output:
	// 32
	// 128
}

func ExampleAddr_Compare() {
	zeroIP4 := netip.MustParseAddr("0.0.0.0")
	zeroIP6 := netip.MustParseAddr("::")
	dnsIP4 := netip.MustParseAddr("8.8.8.8")

	fmt.Println(zeroIP4.Compare(zeroIP4))
	fmt.Println(zeroIP4.Compare(zeroIP6))
	fmt.Println(dnsIP4.Compare(zeroIP4))
	// Output:
	// 0
	// -1
	// 1
}

func ExampleAddr_IPAddrParts() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0004")
	ipv6WithZone := netip.MustParseAddr("fe80::1cc0:3e8c:119f:c2e1%ens18")

	for _, ip := range []netip.Addr{ipv4, ipv6, ipv6WithZone} {
		b, z := ip.IPAddrParts()
		fmt.Printf("slice: %v, zone: %q\n", b, z)
	}
	// Output:
	// slice: [8 8 8 8], zone: ""
	// slice: [0 6 0 0 0 0 0 0 0 0 0 0 0 0 0 4], zone: ""
	// slice: [254 128 0 0 0 0 0 0 28 192 62 140 17 159 194 225], zone: "ens18"
}

func ExampleAddr_Is4() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4.Is4())
	fmt.Println(ipv6.Is4())
	// Output:
	// true
	// false
}

func ExampleAddr_Is4In6() {
	ipv4DNS := netip.MustParseAddr("8.8.8.8")
	ipv6DNS := netip.MustParseAddr("0:0:0:0:0:FFFF:0808:0808")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4DNS.Is4In6())
	fmt.Println(ipv6DNS.Is4In6())
	fmt.Println(ipv6.Is4In6())
	// Output:
	// false
	// true
	// false
}

func ExampleAddr_Is6() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4.Is6())
	fmt.Println(ipv6.Is6())
	// Output:
	// false
	// true
}

func ExampleAddr_IsGlobalUnicast() {
	ipv6Global := netip.MustParseAddr("2000::")
	ipv6UniqLocal := netip.MustParseAddr("2000::")
	ipv6Multi := netip.MustParseAddr("FF00::")

	ipv4Private := netip.MustParseAddr("10.255.0.0")
	ipv4Public := netip.MustParseAddr("8.8.8.8")
	ipv4Broadcast := netip.MustParseAddr("255.255.255.255")

	fmt.Println(ipv6Global.IsGlobalUnicast())
	fmt.Println(ipv6UniqLocal.IsGlobalUnicast())
	fmt.Println(ipv6Multi.IsGlobalUnicast())

	fmt.Println(ipv4Private.IsGlobalUnicast())
	fmt.Println(ipv4Public.IsGlobalUnicast())
	fmt.Println(ipv4Broadcast.IsGlobalUnicast())
	// Output:
	// true
	// true
	// false
	// true
	// true
	// false
}

func ExampleAddr_IsInterfaceLocalMulticast() {
	ipv6InterfaceLocalMulti := netip.MustParseAddr("ff01::1")
	ipv6Global := netip.MustParseAddr("2000::")
	ipv4 := netip.MustParseAddr("255.0.0.0")

	fmt.Println(ipv6InterfaceLocalMulti.IsInterfaceLocalMulticast())
	fmt.Println(ipv6Global.IsInterfaceLocalMulticast())
	fmt.Println(ipv4.IsInterfaceLocalMulticast())
	// Output:
	// true
	// false
	// false
}

func ExampleAddr_IsLinkLocalMulticast() {
	ipv6LinkLocalMulti := netip.MustParseAddr("ff02::2")
	ipv6LinkLocalUni := netip.MustParseAddr("fe80::")
	ipv4LinkLocalMulti := netip.MustParseAddr("224.0.0.0")
	ipv4LinkLocalUni := netip.MustParseAddr("169.254.0.0")

	fmt.Println(ipv6LinkLocalMulti.IsLinkLocalMulticast())
	fmt.Println(ipv6LinkLocalUni.IsLinkLocalMulticast())
	fmt.Println(ipv4LinkLocalMulti.IsLinkLocalMulticast())
	fmt.Println(ipv4LinkLocalUni.IsLinkLocalMulticast())
	// Output:
	// true
	// false
	// true
	// false
}

func ExampleAddr_IsLinkLocalUnicast() {
	ipv6LinkLocalUni := netip.MustParseAddr("fe80::")
	ipv6Global := netip.MustParseAddr("2000::")
	ipv4LinkLocalUni := netip.MustParseAddr("169.254.0.0")
	ipv4LinkLocalMulti := netip.MustParseAddr("224.0.0.0")

	fmt.Println(ipv6LinkLocalUni.IsLinkLocalUnicast())
	fmt.Println(ipv6Global.IsLinkLocalUnicast())
	fmt.Println(ipv4LinkLocalUni.IsLinkLocalUnicast())
	fmt.Println(ipv4LinkLocalMulti.IsLinkLocalUnicast())
	// Output:
	// true
	// false
	// true
	// false
}

func ExampleAddr_IsLoopback() {
	ipv6Lo := netip.MustParseAddr("::1")
	ipv6 := netip.MustParseAddr("ff02::1")
	ipv4Lo := netip.MustParseAddr("127.0.0.0")
	ipv4 := netip.MustParseAddr("128.0.0.0")

	fmt.Println(ipv6Lo.IsLoopback())
	fmt.Println(ipv6.IsLoopback())
	fmt.Println(ipv4Lo.IsLoopback())
	fmt.Println(ipv4.IsLoopback())
	// Output:
	// true
	// false
	// true
	// false
}
func ExampleAddr_IsMulticast() {
	ipv6Multi := netip.MustParseAddr("FF00::")
	ipv6LinkLocalMulti := netip.MustParseAddr("ff02::1")
	ipv6Lo := netip.MustParseAddr("::1")
	ipv4Multi := netip.MustParseAddr("239.0.0.0")
	ipv4LinkLocalMulti := netip.MustParseAddr("224.0.0.0")
	ipv4Lo := netip.MustParseAddr("127.0.0.0")

	fmt.Println(ipv6Multi.IsMulticast())
	fmt.Println(ipv6LinkLocalMulti.IsMulticast())
	fmt.Println(ipv6Lo.IsMulticast())
	fmt.Println(ipv4Multi.IsMulticast())
	fmt.Println(ipv4LinkLocalMulti.IsMulticast())
	fmt.Println(ipv4Lo.IsMulticast())
	// Output:
	// true
	// true
	// false
	// true
	// true
	// false
}

func ExampleAddr_IsPrivate() {
	ipv6Private := netip.MustParseAddr("fc00::")
	ipv6Public := netip.MustParseAddr("fe00::")
	ipv4Private := netip.MustParseAddr("10.255.0.0")
	ipv4Public := netip.MustParseAddr("11.0.0.0")

	fmt.Println(ipv6Private.IsPrivate())
	fmt.Println(ipv6Public.IsPrivate())
	fmt.Println(ipv4Private.IsPrivate())
	fmt.Println(ipv4Public.IsPrivate())
	// Output:
	// true
	// false
	// true
	// false
}

func ExampleAddr_IsUnspecified() {
	ipv6Unspecified := netip.MustParseAddr("::")
	ipv6Specified := netip.MustParseAddr("fe00::")
	ipv4Unspecified := netip.MustParseAddr("0.0.0.0")
	ipv4Specified := netip.MustParseAddr("8.8.8.8")

	fmt.Println(ipv6Unspecified.IsUnspecified())
	fmt.Println(ipv6Specified.IsUnspecified())
	fmt.Println(ipv4Unspecified.IsUnspecified())
	fmt.Println(ipv4Specified.IsUnspecified())
	// Output:
	// true
	// false
	// true
	// false
}

func ExampleAddr_IsValid() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv4Zero := netip.MustParseAddr("0.0.0.0")
	ipv6 := netip.MustParseAddr("0006::0004")
	ipv6Zero := netip.MustParseAddr("::")
	invalidIP, _ := netip.ParseAddr("meow")

	fmt.Println(ipv4.IsValid())
	fmt.Println(ipv4Zero.IsValid())
	fmt.Println(ipv6.IsValid())
	fmt.Println(ipv6Zero.IsValid())
	fmt.Println(invalidIP.IsValid())
	// Output:
	// true
	// true
	// true
	// true
	// false
}

func ExampleAddr_Less() {
	ipv4Smaller := netip.MustParseAddr("8.8.8.8")
	ipv4Larger := netip.MustParseAddr("8.8.8.9")
	ipv6 := netip.MustParseAddr("0006::0004")
	ipv6WithZone := netip.MustParseAddr("0006::0004%ens18")

	fmt.Println(ipv4Smaller.Less(ipv4Larger))
	fmt.Println(ipv4Smaller.Less(ipv6))
	fmt.Println(ipv6WithZone.Less(ipv6))
	// Output:
	// true
	// true
	// false
}

func ExampleAddr_Next() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0001")

	fmt.Println(ipv4.Next())
	fmt.Println(ipv6.Next())
	// Output:
	// 8.8.8.9
	// 6::2
}

func ExampleAddr_Prefix() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	for _, b := range []int{0, 8, 16, 24, 32} {
		p, err := ipv4.Prefix(b)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("IP: %q, Bits: %d, Prefix: %q\n", ipv4, b, p)
	}

	ipv6 := netip.MustParseAddr("1111::2222:3333:4444:5555%ens18")
	for _, b := range []int{0, 16, 32, 48, 64, 80, 96, 112, 128} {
		p, err := ipv6.Prefix(b)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("IP: %q, Bits: %d, Prefix: %q\n", ipv6, b, p)
	}
	// Output:
	// IP: "8.8.8.8", Bits: 0, Prefix: "0.0.0.0/0"
	// IP: "8.8.8.8", Bits: 8, Prefix: "8.0.0.0/8"
	// IP: "8.8.8.8", Bits: 16, Prefix: "8.8.0.0/16"
	// IP: "8.8.8.8", Bits: 24, Prefix: "8.8.8.0/24"
	// IP: "8.8.8.8", Bits: 32, Prefix: "8.8.8.8/32"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 0, Prefix: "::/0"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 16, Prefix: "1111::/16"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 32, Prefix: "1111::/32"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 48, Prefix: "1111::/48"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 64, Prefix: "1111::/64"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 80, Prefix: "1111::2222:0:0:0/80"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 96, Prefix: "1111::2222:3333:0:0/96"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 112, Prefix: "1111::2222:3333:4444:0/112"
	// IP: "1111::2222:3333:4444:5555%ens18", Bits: 128, Prefix: "1111::2222:3333:4444:5555/128"
}

func ExampleAddr_Prev() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0002")

	fmt.Println(ipv4.Prev())
	fmt.Println(ipv6.Prev())
	// Output:
	// 8.8.8.7
	// 6::1
}

func ExampleAddr_String() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0002")

	fmt.Println(ipv4.String())
	fmt.Println(ipv6.String())
	// Output:
	// 8.8.8.8
	// 6::2
}

func ExampleAddr_StringExpanded() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0002")

	fmt.Println(ipv4.StringExpanded())
	fmt.Println(ipv6.StringExpanded())
	// Output:
	// 8.8.8.8
	// 0006:0000:0000:0000:0000:0000:0000:0002
}

func ExampleAddr_Unmap() {
	ipv4DNS := netip.MustParseAddr("8.8.8.8")
	ipv6DNS := netip.MustParseAddr("0:0:0:0:0:FFFF:0808:0808")
	ipv6 := netip.MustParseAddr("0006::0004")

	fmt.Println(ipv4DNS.Unmap())
	fmt.Println(ipv6DNS.Unmap())
	fmt.Println(ipv6.Unmap())
	// Output:
	// 8.8.8.8
	// 8.8.8.8
	// 6::4
}

func ExampleAddr_WithZone() {
	ipv4 := netip.MustParseAddr("8.8.8.8")
	ipv6 := netip.MustParseAddr("0006::0002")
	ipv6WithZone := netip.MustParseAddr("0006::0004%ens18")

	fmt.Println(ipv4.WithZone("a3"))
	fmt.Println(ipv6.WithZone("a3"))
	fmt.Println(ipv6WithZone.WithZone("a3"))
	fmt.Println(ipv6WithZone.WithZone(""))
	// Output:
	// 8.8.8.8
	// 6::2%a3
	// 6::4%a3
	// 6::4
}

func ExampleAddr_Zone() {
	ipv6WithZone := netip.MustParseAddr("0006::0004%ens18")
	ipv6 := netip.MustParseAddr("0006::0002")
	ipv4 := netip.MustParseAddr("8.8.8.8")

	fmt.Printf("IP: %v, Zone: %q\n", ipv6WithZone, ipv6WithZone.Zone())
	fmt.Printf("IP: %v, Zone: %q\n", ipv6, ipv6.Zone())
	fmt.Printf("IP: %v, Zone: %q\n", ipv4, ipv4.Zone())
	// Output:
	// IP: 6::4%ens18, Zone: "ens18"
	// IP: 6::2, Zone: ""
	// IP: 8.8.8.8, Zone: ""
}
