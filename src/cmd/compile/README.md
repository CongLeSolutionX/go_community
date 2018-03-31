cmd/compile contains all of the packages that form the Go compiler. The compiler
may be logically split in two parts: the front-end, and the back-end.

Please keep in mind that this logical split is somewhat arbitrary, and doesn't
always correspond with the layout and hierarchy of packages. Sometimes a
middle-end is also introduced, to separate the parts of the backend that are
architecture-independent from those that aren't.

### Front-end

The main packages involved in this stage are:

* `cmd/compile/internal/syntax` (lexer, parser)
* `cmd/compile/internal/gc` (type-checking, AST transformations)

The front-end begins by parsing the input Go code. If any syntax errors are
found, they are returned. If none are found, it ends up with a Abstract Syntax
Tree, a detailed representation of each Go file. The tree is made of nodes,
which correspond to each component of the file - such as declarations,
statements, and expressions.

The file is then type-checked. The first step is resolving the type of each
expression in the code, and potentially giving type errors such as assigning an
int to a string type.

The front-end also performs a series of extra checks, including "declared and
not used" as well as determining whether or not a function terminates.

Certain transformations are also done on the AST. Some nodes are clarified, such
as string additions being split from the arithmetic addition node type. Other
cases are the copy builtin, which is replaced by memory moves, and append, which
is expanded with simpler components that the rest of the compiler can work with.

Then, having the AST with full type information, the first series of
optimizations are performed. Some good examples are dead code elimination,
function call inlining, and escape analysis. 

Note that the `go/*` family of packages, such as `go/parser` and `go/types`,
have no relation to the compiler. The reason is purely historical. Since the
compiler was initially written in C, the `go/*` packages were developed
to enable writing tools working with Go code, such as `gofmt` and `vet`.

### Back-end

The main packages involved in this stage are:

* `cmd/compile/internal/gc` (converting to SSA)
* `cmd/compile/internal/ssa` (SSA passes and rules)

The back-end begins by converting the AST into Static Single Assignment form, a
much simpler and more restricted language. This has multiple benefits; writing
optimizations is made easier, and the language is much closer to machine code,
since it deals with lower-level ideas such as memory.

Then, a series of machine-independent passes and rules are applied. These,
sometimes called the "middle-end", still do not concern any single computer
architecture, and thus run on all `GOARCH` variants.

Some examples of these generic passes include dead code elimination, removal of
unneeded nil checks, and removal of unused branches. The generic rewrite rules
mainly concern expressions, such as replacing some expressions with constant
values, and optimizing multiplications and float operations.

The machine-dependent part of the backend begins with the "lower" pass, which
rewrites generic values into their machine-specific variants. For example,
TODO: what is a good example here?

TODO: is lower and all the machine-dependent rewrite rules the same thing, or do
they just overlap?

Once the SSA has been "lowered" and is more specific to the target architecture,
the final passes take place. This includes yet another dead code elimination
pass, moving values closer to their uses, the removal of local variables that
are never read from, and register allocation.

The final stage of the back-end is code generation, where the machine-dependent
SSA values are converted to assembly instructions.

TODO: where do intrinsics fit in here? are they one of the SSA passes?

TODO: consider moving the lower/passes/rules details into ssa/README.md instead

To dig deeper into how the SSA package works, including its passes and rules,
head to cmd/compile/internal/ssa/README.md.
