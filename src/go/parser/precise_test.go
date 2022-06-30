package parser_test

// The file defines tests of precise comment parsing.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

// TestPreciseComments ensures that the parser records sufficient
// information in the tree to enable linearization of all tokens and
// comments in the correct order. It does this by inserting numbered
// comments between every pair of tokens and asserting that the
// pretty-printer retains them all.
//
// The testcases were derived from coverage of the parser.
// Inputs with a "!" suffix are subject only to a weaker check that
// all comments are retained but the set and order of other tokens is
// ignored. This is required for cases where the syntax tree does
// not preserve sufficient information about the original tokens
// to prevent simplifications such as this:
//
//	for ; cond; {}    =>    for cond {}
func TestPreciseComments(t *testing.T) {
	// TODO: try a variant in which all permissible ';' are replaced by '\n'.
	var inputs = []string{
		"package p",
		"package p;",
		"package p\nvar x int",
		`package p; import ()`,
		`package p; import "fmt"`,
		`package p; import fmt "fmt"`,
		`package p; import . "fmt"`,
		`package p; import ("fmt")`,
		`package p; import (fmt "fmt")`,
		`package p; import ("fmt"; )`,
		`package p; import ("fmt"; . "fmt")`,
		`package p; import (); import ()`,
		`package p; var ()`,
		`package p; var x int`,
		`package p; var x int = i`,
		`package p; var x = 1`,
		`package p; var x, y int`,
		`package p; var x, y int = 1, 2`,
		`package p; var x, y = 1, 2`,
		`package p; var (x = 1; y = 2)`,
		"package p; var (x = 1\n y = 2)",
		"package p; var (); var ()",
		`package p; const ()`,
		`package p; const x int = 1`,
		`package p; const x = 1`,
		`package p; const x, y int = 1, 2`,
		`package p; const (x = 1; y = 2)`,
		`package p; type ()`,
		`package p; type _ chan int`,
		`package p; type _ chan<- int`,
		`package p; type _ <-chan int`,
		`package p; type _ = <-chan int`,
		`package p; type () `,
		`package p; type (A int)`,
		`package p; type (A int; B int)`,
		`package p; type (); type ()`,
		`package p; func _() {}`,
		`package p; func _() {}; func _() {}`,
		`package p; func _() { for ; ; {} }`,
		`package p; func _() { for ; ; post {} }`,
		`package p; func _() { for ; cond ; {} }` + "!", // !: simplified to "for cond"
		`package p; func _() { for ; cond ; post {} }`,
		`package p; func _() { for ; ; {} }`,
		`package p; func _() { for init ; ; post {} }`,
		`package p; func _() { for init ; cond ; {} }`,
		`package p; func _() { for init ; cond ; post {} }`,
		`package p; func _() { for init ; cond ;  {} }`,
		`package p; func _() { for {} }`,
		`package p; func _() { for cond {} }`,
		`package p; func _() { for range z {} }`,
		`package p; func _() { for x = range z {} }`,
		`package p; func _() { for x := range z {} }`,
		`package p; func _() { for x, y = range z {} }`,
		`package p; func _() { for x, y := range z {} }`,
		`package p; func _()`,
		`package p; func _() ()`,
		`package p; func _() () {}`,
		`package p; func _() (R) {}`,
		`package p; func _() (r R) {}`,
		`package p; func _() (r1, r2 R) {}`,
		`package p; func _(...int) {}`,
		`package p; func _(x...int) {}`,
		`package p; func _(x, ...int) {}`,
		`package p; func _(A) {}`,
		`package p; func _(a A) {}`,
		`package p; func _(a1, a2 A) {}`,
		`package p; func _[T any]() {}`,
		`package p; func _[K, V any]() {}`,
		`package p; func _[K any, V any]() {}`,
		`package p; func (T[K, V]) _() {}`,
		`package p; func (r T[K, V]) _() {}`,
		`package p; func _() { var x int }`,
		`package p; func _() { x := 1 }`,
		`package p; func _() { x, y := 1, 2 }`,
		`package p; func _() { x = 1 }`,
		`package p; func _() { x, y = f() }`,
		`package p; func _() { loop: x = 1 }`,
		`package p; func _() { continue }`,
		`package p; func _() { break }`,
		`package p; func _() { continue loop }`,
		`package p; func _() { break loop }`,
		`package p; func _() { goto label }`,
		`package p; func _() { f() }`,
		`package p; func _() { f(a, b,) }`,
		`package p; func _() { f(a, b, c...) }`,
		`package p; func _() { ch <- x }`,
		`package p; func _() { x++ }`,
		`package p; func _() { defer f() }`,
		`package p; func _() { go f() }`,
		`package p; func _() { {} }`,
		`package p; func _() { {}; }`,
		`package p; func _() { ; }`,
		`package p; func _() { ;; }`,
		`package p; func _() { return }`,
		`package p; func _() { return 1 }`,
		`package p; func _() { return 2, 3 }`,
		`package p; func _() { if cond {} }`,
		`package p; func _() { if cond {} else {} }`,
		`package p; func _() { if cond {} else if cond2 {} }`,
		`package p; func _() { if x := 1; cond {} else if cond2 {} }`,
		`package p; func _() { f(); x = 1; g() }`,
		`package p; func _() { switch {} }`,
		`package p; func _() { switch x {} }`,
		`package p; func _() { switch f(); x {} }`,
		`package p; func _() { switch {case 1: } }`,
		`package p; func _() { switch {case 1: f() } }`,
		`package p; func _() { switch {case 1: case 2: } }`,
		`package p; func _() { switch {case 1: case 2: f() } }`,
		`package p; func _() { switch {case 1, 2: } }`,
		`package p; func _() { switch {case 1, 2: f() } }`,
		`package p; func _() { switch {default: } }`,
		`package p; func _() { switch {default: f() } }`,
		`package p; func _() { switch {default: L: ; } }` + "!", // !: simplified to {L:}
		`package p; func _() { switch {default: L: } }`,
		`package p; func _() { switch x.(type) {} }`,
		`package p; func _() { switch y := x.(type) {} }`,
		`package p; func _() { switch y := x.(type) { case T: } }`,
		`package p; func _() { switch y := x.(type) { case T: f() } }`,
		`package p; func _() { switch y := x.(type) { default: } }`,
		`package p; func _() { switch y := x.(type) { default: f() } }`,
		`package p; func _() { select {} }`,
		`package p; func _() { select { default: } }`,
		`package p; func _() { select { default: f() } }`,
		`package p; func _() { select { case <-ch: } }`,
		`package p; func _() { select { case <-ch: f() } }`,
		`package p; func _() { select { case x := <-ch: } }`,
		`package p; func _() { select { case x := <-ch: f() } }`,
		`package p; func _() { select { case ch <- x: } }`,
		`package p; func _() { select { case ch <- x: f() } }`,
		`package p; func _() { func() {} () }`,
		`package p; func _() { (func() {}) () }`,
		`package p; func _() { (func())(nil)() }`,
		`package p; func _() { L: }`,
		`package p; var _ = 0, 0.0, false, nil, "", '0'`,
		`package p; var _ = func() {}`,
		`package p; var _ = func(T) U {}`,
		`package p; var _ = func(T) (U) {}`,
		`package p; var _ = func(T) (U, V) {}`,
		`package p; var _ = (1)`,
		`package p; var _ = fmt.Println`,
		`package p; var _ = T{}`,
		`package p; var _ = T{{}}`,
		`package p; var _ = &T{}`,
		`package p; var _ = &T{{}}`,
		`package p; var _ = &T{U{}}`,
		`package p; var _ = &T{1}`,
		`package p; var _ = &T{1, 2, 3}`,
		`package p; var _ = &T{a:1, b:2, c:3}`,
		`package p; var _ = x[i]`,
		`package p; var _ = x[i:j]`,
		`package p; var _ = x[:j]`,
		`package p; var _ = x[:]`,
		`package p; var _ = x[i:j:k]`,
		`package p; var _ = x[:j:k]`,
		`package p; var _ = x[K, V]`,
		`package p; var _ = x[K]`,
		`package p; var _ = x.(T)`,
		`package p; var _ = x.(T)`,
		`package p; var _ = *x`,
		`package p; var _ = -x`,
		`package p; var _ = x+y`,
		`package p; var _ chan x`,
		`package p; var _ <- chan x`,
		`package p; type _ = []T`,
		`package p; type _ = [2]T`,
		`package p; type _ []T`,
		`package p; type _ [2]T`,
		`package p; type _[K any] T`,
		`package p; var _ = [...]T{}`,
		`package p; var _ = T{1,}` + "!", // !: simplified to "T{1}"
		`package p; type _ = *T`,
		`package p; type _ = (*T)`,
		`package p; type _ = struct{}`,
		`package p; type _ = struct{int}`,
		`package p; type _ = struct{x int}`,
		`package p; type _ = struct{x, y int}`,
		`package p; type _ = struct{x, y p.Q}`,
		`package p; type _ = struct{x []E}`,
		`package p; type _ = struct{x [P]E}`,
		`package p; type _ = struct{x}`,
		`package p; type _ = struct{pkg.T}`,
		`package p; type _ = struct{pkg.T[K]}`,
		`package p; type _ = struct{*T}`,
		`package p; type _ = struct{*T ""}`,
		`package p; type _ = struct{*pkg.T[K]}`,
		`package p; type _ = interface{}`,
		`package p; type _ = interface{I}`,
		`package p; type _ = interface{I[K]}`,
		`package p; type _ = interface{I[K,V]}`,
		`package p; type _ = interface{p.Q[K,V]}`,
		`package p; type _ = interface{I;}` + "!",      // !: simplified to {I}
		`package p; type _ = interface{(I[K]);}` + "!", // !: simplfied to {(I[K])}
		`package p; type _ = interface{I;J}`,
		`package p; type _ = interface{f()}`,
		`package p; type _ = interface{f(); g()}`,
		`package p; type _ = map[K]V`,
		`package p; type _ = func(T) U`,
		`package p; type _ = pkg.Type`,
		`package p; type _ = pkg.Type[K]`,
		`package p; type _ = pkg.Type[K,]`,
		`package p; type _ = pkg.Type[K, V]`,
		`package p; type _ = pkg.Type[K, V,]`,
		`package p; type _[T any] = x`,
		`package p; type _[T] int`,
		`package p; type P[T ~x] int`,
		`package p; type P[T ~x|y|*p.Q] int`,
	}
	for _, input := range inputs {
		// A "!" suffix disables the token check.
		skipTokenCheck := false
		if rest := strings.TrimSuffix(input, "!"); rest != input {
			skipTokenCheck = true
			input = rest
		}

		// Tokenize, with interleaved /*«%d»*/ comments.
		tokens := tokenize(input, true)

		// Parse the commentized result.
		fset := token.NewFileSet()
		commentized := strings.Join(tokens, "")
		file, err := parser.ParseFile(fset, "", commentized, parser.PreciseComments|parser.ParseComments)
		if err != nil {
			t.Errorf("Failed to parse commentized file: %v:\ninput = %s\ncommentized = %s", err, input, commentized)
			continue
		}

		// Pretty-print the commentized tree and
		// check that it contains all numbered comments.
		var pr printer
		pr.node(file)
		printed := pr.buf.String()
		var missing []int
		ncomments := len(tokens)/2 + 1
		for i := 0; i < ncomments; i++ {
			if !strings.Contains(printed, fmt.Sprintf("/*«%d»*/", i)) {
				missing = append(missing, i)
			}
		}
		if missing != nil {
			// Failure of this assertion means either the parser
			// or printer is missing a comment attachment point.
			t.Errorf("Missing comments!\n"+
				"1: input: %s\n\t\n"+
				"2: commentized tokens of #1:\n\t%s\n"+
				"3: PreciseComments of parse of #2:\n%s"+
				"4: print of parse of #2:\n\t%s\n"+
				"5: missing comments: %v",
				input, commentized, showPreciseComments(file), printed, missing)
			continue
		}

		// Now tokenize the printer output and verify the order
		// of all tokens and comments.
		if skipTokenCheck {
			continue
		}
		tokens2 := tokenize(printed, false)
		if !reflect.DeepEqual(tokens, tokens2) {
			// Failure of this assertion means that a token
			// is missing, or a comment has migrated across tokens.
			// Compare #5 and #6 for details.
			t.Errorf("Token mismatch!\n"+
				"1: input:\n\t%s\n"+
				"2: commentized tokens of #1:\n\t%s\n"+
				"3: PreciseComments of parse of #2:\n%s"+
				"4: print of parse of #2:\n\t%s\n"+
				"5: tokens of #2:\n\t%v\n"+
				"6: tokens of parse of #4:\n\t%v",
				input, commentized, showPreciseComments(file), printed, tokens, tokens2)
		}
	}
}

