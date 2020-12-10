// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package build

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strconv"
	"unicode/utf8"
)

type importReader struct {
	b    *bufio.Reader
	buf  []byte
	peek byte
	err  error
	eof  bool
	nerr int
}

func isIdent(c byte) bool {
	return 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '_' || c >= utf8.RuneSelf
}

var (
	errSyntax = errors.New("syntax error")
	errNUL    = errors.New("unexpected NUL in input")
)

// syntaxError records a syntax error, but only if an I/O error has not already been recorded.
func (r *importReader) syntaxError() {
	if r.err == nil {
		r.err = errSyntax
	}
}

// readByte reads the next byte from the input, saves it in buf, and returns it.
// If an error occurs, readByte records the error in r.err and returns 0.
func (r *importReader) readByte() byte {
	c, err := r.b.ReadByte()
	if err == nil {
		r.buf = append(r.buf, c)
		if c == 0 {
			err = errNUL
		}
	}
	if err != nil {
		if err == io.EOF {
			r.eof = true
		} else if r.err == nil {
			r.err = err
		}
		c = 0
	}
	return c
}

// readByteNoBuf is like readByte but doesn't buffer the byte.
// It exhausts r.buf before reading from r.b.
func (r *importReader) readByteNoBuf() byte {
	if len(r.buf) > 0 {
		c := r.buf[0]
		r.buf = r.buf[1:]
		return c
	}
	c, err := r.b.ReadByte()
	if err == nil && c == 0 {
		err = errNUL
	}
	if err != nil {
		if err == io.EOF {
			r.eof = true
		} else if r.err == nil {
			r.err = err
		}
		c = 0
	}
	return c
}

// peekByte returns the next byte from the input reader but does not advance beyond it.
// If skipSpace is set, peekByte skips leading spaces and comments.
func (r *importReader) peekByte(skipSpace bool) byte {
	if r.err != nil {
		if r.nerr++; r.nerr > 10000 {
			panic("go/build: import reader looping")
		}
		return 0
	}

	// Use r.peek as first input byte.
	// Don't just return r.peek here: it might have been left by peekByte(false)
	// and this might be peekByte(true).
	c := r.peek
	if c == 0 {
		c = r.readByte()
	}
	for r.err == nil && !r.eof {
		if skipSpace {
			// For the purposes of this reader, semicolons are never necessary to
			// understand the input and are treated as spaces.
			switch c {
			case ' ', '\f', '\t', '\r', '\n', ';':
				c = r.readByte()
				continue

			case '/':
				c = r.readByte()
				if c == '/' {
					for c != '\n' && r.err == nil && !r.eof {
						c = r.readByte()
					}
				} else if c == '*' {
					var c1 byte
					for (c != '*' || c1 != '/') && r.err == nil {
						if r.eof {
							r.syntaxError()
						}
						c, c1 = c1, r.readByte()
					}
				} else {
					r.syntaxError()
				}
				c = r.readByte()
				continue
			}
		}
		break
	}
	r.peek = c
	return r.peek
}

// nextByte is like peekByte but advances beyond the returned byte.
func (r *importReader) nextByte(skipSpace bool) byte {
	c := r.peekByte(skipSpace)
	r.peek = 0
	return c
}

// readKeyword reads the given keyword from the input.
// If the keyword is not present, readKeyword records a syntax error.
func (r *importReader) readKeyword(kw string) {
	r.peekByte(true)
	for i := 0; i < len(kw); i++ {
		if r.nextByte(false) != kw[i] {
			r.syntaxError()
			return
		}
	}
	if isIdent(r.peekByte(false)) {
		r.syntaxError()
	}
}

// readIdent reads an identifier from the input.
// If an identifier is not present, readIdent records a syntax error.
func (r *importReader) readIdent() {
	c := r.peekByte(true)
	if !isIdent(c) {
		r.syntaxError()
		return
	}
	for isIdent(r.peekByte(false)) {
		r.peek = 0
	}
}

// readString reads a quoted string literal from the input.
// If an identifier is not present, readString records a syntax error.
func (r *importReader) readString() {
	switch r.nextByte(true) {
	case '`':
		for r.err == nil {
			if r.nextByte(false) == '`' {
				break
			}
			if r.eof {
				r.syntaxError()
			}
		}
	case '"':
		for r.err == nil {
			c := r.nextByte(false)
			if c == '"' {
				break
			}
			if r.eof || c == '\n' {
				r.syntaxError()
			}
			if c == '\\' {
				r.nextByte(false)
			}
		}
	default:
		r.syntaxError()
	}
}

// readImport reads an import clause - optional identifier followed by quoted string -
// from the input.
func (r *importReader) readImport() {
	c := r.peekByte(true)
	if c == '.' {
		r.peek = 0
	} else if isIdent(c) {
		r.readIdent()
	}
	r.readString()
}

// readComments is like io.ReadAll, except that it only reads the leading
// block of comments in the file.
func readComments(f io.Reader) ([]byte, error) {
	r := &importReader{b: bufio.NewReader(f)}
	r.peekByte(true)
	if r.err == nil && !r.eof {
		// Didn't reach EOF, so must have found a non-space byte. Remove it.
		r.buf = r.buf[:len(r.buf)-1]
	}
	return r.buf, r.err
}

