// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package template

import (
	"bytes"
	"fmt"
	"html"
	"html/template/parse"
	"io"
	"strings"
	"text/template"
	textparse "text/template/parse"
)

// escapeTemplate rewrites the named template, which must be
// associated with t, to guarantee that the output of any of the named
// templates is properly escaped. If no error is returned, then the named templates have
// been modified. Otherwise the named templates have been rendered
// unusable.
func escapeTemplate(tmpl *Template, node textparse.Node, name string) error {
	c, _ := tmpl.esc.escapeTree(parse.Context{}, node, name, 0)
	var err error
	if c.Err != nil {
		err, c.Err.Name = c.Err, name
	} else if c.State != parse.StateText {
		err = &parse.Error{parse.ErrEndContext, nil, name, 0, fmt.Sprintf("ends in a non-text context: %v", c)}
	}
	if err != nil {
		// Prevent execution of unsafe templates.
		if t := tmpl.set[name]; t != nil {
			t.escapeErr = err
			t.text.Tree = nil
			t.Tree = nil
		}
		return err
	}
	tmpl.esc.commit()
	if t := tmpl.set[name]; t != nil {
		t.escapeErr = escapeOK
		t.Tree = t.text.Tree
	}
	return nil
}

// evalArgs formats the list of arguments into a string. It is equivalent to
// fmt.Sprint(args...), except that it deferences all pointers.
func evalArgs(args ...interface{}) string {
	// Optimization for simple common case of a single string argument.
	if len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	for i, arg := range args {
		args[i] = indirectToStringerOrError(arg)
	}
	return fmt.Sprint(args...)
}

// funcMap maps command names to functions that render their inputs safe.
var funcMap = template.FuncMap{
	"_html_template_attrescaper":     attrEscaper,
	"_html_template_commentescaper":  commentEscaper,
	"_html_template_cssescaper":      cssEscaper,
	"_html_template_cssvaluefilter":  cssValueFilter,
	"_html_template_htmlnamefilter":  htmlNameFilter,
	"_html_template_htmlescaper":     htmlEscaper,
	"_html_template_jsregexpescaper": jsRegexpEscaper,
	"_html_template_jsstrescaper":    jsStrEscaper,
	"_html_template_jsvalescaper":    jsValEscaper,
	"_html_template_nospaceescaper":  htmlNospaceEscaper,
	"_html_template_rcdataescaper":   rcdataEscaper,
	"_html_template_srcsetescaper":   srcsetFilterAndEscaper,
	"_html_template_urlescaper":      urlEscaper,
	"_html_template_urlfilter":       urlFilter,
	"_html_template_urlnormalizer":   urlNormalizer,
	"_eval_args_":                    evalArgs,
}

// escaper collects type inferences about templates and changes needed to make
// templates injection safe.
type escaper struct {
	// ns is the nameSpace that this escaper is associated with.
	ns *nameSpace
	// output[templateName] is the output context for a templateName that
	// has been mangled to include its input context.
	output map[string]parse.Context
	// derived[c.mangle(name)] maps to a template derived from the template
	// named name templateName for the start context c.
	derived map[string]*template.Template
	// called[templateName] is a set of called mangled template names.
	called map[string]bool
	// xxxNodeEdits are the accumulated edits to apply during commit.
	// Such edits are not applied immediately in case a template set
	// executes a given template in different escaping contexts.
	actionNodeEdits   map[*textparse.ActionNode][]string
	templateNodeEdits map[*textparse.TemplateNode]string
	textNodeEdits     map[*textparse.TextNode][]byte
}

// makeEscaper creates a blank escaper for the given set.
func makeEscaper(n *nameSpace) escaper {
	return escaper{
		n,
		map[string]parse.Context{},
		map[string]*template.Template{},
		map[string]bool{},
		map[*textparse.ActionNode][]string{},
		map[*textparse.TemplateNode]string{},
		map[*textparse.TextNode][]byte{},
	}
}

// filterFailsafe is an innocuous word that is emitted in place of unsafe values
// by sanitizer functions. It is not a keyword in any programming language,
// contains no special characters, is not empty, and when it appears in output
// it is distinct enough that a developer can find the source of the problem
// via a search engine.
const filterFailsafe = "ZgotmplZ"

// escape escapes a template node.
func (e *escaper) escape(c parse.Context, n textparse.Node) parse.Context {
	switch n := n.(type) {
	case *textparse.ActionNode:
		return e.escapeAction(c, n)
	case *textparse.IfNode:
		return e.escapeBranch(c, &n.BranchNode, "if")
	case *textparse.ListNode:
		return e.escapeList(c, n)
	case *textparse.RangeNode:
		return e.escapeBranch(c, &n.BranchNode, "range")
	case *textparse.TemplateNode:
		return e.escapeTemplate(c, n)
	case *textparse.TextNode:
		return e.escapeText(c, n)
	case *textparse.WithNode:
		return e.escapeBranch(c, &n.BranchNode, "with")
	}
	panic("escaping " + n.String() + " is unimplemented")
}