// showPreciseComments shows for each node beneath n, its inventory of
// precise comments and their attachment points.
func showPreciseComments(n ast.Node) string {
	var buf strings.Builder
	depth := 0
	ast.Inspect(n, func(n ast.Node) bool {
		if n != nil {
			var v any
			if f := reflect.ValueOf(n).Elem().FieldByName("PreciseComments"); f.IsValid() {
				v = f.Interface()
			}
			fmt.Fprintf(&buf, "\t%*s%T: %v\n", depth, "", n, v)
			depth++
		} else {
			depth--
		}
		return true
	})
	return buf.String()
}

// tokenize returns the text of all tokens scanned from the input,
// optionally separated by numbered comments to exercise all possible
// comment attachment points.
func tokenize(input string, interleaveComments bool) []string {
	// Tokenize the input.
	var scan scanner.Scanner
	fset := token.NewFileSet()
	tokfile := fset.AddFile("", fset.Base(), len(input))
	scan.Init(tokfile, []byte(input), nil /* no error handler */, scanner.ScanComments)

	// Optionally insert numbered comments between every pair of tokens.
	var tokens []string
	addComment := func() {
		if interleaveComments {
			tokens = append(tokens, fmt.Sprintf("/*«%d»*/", len(tokens)/2))
		}
	}
	for {
		_, tok, lit := scan.Scan()
		if tok == token.EOF {
			break
		}
		addComment()
		if lit == "" {
			lit = tok.String()
		}
		if tok == token.SEMICOLON {
			lit = ";" // normalize implicit semicolons
		}
		tokens = append(tokens, lit)
	}
	addComment()
	return tokens
}

