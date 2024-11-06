/*
 * Test methods for the parser implementation (parse-side only).
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"testing"
)

func TestEmptyWithComments(tst *testing.T) {
	lex := newLexer(
		"  // This is a single line comment\r\n" +
			"  /* This is a single line multi-line comment */\n" +
			"\t/*\n" +
			"   * This is a multi-line comment...\r\n" +
			"   */\r\n" +
			"  // Note that there are tabs in here for whitespace scanning...")

	prs := gesParse(lex)
	if prs != 0 {
		tst.Fatalf("Unexpected error on parse of empty")
	}
}