// escapeAction escapes an action template node.
func (e *escaper) escapeAction(c parse.Context, n *textparse.ActionNode) parse.Context {
	if len(n.Pipe.Decl) != 0 {
		// A local variable assignment, not an interpolation.
		return c
	}
	c = nudge(c)
	// Check for disallowed use of predefined escapers in the pipeline.
	for pos, idNode := range n.Pipe.Cmds {
		node, ok := idNode.Args[0].(*textparse.IdentifierNode)
		if !ok {
			// A predefined escaper "esc" will never be found as an identifier in a
			// Chain or Field node, since:
			// - "esc.x ..." is invalid, since predefined escapers return strings, and
			//   strings do not have methods, keys or fields.
			// - "... .esc" is invalid, since predefined escapers are global functions,
			//   not methods or fields of any types.
			// Therefore, it is safe to ignore these two node types.
			continue
		}
		ident := node.Ident
		if _, ok := predefinedEscapers[ident]; ok {
			if pos < len(n.Pipe.Cmds)-1 ||
				c.State == parse.StateAttr && c.Delim == parse.DelimSpaceOrTagEnd && ident == "html" {
				return parse.Context{
					State: parse.StateError,
					Err:   parse.Errorf(parse.ErrPredefinedEscaper, n, n.Line, "predefined escaper %q disallowed in template", ident),
				}
			}
		}
	}
	switch c.State {
	case parse.StateError:
		return c
	case parse.StateURL, parse.StateCSSDqStr, parse.StateCSSSqStr, parse.StateCSSDqURL, parse.StateCSSSqURL, parse.StateCSSURL:
		if c.URLPart == parse.URLPartUnknown {
			return parse.Context{
				State: parse.StateError,
				Err:   parse.Errorf(parse.ErrAmbigContext, n, n.Line, "%s appears in an ambiguous context within a URL", n),
			}
		}
	case parse.StateJS:
		// A slash after a value starts a div operator.
		c.JSCtx = parse.JSCtxDivOp
	case parse.StateAttrName, parse.StateTag:
		c.State = parse.StateAttrName
	}
	if e.ns.escaperForContext == nil {
		panic("contextual autoescaping behavior is undefined")
	}
	s, err := e.ns.escaperForContext(c)
	if err != nil {
		return parse.Context{
			State: parse.StateError,
			Err:   parse.Errorf(parse.ErrEscapeAction, n, n.Line, "cannot escape action %v: %s", n, err),
		}
	}
	e.editActionNode(n, s)
	return c
}

// defaultEscaperForContext is the default implementation of html/template's contextual
// auto-escaping behavior.
func defaultEscaperForContext(c parse.Context) ([]string, error) {
	s := make([]string, 0, 3)
	switch c.State {
	case parse.StateURL, parse.StateCSSDqStr, parse.StateCSSSqStr, parse.StateCSSDqURL, parse.StateCSSSqURL, parse.StateCSSURL:
		switch c.URLPart {
		case parse.URLPartNone:
			s = append(s, "_html_template_urlfilter")
			fallthrough
		case parse.URLPartPreQuery:
			switch c.State {
			case parse.StateCSSDqStr, parse.StateCSSSqStr:
				s = append(s, "_html_template_cssescaper")
			default:
				s = append(s, "_html_template_urlnormalizer")
			}
		case parse.URLPartQueryOrFrag:
			s = append(s, "_html_template_urlescaper")
		default:
			panic(c.URLPart.String())
		}
	case parse.StateJS:
		s = append(s, "_html_template_jsvalescaper")
	case parse.StateJSDqStr, parse.StateJSSqStr:
		s = append(s, "_html_template_jsstrescaper")
	case parse.StateJSRegexp:
		s = append(s, "_html_template_jsregexpescaper")
	case parse.StateCSS:
		s = append(s, "_html_template_cssvaluefilter")
	case parse.StateText:
		s = append(s, "_html_template_htmlescaper")
	case parse.StateRCDATA:
		s = append(s, "_html_template_rcdataescaper")
	case parse.StateAttr:
		// Handled below in delim check.
	case parse.StateAttrName, parse.StateTag:
		s = append(s, "_html_template_htmlnamefilter")
	case parse.StateSrcset:
		s = append(s, "_html_template_srcsetescaper")
	default:
		if parse.IsComment(c.State) {
			s = append(s, "_html_template_commentescaper")
		} else {
			panic("unexpected state " + c.State.String())
		}
	}
	switch c.Delim {
	case parse.DelimNone:
		// No extra-escaping needed for raw text content.
	case parse.DelimSpaceOrTagEnd:
		s = append(s, "_html_template_nospaceescaper")
	default:
		s = append(s, "_html_template_attrescaper")
	}
	return s, nil
}