// printer is a simple AST printer that preserves comments and token
// order but is otherwise unfussy about formatting.
// It requires an error-free syntax tree.
//
// Observe that it does not need to use token positions to correctly
// linearize the tree (though it would need to look at positions to
// know whether to emit SEMICOLON tokes as ';' or '\n' based on the
// input, for example).
type printer struct {
	buf         strings.Builder
	semi        bool
	switchblock bool // for BlockStmt beneath switch
	nofunc      bool // suppress 'func' token in FuncType
}

func (pr *printer) exprs(exprs []ast.Expr) {
	for i, expr := range exprs {
		if i > 0 {
			pr.token(token.COMMA, 0)
		}
		pr.node(expr)
	}
}

func (pr *printer) stmts(stmts []ast.Stmt) {
	for i, stmt := range stmts {
		if i > 0 {
			pr.semicolon()
		}
		pr.node(stmt)
	}
}

func (pr *printer) idents(idents []*ast.Ident) {
	for i, id := range idents {
		if i > 0 {
			pr.token(token.COMMA, 0)
		}
		pr.node(id)
	}
}

func (pr *printer) fields(n *ast.FieldList, bracket, sep token.Token) {
	nofunc := pr.nofunc

	pr.comments(n.PreciseComments, ast.PointStart)
	if n.Opening != 0 {
		pr.token(bracket, n.Opening)
	}
	for i, field := range n.List {
		if i > 0 {
			pr.token(sep, 0) // comma or semicolon
		}
		pr.idents(field.Names)
		if field.Type != nil {
			pr.nofunc = nofunc
			pr.node(field.Type)
		}
		if field.Tag != nil {
			pr.node(field.Tag)
		}
	}
	pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
	if n.Closing != 0 {
		pr.token(bracket-token.LPAREN+token.RPAREN, n.Closing)
	}
	pr.comments(n.PreciseComments, ast.PointEnd)

	pr.nofunc = false
}

