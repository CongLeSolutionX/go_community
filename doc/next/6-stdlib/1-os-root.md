### Directory-limited filesystem access

<!-- go.dev/issue/67002 -->
The new [Root] type provides the ability to perform filesystem
operations within a specific directory.

The [OpenRoot] function opens a directory and returns a [Root].
Methods on [Root] operate within the directory and do not permit
paths that refer to locations outside the directory, including
ones that follow symbolic links out of the directory.

- [Root.Open] opens a file for reading.
- [Root.Create] creates a file.
- [Root.OpenFile] is the generalized open call.
- [Root.Mkdir] creates a directory.

