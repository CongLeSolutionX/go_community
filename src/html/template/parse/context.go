// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package parse defines the internal state of the HTML parser used
// to parse and escape templates as defined by html/template. These data
// structures are meant for clients to use when defining
// template.EscaperForContext functions.
package parse

import (
	"strings"
)

// Context describes the state an HTML parser must be in when it reaches the
// portion of HTML produced by evaluating a particular template node.
//
// The zero value of type Context is the start context for a template that
// produces an HTML fragment as defined at
// http://www.w3.org/TR/html5/syntax.html#the-end
// where the context element is null.
type Context struct {
	State   State
	Delim   Delim
	URLPart URLPart
	JSCtx   JSCtx
	Element Element
	Attr    Attr
	Err     *Error
	// ScriptType is the lowercase value of the "type" attribute inside the current "script"
	// element (see https://dev.w3.org/html5/spec-preview/the-script-element.html#attr-script-type).
	// This field will be empty if the parser is currently not in a script element,
	// the type attribute has not already been parsed in the current element, or if the
	// value of the type attribute cannot be determined at parse time.
	ScriptType string
	// LinkRel is the value of the "rel" attribute inside the current "link"
	// element (see https://html.spec.whatwg.org/multipage/semantics.html#attr-link-rel).
	// This value has been normalized to lowercase with exactly one space between tokens
	// and exactly one space at start and end, so that a lookup of any token foo can
	// be performed by searching for the substring " foo ".
	// This field will be empty if the parser is currently not in a link element,
	// the rel attribute has not already been parsed in the current element, or if the
	// value of the rel attribute cannot be determined at parse time.
	LinkRel string
}

// Eq returns whether Context c is equal to Context d.
func (c Context) Eq(d Context) bool {
	return c.State == d.State &&
		c.Delim == d.Delim &&
		c.URLPart == d.URLPart &&
		c.JSCtx == d.JSCtx &&
		c.Element.Eq(d.Element) &&
		c.Attr.Eq(d.Attr) &&
		c.Err == d.Err &&
		c.ScriptType == d.ScriptType &&
		c.LinkRel == d.LinkRel
}

// State describes a high-level HTML parser state.
//
// It bounds the top of the element stack, and by extension the HTML insertion
// mode, but also contains state that does not correspond to anything in the
// HTML5 parsing algorithm because a single token production in the HTML
// grammar may contain embedded actions in a template. For instance, the quoted
// HTML attribute produced by
//     <div title="Hello {{.World}}">
// is a single token in HTML's grammar but in a template spans several nodes.
type State uint8

//go:generate stringer -type State

const (
	// StateText is parsed character data. An HTML parser is in
	// this state when its parse position is outside an HTML tag,
	// directive, comment, and special element body.
	StateText State = iota
	// StateTag occurs before an HTML attribute or the end of a tag.
	StateTag
	// StateAttrName occurs inside an attribute name.
	// It occurs between the ^'s in ` ^name^ = value`.
	StateAttrName
	// StateAfterName occurs after an attr name has ended but before any
	// equals sign. It occurs between the ^'s in ` name^ ^= value`.
	StateAfterName
	// StateBeforeValue occurs after the equals sign but before the value.
	// It occurs between the ^'s in ` name =^ ^value`.
	StateBeforeValue
	// StateHTMLCmt occurs inside an <!-- HTML comment -->.
	StateHTMLCmt
	// StateRCDATA occurs inside an RCDATA element (<textarea> or <title>)
	// as described at http://www.w3.org/TR/html5/syntax.html#elements-0
	StateRCDATA
	// StateAttr occurs inside an HTML attribute whose content is text.
	StateAttr
	// StateURL occurs inside an HTML attribute whose content is a URL.
	StateURL
	// StateSrcset occurs inside an HTML srcset attribute.
	StateSrcset
	// StateJS occurs inside an event handler or script element.
	StateJS
	// StateJSDqStr occurs inside a JavaScript double quoted string.
	StateJSDqStr
	// StateJSSqStr occurs inside a JavaScript single quoted string.
	StateJSSqStr
	// StateJSRegexp occurs inside a JavaScript regexp literal.
	StateJSRegexp
	// StateJSBlockCmt occurs inside a JavaScript /* block comment */.
	StateJSBlockCmt
	// StateJSLineCmt occurs inside a JavaScript // line comment.
	StateJSLineCmt
	// StateCSS occurs inside a <style> element or style attribute.
	StateCSS
	// StateCSSDqStr occurs inside a CSS double quoted string.
	StateCSSDqStr
	// StateCSSSqStr occurs inside a CSS single quoted string.
	StateCSSSqStr
	// StateCSSDqURL occurs inside a CSS double quoted url("...").
	StateCSSDqURL
	// StateCSSSqURL occurs inside a CSS single quoted url('...').
	StateCSSSqURL
	// StateCSSURL occurs inside a CSS unquoted url(...).
	StateCSSURL
	// StateCSSBlockCmt occurs inside a CSS /* block comment */.
	StateCSSBlockCmt
	// StateCSSLineCmt occurs inside a CSS // line comment.
	StateCSSLineCmt
	// StateError is an infectious error state outside any valid
	// HTML/CSS/JS construct.
	StateError
)