// ensurePipelineContains ensures that the pipeline ends with the commands with
// the identifiers in s in order. If the pipeline ends with a predefined escaper
// (i.e. "html" or "urlquery"), merge it with the identifiers in s.
func ensurePipelineContains(p *textparse.PipeNode, s []string) {
	if len(s) == 0 {
		// Do not rewrite pipeline if we have no escapers to insert.
		return
	}
	// Precondition: p.Cmds contains at most one predefined escaper and the
	// escaper will be present at p.Cmds[len(p.Cmds)-1]. This precondition is
	// always true because of the checks in escapeAction.
	pipelineLen := len(p.Cmds)
	if pipelineLen > 0 {
		lastCmd := p.Cmds[pipelineLen-1]
		if idNode, ok := lastCmd.Args[0].(*textparse.IdentifierNode); ok {
			if esc := idNode.Ident; predefinedEscapers[esc] {
				// Pipeline ends with a predefined escaper.
				if len(p.Cmds) == 1 && len(lastCmd.Args) > 1 {
					// Special case: pipeline is of the form {{ esc arg1 arg2 ... argN }},
					// where esc is the predefined escaper, and arg1...argN are its arguments.
					// Convert this into the equivalent form
					// {{ _eval_args_ arg1 arg2 ... argN | esc }}, so that esc can be easily
					// merged with the escapers in s.
					lastCmd.Args[0] = textparse.NewIdentifier("_eval_args_").SetTree(nil).SetPos(lastCmd.Args[0].Position())
					p.Cmds = appendCmd(p.Cmds, newIdentCmd(esc, p.Position()))
					pipelineLen++
				}
				// If any of the commands in s that we are about to insert is equivalent
				// to the predefined escaper, use the predefined escaper instead.
				dup := false
				for i, escaper := range s {
					if escFnsEq(esc, escaper) {
						s[i] = idNode.Ident
						dup = true
					}
				}
				if dup {
					// The predefined escaper will already be inserted along with the
					// escapers in s, so do not copy it to the rewritten pipeline.
					pipelineLen--
				}
			}
		}
	}
	// Rewrite the pipeline, creating the escapers in s at the end of the pipeline.
	newCmds := make([]*textparse.CommandNode, pipelineLen, pipelineLen+len(s))
	insertedIdents := make(map[string]bool)
	for i := 0; i < pipelineLen; i++ {
		cmd := p.Cmds[i]
		newCmds[i] = cmd
		if idNode, ok := cmd.Args[0].(*textparse.IdentifierNode); ok {
			insertedIdents[normalizeEscFn(idNode.Ident)] = true
		}
	}
	for _, name := range s {
		if !insertedIdents[normalizeEscFn(name)] {
			// When two templates share an underlying parse tree via the use of
			// AddParseTree and one template is executed after the other, this check
			// ensures that escapers that were already inserted into the pipeline on
			// the first escaping pass do not get inserted again.
			newCmds = appendCmd(newCmds, newIdentCmd(name, p.Position()))
		}
	}
	p.Cmds = newCmds
}

// predefinedEscapers contains template predefined escapers that are equivalent
// to some contextual escapers. Keep in sync with equivEscapers.
var predefinedEscapers = map[string]bool{
	"html":     true,
	"urlquery": true,
}

// equivEscapers matches contextual escapers to equivalent predefined
// template escapers.
var equivEscapers = map[string]string{
	// The following pairs of HTML escapers provide equivalent security
	// guarantees, since they all escape '\000', '\'', '"', '&', '<', and '>'.
	"_html_template_attrescaper":   "html",
	"_html_template_htmlescaper":   "html",
	"_html_template_rcdataescaper": "html",
	// These two URL escapers produce URLs safe for embedding in a URL query by
	// percent-encoding all the reserved characters specified in RFC 3986 Section
	// 2.2
	"_html_template_urlescaper": "urlquery",
	// These two functions are not actually equivalent; urlquery is stricter as it
	// escapes reserved characters (e.g. '#'), while _html_template_urlnormalizer
	// does not. It is therefore only safe to replace _html_template_urlnormalizer
	// with urlquery (this happens in ensurePipelineContains), but not the otherI've
	// way around. We keep this entry around to preserve the behavior of templates
	// written before Go 1.9, which might depend on this substitution taking place.
	"_html_template_urlnormalizer": "urlquery",
}

