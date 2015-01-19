// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Example-based syntax error messages.
// See bisonerrors, Makefile, go.y.

static struct {
	int yystate;
	int yychar;
	char *msg;
} yymsg[] = {
	// Each line of the form % token list
	// is converted by bisonerrors into the yystate and yychar caused
	// by that token list.

	{240, ',',
	"unexpected comma during import block"},

	{42, ';',
	"missing import path; require quoted string"},

	{402, ';',
	"missing { after if clause"},

	{423, ';',
	"missing { after switch clause"},

	{257, ';',
	"missing { after for clause"},

	{504, LBODY,
	"missing { after for clause"},

	{28, '{',
	"unexpected semicolon or newline before {"},

	{159, ';',
	"unexpected semicolon or newline in type declaration"},

	{47, '}',
	"unexpected } in channel type"},
	
	{47, ')',
	"unexpected ) in channel type"},
	
	{47, ',',
	"unexpected comma in channel type"},

	{463, LELSE,
	"unexpected semicolon or newline before else"},

	{277, ',',
	"name list not allowed in interface type"},

	{257, LVAR,
	"var declaration not allowed in for initializer"},

	{75, '{',
	"unexpected { at end of statement"},

	{401, '{',
	"unexpected { at end of statement"},
	
	{140, ';',
	"argument to go/defer must be function call"},
	
	{450, ';',
	"need trailing comma before newline in composite literal"},
	
	{461, ';',
	"need trailing comma before newline in composite literal"},
	
	{127, LNAME,
	"nested func not allowed"},

	{685, ';',
	"else must be followed by if or statement block"}
};
