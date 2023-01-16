---
title: "Go, Backwards Compatibility, and GODEBUG"
layout: article
---

<!--
This document is kept in the Go repo, not x/website,
because it documents the full list of known GODEBUG settings,
which are tied to a specific release.
-->

## Introduction {#intro}

Go's emphasis on backwards compatibility is one of its key strengths.
There are, however, times when we cannot maintain complete compatibility.
If code depends on buggy (including insecure) behavior,
then fixing the bug will break that code.
Sometimes new features can have similar impacts:
enabling the HTTP/2 use by the HTTP client broke programs
connecting to servers with buggy HTTP/2 implementations.
These kinds of changes are unavoidable and
[permitted by the Go 1 compatibillity rules](/doc/go1compat).
Even so, Go provides a mechanism called GODEBUG to
reduce the impact such changes have on Go developers
using newer toolchains to compile old code.

A GODEBUG setting is a `key=value` pair
that controls the execution of certain parts of a Go program.
The environment variable `$GODEBUG`
(`%GODEBUG%` on Windows)
can hold a comma-separated list of these settings.
For example, if a Go program is running in an environment that contains

	GODEBUG=http2client=0,http2server=0

then that Go program will disable the use of HTTP/2 in both
the HTTP client and the HTTP server.
It is also possible to set the default `GODEBUG` for a given program
(discussed below).

When preparing any change that is permitted by Go 1 compatibility
but may nonetheless break some existing programs,
we first engineer the change to keep as many existing programs working as possible.
For the remaining programs,
we define a new GODEBUG setting that
allows individual programs to opt back in to the old behavior.
A GODEBUG setting may not be added if doing so is infeasible,
but that should be extremely rare.

GODEBUG settings added for compatibility will be maintained
for a minimum of two years (four Go releases).
Some, such as `http2client` and `http2server`,
will be maintained much longer, even indefinitely.

When possible, each GODEBUG setting has an associated
[runtime/metrics](/pkg/runtime/metrics/) counter
named `/godebug/non-default-behavior/<name>:events`
that counts the number of times a particular program's
behavior has changed based on a non-default value
for that setting.
For example, when `GODEBUG=http2client=0` is set,
`/godebug/non-default-behavior/http2client:events`
counts the number of HTTP transports that the program
has configured without HTTP/2 support.

## Default GODEBUG Values {#default}

When a GODEBUG setting is not listed in the environment variable,
its implicit default value is defined by the Go toolchain in use.
The [GODEBUG History](#history) gives the exact defaults for each Go toolchain version.
For example, Go 1.21 introduces the `panicnil` setting,
controlling whether `panic(nil)` is allowed;
it defaults to `panicnil=0`, making `panic(nil)` a run-time error.
Using `panicnil=1` restores the behavior of Go 1.20 and earlier.

Each compiled program can also contain a set of explicit defaults.
These explicit defaults match the Go version listed in the work module
or workspace used when compiling the program.
For example, when a Go 1.21 toolchain compiles a program,
if the work module's `go.mod` or the workspace's `go.work`
says `go` `1.20`, then the program will contain a default `panicnil=0`,
matching Go 1.20 instead of Go 1.21.
Because Go 1.21 introduced this method of setting GODEBUG defaults,
programs listing versions of Go earlier than Go 1.20 are configured only to match Go 1.20,
not the older version.

To override these explicit defaults, a main package's source files
can include one or more `//go:debug` directive at the top of the file
(preceding the `package` statement).
Continuing the `panicnil` example, if the module or workspace is updated
to say `go` `1.21`, the program can opt back into the old `panic(nil)`
behavior by including this directive:

	//go:debug panicnil=1

Starting in Go 1.21, the Go toolchain treats a `//go:debug` directive
with an unrecognized GODEBUG settings as an invalid program.
(Older toolchains ignore `//go:debug` directives entirely.)

The explicit defaults that will be compiled into a main package
are reported by the command:

	go list -f '{{.DefaultGODEBUG}}' my/main/package

When testing a package, `//go:debug` lines in the `*_test.go`
files are treated as directives for the test's main package.
In any other context, `//go:debug` lines are ignored by the toolchain;
`go` `vet` reports such lines as misplaced.

## GODEBUG History {#history}

This section documents the GODEBUG settings introduced and removed in each major Go release.

### Go 1.21

Go 1.21 made it a run-time error to call `panic` with a nil interface value,
controlled by the [`panicnil` setting](/pkg/builtin/#panic).
There is no plan to remove this setting.

### Go 1.20

Go 1.20 introduced support for rejecting insecure paths in tar and zip archives,
controlled by the [`tarinsecurepath` setting](/pkg/archive/tar/#Reader.Next)
and the [`zipinsecurepath` setting](/pkg/archive/zip/#NewReader).
These default to `tarinsecurepath=1` and `zipinsecurepath=1`,
preserving the behavior of earlier versions of Go.
A future version of Go may change the defaults to
`tarinsecurepath=0` and `zipinsecurepath=0`.

Go 1.20 introduced automatic seeding of the
[`math/rand`](/pkg/math/rand) global random number generator,
controlled by the [`randautoseed` setting](/pkg/math/rand/#Seed).

Go 1.20 introduced the concept of fallback roots for use during certificate verification,
controlled by the [`x509usefallbackroots` setting](/pkg/crypto/x509/#SetFallbackRoots).

Go 1.20 removed the preinstalled `.a` files for the standard library
from the Go distribution.
Installations now build and cache the standard library like
packages in other modules.
The [`installgoroot` setting](/cmd/go#hdr-Compile_and_install_packages_and_dependencies)
restores the installation and use of preinstalled `.a` files.

There is no plan to remove any of these settings.

### Go 1.19

Go 1.19 made it an error for path lookups to resolve to binaries in the current directory,
controlled by the [`execerrdot` setting](/pkg/os/exec#hdr-Executables_in_the_current_directory).
There is no plan to remove this setting.

### Go 1.18

Go 1.18 removed support for SHA1 in most X.509 certificates,
controlled by the [`x509sha1` setting](/crypto/x509#InsecureAlgorithmError).
This setting will be removed in a future release, Go 1.22 at the earliest.

### Go 1.6

Go 1.6 introduced transparent support for HTTP/2,
controlled by the [`http2client`, `http2server`, and `http2debug` settings](/pkg/net/http/#hdr-HTTP_2).
There is no plan to remove this setting.

### Go 1.5

Go 1.5 introduced a pure Go DNS resolver,
controlled by the [`netdns` setting](/pkg/net/#hdr-Name_Resolution).
There is no plan to remove this setting.
