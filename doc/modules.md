<!--{
  "Title": "Go Modules Reference",
  "Subtitle": "Version of Sep 4, 2019",
  "Path": "/ref/modules"
}-->
<!-- TODO(jayconrod): ensure golang.org/x/website can render Markdown or convert
this document to HTML before Go 1.14. -->
<!-- TODO(jayconrod): ensure anchors work correctly after Markdown rendering -->

<a id="introduction"></a>
## Introduction

<a id="modules-packages-and-versions"></a>
## Modules, packages, and versions

<a id="versions"></a>
### Versions

<a id="major-version-suffixes"></a>
### Major version suffixes

<a id="resolving-a-package-to-a-module"></a>
### Resolving a package to a module

<a id="go.mod-files"></a>
## go.mod files

<a id="go.mod-file-format"></a>
### go.mod file format

<a id="minimal-version-selection"></a>
### Minimal version selection (MVS)

<a id="compatibility-with-non-module-repositories"></a>
### Compatibility with non-module repositories

<a id="module-aware-build-commands"></a>
## Module-aware build commands

<a id="enabling-modules"></a>
### Enabling modules

<a id="initializing-modules"></a>
### Initializing modules

<a id="build-commands"></a>
### Build commands

<a id="vendoring"></a>
### Vendoring

<a id="go-mod-download"></a>
### `go mod download`

<a id="go-mod-verify"></a>
### `go mod verify`

<a id="go-mod-edit"></a>
### `go mod edit`

<a id="go-clean-modcache"></a>
### `go clean -modcache`

<a id="module-commands-outside-a-module"></a>
### Module commands outside a module

<a id="retrieving-modules"></a>
## Retrieving modules

<a id="goproxy-protocol"></a>
### GOPROXY protocol

<a id="communicating-with-proxies"></a>
### Communicating with proxies

<a id="communicating-with-repositories"></a>
### Communicating with repositories

<a id="custom-import-paths"></a>
### Custom import paths

<!-- TODO(jayconrod): custom import paths, details of direct mode -->

<a id="module-zip-requirements"></a>
### Module zip requirements

<a id="communicating-privacy"></a>
### Privacy

<a id="private-modules"></a>
### Private modules

<a id="authenticating-modules"></a>
## Authenticating modules

<a id="go.sum-file-format"></a>
### go.sum file format

<a id="checksum-database"></a>
### Checksum database

<a id="authenticating-privacy"></a>
### Privacy

<a id="environment-variables"></a>
## Environment variables

<a id="glossary">
## Glossary

<a id="glos-build-list"></a>
**build list** - List of module versions that will be used for a build
command. Determined from the [go.mod file](#glos-go.mod-file) using
[minimal version selection](#minimal-version-selection).

<a id="glos-go.mod-file"></a>
**go.mod file** - File that appears in a module's root directory and defines the
module's path, requirements, and other metadata. See
[go.mod files](#go.mod-files).

<a id="glos-import-path"></a>
**import path** - A path that identifies a package. An import path is a module
path joined with a subdirectory within the module. For example,
`"golang.org/x/net/html"` is a package in the `"html"` directory in module
`"golang.org/x/net"`.

<a id="glos-main-module"></a>
**main module** - The module defined in the directory where the go command is
invoked (or a parent directory).

<a id="glos-major-version-number"></a>
**major version number** - The first number in a semantic version number (`1` in
`v1.2.3`). Must be incremented in a release with incompatible changes. Versions
with major version 0 are considered unstable.

<a id="glos-major-version-suffix"></a>
**major version suffix** - A module path suffix that matches the major version
number. For example, `/v2` in `example.com/mod/v2`. Required at `v2.0.0` or
later, not allowed for earlier versions. See
[Major version suffixes](#major-version-suffixes).

<a id="glos-minimal-version-selection"></a>
**minimal version selection** - Algorithm used to determine the versions of all
modules that will be used in a build. See
[Minimal version selection](#minimal-version-selection) for details.

<a id="glos-minor-version-number"></a>
**minor version number** - The second number in a semantic version number (`2`
in `v1.2.3`). Must be incremented in a release with compatible changes.

<a id="glos-module"></a>
**module** - A collection of packages that are released, versioned, and
distributed together.

<a id="module-path"></a>
**module path** - A path that identifies a module and acts as a prefix for
package import paths within the module. For example, `"golang.org/x/net"`.

<a id="glos-module-root-directory"></a>
**module root directory** - The directory that contains the go.mod file that
defines a module.

<a id="glos-package"></a>
**package** - A collection of source files in the same directory that are
compiled together.

<a id="glos-patch-version-number"></a>
**patch version number** - The third number in a semantic version number (`3` in
`v1.2.3`). Must be incremented in a release with no changes to the public
interface of a module.

<a id="glos-prerelease-suffix"></a>
**prerelease suffix** - An optional suffix for a semantic version (`-pre` in
`v1.2.3-pre`). Prerelease versions are considered unstable. A prerelease version
sorts before its corresponding non-prerelease version (`v1.2.3-pre` is before
`v1.2.3`).

<a id="glos-version"></a>
**version** - An identifier for an immutable snapshot of a module, written as
the letter `v` followed by a semantic version number. See [Versions](#versions).