// escFnsEq reports whether the two escaping functions are equivalent.
func escFnsEq(a, b string) bool {
	return normalizeEscFn(a) == normalizeEscFn(b)
}

// normalizeEscFn(a) is equal to normalizeEscFn(b) for any pair of names of
// escaper functions a and b that are equivalent.
func normalizeEscFn(e string) string {
	if norm := equivEscapers[e]; norm != "" {
		return norm
	}
	return e
}

// redundantFuncs[a][b] implies that funcMap[b](funcMap[a](x)) == funcMap[a](x)
// for all x.
var redundantFuncs = map[string]map[string]bool{
	"_html_template_commentescaper": {
		"_html_template_attrescaper":    true,
		"_html_template_nospaceescaper": true,
		"_html_template_htmlescaper":    true,
	},
	"_html_template_cssescaper": {
		"_html_template_attrescaper": true,
	},
	"_html_template_jsregexpescaper": {
		"_html_template_attrescaper": true,
	},
	"_html_template_jsstrescaper": {
		"_html_template_attrescaper": true,
	},
	"_html_template_urlescaper": {
		"_html_template_urlnormalizer": true,
	},
}

// appendCmd appends the given command to the end of the command pipeline
// unless it is redundant with the last command.
func appendCmd(cmds []*textparse.CommandNode, cmd *textparse.CommandNode) []*textparse.CommandNode {
	if n := len(cmds); n != 0 {
		last, okLast := cmds[n-1].Args[0].(*textparse.IdentifierNode)
		next, okNext := cmd.Args[0].(*textparse.IdentifierNode)
		if okLast && okNext && redundantFuncs[last.Ident][next.Ident] {
			return cmds
		}
	}
	return append(cmds, cmd)
}

// newIdentCmd produces a command containing a single identifier node.
func newIdentCmd(identifier string, pos textparse.Pos) *textparse.CommandNode {
	return &textparse.CommandNode{
		NodeType: textparse.NodeCommand,
		Args:     []textparse.Node{textparse.NewIdentifier(identifier).SetTree(nil).SetPos(pos)}, // TODO: SetTree.
		Pos:      pos,
	}
}

// nudge returns the context that would result from following empty string
// transitions from the input context.
// For example, parsing:
//     `<a href=`
// will end in Context{StateBeforeValue, AttrURL}, but parsing one extra rune:
//     `<a href=x`
// will end in Context{StateURL, DelimSpaceOrTagEnd, ...}.
// There are two transitions that happen when the 'x' is seen:
// (1) Transition from a before-value state to a start-of-value state without
//     consuming any character.
// (2) Consume 'x' and transition past the first value character.
// In this case, nudging produces the context after (1) happens.
func nudge(c parse.Context) parse.Context {
	switch c.State {
	case parse.StateTag:
		// In `<foo {{.}}`, the action should emit an attribute.
		c.State = parse.StateAttrName
	case parse.StateBeforeValue:
		// In `<foo bar={{.}}`, the action is an undelimited value.
		c.State, c.Delim = attrStartState(c.Attr), parse.DelimSpaceOrTagEnd
	case parse.StateAfterName:
		// In `<foo bar {{.}}`, the action is an attribute name.
		c.State = parse.StateAttrName
	}
	return c
}

// join joins the two contexts of a branch template node. The result is an
// error context if either of the input contexts are error contexts, or if the
// input contexts differ.
func join(a, b parse.Context, node textparse.Node, nodeName string) parse.Context {
	if a.State == parse.StateError {
		return a
	}
	if b.State == parse.StateError {
		return b
	}

	// Accumulate the result of context-joining elements and attributes in a, since the
	// contents of a are always returned.
	a.Element.Names = joinNames(a.Element.Name, b.Element.Name, a.Element.Names, b.Element.Names)
	a.Attr.Names = joinNames(a.Attr.Name, b.Attr.Name, a.Attr.Names, b.Attr.Names)
	if a.Attr.Value != b.Attr.Value {
		a.Attr.AmbiguousValue = true
	}

	if a.Eq(b) {
		return a
	}

	c := a
	c.Element.Name = b.Element.Name
	if c.Eq(b) && stateForElement[a.Element.Name] == stateForElement[b.Element.Name] {
		// The contexts differ only by their element names, which are both treated the same way by the parser and escaper.
		return c
	}

	c = a
	c.Attr.Name = b.Attr.Name
	if c.Eq(b) && attrStartState(a.Attr) == attrStartState(b.Attr) {
		// The contexts differ only by their attribute names, which both map to the same parser state.
		return c
	}

	c = a
	c.URLPart = b.URLPart
	if c.Eq(b) {
		// The contexts differ only by URLPart.
		c.URLPart = parse.URLPartUnknown
		return c
	}

	c = a
	c.JSCtx = b.JSCtx
	if c.Eq(b) {
		// The contexts differ only by JSCtx.
		c.JSCtx = parse.JSCtxUnknown
		return c
	}

	// Allow a nudged context to join with an unnudged one.
	// This means that
	//   <p title={{if .C}}{{.}}{{end}}
	// ends in an unquoted value State even though the else branch
	// ends in StateBeforeValue.
	if c, d := nudge(a), nudge(b); !(c.Eq(a) && d.Eq(b)) {
		if e := join(c, d, node, nodeName); e.State != parse.StateError {
			return e
		}
	}

	return parse.Context{
		State: parse.StateError,
		Err:   parse.Errorf(parse.ErrBranchEnd, node, 0, "{{%s}} branches end in different contexts: %v, %v", nodeName, a, b),
	}
}

