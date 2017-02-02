// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package url_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strings"
)

func ExampleValues() {
	v := url.Values{}
	v.Set("name", "Ava")
	v.Add("friend", "Jess")
	v.Add("friend", "Sarah")
	v.Add("friend", "Zoe")
	// v.Encode() == "name=Ava&friend=Jess&friend=Sarah&friend=Zoe"
	fmt.Println(v.Get("name"))
	fmt.Println(v.Get("friend"))
	fmt.Println(v["friend"])
	// Output:
	// Ava
	// Jess
	// [Jess Sarah Zoe]
}

func ExampleURL() {
	u, err := url.Parse("http://bing.com/search?q=dotnet")
	if err != nil {
		log.Fatal(err)
	}
	u.Scheme = "https"
	u.Host = "google.com"
	q := u.Query()
	q.Set("q", "golang")
	u.RawQuery = q.Encode()
	fmt.Println(u)
	// Output: https://google.com/search?q=golang
}

func ExampleURL_roundtrip() {
	// Parse + String preserve the original encoding.
	u, err := url.Parse("https://example.com/foo%2fbar")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u.Path)
	fmt.Println(u.RawPath)
	fmt.Println(u.String())
	// Output:
	// /foo/bar
	// /foo%2fbar
	// https://example.com/foo%2fbar
}

func ExampleURL_ResolveReference() {
	u, err := url.Parse("../../..//search?q=dotnet")
	if err != nil {
		log.Fatal(err)
	}
	base, err := url.Parse("http://example.com/directory/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(base.ResolveReference(u))
	// Output:
	// http://example.com/search?q=dotnet
}

func ExampleParse() {
	examples := []struct {
		url  string
		want *url.URL
	}{{
		// A typical URL comes with at least the Scheme, Host, and Path.
		url:  "https://golang.org/pkg/net/url/?m=all",
		want: &url.URL{Scheme: "https", Host: "golang.org", Path: "/pkg/net/url/", RawQuery: "m=all"},
	}, {
		// A path-only URL may omit the Scheme and Host.
		// This form must not start with a "//" prefix.
		url:  "/pkg/net/url/#URL",
		want: &url.URL{Path: "/pkg/net/url/", Fragment: "URL"},
	}, {
		// Use the "network-path reference" form when including the Host.
		// This form differs from path-only URLs by having a "//" prefix
		// and ensures that the URL is parsed into the Host and Path fields.
		url:  "//golang.org/pkg/net/url/",
		want: &url.URL{Host: "golang.org", Path: "/pkg/net/url/"},
	}, {
		// Warning: The first path segment is not considered the Host.
		// In this case, the entire URL is parsed as the Path.
		// Use the "network-host reference" form shown above instead.
		url:  "golang.org/pkg/net/url/", // Warning: probably incorrect
		want: &url.URL{Path: "golang.org/pkg/net/url/"},
	}, {
		// An optional userinfo can be passed for authentication purposes.
		// The authentication information is stored in the User field.
		url:  "ftps://user:pass@golang.org",
		want: &url.URL{Scheme: "ftps", User: url.UserPassword("user", "pass"), Host: "golang.org"},
	}, {
		// An URL may use the mailto URI scheme specified in RFC 6068.
		// The target email address is stored in the Opaque field.
		url:  "mailto:example@golang.org?subject=hello,%20world!",
		want: &url.URL{Scheme: "mailto", Opaque: "example@golang.org", RawQuery: "subject=hello,%20world!"},
	}, {
		// An URL may use the data URI scheme specified in RFC 2397.
		// It is the user's responsibility to parse the media type and
		// data out of the Opaque field.
		url:  "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP__AA8AA_6P688AAAAASUVORK5CYII=",
		want: &url.URL{Scheme: "data", Opaque: "image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP__AA8AA_6P688AAAAASUVORK5CYII="},
	}}

	for _, ex := range examples {
		got, err := url.Parse(ex.url)
		if err != nil {
			log.Fatal(err)
		}
		if !reflect.DeepEqual(got, ex.want) {
			fmt.Printf("url.Parse(%q):\n\tgot  %#v\n\twant %#v\n", ex.url, got, ex.want)
		} else {
			fmt.Printf("url.Parse(%q):\n\t%s\n", ex.url, formatURL(got))
		}
	}

	// Output:
	// url.Parse("https://golang.org/pkg/net/url/?m=all"):
	// 	&url.URL{Scheme:"https", Host:"golang.org", Path:"/pkg/net/url/", RawQuery:"m=all"}
	// url.Parse("/pkg/net/url/#URL"):
	// 	&url.URL{Path:"/pkg/net/url/", Fragment:"URL"}
	// url.Parse("//golang.org/pkg/net/url/"):
	// 	&url.URL{Host:"golang.org", Path:"/pkg/net/url/"}
	// url.Parse("golang.org/pkg/net/url/"):
	// 	&url.URL{Path:"golang.org/pkg/net/url/"}
	// url.Parse("ftps://user:pass@golang.org"):
	// 	&url.URL{Scheme:"ftps", User:&url.Userinfo{username:"user", password:"pass", passwordSet:true}, Host:"golang.org"}
	// url.Parse("mailto:example@golang.org?subject=hello,%20world!"):
	// 	&url.URL{Scheme:"mailto", Opaque:"example@golang.org", RawQuery:"subject=hello,%20world!"}
	// url.Parse("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP__AA8AA_6P688AAAAASUVORK5CYII="):
	// 	&url.URL{Scheme:"data", Opaque:"image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR4nGJiYGBgAAQAAP__AA8AA_6P688AAAAASUVORK5CYII="}
}

// formatURL formats the URL similar to fmt.Sprintf("%#v", u),
// but ignores zero value fields.
func formatURL(u *url.URL) string {
	var ss []string
	v := reflect.ValueOf(*u)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		z := reflect.Zero(f.Type())
		if !reflect.DeepEqual(f.Interface(), z.Interface()) {
			ss = append(ss, fmt.Sprintf("%s:%#v", v.Type().Field(i).Name, f))
		}
	}
	return fmt.Sprintf("&%T{%s}", *u, strings.Join(ss, ", "))
}

func ExampleParseQuery() {
	m, err := url.ParseQuery(`x=1&y=2&y=3;z`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(toJSON(m))
	// Output:
	// {"x":["1"], "y":["2", "3"], "z":[""]}
}

func toJSON(m interface{}) string {
	js, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(string(js), ",", ", ", -1)
}
