// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

// This file defines the "precise" (token-accurate) comments interface.

import (
	"fmt"
	"reflect"
	"strings"
)

/*
   Precise comments

   This is a newer representation of comments that permits faithful
   recording of the relative order of tokens and comments
   (CommentGroup) within the syntax tree.

   Apart from doc comments on Decls, Specs, and Fields, all comments
   in the original comment mechanism are kept in a list in the root of
   the tree (File.Comments). This means a formatter must combine the
   tokens it finds from its traversal of the tree with the comments it
   finds from iterating the list, using position information to
   interleave them in the appropriate order.

   While this works well for trees produced by the parser, it makes it
   hard to modify syntax trees (as in a refactoring tools) because
   subtrees synthesized by the tool do not have position info, causing
   comments to be misplaced. Also, synthesized trees cannot themselves
   have comments.

   Even for trees produced by the parser, the correct token order is
   sometimes lost, causing comments to migrate during formatting. For
   example:

       f(/*a=* /1, /*b=* /2)

   is reformatted as:

       f( /*a=* / 1 /*b=* /, 2)

   The PreciseComments approach records all comments (each still
   represented as a CommentGroup) within the syntax tree, permitting a
   formatter to print tokens and comments in the correct order by
   traversing only the tree, without consulting position information.
   (A formatter may nonetheless exploit position information to
   preserve linebreaks in the source.)

   Conceptually, every Node struct has a number of new fields, each of
   type []*CommentGroup, one per place between tokens at which a list
   of comments may appear. Such places are called "Points". Points
   have descriptive names like Start, AfterColon,
   BeforeClosingBracket, and so on.

   For space efficiency, rather than one actual field per point, we
   use a single field per Node struct that holds a mapping from Points
   to []*CommentGroup. The mapping is concretely represented as
   a slice of (Point, *CommentGroup) pairs for compactness,
   and has type PreciseComments.

   Again for space efficiency, we add the minimal number of
   PreciseComments fields to nodes, exploiting the invariant that
   every Decl, Spec, Stmt, FieldList, and Expr (both terms and types)
   has a comment attachment point at its start and end, so there is no
   need to add redundant ones next to those nonterminals.

   For example, in the BinaryExpr "x + y", x and y are Exprs and thus
   may hold comments before and after, so there is no need for
   BinaryExpr to associate comments with the points at start or end,
   or before and after the "+" token:

      ◆ x ◆ + ◆ y ◆          =>      x + y

   In other words, BinaryExpr has no points, and therefore its
   point-to-comment mapping is always empty, and it does not need
   a *PreciseComments field.

   By contrast, the UnaryExpr "- x" may push comments surrounding the
   x operand down into the subexpression, but it must accommodate a
   comment appearing before the operator:

      ◆ - ◆ x ◆              =>      ◆ - x

   In other words, UnaryExpr has a single point, Start, and thus
   requires a PreciseComments field.

   When a syntax node has optional subtrees, a comment attachment
   point may be needed only in the case when the subtree is
   absent. Consider ArrayType, whose Len field is optional. In the
   first expression below, the comment can be attached to the literal
   "3", but in the second, it must be associated with the space after
   the "[" token.

      [/*len=* /3]int
      [/*no len* /]int        =>     ◆ [ Len ◆ ] Type

   This necessarily leads to redundancy in the representation: there
   may be a comment associated with the literal "3" and with the space
   after the "[" token. The parser will consistently produce trees
   that prefer one (unspecified) point over another, but programmatically
   constructed syntax trees may have comments in both. A formatter
   must print comments from all the points in the appropriate order,
   even when there is apparent redundancy.

*/

// A PreciseComments is a mapping from a Point to a list of comments
// at that point.
//
// The set of valid points that a given PreciseComments may contain is
// determined by the syntax node to which the PreciseComments belongs.
//
// For example, an Ident node has points {Start, End}; these are the
// only possible points of comment attachment for that syntax
// node. When a formatter prints an Ident node, it will ignore List
// elements with other point values.
//
// For space efficiency, all key/value pairs of the mapping are
// combined in a single slice.  The order of elements of the same
// Point determines the order in which they are emitted by the
// formatter. The relative order of elements with different Points is
// immaterial.
type PreciseComments struct {
	List []PreciseComment
}

// At returns the slice of comments attached at point.
// It may alias the array of pcs.List, so callers should not modify it.
func (pcs *PreciseComments) At(point Point) (res []PreciseComment) {
	if pcs != nil {
		for _, pc := range pcs.List {
			if pc.Point == point {
				// TODO(adonovan): opt: if all result elements are contiguous,
				// return a slice of the original array to avoid allocation.
				res = append(res, pc)
			}
		}
	}
	return res
}

// NodePreciseComments returns the PreciseComments field of node n, if any.
func NodePreciseComments(n Node) *PreciseComments {
	f := reflect.ValueOf(n).Elem().FieldByName("PreciseComments")
	if !f.IsValid() {
		return nil
	}
	return f.Interface().(*PreciseComments)
}