// joinNames returns the slice of all possible names that an Element or Attr could
// assume after context joining the Element or Attr containing aName and aNames with the
// Element or Attr containing bName and bNames.
func joinNames(aName, bName string, aNames, bNames []string) []string {
	var ret []string
	if aName != bName {
		ret = append(ret, aName, bName)
	}
	aNamesSet := make(map[string]bool)
	for _, name := range aNames {
		aNamesSet[name] = true
	}
	for _, name := range bNames {
		if !aNamesSet[name] {
			ret = append(ret, name)
		}
	}
	return ret
}

// escapeBranch escapes a branch template node: "if", "range" and "with".
func (e *escaper) escapeBranch(c parse.Context, n *textparse.BranchNode, nodeName string) parse.Context {
	c0 := e.escapeList(c, n.List)
	if nodeName == "range" && c0.State != parse.StateError {
		// The "true" branch of a "range" node can execute multiple times.
		// We check that executing n.List once results in the same context
		// as executing n.List twice.
		c1, _ := e.escapeListConditionally(c0, n.List, nil)
		c0 = join(c0, c1, n, nodeName)
		if c0.State == parse.StateError {
			// Make clear that this is a problem on loop re-entry
			// since developers tend to overlook that branch when
			// debugging templates.
			c0.Err.Line = n.Line
			c0.Err.Description = "on range loop re-entry: " + c0.Err.Description
			return c0
		}
	}
	c1 := e.escapeList(c, n.ElseList)
	return join(c0, c1, n, nodeName)
}

// escapeList escapes a list template node.
func (e *escaper) escapeList(c parse.Context, n *textparse.ListNode) parse.Context {
	if n == nil {
		return c
	}
	for _, m := range n.Nodes {
		c = e.escape(c, m)
	}
	return c
}

// escapeListConditionally escapes a list node but only preserves edits and
// inferences in e if the inferences and output context satisfy filter.
// It returns the best guess at an output context, and the result of the filter
// which is the same as whether e was updated.
func (e *escaper) escapeListConditionally(c parse.Context, n *textparse.ListNode, filter func(*escaper, parse.Context) bool) (parse.Context, bool) {
	e1 := makeEscaper(e.ns)
	// Make type inferences available to f.
	for k, v := range e.output {
		e1.output[k] = v
	}
	c = e1.escapeList(c, n)
	ok := filter != nil && filter(&e1, c)
	if ok {
		// Copy inferences and edits from e1 back into e.
		for k, v := range e1.output {
			e.output[k] = v
		}
		for k, v := range e1.derived {
			e.derived[k] = v
		}
		for k, v := range e1.called {
			e.called[k] = v
		}
		for k, v := range e1.actionNodeEdits {
			e.editActionNode(k, v)
		}
		for k, v := range e1.templateNodeEdits {
			e.editTemplateNode(k, v)
		}
		for k, v := range e1.textNodeEdits {
			e.editTextNode(k, v)
		}
	}
	return c, ok
}

// escapeTemplate escapes a {{template}} call node.
func (e *escaper) escapeTemplate(c parse.Context, n *textparse.TemplateNode) parse.Context {
	c, name := e.escapeTree(c, n, n.Name, n.Line)
	if name != n.Name {
		e.editTemplateNode(n, name)
	}
	return c
}

// mangle produces an identifier that includes a suffix that distinguishes it
// from template names mangled with different contexts.
func mangle(c parse.Context, templateName string) string {
	// The mangled name for the default Context is the input templateName.
	if c.State == parse.StateText {
		return templateName
	}
	s := templateName + "$htmltemplate_" + c.State.String()
	if c.Delim != 0 {
		s += "_" + c.Delim.String()
	}
	if c.URLPart != 0 {
		s += "_" + c.URLPart.String()
	}
	if c.JSCtx != 0 {
		s += "_" + c.JSCtx.String()
	}
	if c.Attr.Name != "" {
		s += "_" + c.Attr.String()
	}
	if c.Element.Name != "" {
		s += "_" + c.Element.String()
	}
	return s
}

