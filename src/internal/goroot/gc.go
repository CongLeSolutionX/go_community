// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gc

package goroot

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var std = strings.Fields(`
archive/tar
archive/zip
bufio
bytes
compress/bzip2
compress/flate
compress/gzip
compress/lzw
compress/zlib
container/heap
container/list
container/ring
context
crypto
crypto/aes
crypto/cipher
crypto/des
crypto/dsa
crypto/ecdsa
crypto/ed25519
crypto/ed25519/internal/edwards25519
crypto/ed25519/internal/edwards25519/field
crypto/elliptic
crypto/elliptic/internal/fiat
crypto/elliptic/internal/nistec
crypto/hmac
crypto/internal/boring
crypto/internal/boring/bbig
crypto/internal/boring/sig
crypto/internal/randutil
crypto/internal/subtle
crypto/md5
crypto/rand
crypto/rc4
crypto/rsa
crypto/sha1
crypto/sha256
crypto/sha512
crypto/subtle
crypto/tls
crypto/x509
crypto/x509/internal/macos
crypto/x509/pkix
database/sql
database/sql/driver
debug/buildinfo
debug/dwarf
debug/elf
debug/gosym
debug/macho
debug/pe
debug/plan9obj
embed
embed/internal/embedtest
encoding
encoding/ascii85
encoding/asn1
encoding/base32
encoding/base64
encoding/binary
encoding/csv
encoding/gob
encoding/hex
encoding/json
encoding/pem
encoding/xml
errors
expvar
flag
fmt
go/ast
go/build
go/build/constraint
go/constant
go/doc
go/doc/comment
go/format
go/importer
go/internal/gccgoimporter
go/internal/gcimporter
go/internal/srcimporter
go/internal/typeparams
go/parser
go/printer
go/scanner
go/token
go/types
hash
hash/adler32
hash/crc32
hash/crc64
hash/fnv
hash/maphash
html
html/template
image
image/color
image/color/palette
image/draw
image/gif
image/internal/imageutil
image/jpeg
image/png
index/suffixarray
internal/abi
internal/buildcfg
internal/buildinternal
internal/bytealg
internal/cfg
internal/cpu
internal/diff
internal/execabs
internal/fmtsort
internal/fuzz
internal/goarch
internal/godebug
internal/goexperiment
internal/goos
internal/goroot
internal/goversion
internal/intern
internal/itoa
internal/lazyregexp
internal/lazytemplate
internal/nettrace
internal/obscuretestdata
internal/oserror
internal/pkgbits
internal/poll
internal/profile
internal/race
internal/reflectlite
internal/singleflight
internal/syscall/execenv
internal/syscall/unix
internal/sysinfo
internal/testenv
internal/testlog
internal/trace
internal/txtar
internal/unsafeheader
internal/xcoff
io
io/fs
io/ioutil
log
log/syslog
math
math/big
math/bits
math/cmplx
math/rand
mime
mime/multipart
mime/quotedprintable
net
net/http
net/http/cgi
net/http/cookiejar
net/http/fcgi
net/http/httptest
net/http/httptrace
net/http/httputil
net/http/internal
net/http/internal/ascii
net/http/internal/testcert
net/http/pprof
net/internal/socktest
net/mail
net/netip
net/rpc
net/rpc/jsonrpc
net/smtp
net/textproto
net/url
os
os/exec
os/exec/internal/fdtest
os/signal
os/signal/internal/pty
os/user
path
path/filepath
plugin
reflect
reflect/internal/example1
reflect/internal/example2
regexp
regexp/syntax
runtime
runtime/cgo
runtime/debug
runtime/internal/atomic
runtime/internal/math
runtime/internal/sys
runtime/metrics
runtime/pprof
runtime/race
runtime/trace
sort
strconv
strings
sync
sync/atomic
syscall
testing
testing/fstest
testing/internal/testdeps
testing/iotest
testing/quick
text/scanner
text/tabwriter
text/template
text/template/parse
time
time/tzdata
unicode
unicode/utf16
unicode/utf8
unsafe
vendor/golang.org/x/crypto/chacha20
vendor/golang.org/x/crypto/chacha20poly1305
vendor/golang.org/x/crypto/cryptobyte
vendor/golang.org/x/crypto/cryptobyte/asn1
vendor/golang.org/x/crypto/curve25519
vendor/golang.org/x/crypto/curve25519/internal/field
vendor/golang.org/x/crypto/hkdf
vendor/golang.org/x/crypto/internal/poly1305
vendor/golang.org/x/crypto/internal/subtle
vendor/golang.org/x/net/dns/dnsmessage
vendor/golang.org/x/net/http/httpguts
vendor/golang.org/x/net/http/httpproxy
vendor/golang.org/x/net/http2/hpack
vendor/golang.org/x/net/idna
vendor/golang.org/x/net/nettest
vendor/golang.org/x/net/route
vendor/golang.org/x/sys/cpu
vendor/golang.org/x/text/secure/bidirule
vendor/golang.org/x/text/transform
vendor/golang.org/x/text/unicode/bidi
vendor/golang.org/x/text/unicode/norm
`)