func (pr *printer) node(n ast.Node) {
	switch n := n.(type) {
	case *ast.Field, *ast.FieldList:
		panic("call printer.fields instead")

	case *ast.Ident:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.string(n.Name, n.NamePos)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.BasicLit:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.string(n.Value, n.ValuePos)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.Ellipsis:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.ELLIPSIS, n.Ellipsis)
		if n.Elt != nil {
			pr.node(n.Elt)
		}
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.FuncLit:
		pr.node(n.Type)
		pr.node(n.Body)

	case *ast.CompositeLit:
		pr.comments(n.PreciseComments, ast.PointStart)
		if n.Type != nil {
			pr.node(n.Type)
		}
		pr.token(token.LBRACE, n.Lbrace)
		pr.exprs(n.Elts)
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RBRACE, n.Rbrace)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.ParenExpr:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.LPAREN, n.Lparen)
		pr.node(n.X)
		pr.token(token.RPAREN, n.Rparen)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.SelectorExpr:
		pr.node(n.X)
		pr.token(token.PERIOD, 0)
		pr.node(n.Sel)

	case *ast.IndexExpr:
		pr.node(n.X)
		pr.token(token.LBRACK, n.Lbrack)
		pr.node(n.Index)
		if hasComment(n.PreciseComments, ast.PointBeforeCloseBracket) {
			pr.token(token.COMMA, 0) // see Go issue #55007
			pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		}
		pr.token(token.RBRACK, n.Rbrack)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.IndexListExpr:
		pr.node(n.X)
		pr.token(token.LBRACK, n.Lbrack)
		pr.exprs(n.Indices)
		if len(n.Indices) > 0 && hasComment(n.PreciseComments, ast.PointBeforeCloseBracket) {
			pr.token(token.COMMA, 0)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RBRACK, n.Rbrack)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.SliceExpr:
		pr.node(n.X)
		if n.Lbrack != 0 {
			pr.token(token.LBRACK, n.Lbrack)
		}
		if n.Low != nil {
			pr.node(n.Low)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeColon)
		pr.token(token.COLON, 0)
		pr.comments(n.PreciseComments, ast.PointAfterColon)
		if n.High != nil {
			pr.node(n.High)
		}
		if n.Slice3 {
			pr.token(token.COLON, 0)
			if n.Max != nil {
				pr.node(n.Max)
			}
		}
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RBRACK, n.Rbrack)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.TypeAssertExpr:
		pr.node(n.X)
		pr.token(token.PERIOD, 0)
		pr.comments(n.PreciseComments, ast.PointBeforeOpenBracket)
		pr.token(token.LPAREN, n.Lparen)
		pr.comments(n.PreciseComments, ast.PointAfterOpenBracket)
		if n.Type != nil {
			pr.node(n.Type)
		} else {
			pr.token(token.TYPE, 0)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RPAREN, n.Rparen)
		pr.comments(n.PreciseComments, ast.PointAfterCloseBracket)

	case *ast.CallExpr:
		pr.node(n.Fun)
		pr.token(token.LPAREN, n.Lparen)
		pr.exprs(n.Args)
		if n.Ellipsis != 0 {
			pr.token(token.ELLIPSIS, n.Ellipsis)
		} else if len(n.Args) > 0 && hasComment(n.PreciseComments, ast.PointBeforeCloseBracket) {
			pr.token(token.COMMA, 0)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RPAREN, n.Rparen)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.StarExpr:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.MUL, n.Star)
		pr.node(n.X)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.UnaryExpr:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(n.Op, n.OpPos)
		pr.node(n.X)

	case *ast.BinaryExpr:
		pr.node(n.X)
		pr.token(n.Op, n.OpPos)
		pr.node(n.Y)

	case *ast.KeyValueExpr:
		pr.node(n.Key)
		pr.token(token.COLON, n.Colon)
		pr.node(n.Value)

	case *ast.ArrayType:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.LBRACK, 0)
		if n.Len != nil {
			pr.node(n.Len)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		pr.token(token.RBRACK, 0)
		pr.node(n.Elt)

	case *ast.StructType:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.STRUCT, n.Struct)
		pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		pr.fields(n.Fields, token.LBRACE, token.SEMICOLON)

	case *ast.FuncType:
		if pr.nofunc {
			pr.nofunc = false
		} else {
			pr.comments(n.PreciseComments, ast.PointStart)
			pr.token(token.FUNC, n.Func)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		}
		if n.TypeParams != nil {
			pr.fields(n.TypeParams, token.LBRACK, token.COMMA)
		}
		pr.fields(n.Params, token.LPAREN, token.COMMA)
		if n.Results != nil { // includes (/*...*/)
			pr.fields(n.Results, token.LPAREN, token.COMMA)
		}

	case *ast.InterfaceType:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.INTERFACE, n.Interface)
		pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		pr.nofunc = true // suppress 'func'
		pr.fields(n.Methods, token.LBRACE, token.SEMICOLON)

	case *ast.MapType:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.MAP, n.Map)
		pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		pr.token(token.LBRACK, 0)
		pr.node(n.Key)
		pr.token(token.RBRACK, 0)
		pr.node(n.Value)

	case *ast.ChanType:
		pr.comments(n.PreciseComments, ast.PointStart)
		switch n.Dir {
		case ast.SEND | ast.RECV:
			pr.token(token.CHAN, n.Begin)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		case ast.SEND:
			pr.token(token.CHAN, n.Begin)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
			pr.token(token.ARROW, n.Arrow)
		case ast.RECV:
			pr.token(token.ARROW, n.Begin)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
			pr.token(token.CHAN, n.Arrow)
		}
		pr.node(n.Value)

	case *ast.DeclStmt:
		pr.node(n.Decl)

	case *ast.EmptyStmt:
		pr.comments(n.PreciseComments, ast.PointStart)

	case *ast.LabeledStmt:
		pr.node(n.Label)
		pr.token(token.COLON, n.Colon)
		pr.node(n.Stmt)

	case *ast.ExprStmt:
		pr.node(n.X)

	case *ast.SendStmt:
		pr.node(n.Chan)
		pr.token(token.ARROW, n.Arrow)
		pr.node(n.Value)

	case *ast.IncDecStmt:
		pr.node(n.X)
		pr.token(n.Tok, n.TokPos)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.AssignStmt:
		pr.exprs(n.Lhs)
		pr.token(n.Tok, n.TokPos)
		pr.exprs(n.Rhs)

	case *ast.GoStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.GO, n.Go)
		pr.node(n.Call)

	case *ast.DeferStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.DEFER, n.Defer)
		pr.node(n.Call)

	case *ast.ReturnStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.RETURN, n.Return)
		pr.exprs(n.Results)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.BranchStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(n.Tok, n.TokPos)
		if n.Label != nil {
			pr.node(n.Label)
		}
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.BlockStmt:
		switchblock := pr.switchblock
		pr.switchblock = false

		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.LBRACE, n.Lbrace)
		for i, stmt := range n.List {
			if i > 0 && !switchblock {
				pr.semicolon()
			}
			pr.node(stmt)
		}
		if hasComment(n.PreciseComments, ast.PointBeforeCloseBracket) {
			if len(n.List) > 0 && !switchblock {
				pr.semicolon()
			}
			pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		}
		pr.token(token.RBRACE, n.Rbrace)
		pr.comments(n.PreciseComments, ast.PointEnd)

	case *ast.IfStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.IF, n.If)
		if n.Init != nil {
			pr.node(n.Init)
			pr.semicolon()
		}
		if n.Cond != nil {
			pr.node(n.Cond)
		}
		pr.node(n.Body)
		if n.Else != nil {
			pr.token(token.ELSE, 0)
			pr.node(n.Else)
		}

	case *ast.CaseClause:
		pr.comments(n.PreciseComments, ast.PointStart)
		if n.List != nil {
			pr.token(token.CASE, n.Case)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
			pr.exprs(n.List)
		} else {
			pr.token(token.DEFAULT, n.Case)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		}
		pr.token(token.COLON, n.Colon)
		pr.comments(n.PreciseComments, ast.PointAfterColon)
		if n.Body != nil {
			pr.stmts(n.Body)
		}

	case *ast.SwitchStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.SWITCH, n.Switch)
		if n.Init != nil {
			pr.node(n.Init)
			pr.semicolon()
		}
		if n.Tag != nil {
			pr.node(n.Tag)
		}
		pr.switchblock = true // disable semicolons between cases
		pr.node(n.Body)

	case *ast.TypeSwitchStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.SWITCH, n.Switch)
		if n.Init != nil {
			pr.node(n.Init)
			pr.semicolon()
		}
		pr.node(n.Assign)
		pr.switchblock = true // disable semicolons between cases
		pr.node(n.Body)

	case *ast.CommClause:
		pr.comments(n.PreciseComments, ast.PointStart)
		if n.Comm != nil {
			pr.token(token.CASE, n.Case)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
			pr.node(n.Comm)
		} else {
			pr.token(token.DEFAULT, n.Case)
			pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		}
		pr.token(token.COLON, n.Colon)
		pr.comments(n.PreciseComments, ast.PointAfterColon)
		if n.Body != nil {
			pr.stmts(n.Body)
		}

	case *ast.SelectStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.SELECT, n.Select)
		pr.node(n.Body)

	case *ast.ForStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.FOR, n.For)
		pr.comments(n.PreciseComments, ast.PointAfterInitialToken)
		if n.Init != nil || n.Post != nil ||
			hasComment(n.PreciseComments, ast.PointAfterFirstSemicolon) {
			// 3-clause
			if n.Init != nil {
				pr.node(n.Init)
			}
			pr.semicolon()

			pr.comments(n.PreciseComments, ast.PointAfterFirstSemicolon)
			if n.Cond != nil {
				pr.node(n.Cond)
			}
			pr.semicolon()

			if n.Post != nil {
				pr.node(n.Post)
			}
		} else {
			// 0 or 1 clause
			if n.Cond != nil {
				pr.node(n.Cond)
			}
		}
		pr.node(n.Body)

	case *ast.RangeStmt:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.FOR, n.For)
		if n.Key != nil {
			pr.node(n.Key)
			if n.Value != nil {
				pr.token(token.COMMA, 0)
				pr.node(n.Value)
			}
			pr.token(n.Tok, n.TokPos)
		}
		pr.comments(n.PreciseComments, ast.PointBeforeRange)
		pr.token(token.RANGE, 0)
		pr.node(n.X)
		pr.node(n.Body)

	case *ast.ImportSpec:
		if n.Name != nil {
			pr.node(n.Name)
		}
		pr.node(n.Path)

	case *ast.ValueSpec:
		pr.idents(n.Names)
		if n.Type != nil {
			pr.node(n.Type)
		}
		if n.Values != nil {
			pr.token(token.ASSIGN, 0)
			pr.exprs(n.Values)
		}

	case *ast.TypeSpec:
		pr.node(n.Name)
		if n.TypeParams != nil {
			pr.fields(n.TypeParams, token.LBRACK, token.COMMA)
		}
		if n.Assign != 0 {
			pr.token(token.ASSIGN, n.Assign)
		}
		pr.node(n.Type)

	case *ast.GenDecl:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(n.Tok, n.TokPos)
		pr.comments(n.PreciseComments, ast.PointBeforeOpenBracket)
		if n.Lparen != 0 {
			pr.token(token.LPAREN, n.Lparen)
		}
		for i, spec := range n.Specs {
			if i > 0 {
				pr.semicolon()
			}
			pr.node(spec)
		}
		if hasComment(n.PreciseComments, ast.PointBeforeCloseBracket) {
			if len(n.Specs) > 0 {
				pr.semicolon()
			}
			pr.comments(n.PreciseComments, ast.PointBeforeCloseBracket)
		}
		if n.Rparen != 0 {
			pr.token(token.RPAREN, n.Rparen)
		}
		pr.comments(n.PreciseComments, ast.PointAfterCloseBracket)

	case *ast.FuncDecl:
		// (These steps are suppressed within FuncType)
		pr.comments(n.Type.PreciseComments, ast.PointStart)
		pr.token(token.FUNC, n.Type.Func)
		pr.comments(n.Type.PreciseComments, ast.PointAfterInitialToken)

		if n.Recv != nil {
			pr.fields(n.Recv, token.LPAREN, token.ILLEGAL)
		}
		pr.node(n.Name)
		pr.nofunc = true // suppress 'func'
		pr.node(n.Type)
		if n.Body != nil {
			pr.node(n.Body)
		}

	case *ast.File:
		pr.comments(n.PreciseComments, ast.PointStart)
		pr.token(token.PACKAGE, n.Package)
		pr.node(n.Name)
		pr.semicolon()
		for _, decl := range n.Decls {
			pr.node(decl)
			pr.semicolon()
		}
		pr.comments(n.PreciseComments, ast.PointEnd)

	default: // including Bad{Decl,Stmt,Expr}
		panic(fmt.Sprintf("unexpected node type %T", n))
	}
	pr.nofunc = false
}

func (pr *printer) semicolon() {
	// A smarter printer would use the line number of
	// the next token to choose '\n' vs ';'.
	pr.token(token.SEMICOLON, 0)
}

func (pr *printer) string(s string, pos token.Pos) {
	pr.buf.WriteString(s)
	pr.buf.WriteByte(' ')
}

func (pr *printer) token(tok token.Token, pos token.Pos) {
	pr.string(tok.String(), pos)
}

func (pr *printer) comments(pcs *ast.PreciseComments, point ast.Point) {
	if pcs != nil {
		for _, pc := range pcs.List {
			if pc.Point == point {
				fmt.Fprintf(&pr.buf, "/*%s*/ ", strings.TrimSpace(pc.Comment.Text()))
			}
		}
	}
}

func hasComment(pcs *ast.PreciseComments, point ast.Point) bool {
	if pcs != nil {
		for _, pc := range pcs.List {
			if pc.Point == point {
				return true
			}
		}
	}
	return false
}