// escapeTree escapes the named template starting in the given context as
// necessary and returns its output context.
func (e *escaper) escapeTree(c parse.Context, node textparse.Node, name string, line int) (parse.Context, string) {
	// Mangle the template name with the input context to produce a reliable
	// identifier.
	dname := mangle(c, name)
	e.called[dname] = true
	if out, ok := e.output[dname]; ok {
		// Already escaped.
		return out, dname
	}
	t := e.template(name)
	if t == nil {
		// Two cases: The template exists but is empty, or has never been mentioned at
		// all. Distinguish the cases in the error messages.
		if e.ns.set[name] != nil {
			return parse.Context{
				State: parse.StateError,
				Err:   parse.Errorf(parse.ErrNoSuchTemplate, node, line, "%q is an incomplete or empty template", name),
			}, dname
		}
		return parse.Context{
			State: parse.StateError,
			Err:   parse.Errorf(parse.ErrNoSuchTemplate, node, line, "no such template %q", name),
		}, dname
	}
	if dname != name {
		// Use any template derived during an earlier call to escapeTemplate
		// with different top level templates, or clone if necessary.
		dt := e.template(dname)
		if dt == nil {
			dt = template.New(dname)
			dt.Tree = &textparse.Tree{Name: dname, Root: t.Root.CopyList()}
			e.derived[dname] = dt
		}
		t = dt
	}
	return e.computeOutCtx(c, t), dname
}

// computeOutCtx takes a template and its start context and computes the output
// context while storing any inferences in e.
func (e *escaper) computeOutCtx(c parse.Context, t *template.Template) parse.Context {
	// Propagate context over the body.
	c1, ok := e.escapeTemplateBody(c, t)
	if !ok {
		// Look for a fixed point by assuming c1 as the output context.
		if c2, ok2 := e.escapeTemplateBody(c1, t); ok2 {
			c1, ok = c2, true
		}
		// Use c1 as the error context if neither assumption worked.
	}
	if !ok && c1.State != parse.StateError {
		return parse.Context{
			State: parse.StateError,
			Err:   parse.Errorf(parse.ErrOutputContext, t.Tree.Root, 0, "cannot compute output context for template %s", t.Name()),
		}
	}
	return c1
}

// escapeTemplateBody escapes the given template assuming the given output
// context, and returns the best guess at the output context and whether the
// assumption was correct.
func (e *escaper) escapeTemplateBody(c parse.Context, t *template.Template) (parse.Context, bool) {
	filter := func(e1 *escaper, c1 parse.Context) bool {
		if c1.State == parse.StateError {
			// Do not update the input escaper, e.
			return false
		}
		if !e1.called[t.Name()] {
			// If t is not recursively called, then c1 is an
			// accurate output context.
			return true
		}
		// c1 is accurate if it matches our assumed output context.
		return c.Eq(c1)
	}
	// We need to assume an output context so that recursive template calls
	// take the fast path out of escapeTree instead of infinitely recursing.
	// Naively assuming that the input context is the same as the output
	// works >90% of the time.
	e.output[t.Name()] = c
	return e.escapeListConditionally(c, t.Tree.Root, filter)
}

// delimEnds maps each Delim to a string of characters that terminate it.
var delimEnds = [...]string{
	parse.DelimDoubleQuote: `"`,
	parse.DelimSingleQuote: "'",
	// Determined empirically by running the below in various browsers.
	// var div = document.createElement("DIV");
	// for (var i = 0; i < 0x10000; ++i) {
	//   div.innerHTML = "<span title=x" + String.fromCharCode(i) + "-bar>";
	//   if (div.getElementsByTagName("SPAN")[0].title.indexOf("bar") < 0)
	//     document.write("<p>U+" + i.toString(16));
	// }
	parse.DelimSpaceOrTagEnd: " \t\n\f\r>",
}

var doctypeBytes = []byte("<!DOCTYPE")