// A PreciseComment represents a comment and the quality of its
// leading and trailing space (none, space, newline, blank line).
//
// TODO(adonovan): implement Space. Can we move it into CommentGroup?
type PreciseComment struct {
	Point   Point         // see "Points"
	Space   uint8         // kind of whitespace before/after this comment
	Comment *CommentGroup // one or more comments
}

func (pc PreciseComment) String() string {
	return fmt.Sprintf("%s:/*%s*/", pc.Point, strings.TrimSpace(pc.Comment.Text()))
}

// A Point represents a point in the syntax tree between two tokens,
// where comments may be attached. The set of Points that are allowed
// in a PreciseComments and their exact interpretation depend on
// the type of the Node that holds it.
type Point uint8

const (
	PointStart Point = iota
	PointEnd
	PointAfterInitialToken
	PointBeforeOpenBracket // bracket may be '(', '[', or '{', depending on context
	PointAfterOpenBracket
	PointBeforeCloseBracket
	PointAfterCloseBracket
	PointBeforeColon         // SliceExpr only
	PointAfterColon          // {Case,Comm}Clause only
	PointAfterFirstSemicolon // ForStmt only
	PointBeforeRange         // RangeStmt only
)

var pointName = [...]string{
	PointStart:               "Start",
	PointEnd:                 "End",
	PointAfterInitialToken:   "AfterInitialToken",
	PointBeforeOpenBracket:   "BeforeOpenBracket",
	PointAfterOpenBracket:    "AfterOpenBracket",
	PointBeforeCloseBracket:  "BeforeCloseBracket",
	PointAfterCloseBracket:   "AfterCloseBracket",
	PointBeforeColon:         "BeforeColon",
	PointAfterColon:          "AfterColon",
	PointAfterFirstSemicolon: "AfterFirstSemicolon",
	PointBeforeRange:         "BeforeRange",
}

func (p Point) String() string { return pointName[p] }

// AddComment adds a comment before or after the specified node.
//
// The c.Point field is ignored; the copy of it retained in the tree
// will have its Point field set apropriately.
//
// The Node must be first-class syntax: AddComment panics if c is a
// Comment, CommentGroup, or Package.
func AddComment(n Node, c PreciseComment, before bool) {
	pcPtr, point := attachmentPoint(n, before)
	pc := *pcPtr
	if pc == nil {
		pc = &PreciseComments{}
		*pcPtr = pc
	}
	c.Point = point
	pc.List = append(pc.List, c)
}

// GetComments returns a list of comments attached before or after the specified node.
//
// TODO(adonovan): because there is redundancy in the representation,
// the list of comments may come from more than one point. Consider a
// file consisting of "package main /*a*/ /*b*/": the comments could
// be associated with the End point of the Ident "main", or with the
// End point of the File, or one of each. Perhaps GetComments needs to
// iterate over the list of possibilities?
func GetComments(n Node, before bool) []PreciseComment {
	pcPtr, point := attachmentPoint(n, before)
	pc := *pcPtr
	if pc == nil {
		return nil
	}
	return pc.At(point) // note: array may alias AST
}

