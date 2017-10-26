// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"strings"
	"testing"
)

type dnsNameTest struct {
	name   string
	result bool
}

var dnsNameTests = []dnsNameTest{
	// RFC 2181, section 11.

	// Names with underscores (non-LDH).
	{"_xmpp-server._tcp.google.com", true},

	// LDH names, both matching and not matching the preferred name syntax.
	{"foo.com", true},
	{"1foo.com", true},
	{"26.0.0.73.com", true},
	{"fo-o.com", true},
	{"fo1o.com", true},
	{"foo1.com", true},
	{"a.b..com", false},
	{"a.b-.com", true},
	{"a.b.com-", true},
	{"a.b..", false},
	{"b.com.", true},
	{"example", true},

	// Names with unescaped characters that are special in RFC 1035 section 5.1 master files.
	{"\t.example", false},
	{"\n.example", false},
	{"\r.example", false},
	{" .example", false},
	{"$.example", false},
	{"a$.example", true},
	{"a.$.example", true},
	{"..example", false},
	{"\".example", false},
	{"a\".example", false},
	{"@.example", true},
	{"(.example", false},
	{"a(.example", false},
	{").example", false},
	{"a).example", false},
	{";.example", false},
	{"a;.example", false},

	// Names with other unescaped non-LDH octets.
	{"\x00.example", false},
	{"\x01.example", false},
	{"\x1F.example", false},
	{"#.example", true},
	{"*.example", true},
	{"/*.example", true},
	{"A*.example", true},
	{"a.*.example", true},
	{"*.*.example", true},
	{"~.example", true},
	{"\x7F.example", false},
	{"\x80.example", false},
	{"\xC0.example", false},
	{"\xE0.example", false},
	{"\xF0.example", false},
	{"\xF8.example", false},
	{"\xFC.example", false},
	{"\xFE.example", false},
	{"\xFF.example", false},

	// Names with escaped characters that are special in RFC 1035 section 5.1 master files.
	{"\\\t.example", true},
	{"\\\n.example", false},
	{"\\\r.example", false},
	{`\ .example`, true},
	{`\$.example`, true},
	{`a\$.example`, true},
	{`a.\$.example`, true},
	{`\..example`, true},
	{`\".example`, true},
	{`a\".example`, true},
	{`\@.example`, true},
	{`\(.example`, true},
	{`a\(.example`, true},
	{`\).example`, true},
	{`a\).example`, true},
	{`\;.example`, true},
	{`a\;.example`, true},

	// Names with other escaped non-LDH octets.
	{"\\\x00.example", false},
	{"\\\x01.example", false},
	{"\\\x1F.example", false},
	{"\\#.example", true},
	{"\\*.example", true},
	{"\\/\\*.example", true},
	{"\\A\\*.example", true},
	{"a.\\*.example", true},
	{"\\*.\\*.example", true},
	{"\\~.example", true},
	{"\\\x7F.example", false},
	{"\\\x80.example", false},
	{"\\\xC0.example", false},
	{"\\\xE0.example", false},
	{"\\\xF0.example", false},
	{"\\\xF8.example", false},
	{"\\\xFC.example", false},
	{"\\\xFE.example", false},
	{"\\\xFF.example", false},

	// Names with decimal-escaped characters that are special in RFC 1035 section 5.1 master files.
	{`\009.example`, true},
	{`\010.example`, true},
	{`\013.example`, true},
	{`\032.example`, true},
	{`\036.example`, true},
	{`a\036.example`, true},
	{`a.\036.example`, true},
	{`\046.example`, true},
	{`\034.example`, true},
	{`a\034.example`, true},
	{`\064.example`, true},
	{`\040.example`, true},
	{`a\040.example`, true},
	{`\041.example`, true},
	{`a\041.example`, true},
	{`\059.example`, true},
	{`a\059.example`, true},

	// Names with other decimal-escaped non-LDH octets.
	{`\000.example`, true},
	{`\001.example`, true},
	{`\031.example`, true},
	{`\035.example`, true},
	{`\042.example`, true},
	{`\047\042.example`, true},
	{`\065\042.example`, true},
	{`a.\042.example`, true},
	{`\042.\042.example`, true},
	{`\126.example`, true},
	{`\127.example`, true},
	{`\128.example`, true},
	{`\192.example`, true},
	{`\224.example`, true},
	{`\240.example`, true},
	{`\248.example`, true},
	{`\252.example`, true},
	{`\254.example`, true},
	{`\255.example`, true},
	{`\0255.example`, true},
	{`\0256.example`, true},

	// Names with bad escapes.
	{`example\`, false},
	{`example\0`, false},
	{`example\00`, false},
	{`\0.example`, false},
	{`\00.example`, false},
	{`\0  .example`, false},
	{`\00a.example`, false},
	{`\7up.example`, false},
	{`\256.example`, false},

	// Names testing whole-domain requirements.
	{"@", false},
	{`\@`, true},
	{`\064`, true},
	{`42`, false},
	{`\0522`, true},
	{`0.0`, false},
	{`\048.0`, true},
	{`decimal.0.0`, true},
}

func emitDNSNameTest(ch chan<- dnsNameTest) {
	defer close(ch)
	var char63 = ""
	var escape63 = ""
	for i := 0; i < 63; i++ {
		char63 += "a"
		escape63 += `\128`
	}
	char64 := char63 + "a"
	escape64 := escape63 + "0"
	longDomain := strings.Repeat(char63+".", 4)[2:]
	longestDomain := strings.Repeat(escape63+".", 4)[8:]

	for _, tc := range dnsNameTests {
		ch <- tc
	}

	ch <- dnsNameTest{char63 + ".com", true}
	ch <- dnsNameTest{escape63 + ".com", true}
	ch <- dnsNameTest{char64 + ".com", false}
	ch <- dnsNameTest{escape64 + ".com", false}

	// Remember: wire format is two octets longer than presentation
	// (length octets for the first and [root] last labels).
	// 253 is fine:
	ch <- dnsNameTest{longDomain[:len(longDomain)-1], true}
	// A terminal dot doesn't contribute to length:
	ch <- dnsNameTest{longDomain, true}
	// 254 is bad:
	ch <- dnsNameTest{"a" + longDomain[:len(longDomain)-1], false}
	ch <- dnsNameTest{"a" + longDomain, false}
	ch <- dnsNameTest{"a." + longDomain[1:len(longDomain)-1], false}
	ch <- dnsNameTest{"a." + longDomain[1:], false}

	ch <- dnsNameTest{longestDomain[:len(longestDomain)-1], true}
	ch <- dnsNameTest{longestDomain, true}
	ch <- dnsNameTest{"a" + longestDomain[:len(longestDomain)-1], false}
	ch <- dnsNameTest{"a" + longestDomain, false}
	ch <- dnsNameTest{"a." + longestDomain[4:len(longestDomain)-1], false}
	ch <- dnsNameTest{"a." + longestDomain[4:], false}
}

func TestDNSName(t *testing.T) {
	ch := make(chan dnsNameTest)
	go emitDNSNameTest(ch)
	for tc := range ch {
		if isDomainName(tc.name) != tc.result {
			t.Errorf("isDomainName(%q) = %v; want %v", tc.name, !tc.result, tc.result)
		}
	}
}

func BenchmarkDNSName(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	benchmarks := append(dnsNameTests, []dnsNameTest{
		{strings.Repeat("a", 63), true},
		{strings.Repeat("a", 64), false},
	}...)
	for n := 0; n < b.N; n++ {
		for _, tc := range benchmarks {
			if isDomainName(tc.name) != tc.result {
				b.Errorf("isDomainName(%q) = %v; want %v", tc.name, !tc.result, tc.result)
			}
		}
	}
}