// escapeText escapes a text template node.
func (e *escaper) escapeText(c parse.Context, n *textparse.TextNode) parse.Context {
	s, written, i, b := n.Text, 0, 0, new(bytes.Buffer)
	if e.ns.cspCompatible && bytes.Contains(s, []byte("javascript:")) {
		// This substring search is a hack, but it is unlikely that this substring will
		// exist in template text for any other reason than to specify a javascript URI.
		return parse.Context{
			State: parse.StateError,
			Err:   parse.Errorf(parse.ErrCSPCompatibility, n, 0, `"javascript:" URI disallowed for CSP compatibility`),
		}
	}
	for i != len(s) {
		if e.ns.cspCompatible && strings.HasPrefix(c.Attr.Name, "on") {
			return parse.Context{
				State: parse.StateError,
				Err:   parse.Errorf(parse.ErrCSPCompatibility, n, 0, "inline event handler %q is disallowed for CSP compatibility", c.Attr.Name),
			}
		}
		c1, nread := contextAfterText(c, s[i:])
		i1 := i + nread
		if c.State == parse.StateText || c.State == parse.StateRCDATA {
			end := i1
			if c1.State != c.State {
				for j := end - 1; j >= i; j-- {
					if s[j] == '<' {
						end = j
						break
					}
				}
			}
			for j := i; j < end; j++ {
				if s[j] == '<' && !bytes.HasPrefix(bytes.ToUpper(s[j:]), doctypeBytes) {
					b.Write(s[written:j])
					b.WriteString("&lt;")
					written = j + 1
				}
			}
		} else if parse.IsComment(c.State) && c.Delim == parse.DelimNone {
			switch c.State {
			case parse.StateJSBlockCmt:
				// http://es5.github.com/#x7.4:
				// "Comments behave like white space and are
				// discarded except that, if a MultiLineComment
				// contains a line terminator character, then
				// the entire comment is considered to be a
				// LineTerminator for purposes of parsing by
				// the syntactic grammar."
				// TODO(samueltan): replace this with bytes.ContainsAny once AppEngine supports go1.8.
				if strings.ContainsAny(string(s[written:i1]), "\n\r\u2028\u2029") {
					b.WriteByte('\n')
				} else {
					b.WriteByte(' ')
				}
			case parse.StateCSSBlockCmt:
				b.WriteByte(' ')
			}
			written = i1
		}
		if c.State != c1.State && parse.IsComment(c1.State) && c1.Delim == parse.DelimNone {
			// Preserve the portion between written and the comment start.
			cs := i1 - 2
			if c1.State == parse.StateHTMLCmt {
				// "<!--" instead of "/*" or "//"
				cs -= 2
			}
			b.Write(s[written:cs])
			written = i1
		}
		if i == i1 && c.State == c1.State {
			panic(fmt.Sprintf("infinite loop from %v to %v on %q..%q", c, c1, s[:i], s[i:]))
		}
		c, i = c1, i1
	}

	if written != 0 && c.State != parse.StateError {
		if !parse.IsComment(c.State) || c.Delim != parse.DelimNone {
			b.Write(n.Text[written:])
		}
		e.editTextNode(n, b.Bytes())
	}
	return c
}

// contextAfterText starts in context c, consumes some tokens from the front of
// s, then returns the context after those tokens and the unprocessed suffix.
func contextAfterText(c parse.Context, s []byte) (parse.Context, int) {
	if c.Delim == parse.DelimNone {
		c1, i := tSpecialTagEnd(c, s)
		if i == 0 {
			// A special end tag (`</script>`) has been seen and
			// all content preceding it has been consumed.
			return c1, 0
		}
		// Consider all content up to any end tag.
		return transitionFunc[c.State](c, s[:i])
	}

	// We are at the beginning of an attribute value.

	i := bytes.IndexAny(s, delimEnds[c.Delim])
	if i == -1 {
		i = len(s)
	}
	if c.Delim == parse.DelimSpaceOrTagEnd {
		// http://www.w3.org/TR/html5/syntax.html#attribute-value-(unquoted)-state
		// lists the runes below as error characters.
		// Error out because HTML parsers may differ on whether
		// "<a id= onclick=f("     ends inside id's or onclick's value,
		// "<a class=`foo "        ends inside a value,
		// "<a style=font:'Arial'" needs open-quote fixup.
		// IE treats '`' as a quotation character.
		if j := bytes.IndexAny(s[:i], "\"'<=`"); j >= 0 {
			return parse.Context{
				State: parse.StateError,
				Err:   parse.Errorf(parse.ErrBadHTML, nil, 0, "%q in unquoted attr: %q", s[j:j+1], s[:i]),
			}, len(s)
		}
	}
	if i == len(s) {
		c.Attr.Value += string(s)
		// Remain inside the attribute.
		// Decode the value so non-HTML rules can easily handle
		//     <button onclick="alert(&quot;Hi!&quot;)">
		// without having to entity decode token boundaries.
		for u := []byte(html.UnescapeString(string(s))); len(u) != 0; {
			c1, i1 := transitionFunc[c.State](c, u)
			c, u = c1, u[i1:]
		}
		return c, len(s)
	}

	// On exiting an attribute, we discard all state information
	// except the State, Element, ScriptType, and LinkRel.
	ret := parse.Context{
		State:      parse.StateTag,
		Element:    c.Element,
		ScriptType: c.ScriptType,
		LinkRel:    c.LinkRel,
	}
	// Save the script element's type attribute value if we are parsing it for the first time.
	if c.State == parse.StateAttr && c.Element.Name == "script" && c.Attr.Name == "type" {
		ret.ScriptType = strings.ToLower(string(s[:i]))
	}
	// Save the link element's rel attribute value if we are parsing it for the first time.
	if c.State == parse.StateAttr && c.Element.Name == "link" && c.Attr.Name == "rel" {
		ret.LinkRel = " " + strings.Join(strings.Fields(strings.TrimSpace(strings.ToLower(string(s[:i])))), " ") + " "
	}
	if c.Delim != parse.DelimSpaceOrTagEnd {
		// Consume any quote.
		i++
	}
	return ret, i
}

