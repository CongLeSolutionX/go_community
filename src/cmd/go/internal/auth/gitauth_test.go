package auth

import (
	"net/url"
	"reflect"
	"testing"
)

type want struct {
	data     string
	prefix   *url.URL
	username string
	password string
}

func buildURL(protocol, host, path string) *url.URL {
	prefix := new(url.URL)
	prefix.Scheme = protocol
	prefix.Host = host
	prefix.Path = path
	return prefix
}

func TestParseGitAuth(t *testing.T) {
	wants := []want{
		{ // Standard case.
			data: `
protocol=https
host=example.com
username=bob
password=secr3t
`,
			prefix:   buildURL("https", "example.com", ""),
			username: "bob",
			password: "secr3t",
		},
		{ // Should not use an invalid url.
			data: `
protocol=https
host=example.com
username=bob
password=secr3t
url=invalid
`,
			prefix:   buildURL("https", "example.com", ""),
			username: "bob",
			password: "secr3t",
		},
		{ // Should use the new url.
			data: `
protocol=https
host=example.com
username=bob
password=secr3t
url=https://go.dev
`,
			prefix:   buildURL("https", "go.dev", ""),
			username: "bob",
			password: "secr3t",
		},
		{ // Empty data.
			data: `
`,
			prefix:   buildURL("", "", ""),
			username: "",
			password: "",
		},
		{ // Does not follow format.
			data: `
protocol:https
host:example.com
username:bob
password:secr3t
`,
			prefix:   buildURL("", "", ""),
			username: "",
			password: "",
		},
	}
	for _, want := range wants {
		gotPrefix, gotUsername, gotPassword := parseGitAuth([]byte(want.data))
		if !reflect.DeepEqual(gotPrefix, want.prefix) {
			t.Errorf("parseGitAuth:\nhave %q\nwant %q", gotPrefix, want.prefix)
		}
		if gotUsername != want.username {
			t.Errorf("parseGitAuth:\nhave %q\nwant %q", gotUsername, want.username)
		}
		if gotPassword != want.password {
			t.Errorf("parseGitAuth:\nhave %q\nwant %q", gotPassword, want.password)
		}
	}
}