var stdMap = func() map[string]bool {
	m := map[string]bool{}
	for _, k := range std {
		m[k] = true
	}
	return m
}()

// IsStandardPackage reports whether path is a standard package,
// given goroot and compiler.
func IsStandardPackage(goroot, compiler, path string) bool {
	return stdMap[path] || strings.HasPrefix(path, "cmd/")
	switch compiler {
	case "gc":
		dir := filepath.Join(goroot, "src", path)
		_, err := os.Stat(dir)
		return err == nil
	case "gccgo":
		return gccgoSearch.isStandard(path)
	default:
		panic("unknown compiler " + compiler)
	}
}

// gccgoSearch holds the gccgo search directories.
type gccgoDirs struct {
	once sync.Once
	dirs []string
}

// gccgoSearch is used to check whether a gccgo package exists in the
// standard library.
var gccgoSearch gccgoDirs

// init finds the gccgo search directories. If this fails it leaves dirs == nil.
func (gd *gccgoDirs) init() {
	gccgo := os.Getenv("GCCGO")
	if gccgo == "" {
		gccgo = "gccgo"
	}
	bin, err := exec.LookPath(gccgo)
	if err != nil {
		return
	}

	allDirs, err := exec.Command(bin, "-print-search-dirs").Output()
	if err != nil {
		return
	}
	versionB, err := exec.Command(bin, "-dumpversion").Output()
	if err != nil {
		return
	}
	version := strings.TrimSpace(string(versionB))
	machineB, err := exec.Command(bin, "-dumpmachine").Output()
	if err != nil {
		return
	}
	machine := strings.TrimSpace(string(machineB))

	dirsEntries := strings.Split(string(allDirs), "\n")
	const prefix = "libraries: ="
	var dirs []string
	for _, dirEntry := range dirsEntries {
		if strings.HasPrefix(dirEntry, prefix) {
			dirs = filepath.SplitList(strings.TrimPrefix(dirEntry, prefix))
			break
		}
	}
	if len(dirs) == 0 {
		return
	}

	var lastDirs []string
	for _, dir := range dirs {
		goDir := filepath.Join(dir, "go", version)
		if fi, err := os.Stat(goDir); err == nil && fi.IsDir() {
			gd.dirs = append(gd.dirs, goDir)
			goDir = filepath.Join(goDir, machine)
			if fi, err = os.Stat(goDir); err == nil && fi.IsDir() {
				gd.dirs = append(gd.dirs, goDir)
			}
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			lastDirs = append(lastDirs, dir)
		}
	}
	gd.dirs = append(gd.dirs, lastDirs...)
}

// isStandard reports whether path is a standard library for gccgo.
func (gd *gccgoDirs) isStandard(path string) bool {
	// Quick check: if the first path component has a '.', it's not
	// in the standard library. This skips most GOPATH directories.
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	if strings.Contains(path[:i], ".") {
		return false
	}

	if path == "unsafe" {
		// Special case.
		return true
	}

	gd.once.Do(gd.init)
	if gd.dirs == nil {
		// We couldn't find the gccgo search directories.
		// Best guess, since the first component did not contain
		// '.', is that this is a standard library package.
		return true
	}

	for _, dir := range gd.dirs {
		full := filepath.Join(dir, path) + ".gox"
		if fi, err := os.Stat(full); err == nil && !fi.IsDir() {
			return true
		}
	}

	return false
}
