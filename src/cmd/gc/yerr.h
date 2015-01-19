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

	{695, ',',
	"unexpected comma during import block"},

	{669, ';',
	"missing import path; require quoted string"},

	{465, ';',
	"missing { after if clause"},

	{485, ';',
	"missing { after switch clause"},

	{343, ';',
	"missing { after for clause"},

	{538, LBODY,
	"missing { after for clause"},

	{656, '{',
	"unexpected semicolon or newline before {"},

	{242, ';',
	"unexpected semicolon or newline in type declaration"},

	{113, '}',
	"unexpected } in channel type"},
	
	{113, ')',
	"unexpected ) in channel type"},
	
	{113, ',',
	"unexpected comma in channel type"},

	{305, LELSE,
	"unexpected semicolon or newline before else"},

	{360, ',',
	"name list not allowed in interface type"},

	{343, LVAR,
	"var declaration not allowed in for initializer"},

	{673, '{',
	"unexpected { at end of statement"},

	{169, '{',
	"unexpected { at end of statement"},
	
	{225, ';',
	"argument to go/defer must be function call"},
	
	{513, ';',
	"need trailing comma before newline in composite literal"},
	
	{524, ';',
	"need trailing comma before newline in composite literal"},
	
	{119, LNAME,
	"nested func not allowed"},

	{657, ';',
	"else must be followed by if or statement block"}
};
