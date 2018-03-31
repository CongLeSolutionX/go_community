cmd/compile contains the main packages that form the Go compiler. The compiler
may be logically split in four steps, which we will briefly describe alongisde
the list of packages that contain their code.

You may sometimes hear the terms "front-end" and "back-end" when referring to
the compiler. Roughly speaking, these translate to the first two and last two
steps we are going to list here. A third term, "middle-end", often refers to
much of the work that happens in the second step.

Note that the `go/*` family of packages, such as `go/parser` and `go/types`,
have no relation to the compiler. The reason is mostly historical. Since the
compiler was initially written in C, the `go/*` packages were developed to
enable writing tools working with Go code, such as `gofmt` and `vet`.

### 1. Parsing

* `cmd/compile/internal/syntax` (lexer, parser, AST)

This step tokenizes and parses the input code. If any syntax errors are found,
they are returned. Otherwise, it builds an Abstract Syntax Tree, a detailed
representation of each Go file.

The tree is made of nodes, which correspond to each component of the file - such
as declarations, statements, and expressions. The AST includes position
information, which is used by later stages of the compiler.

### 2. Type-checking and AST transformations

* `cmd/compile/internal/gc` (old AST, type checking, AST transformations)

The gc package includes an AST definition carried over from when it was written
in C. All of its code is written in terms of it, so the first thing that the gc
package must do is convert the syntax package's AST to its own representation.
This extra step will likely be refactored away in the future.

The AST is then type-checked. The first step is resolving the type of each
expression in the code, and potentially giving type errors such as assigning an
int to a string type. Type-checking includes certain extra checks, such as
"declared and not used" as well as determining whether or not a function
terminates.

Certain transformations are also done on the AST. Some nodes are refined based
on type information, such as string additions being split from the arithmetic
addition node type. Some other examples are dead code elimination, function call
inlining, and escape analysis.

Finally, certain nodes are lowered into simpler components that the rest of the
compiler can work with. For instance, the copy builtin is replaced by memory
moves, and the append builtin is expanded into code that performs the opeartion.

It should be clarified that the name "gc" stands for "Go compiler", and has
little to do with uppercase GC, which stands for garbage collection.

### 3. Generic SSA

* `cmd/compile/internal/gc` (converting to SSA)
* `cmd/compile/internal/ssa` (SSA passes and rules)

This step begins by converting the AST into Static Single Assignment form, a
much simpler and more restricted language. This has multiple benefits; writing
optimizations is made easier, and the language is much closer to machine code,
since it deals with lower-level ideas such as memory.

During this conversion, function intrinsics are applied. These are special
functions that the compiler has been taught to replace with heavily optimized
code on a case-by-case basis.

Then, a series of machine-independent passes and rules are applied. These do not
concern any single computer architecture, and thus run on all `GOARCH` variants.

Some examples of these generic passes include dead code elimination, removal of
unneeded nil checks, and removal of unused branches. The generic rewrite rules
mainly concern expressions, such as replacing some expressions with constant
values, and optimizing multiplications and float operations.

### 4. Exporting machine code

* `cmd/compile/internal/ssa` (SSA lowering and arch-specific passes)
* `cmd/internal/obj` (machine code generation)

The machine-dependent part of the compiler begins with the "lower" pass, which
rewrites generic values into their machine-specific variants. For example, on
amd64 memory operands are possible, so many load-store operations may be combined.

Note that the lower step runs all machine-specific rewrite rules, and thus it
currently applies lots of optimizations too.

Once the SSA has been "lowered" and is more specific to the target architecture,
the final passes take place. This includes yet another dead code elimination
pass, moving values closer to their uses, the removal of local variables that
are never read from, and register allocation.

Other important pieces of work done as part of this step include stack frame
layout, which assigns stack offsets to local variables, and pointer liveness
analysis, which computes which on-stack pointers are live at each GC safe point.

At the end of SSA, Go functions have been transformed into a series of obj.Prog
instructions. These are passed to the assembler (cmd/internal/obj), which turns
them into machine code and writes out the final object file. The object file
will also contain reflect data, export data, and debugging information.

### Further reading

To dig deeper into how the SSA package works, including its passes and rules,
head to cmd/compile/internal/ssa/README.md.