// attachmentPoint returns the address of the *PreciseComments field
// in the given subtree of n that is the appropriate attachment point
// for a comment before or after that node, along with its Point. The
// value of the field may be nil if it has not yet been populated.
func attachmentPoint(n Node, before bool) (**PreciseComments, Point) {
	switch n := n.(type) {
	case *Field:
		if before {
			if len(n.Names) > 0 {
				return attachmentPoint(n.Names[0], before)
			} else {
				return attachmentPoint(n.Type, before)
			}
		} else {
			if n.Tag != nil {
				return attachmentPoint(n.Tag, before)
			} else {
				return attachmentPoint(n.Type, before)
			}
		}

	case *FieldList:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *BadExpr:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *Ident:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *BasicLit:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *Ellipsis:
		if before {
			return &n.PreciseComments, PointStart
		} else if n.Elt != nil {
			return attachmentPoint(n.Elt, false)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *FuncLit:
		return attachmentPoint(cond[Node](before, n.Type, n.Body), before)

	case *CompositeLit:
		if before {
			if n.Type != nil {
				return attachmentPoint(n.Type, before)
			} else {
				return &n.PreciseComments, PointStart
			}
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *ParenExpr:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *SelectorExpr:
		return attachmentPoint(cond[Node](before, n.X, n.Sel), before)

	case *IndexExpr:
		if before {
			return attachmentPoint(n.X, before)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *IndexListExpr:
		if before {
			return attachmentPoint(n.X, before)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *SliceExpr:
		if before {
			return attachmentPoint(n.X, before)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *TypeAssertExpr:
		if before {
			return attachmentPoint(n.X, before)
		} else {
			return &n.PreciseComments, PointAfterCloseBracket
		}

	case *CallExpr:
		if before {
			return attachmentPoint(n.Fun, before)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *StarExpr:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.X, before)
		}

	case *UnaryExpr:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.X, before)
		}

	case *BinaryExpr:
		return attachmentPoint(cond(before, n.X, n.Y), before)

	case *KeyValueExpr:
		return attachmentPoint(cond(before, n.Key, n.Value), before)

	case *ArrayType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Elt, before)
		}

	case *StructType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Fields, before)
		}

	case *FuncType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(cond[Node](n.Results != nil, n.Results, n.Params), before)
		}

	case *InterfaceType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Methods, before)
		}

	case *MapType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Value, before)
		}

	case *ChanType:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Value, before)
		}

	case *BadStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *DeclStmt:
		return attachmentPoint(n.Decl, before)

	case *EmptyStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *LabeledStmt:
		return attachmentPoint(cond[Node](before, n.Label, n.Stmt), before)

	case *ExprStmt:
		return attachmentPoint(n.X, before)

	case *SendStmt:
		return attachmentPoint(cond(before, n.Chan, n.Value), before)

	case *IncDecStmt:
		if before {
			return attachmentPoint(n.X, before)
		} else {
			return &n.PreciseComments, PointEnd
		}

	case *AssignStmt:
		return attachmentPoint(cond(before, n.Lhs[0], n.Rhs[len(n.Rhs)-1]), before)

	case *GoStmt:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Call, before)
		}

	case *DeferStmt:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(n.Call, before)
		}

	case *ReturnStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *BranchStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *BlockStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *IfStmt:
		if before {
			return &n.PreciseComments, PointStart
		} else {
			return attachmentPoint(cond[Node](n.Else != nil, n.Else, n.Body), before)
		}

	case *CaseClause:
		if before {
			return &n.PreciseComments, PointStart
		} else if last := len(n.Body) - 1; last >= 0 {
			return attachmentPoint(n.Body[last], before)
		} else {
			return &n.PreciseComments, PointAfterColon
		}

	case *SwitchStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *TypeSwitchStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *CommClause:
		if before {
			return &n.PreciseComments, PointStart
		} else if last := len(n.Body) - 1; last >= 0 {
			return attachmentPoint(n.Body[last], before)
		} else {
			return &n.PreciseComments, PointAfterColon
		}

	case *SelectStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *ForStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *RangeStmt:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *ImportSpec:
		if before {
			return attachmentPoint(cond[Node](n.Name != nil, n.Name, n.Path), before)
		} else {
			return attachmentPoint(n.Path, before)
		}

	case *ValueSpec:
		if before {
			return attachmentPoint(n.Names[0], before)
		} else {
			if last := len(n.Values) - 1; last >= 0 {
				return attachmentPoint(n.Values[last], before)
			} else {
				return attachmentPoint(n.Type, before)
			}
		}

	case *TypeSpec:
		return attachmentPoint(cond[Node](before, n.Name, n.Type), before)

	case *BadDecl:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *GenDecl:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)

	case *FuncDecl:
		if before {
			return attachmentPoint(n.Type, before)
		} else if n.Body != nil {
			return attachmentPoint(cond[Node](n.Body != nil, n.Body, n.Type), before)
		}

	case *File:
		return &n.PreciseComments, cond(before, PointStart, PointEnd)
	}
	// includes *Comment, *CommentGroup, *Package
	panic(fmt.Sprintf("unexpected node type %T", n))
}

func cond[T any](cond bool, x, y T) T {
	if cond {
		return x
	} else {
		return y
	}
}

// TODO: Next steps:
// - test insertion API
// - test extraction API
// - benchmark CPU/real cost (enabled and disabled) on std.
//
// Final size distribution of List in non-empty pcs across std:
// 60%   = 1
// 97%  <= 2
// 99%  <= 3
// 100% <= 7
//
// Number of distinct Point values in a PCS slice: 1 (99%) 2(1%).
//
// Nodes without a PreciseComments field and their frequency in std:
// 329466	SelectorExpr (7%)
// 173976	BinaryExpr (3.8%)
// 146308	AssignStmt (3.2%)
// 127875	Field (2.8%)
// 115402	Comment
// 78510	KeyValueExpr
// 74106	ExprStmt
// 41876	CommentGroup
// 40666	ValueSpec
// 39992	FuncDecl
// 17160	DeclStmt
// 12050	ImportSpec
// 7365		TypeSpec
// 6993		FuncLit
// 821		SendStmt
// 430		LabeledStmt
// 1212996 total (27% of 4555851 total nodes)
//
// Comments are only about 2.5% and the actual objects for
// precisecomments seem barely measuraable.  so why not have it always
// on?
//
// We should consider increased size classes too:
// unfortunately Ident (39%) and BasicLit (11%) will
// bump up into the next class.
//
// std LoadSyntax, approx 94218 File.Comments groups (116K Comment nodes)
//
// heap = 198MB baseline go/packages LoadSyntax before Precisecomments CL
// heap = 232MB PreciseComments off (+17%)
// heap = 235MB PreciseComments on (a further +1.2%) -- so why not have it always on?
// head = 240MB PreciseComments off, with an unused *PreciseComments field
//              added to each of the node types above.
//              (Strangely adding a struct{} instead had the same effect.)

// TODO Measure CPU cost. Ensure the fast path mode check in parser.preciseComments is inlined.
