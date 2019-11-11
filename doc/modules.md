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

A *module* is a collection of packages that are released, versioned, and
distributed together. A module is identified by a *module path*, which is
declared in a [`go.mod` file](#go.mod-files), together with information about
the module's dependencies. The directory that contains the `go.mod` file is
called the *module root directory*. The *main module* is the module defined in
the directory where the `go` command is invoked (or a parent directory).

Each *package* within a module is a collection of source files in the same
directory that are compiled together. A package is identified by an *import
path*, which is determined by joining the module path and the subdirectory
under the module root that contains the package. For example, the module
`"golang.org/x/net"` contains a package in the directory `"html"`. Its import
path is `"golang.org/x/net/html"`.

<a id="versions"></a>
### Versions

Modules are released and distributed using versions. A *version* identifies an
immutable snapshot of a module. Each version starts with the letter `v`,
followed by a semantic version number. See [Semantic Versioning
2.0.0](https://semver.org/spec/v2.0.0.html) for details on how versions are
formatted, interpreted, and compared.

To summarize, a semantic version consists of three non-negative integers (the
major, minor, and patch version numbers, from left to right) separated by
dots. The patch number may be followed by an optional prerelease string starting
with a dash. For example, `v0.0.0`, `v1.12.134`, and `v8.0.5-beta.1` are valid
versions.

Go deviates from the semantic versioning specification in two ways. First, the
letter `v` is a required prefix of every version. Second, build tags are not
allowed except for `+incompatible`, which is defined in [Compatibility with
non-module repositories](#compatibility-with-non-module-repositories).

Each part of the version indicates whether a version is stable and whether it is
compatible with previous versions.

* The *major version number* must be incremented and the minor and patch version
  numbers must be set to zero after a backwards incompatible change is made to
  the module's public interface or documented functionality, for example, after
  a package is removed.
* The *minor version number* must be incremented and the patch version number
  set to zero after a backwards compatible change, for example, after a new
  function is added.
* The *patch version number* must be incremented after a change that does not
  affect the module's public interface, such as a bug fix or optimization.

A version is considered unstable if the major version number is 0 or if a
prerelease suffix is present. Unstable versions are not subject to compatibility
requirements. For example, `v0.2.0` may not be compatible with `v0.1.0`, and
`v1.5.0-beta` may not be compatible with `v1.5.0`.

Go may still access modules in version control systems usings tags, branches, or
revisions that don't follow these conventions. However, within the main module,
the `go` command will automatically convert revision names that don't follow
this standard into canonical versions, possibly resulting in a
pseudo-version. For example, the command `go get -d golang.org/x/net@master`
will convert the branch name `master` into a pseudo-version like
`v0.0.0-20191109021931-daa7c04131f5`.

<a id="major-version-suffixes"></a>
### Major version suffixes

Starting with version `v2.0.0`, module paths must have a *major version suffix*
like `/v2` that matches the major version number. For example, if a module has
the path `example.com/mod` at `v1.0.0`, it must have the path
`example.com/mod/v2` at version `v2.0.0`.

Major version suffixes are used to enforce the [*import compatibility
rule*](https://research.swtch.com/vgo-import):

> If an old package and a new package have the same import path,
> the new package must be backwards compatible with the old package.

By definition, a package in new major version of a module is not backwards
compatible with the corresponding package in the previous major version.
Consequently, starting with `v2.0.0`, packages need new import paths. This is
accomplished by adding a major version suffix to the module path, which is a
prefix of the the import path for each package within the module.

Major version suffixes are not allowed at major versions `v0` or `v1`. There
is no need to change the module path between `v0` and `v1` because `v0` versions
are unstable and have no compatibility guarantee.

As a special case, modules paths starting with `gopkg.in/` must always have a
major version suffix, even at `v0` and `v1`. The suffix must start with a dot
rather than a slash (for example, `gopkg.in/yaml.v2`). Modules starting with
`gopkg.in/` must have exactly two path components. Packages within those modules
may, of course, have more.

Major version suffixes let multiple major versions of a module coexist in the
same build. This may be necessary due to a [diamond dependency
problem](https://research.swtch.com/vgo-import#dependency_story). Ordinarily, if
a module is required at two different versions by transitive dependencies, the
later version will be used. However, if the two versions are incompatible,
neither version will satisfy all clients. Since incompatible versions must have
different major version numbers, they must also have different module paths due
to major version suffixes. This resolves the conflict: modules with distinct
paths are treated as separate modules, even if they are different versions of
the same set of packages.

Many Go projects released versions at `v2.0.0` or later without using a major
version suffix before migrating to modules (perhaps before modules were even
introduced). These versions are annotated with a `+incompatible` build tag (for
example, `v2.0.0+incompatible`). See [Compatibility with non-module
repositories](#compatibility-with-non-module-repositories) for more information.

<a id="resolving-a-package-to-a-module"></a>
### Resolving a package to a module

<a id="go.mod-files"></a>
## `go.mod` files

<a id="go.mod-file-format"></a>
### `go.mod` file format

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
**build list:** List of module versions that will be used for a build command
such as `go build`, `go list`, or `go test`. Determined from the [`go.mod`
file](#glos-go.mod-file) using [minimal version
selection](#minimal-version-selection).

<a id="glos-go.mod-file"></a>
**`go.mod` file:** File that appears in a module's root directory and defines the
module's path, requirements, and other metadata. See
[`go.mod` files](#go.mod-files).

<a id="glos-import-path"></a>
**import path:** A path that identifies a package. An import path is a module
path joined with a subdirectory within the module. For example,
`"golang.org/x/net/html"` is a package in the module `"golang.org/x/net"` in
the `"html"` subdirectory.

<a id="glos-main-module"></a>
**main module:** The module defined in the directory where the `go` command is
invoked (or a parent directory).

<a id="glos-major-version-number"></a>
**major version number:** The first number in a semantic version number (`1` in
`v1.2.3`). In a release with incompatible changes, the major version number
must be incremented, and the minor and patch version numbers must be set to 0.
Versions with major version 0 are considered unstable.

<a id="glos-major-version-suffix"></a>
**major version suffix:** A module path suffix that matches the major version
number. For example, `/v2` in `example.com/mod/v2`. Required at `v2.0.0` or
later, not allowed for earlier versions. See
[Major version suffixes](#major-version-suffixes).

<a id="glos-minimal-version-selection"></a>
**minimal version selection:** Algorithm used to determine the versions of all
modules that will be used in a build. See
[Minimal version selection](#minimal-version-selection) for details.

<a id="glos-minor-version-number"></a>
**minor version number:** The second number in a semantic version number (`2`
in `v1.2.3`). In a release with compatible changes, the minor version number
must be incremented, and the patch version number must be set to 0.

<a id="glos-module"></a>
**module:** A collection of packages that are released, versioned, and
distributed together.

<a id="module-path"></a>
**module path:** A path that identifies a module and acts as a prefix for
package import paths within the module. For example, `"golang.org/x/net"`.

<a id="glos-module-root-directory"></a>
**module root directory:** The directory that contains the `go.mod` file that
defines a module.

<a id="glos-package"></a>
**package:** A collection of source files in the same directory that are
compiled together.

<a id="glos-patch-version-number"></a>
**patch version number:** The third number in a semantic version number (`3` in
`v1.2.3`). In a release with no changes to the module's public interface, the
patch version number must be incremented.

<a id="glos-prerelease-suffix"></a>
**prerelease suffix:** An optional suffix for a semantic version (`-pre` in
`v1.2.3-pre`). Prerelease versions are considered unstable. A prerelease version
sorts before its corresponding non-prerelease version (`v1.2.3-pre` is before
`v1.2.3`).

<a id="glos-pseudo-version"></a>
**pseudo-version:** A version that encodes a revision number and timestamp from
a version control system. For example,
`v0.0.0-20191109021931-daa7c04131f5`. Used for [compatibility with non-module
repositories](#compatibility-with-non-module-repositories) and in other
situations when a released version is not available.

<a id="glos-version"></a>
**version:** An identifier for an immutable snapshot of a module, written as
the letter `v` followed by a semantic version number. See [Versions](#versions).