// editActionNode records a change to an action pipeline for later commit.
func (e *escaper) editActionNode(n *textparse.ActionNode, cmds []string) {
	if _, ok := e.actionNodeEdits[n]; ok {
		panic(fmt.Sprintf("node %s shared between templates", n))
	}
	e.actionNodeEdits[n] = cmds
}

// editTemplateNode records a change to a {{template}} callee for later commit.
func (e *escaper) editTemplateNode(n *textparse.TemplateNode, callee string) {
	if _, ok := e.templateNodeEdits[n]; ok {
		panic(fmt.Sprintf("node %s shared between templates", n))
	}
	e.templateNodeEdits[n] = callee
}

// editTextNode records a change to a text node for later commit.
func (e *escaper) editTextNode(n *textparse.TextNode, text []byte) {
	if _, ok := e.textNodeEdits[n]; ok {
		panic(fmt.Sprintf("node %s shared between templates", n))
	}
	e.textNodeEdits[n] = text
}

// commit applies changes to actions and template calls needed to contextually
// autoescape content and adds any derived templates to the set.
func (e *escaper) commit() {
	for name := range e.output {
		e.template(name).Funcs(funcMap)
	}
	// Any template from the name space associated with this escaper can be used
	// to add derived templates to the underlying text/template name space.
	tmpl := e.arbitraryTemplate()
	for _, t := range e.derived {
		if _, err := tmpl.text.AddParseTree(t.Name(), t.Tree); err != nil {
			panic("error adding derived template")
		}
	}
	for n, s := range e.actionNodeEdits {
		ensurePipelineContains(n.Pipe, s)
	}
	for n, name := range e.templateNodeEdits {
		n.Name = name
	}
	for n, s := range e.textNodeEdits {
		n.Text = s
	}
	// Reset state that is specific to this commit so that the same changes are
	// not re-applied to the template on subsequent calls to commit.
	e.called = make(map[string]bool)
	e.actionNodeEdits = make(map[*textparse.ActionNode][]string)
	e.templateNodeEdits = make(map[*textparse.TemplateNode]string)
	e.textNodeEdits = make(map[*textparse.TextNode][]byte)
}

// template returns the named template given a mangled template name.
func (e *escaper) template(name string) *template.Template {
	// Any template from the name space associated with this escaper can be used
	// to look up templates in the underlying text/template name space.
	t := e.arbitraryTemplate().text.Lookup(name)
	if t == nil {
		t = e.derived[name]
	}
	return t
}

// arbitraryTemplate returns an arbitrary template from the name space
// associated with e and panics if no templates are found.
func (e *escaper) arbitraryTemplate() *Template {
	for _, t := range e.ns.set {
		return t
	}
	panic("no templates in name space")
}

// Forwarding functions so that clients need only import this package
// to reach the general escaping functions of text/template.

// HTMLEscape writes to w the escaped HTML equivalent of the plain text data b.
func HTMLEscape(w io.Writer, b []byte) {
	template.HTMLEscape(w, b)
}

// HTMLEscapeString returns the escaped HTML equivalent of the plain text data s.
func HTMLEscapeString(s string) string {
	return template.HTMLEscapeString(s)
}

// HTMLEscaper returns the escaped HTML equivalent of the textual
// representation of its arguments.
func HTMLEscaper(args ...interface{}) string {
	return template.HTMLEscaper(args...)
}

// JSEscape writes to w the escaped JavaScript equivalent of the plain text data b.
func JSEscape(w io.Writer, b []byte) {
	template.JSEscape(w, b)
}

// JSEscapeString returns the escaped JavaScript equivalent of the plain text data s.
func JSEscapeString(s string) string {
	return template.JSEscapeString(s)
}

// JSEscaper returns the escaped JavaScript equivalent of the textual
// representation of its arguments.
func JSEscaper(args ...interface{}) string {
	return template.JSEscaper(args...)
}

// URLQueryEscaper returns the escaped value of the textual representation of
// its arguments in a form suitable for embedding in a URL query.
func URLQueryEscaper(args ...interface{}) string {
	return template.URLQueryEscaper(args...)
}