// readGoInfo expects a Go file as input and reads the file up to and including the import section.
// It records what it learned in *info.
// If info.fset is non-nil, readGoInfo parses the file and sets info.parsed, info.parseErr,
// info.imports, info.embeds, and info.embedErr.
//
// It only returns an error if there are problems reading the file,
// not for syntax errors in the file itself.
func readGoInfo(f io.Reader, info *fileInfo) error {
	r := &importReader{b: bufio.NewReader(f)}

	r.readKeyword("package")
	r.readIdent()
	for r.peekByte(true) == 'i' {
		r.readKeyword("import")
		if r.peekByte(true) == '(' {
			r.nextByte(false)
			for r.peekByte(true) != ')' && r.err == nil {
				r.readImport()
			}
			r.nextByte(false)
		} else {
			r.readImport()
		}
	}

	info.header = r.buf

	// If we stopped successfully before EOF, we read a byte that told us we were done.
	// Return all but that last byte, which would cause a syntax error if we let it through.
	if r.err == nil && !r.eof {
		info.header = r.buf[:len(r.buf)-1]
	}

	// If we stopped for a syntax error, consume the whole file so that
	// we are sure we don't change the errors that go/parser returns.
	if r.err == errSyntax {
		r.err = nil
		for r.err == nil && !r.eof {
			r.readByte()
		}
		info.header = r.buf
	}
	if r.err != nil {
		return r.err
	}

	if info.fset == nil {
		return nil
	}

	// Parse file header & record imports.
	info.parsed, info.parseErr = parser.ParseFile(info.fset, info.name, info.header, parser.ImportsOnly|parser.ParseComments)
	if info.parseErr != nil {
		return nil
	}

	var embedPkgNames map[string]bool
	for _, decl := range info.parsed.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, dspec := range d.Specs {
			spec, ok := dspec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			quoted := spec.Path.Value
			path, err := strconv.Unquote(quoted)
			if err != nil {
				return fmt.Errorf("parser returned invalid quoted string: <%s>", quoted)
			}
			if path == "embed" {
				if embedPkgNames == nil {
					embedPkgNames = make(map[string]bool)
				}
				pkgName := "embed"
				if spec.Name != nil {
					pkgName = spec.Name.Name
				}
				embedPkgNames[pkgName] = true
			}

			doc := spec.Doc
			if doc == nil && len(d.Specs) == 1 {
				doc = d.Doc
			}
			info.imports = append(info.imports, fileImport{path, spec.Pos(), doc})
		}
	}

	// If the file imports "embed",
	// we have to look for //go:embed comments
	// in the remainder of the file.
	// The compiler will enforce the mapping of comments to
	// declared variables. We just need to know the patterns.
	// If there were //go:embed comments earlier in the file
	// (near the package statement or imports), the compiler
	// will reject them. They can be (and have already been) ignored.
	if len(embedPkgNames) != 0 {
		for r.err == nil && !r.eof {
			r.readByte()
		}
		info.header = r.buf

		info.parsed, info.parseErr = parser.ParseFile(info.fset, info.name, info.header, parser.ParseComments)
		if info.parseErr != nil {
			return nil
		}

		v := embedVisitor{
			pkgNames:  embedPkgNames,
			dotImport: embedPkgNames["."],
			info:      info,
		}
		ast.Walk(&v, info.parsed)
	}

	return nil
}

type embedVisitor struct {
	info      *fileInfo
	pkgNames  map[string]bool
	decls     []string
	dotImport bool
}

func (v *embedVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.CallExpr:
		v.visitCall(n)

	case *ast.ValueSpec:
		for _, name := range n.Names {
			v.decl(name)
		}
	case *ast.TypeSpec:
		v.decl(n.Name)

	case *ast.FuncDecl:
		return v.push(n.Recv, n.Type.Params, n.Type.Results)
	case *ast.FuncLit:
		return v.push(n.Type.Params, n.Type.Results)

	case *ast.SwitchStmt:
		v := v.push()
		if id, _ := n.Tag.(*ast.Ident); id != nil {
			v.decl(id)
		}
		return v

	case *ast.BlockStmt, *ast.IfStmt, *ast.ForStmt, *ast.CaseClause, *ast.CommClause:
		return v.push()
	}
	return v
}

func (v *embedVisitor) push(lists ...*ast.FieldList) *embedVisitor {
	c := *v

	for _, list := range lists {
		if list == nil {
			continue
		}
		for _, field := range list.List {
			for _, name := range field.Names {
				c.decl(name)
			}
		}
	}

	return &c
}

func (v *embedVisitor) decl(id *ast.Ident) {
	if v.pkgNames[id.Name] || v.dotImport && isEmbedDeclName(id) {
		v.decls = append(v.decls, id.Name)
	}
}

func (v *embedVisitor) isShadowed(id *ast.Ident) bool {
	for _, decl := range v.decls {
		if decl == id.Name {
			return true
		}
	}
	return false
}

func (v *embedVisitor) visitCall(call *ast.CallExpr) {
	switch n := call.Fun.(type) {
	case *ast.Ident:
		if !v.dotImport {
			return
		}
		if !isEmbedDeclName(n) || v.isShadowed(n) {
			return
		}
	case *ast.SelectorExpr:
		pkg, ok := n.X.(*ast.Ident)
		if !ok || !v.pkgNames[pkg.Name] {
			return
		}
		if !isEmbedDeclName(n.Sel) || v.isShadowed(pkg) {
			return
		}
	default:
		return
	}

	for _, arg := range call.Args {
		arg, ok := arg.(*ast.BasicLit)
		if !ok || arg.Kind != token.STRING {
			continue
		}
		if arg, err := strconv.Unquote(arg.Value); err == nil {
			v.info.embeds = append(v.info.embeds, arg)
		}
	}
}

func isEmbedDeclName(id *ast.Ident) bool {
	switch id.Name {
	case "Bytes", "String", "Files":
		return true
	}
	return false
}
