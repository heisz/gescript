/*
 * Test methods for the lexical scanner.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"testing"
)

func TestComments(tst *testing.T) {
	var lval symType
	lex := newLexer(
		"  // This is a single line comment\r\n" +
			"  /* This is a single line multi-line comment */\n" +
			"\t/*\n" +
			"   * This is a multi-line comment...\r\n" +
			"   */\r\n" +
			"  // Note that there are tabs in here for whitespace scanning...")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for empty file of comments")
	}
	if lex.lineNumber != 6 {
		tst.Fatalf("Incorrect eof line number for comments: %d", lex.lineNumber)
	}
}

func TestKeywords(tst *testing.T) {
	var lval symType
	lex := newLexer("\tif\nvar while")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_IF) || (err != nil) {
		tst.Fatalf("Failed to parse IF token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_VAR) || (err != nil) {
		tst.Fatalf("Failed to parse VAR token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_WHILE) || (err != nil) {
		tst.Fatalf("Failed to parse WHILE token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestMixedWords(tst *testing.T) {
	var lval symType
	lex := newLexer("\tif\nabcd while")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_IF) || (err != nil) {
		tst.Fatalf("Failed to parse IF token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_IDENTIFIER) || (err != nil) ||
		(lval.identifier != "abcd") {
		tst.Fatalf("Failed to parse ABCD identifier")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_WHILE) || (err != nil) {
		tst.Fatalf("Failed to parse WHILE token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestNumerics(tst *testing.T) {
	var lval symType
	lex := newLexer("0 12a 007a 0x0aBc 0xyz 123456789")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 0) {
		tst.Fatalf("Failed to parse integer 0 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 12) {
		tst.Fatalf("Failed to parse integer 12 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_IDENTIFIER) || (err != nil) ||
		(lval.identifier != "a") {
		tst.Fatalf("Failed to parse stray a identifier")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 7) {
		tst.Fatalf("Failed to parse octal 7 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_IDENTIFIER) || (err != nil) ||
		(lval.identifier != "a") {
		tst.Fatalf("Failed to parse stray a identifier")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 2748) {
		tst.Fatalf("Failed to parse hex 0aBc token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 0) {
		tst.Fatalf("Failed to parse invalid hex leader token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_IDENTIFIER) || (err != nil) ||
		(lval.identifier != "xyz") {
		tst.Fatalf("Failed to parse errant hex identifier")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(int64) != 123456789) {
		tst.Fatalf("Failed to parse larger integer token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("0. 12.0 1.234e3")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(float64) != 0.0) {
		tst.Fatalf("Failed to parse float 0.0 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(float64) != 12.0) {
		tst.Fatalf("Failed to parse float 12.0 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(float64) != 1234.0) {
		tst.Fatalf("Failed to parse float 1.234e3 token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestStrings(tst *testing.T) {
	var lval symType
	lex := newLexer("'abc' \"def\"")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(string) != "abc") {
		tst.Fatalf("Failed to parse 'abc' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(string) != "def") {
		tst.Fatalf("Failed to parse \"def\" token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("'\\b\\f\\n\\t\\v\\\\\\'\\\"\\\r\n\\\na'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(string) != "\b\f\n\t\v\\'\"  a") {
		tst.Fatalf("Failed to parse '<mega-escape>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("'J\\x65ff \\X77u\\x7A h\\145re\\41'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(string) != "Jeff wuz here!") {
		tst.Fatalf("Failed to parse '<hex/octal>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	// This has native UTF-8, Golang encoded and parse encoded
	lex = newLexer("'Τ\u03B6\\u03B5φ'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_LITERAL) || (err != nil) ||
		(lval.literal.Native().(string) != "Τζεφ") {
		tst.Fatalf("Failed to parse '<unicode>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestTemplates(tst *testing.T) {
	var lval symType
	lex := newLexer("`abc` `de\nf`")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_TEMPLATE) || (err != nil) ||
		(lval.literal.Native().(string) != "abc") {
		tst.Fatalf("Failed to parse `abc` token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_TEMPLATE) || (err != nil) ||
		(lval.literal.Native().(string) != "de\nf") {
		tst.Fatalf("Failed to parse `de\\nf` token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestRegex(tst *testing.T) {
	var lval symType
	lex := newLexer("/abc/ if /def/")

	lex.regexValid = true
	tok, err := lex.lex(&lval)
	if (tok != GTOK_REGEXP) || (err != nil) {
		tst.Fatalf("Failed to parse regex instance")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_IF) || (err != nil) {
		tst.Fatalf("Failed to parse if token")
	}
	lex.regexValid = true
	tok, err = lex.lex(&lval)
	if (tok != GTOK_REGEXP) || (err != nil) {
		tst.Fatalf("Failed to parse regex instance")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestOperators(tst *testing.T) {
	var lval symType
	lex := newLexer("=> { } ( ) [ ] . ... ; , ~ ? :")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_ARROW) || (err != nil) {
		tst.Fatalf("Failed to parse '=>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LC) || (err != nil) {
		tst.Fatalf("Failed to parse '{' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_RC) || (err != nil) {
		tst.Fatalf("Failed to parse '}' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LP) || (err != nil) {
		tst.Fatalf("Failed to parse '(' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_RP) || (err != nil) {
		tst.Fatalf("Failed to parse ')' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LB) || (err != nil) {
		tst.Fatalf("Failed to parse '[' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_RB) || (err != nil) {
		tst.Fatalf("Failed to parse ']' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_DOT) || (err != nil) {
		tst.Fatalf("Failed to parse '.' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ELLIPSIS) || (err != nil) {
		tst.Fatalf("Failed to parse '...' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_SEMI) || (err != nil) {
		tst.Fatalf("Failed to parse ';' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_COMMA) || (err != nil) {
		tst.Fatalf("Failed to parse ',' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_TILDE) || (err != nil) {
		tst.Fatalf("Failed to parse '~' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_QMARK) || (err != nil) {
		tst.Fatalf("Failed to parse '?' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_COLON) || (err != nil) {
		tst.Fatalf("Failed to parse ':' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("< <= << <<= > >= >> >>= >>> >>>=")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_LT) || (err != nil) {
		tst.Fatalf("Failed to parse '<' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LTEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '<=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_LTLT) || (err != nil) {
		tst.Fatalf("Failed to parse '<<' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_LTLT) || (err != nil) {
		tst.Fatalf("Failed to parse '<<=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_GT) || (err != nil) {
		tst.Fatalf("Failed to parse '>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_GTEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '>=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_GTGT) || (err != nil) {
		tst.Fatalf("Failed to parse '>>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_GTGT) || (err != nil) {
		tst.Fatalf("Failed to parse '>>=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_GTGTGT) || (err != nil) {
		tst.Fatalf("Failed to parse '>>>' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_GTGTGT) ||
		(err != nil) {
		tst.Fatalf("Failed to parse '>>>=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("= == === =~ ! != !== !~")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGN) || (err != nil) {
		tst.Fatalf("Failed to parse '=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EQEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '==' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EQEQEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '===' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_REGEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '=~' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_NOT) || (err != nil) {
		tst.Fatalf("Failed to parse '!' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_NOTEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '!=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_NOTEQEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '!==' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_REGNOTEQ) || (err != nil) {
		tst.Fatalf("Failed to parse '!~' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("+ ++ += - -- -= * *= / /=")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ADD) || (err != nil) {
		tst.Fatalf("Failed to parse '+' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_INCR) || (err != nil) {
		tst.Fatalf("Failed to parse '++' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_ADD) || (err != nil) {
		tst.Fatalf("Failed to parse '+=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_SUB) || (err != nil) {
		tst.Fatalf("Failed to parse '-' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_DECR) || (err != nil) {
		tst.Fatalf("Failed to parse '--' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_SUB) || (err != nil) {
		tst.Fatalf("Failed to parse '-=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_MULT) || (err != nil) {
		tst.Fatalf("Failed to parse '*' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_MULT) || (err != nil) {
		tst.Fatalf("Failed to parse '*=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_DIV) || (err != nil) {
		tst.Fatalf("Failed to parse '/' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_DIV) || (err != nil) {
		tst.Fatalf("Failed to parse '/=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("% %= & && &= | || |= ^ ^=")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_MOD) || (err != nil) {
		tst.Fatalf("Failed to parse '%%' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_MOD) || (err != nil) {
		tst.Fatalf("Failed to parse '%%=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_AND) || (err != nil) {
		tst.Fatalf("Failed to parse '&' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ANDAND) || (err != nil) {
		tst.Fatalf("Failed to parse '&&' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_AND) || (err != nil) {
		tst.Fatalf("Failed to parse '&=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_OR) || (err != nil) {
		tst.Fatalf("Failed to parse '|' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_OROR) || (err != nil) {
		tst.Fatalf("Failed to parse '||' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_OR) || (err != nil) {
		tst.Fatalf("Failed to parse '|=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_XOR) || (err != nil) {
		tst.Fatalf("Failed to parse '^' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ASSIGNOP) || (lval.assignOp != GTOK_XOR) || (err != nil) {
		tst.Fatalf("Failed to parse '^=' token")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_EOF) || (err != nil) {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestErrors(tst *testing.T) {
	var lval symType
	lex := newLexer("/* This is a comment that never ends...")

	tok, err := lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for unterminated comment")
	}

	lex = newLexer("9999999999999999999999999999999999999999999999999999999999")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for integer overflow")
	}

	lex = newLexer("9.9.e1.2")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for invalid float")
	}

	lex = newLexer("`This is a string that never ends...")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for unterminated template")
	}

	lex = newLexer("'This is a string that never ends...")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for unterminated string")
	}

	lex = newLexer("'This is a string that\nsorta ends'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for unescaped newline in string")
	}

	lex = newLexer("'Invalid hex \\xaz'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for invalid hex character sequence")
	}

	lex = newLexer("'Invalid Unicode \\u000z'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for invalid Unicode character sequence")
	}

	lex = newLexer("'Invalid escape \\z'")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for invalid escape character sequence")
	}

	lex = newLexer("# abc")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for invalid token")
	}

	lex = newLexer("case /abc")

	tok, err = lex.lex(&lval)
	if (tok != GTOK_CASE) || (err != nil) {
		tst.Fatalf("Failed to read case leader (enable regex)")
	}
	tok, err = lex.lex(&lval)
	if (tok != GTOK_ERROR) || (err == nil) {
		tst.Fatalf("Failed to error for unterminated regex")
	}
}
