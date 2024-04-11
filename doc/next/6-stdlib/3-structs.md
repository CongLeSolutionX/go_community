### New structs package


The new [structs](/pkg/structs) package provides
types for struct fields that modify properties of
the containing struct type (for example, properties
like field memory order or intended struct copy-ability).

In this release, the only such type is `HostLayout`
which indicates that a structure with a field of that
type has a layout that conforms to host platform
expectations.