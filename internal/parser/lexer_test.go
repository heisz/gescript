/*
 * Test methods for the lexical scanner.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"testing"
)

func TestComments(tst *testing.T) {
	var lval gesSymType
	lex := newLexer(
		"  // This is a single line comment\r\n" +
			"  /* This is a single line multi-line comment */\n" +
			"\t/*\n" +
			"   * This is a multi-line comment...\r\n" +
			"   */\r\n" +
			"  // Note that there are tabs in here for whitespace scanning...")

	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for empty file of comments")
	}
	if lex.lineNumber != 6 {
		tst.Fatalf("Incorrect eof line number for comments: %d", lex.lineNumber)
	}
}

func TestKeywords(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("\tif\nvar while")

	if lex.Lex(&lval) != GTOK_IF {
		tst.Fatalf("Failed to parse IF token")
	}
	if lex.Lex(&lval) != GTOK_VAR {
		tst.Fatalf("Failed to parse VAR token")
	}
	if lex.Lex(&lval) != GTOK_WHILE {
		tst.Fatalf("Failed to parse WHILE token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestMixedWords(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("\tif\nabcd while")

	if lex.Lex(&lval) != GTOK_IF {
		tst.Fatalf("Failed to parse IF token")
	}
	if (lex.Lex(&lval) != GTOK_IDENTIFIER) || (lval.identifier != "abcd") {
		tst.Fatalf("Failed to parse ABCD identifier")
	}
	if lex.Lex(&lval) != GTOK_WHILE {
		tst.Fatalf("Failed to parse WHILE token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestNumerics(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("0 12a 007a 0x0aBc 0xyz 123456789")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 0) {
		tst.Fatalf("Failed to parse integer 0 token")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 12) {
		tst.Fatalf("Failed to parse integer 12 token")
	}
	if (lex.Lex(&lval) != GTOK_IDENTIFIER) || (lval.identifier != "a") {
		tst.Fatalf("Failed to parse stray a identifier")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 7) {
		tst.Fatalf("Failed to parse octal 7 token")
	}
	if (lex.Lex(&lval) != GTOK_IDENTIFIER) || (lval.identifier != "a") {
		tst.Fatalf("Failed to parse stray a identifier")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 2748) {
		tst.Fatalf("Failed to parse hex 0aBc token")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 0) {
		tst.Fatalf("Failed to parse invalid hex leader token")
	}
	if (lex.Lex(&lval) != GTOK_IDENTIFIER) || (lval.identifier != "xyz") {
		tst.Fatalf("Failed to parse errant hex identifier")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(int64) != 123456789) {
		tst.Fatalf("Failed to parse larger integer token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("0. 12.0 1.234e3")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(float64) != 0.0) {
		tst.Fatalf("Failed to parse float 0.0 token")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(float64) != 12.0) {
		tst.Fatalf("Failed to parse float 12.0 token")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(float64) != 1234.0) {
		tst.Fatalf("Failed to parse float 1.234e3 token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestStrings(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("'abc' \"def\"")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(string) != "abc") {
		tst.Fatalf("Failed to parse 'abc' token")
	}
	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(string) != "def") {
		tst.Fatalf("Failed to parse \"def\" token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("'\\b\\f\\n\\t\\v\\\\\\'\\\"\\\r\n\\\na'")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(string) != "\b\f\n\t\v\\'\"  a") {
		tst.Fatalf("Failed to parse '<mega-escape>' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("'J\\x65ff \\X77u\\x7A h\\145re\\41'")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(string) != "Jeff wuz here!") {
		tst.Fatalf("Failed to parse '<hex/octal>' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	// This has native UTF-8, Golang encoded and parse encoded
	lex = newLexer("'Τ\u03B6\\u03B5φ'")

	if (lex.Lex(&lval) != GTOK_LITERAL) ||
		(lval.literal.Native().(string) != "Τζεφ") {
		tst.Fatalf("Failed to parse '<unicode>' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestRegex(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("/abc/ if /def/")
	lex.regexValid = true

	if lex.Lex(&lval) != GTOK_REGEXP {
		tst.Fatalf("Failed to parse regex instance")
	}
	if lex.Lex(&lval) != GTOK_IF {
		tst.Fatalf("Failed to parse if token")
	}
	lex.regexValid = true
	if lex.Lex(&lval) != GTOK_REGEXP {
		tst.Fatalf("Failed to parse regex instance")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestOperators(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("{ } ( ) [ ] . ; , ~ ? :")

	if lex.Lex(&lval) != GTOK_LC {
		tst.Fatalf("Failed to parse '{' token")
	}
	if lex.Lex(&lval) != GTOK_RC {
		tst.Fatalf("Failed to parse '}' token")
	}
	if lex.Lex(&lval) != GTOK_LP {
		tst.Fatalf("Failed to parse '(' token")
	}
	if lex.Lex(&lval) != GTOK_RP {
		tst.Fatalf("Failed to parse ')' token")
	}
	if lex.Lex(&lval) != GTOK_LB {
		tst.Fatalf("Failed to parse '[' token")
	}
	if lex.Lex(&lval) != GTOK_RB {
		tst.Fatalf("Failed to parse ']' token")
	}
	if lex.Lex(&lval) != GTOK_DOT {
		tst.Fatalf("Failed to parse '.' token")
	}
	if lex.Lex(&lval) != GTOK_SEMI {
		tst.Fatalf("Failed to parse ';' token")
	}
	if lex.Lex(&lval) != GTOK_COMMA {
		tst.Fatalf("Failed to parse ',' token")
	}
	if lex.Lex(&lval) != GTOK_TILDE {
		tst.Fatalf("Failed to parse '~' token")
	}
	if lex.Lex(&lval) != GTOK_QMARK {
		tst.Fatalf("Failed to parse '?' token")
	}
	if lex.Lex(&lval) != GTOK_COLON {
		tst.Fatalf("Failed to parse ':' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("< <= << <<= > >= >> >>= >>> >>>=")

	if lex.Lex(&lval) != GTOK_LT {
		tst.Fatalf("Failed to parse '<' token")
	}
	if lex.Lex(&lval) != GTOK_LTEQ {
		tst.Fatalf("Failed to parse '<=' token")
	}
	if lex.Lex(&lval) != GTOK_LTLT {
		tst.Fatalf("Failed to parse '<<' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_LTLT {
		tst.Fatalf("Failed to parse '<<=' token")
	}
	if lex.Lex(&lval) != GTOK_GT {
		tst.Fatalf("Failed to parse '>' token")
	}
	if lex.Lex(&lval) != GTOK_GTEQ {
		tst.Fatalf("Failed to parse '>=' token")
	}
	if lex.Lex(&lval) != GTOK_GTGT {
		tst.Fatalf("Failed to parse '>>' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_GTGT {
		tst.Fatalf("Failed to parse '>>=' token")
	}
	if lex.Lex(&lval) != GTOK_GTGTGT {
		tst.Fatalf("Failed to parse '>>>' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_GTGTGT {
		tst.Fatalf("Failed to parse '>>>=' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("= == === =~ ! != !== !~")

	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != 0 {
		tst.Fatalf("Failed to parse '=' token")
	}
	if lex.Lex(&lval) != GTOK_EQEQ {
		tst.Fatalf("Failed to parse '==' token")
	}
	if lex.Lex(&lval) != GTOK_EQEQEQ {
		tst.Fatalf("Failed to parse '===' token")
	}
	if lex.Lex(&lval) != GTOK_REGEQ {
		tst.Fatalf("Failed to parse '=~' token")
	}
	if lex.Lex(&lval) != GTOK_NOT {
		tst.Fatalf("Failed to parse '!' token")
	}
	if lex.Lex(&lval) != GTOK_NOTEQ {
		tst.Fatalf("Failed to parse '!=' token")
	}
	if lex.Lex(&lval) != GTOK_NOTEQEQ {
		tst.Fatalf("Failed to parse '!==' token")
	}
	if lex.Lex(&lval) != GTOK_REGNOTEQ {
		tst.Fatalf("Failed to parse '!~' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("+ ++ += - -- -= * *= / /=")

	if lex.Lex(&lval) != GTOK_ADD {
		tst.Fatalf("Failed to parse '+' token")
	}
	if lex.Lex(&lval) != GTOK_INCR {
		tst.Fatalf("Failed to parse '++' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_ADD {
		tst.Fatalf("Failed to parse '+=' token")
	}
	if lex.Lex(&lval) != GTOK_SUB {
		tst.Fatalf("Failed to parse '-' token")
	}
	if lex.Lex(&lval) != GTOK_DECR {
		tst.Fatalf("Failed to parse '--' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_SUB {
		tst.Fatalf("Failed to parse '-=' token")
	}
	if lex.Lex(&lval) != GTOK_MULT {
		tst.Fatalf("Failed to parse '*' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_MULT {
		tst.Fatalf("Failed to parse '*=' token")
	}
	if lex.Lex(&lval) != GTOK_DIV {
		tst.Fatalf("Failed to parse '/' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_DIV {
		tst.Fatalf("Failed to parse '/=' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}

	lex = newLexer("% %= & && &= | || |= ^ ^=")

	if lex.Lex(&lval) != GTOK_MOD {
		tst.Fatalf("Failed to parse '%%' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_MOD {
		tst.Fatalf("Failed to parse '%%=' token")
	}
	if lex.Lex(&lval) != GTOK_AND {
		tst.Fatalf("Failed to parse '&' token")
	}
	if lex.Lex(&lval) != GTOK_ANDAND {
		tst.Fatalf("Failed to parse '&&' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_AND {
		tst.Fatalf("Failed to parse '&=' token")
	}
	if lex.Lex(&lval) != GTOK_OR {
		tst.Fatalf("Failed to parse '|' token")
	}
	if lex.Lex(&lval) != GTOK_OROR {
		tst.Fatalf("Failed to parse '||' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_OR {
		tst.Fatalf("Failed to parse '|=' token")
	}
	if lex.Lex(&lval) != GTOK_XOR {
		tst.Fatalf("Failed to parse '^' token")
	}
	if lex.Lex(&lval) != GTOK_ASSIGNOP || lval.assignOp != GTOK_XOR {
		tst.Fatalf("Failed to parse '^=' token")
	}
	if lex.Lex(&lval) != 0 {
		tst.Fatalf("Invalid Lex return for eof")
	}
}

func TestErrors(tst *testing.T) {
	var lval gesSymType
	lex := newLexer("/* This is a comment that never ends...")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for unterminated comment")
	}

	lex = newLexer("9999999999999999999999999999999999999999999999999999999999")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for integer overflow")
	}

	lex = newLexer("9.9.e1.2")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for invalid float")
	}

	lex = newLexer("'This is a string that never ends...")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for unterminated string")
	}

	lex = newLexer("'Invalid hex \\xaz'")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for invalid hex character sequence")
	}

	lex = newLexer("'Invalid Unicode \\u000z'")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for invalid Unicode character sequence")
	}

	lex = newLexer("'Invalid escape \\z'")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for invalid escape character sequence")
	}

	lex = newLexer("# abc")

	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for invalid token")
	}

	lex = newLexer("case /abc")

	if lex.Lex(&lval) != GTOK_CASE {
		tst.Fatalf("Failed to read case leader (enable regex)")
	}
	if lex.Lex(&lval) != GTOK_ERROR {
		tst.Fatalf("Failed to error for unterminated regex")
	}
}