// IsComment reports whether a state contains content meant for template
// authors & maintainers, not for end-users or machines.
func IsComment(s State) bool {
	switch s {
	case StateHTMLCmt, StateJSBlockCmt, StateJSLineCmt, StateCSSBlockCmt, StateCSSLineCmt:
		return true
	}
	return false
}

// IsInTag reports whether s occurs solely inside an HTML tag.
func IsInTag(s State) bool {
	switch s {
	case StateTag, StateAttrName, StateAfterName, StateBeforeValue, StateAttr:
		return true
	}
	return false
}

// Delim is the delimiter that will end the current HTML attribute.
type Delim uint8

//go:generate stringer -type Delim

const (
	// DelimNone occurs outside any attribute.
	DelimNone Delim = iota
	// DelimDoubleQuote occurs when a double quote (") closes the attribute.
	DelimDoubleQuote
	// DelimSingleQuote occurs when a single quote (') closes the attribute.
	DelimSingleQuote
	// DelimSpaceOrTagEnd occurs when a space or right angle bracket (>)
	// closes the attribute.
	DelimSpaceOrTagEnd
)

// URLPart identifies a part in an RFC 3986 hierarchical URL to allow different
// encoding strategies.
type URLPart uint8

//go:generate stringer -type URLPart

const (
	// URLPartNone occurs when not in a URL, or possibly at the start:
	// ^ in "^http://auth/path?k=v#frag".
	URLPartNone URLPart = iota
	// URLPartPreQuery occurs in the scheme, authority, or path; between the
	// ^s in "h^ttp://auth/path^?k=v#frag".
	URLPartPreQuery
	// URLPartQueryOrFrag occurs in the query portion between the ^s in
	// "http://auth/path?^k=v#frag^".
	URLPartQueryOrFrag
	// URLPartUnknown occurs due to joining of contexts both before and
	// after the query separator.
	URLPartUnknown
)

// JSCtx determines whether a '/' starts a regular expression literal or a
// division operator.
type JSCtx uint8

//go:generate stringer -type JSCtx

const (
	// JSCtxRegexp occurs where a '/' would start a regexp literal.
	JSCtxRegexp JSCtx = iota
	// JSCtxDivOp occurs where a '/' would start a division operator.
	JSCtxDivOp
	// JSCtxUnknown occurs where a '/' is ambiguous due to context joining.
	JSCtxUnknown
)

type Element struct {
	// Name is the lowercase name of the element. If context joining has occurred, Name
	// will be arbitrarily assigned the element name from one of the joined contexts.
	Name string
	// Names contains all possible names the element could assume because of context joining.
	// For example, after joining the contexts in the "if" and "else" branches of
	//     {{if .C}}<img{{else}}<audio{{end}} src="/some/path">`,
	// Names will contain "img" and "audio".
	// Names can also contain empty strings, which represent joined contexts with no element name.
	// Names will be empty if no context joining occured.
	Names []string
}

// Eq reports whether a and b have the same Name. All other fields are ignored.
func (e Element) Eq(d Element) bool {
	return e.Name == d.Name
}

// String returns the string representation of the Element.
func (e Element) String() string {
	return "Element" + strings.Title(e.Name)
}

// Attr represents the attribute that the parser is in, that is,
// starting from StateAttrName until StateTag/StateText (exclusive).
type Attr struct {
	// Name is the lowercase name of the attribute. If context joining has occurred, Name
	// will be arbitrarily assigned the attribute name from one of the joined contexts.
	Name string
	// Value is the value of the attribute. If context joining has occurred, Value
	// will be arbitrarily assigned the attribute value from one of the joined contexts.
	// If there are multiple actions in the attribute value, Value will contain the
	// concatenation of all values seen so far. For example, in
	//    <a name="foo{{.X}}bar{{.Y}}">
	// Value is "foo" at "{{.X}}" and "foobar" at "{{.Y}}".
	Value string
	// AmbiguousValue indicates whether Value contains an ambiguous value due to context-joining.
	AmbiguousValue bool
	// Names contains all possible names the attribute could assume because of context joining.
	// For example, after joining the contexts in the "if" and "else" branches of
	//     <a {{if .C}}title{{else}}name{{end}}="foo">
	// Names will contain "title" and "name".
	// Names can also contain empty strings, which represent joined contexts with no attribute name.
	// Names will be empty if no context joining occured.
	Names []string
}

// Eq reports whether a and b have the same Name. All other fields are ignored.
func (a Attr) Eq(b Attr) bool {
	return a.Name == b.Name
}

// String returns the string representation of the Attr.
func (a Attr) String() string {
	return "Attr" + strings.Title(a.Name)
}
