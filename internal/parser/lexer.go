/*
 * Lexical scanner to support the ECMAScript grammar.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

// Perform the automated yacc generation from this origin
//go:generate goyacc -l -o grammar.go -p ges grammar.y

import (
	"bytes"
	"strconv"

	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/types"
)

// Note: only supports Unicode in string literals (binary sequences)
type gesLex struct {
	// The first sections are for tracking the lexical context
	source     []byte
	offset     int
	lineNumber int
	regexValid bool
	error      *string

	// Second part carries the reference for the working parsing context
	ctx *parsingContext
}

// Initialize a new lexer instance, for test and parse wrapping
func newLexer(source string) *gesLex {
	lex := &gesLex{
		source:     append([]byte(source), 0),
		offset:     0,
		lineNumber: 1,
		regexValid: false,
		error:      nil,

		ctx: &parsingContext{
			body: &engine.Function{},
		},
	}

	return lex
}

var keywords = []struct {
	word  string
	token int
}{
	// Current reserved keywords according to the specification
	{"break", GTOK_BREAK},
	{"case", GTOK_CASE},
	{"catch", GTOK_CATCH},
	{"class", GTOK_CLASS},
	{"const", GTOK_CONST},
	{"continue", GTOK_CONTINUE},
	{"debugger", GTOK_DEBUGGER},
	{"default", GTOK_DEFAULT},
	{"delete", GTOK_DELETE},
	{"do", GTOK_DO},
	{"else", GTOK_ELSE},
	{"export", GTOK_EXPORT},
	{"extends", GTOK_EXTENDS},
	{"finally", GTOK_FINALLY},
	{"for", GTOK_FOR},
	{"function", GTOK_FUNCTION},
	{"if", GTOK_IF},
	{"import", GTOK_IMPORT},
	{"in", GTOK_IN},
	{"instanceof", GTOK_INSTANCEOF},
	{"let", GTOK_LET},
	{"new", GTOK_NEW},
	{"return", GTOK_RETURN},
	{"super", GTOK_SUPER},
	{"switch", GTOK_SWITCH},
	{"this", GTOK_THIS},
	{"throw", GTOK_THROW},
	{"try", GTOK_TRY},
	{"typeof", GTOK_TYPEOF},
	{"var", GTOK_VAR},
	{"void", GTOK_VOID},
	{"while", GTOK_WHILE},
	{"with", GTOK_WITH},
	{"yield", GTOK_YIELD},

	// Future reserved keywords (including strict mode)
	{"await", GTOK_AWAIT},
	{"enum", GTOK_ENUM},
	{"implements", GTOK_IMPLEMENTS},
	{"interface", GTOK_INTERFACE},
	{"package", GTOK_PACKAGE},
	{"private", GTOK_PRIVATE},
	{"protected", GTOK_PROTECTED},
	{"public", GTOK_PUBLIC},

	// Reserved words
	{"null", GTOK_NULL},
	{"true", GTOK_TRUE},
	{"false", GTOK_FALSE},
}

// Platform-independent hex handling
func isHex(ch byte) bool {
	if ((ch >= '0') && (ch <= '9')) ||
		((ch >= 'a') && (ch <= 'f')) ||
		((ch >= 'A') && (ch <= 'F')) {
		return true
	}

	return false
}

func hexToInt(ch byte) int64 {
	if (ch >= '0') && (ch <= '9') {
		return int64(ch) - 48
	}
	if (ch >= 'a') && (ch <= 'f') {
		return int64(ch) - 87
	}
	if (ch >= 'A') && (ch <= 'F') {
		return int64(ch) - 55
	}
	return 0
}

// Wrapped internal function parses the raw lexical element
func (ctx *gesLex) lex(lval *gesSymType) int {
	lval.parseType = PARSED_UNDEFINED
	for true {
		ch := ctx.source[ctx.offset]
		if ch == 0 {
			break
		}
		nch := ctx.source[ctx.offset+1]

		// Consume white space (non-Unicode)
		if (ch == ' ') || (ch == '\t') {
			ctx.offset++
			continue
		}
		if (ch == '\r') && (nch == '\n') {
			// Move to newline to avoid double counting
			ctx.offset++
			ch = ctx.source[ctx.offset]
		}
		if (ch == '\r') || (ch == '\n') {
			ctx.lineNumber++
			ctx.offset++
			continue
		}

		// Consume single line comment
		if (ch == '/') && (nch == '/') {
			ctx.offset += 2
			for (ctx.source[ctx.offset] != 0) &&
				(ctx.source[ctx.offset] != '\r') &&
				(ctx.source[ctx.offset] != '\n') {
				ctx.offset++
			}
			if (ctx.source[ctx.offset] == '\r') &&
				(ctx.source[ctx.offset+1] == '\n') {
				ctx.offset++
			}
			if ctx.source[ctx.offset] != 0 {
				ctx.lineNumber++
				ctx.offset++
			}
			continue
		}

		// And multi-line comment
		if (ch == '/') && (nch == '*') {
			ctx.offset += 2
			for (ctx.source[ctx.offset] != 0) &&
				((ctx.source[ctx.offset] != '*') ||
					(ctx.source[ctx.offset+1] != '/')) {
				if ctx.source[ctx.offset] == '\n' {
					ctx.lineNumber++
				}
				if ctx.source[ctx.offset] == '\r' {
					if ctx.source[ctx.offset+1] == '\n' {
						ctx.offset++
					}
					ctx.lineNumber++
				}
				ctx.offset++
			}
			if ctx.source[ctx.offset] == 0 {
				errstr := "Unterminated multi-line comment"
				ctx.error = &errstr
				return GTOK_ERROR
			} else {
				ctx.offset += 2
			}
			continue
		}

		// Identifer/keywords (NOTE: Unicode escape names not supported)
		if ((ch >= 'a') && (ch <= 'z')) || ((ch >= 'A') && (ch <= 'Z')) ||
			(ch == '$') || (ch == '_') {
			eso := ctx.offset
			for ctx.source[eso] != 0 {
				ch = ctx.source[eso]
				if ((ch < 'a') || (ch > 'z')) && ((ch < 'A') || (ch > 'Z')) &&
					((ch < '0') || (ch > '9')) && (ch != '$') && (ch != '_') {
					break
				}
				eso++
			}

			ln := eso - ctx.offset
			for _, keywd := range keywords {
				if (len(keywd.word) == ln) &&
					bytes.Equal(ctx.source[ctx.offset:eso],
						[]byte(keywd.word)) {
					ctx.offset = eso
					if keywd.token == GTOK_CASE {
						ctx.regexValid = true
					}
					return keywd.token
				}
			}

			// TODO - support macros?

			// Non-keyword identifier
			lval.parseType = PARSED_IDENTIFIER
			lval.identifier = string(ctx.source[ctx.offset:eso])
			ctx.offset = eso
			return GTOK_IDENTIFIER
		}

		// Numeric literals
		if ((ch >= '0') && (ch <= '9')) ||
			((ch == '.') && ((nch >= '0') && (nch <= '9'))) {
			ival := int64(0)

			// Assume decimal until told otherwise
			radix := 10
			if ch == '0' {
				ctx.offset++
				ch = ctx.source[ctx.offset]
				if (nch == 'x') || (nch == 'X') {
					ctx.offset++
					ch = ctx.source[ctx.offset]
					radix = 16
					ival = 0
				} else if (ch >= '0') && (ch <= '9') {
					radix = 8
					ival = 0
				}
			}

			// Process digits (potentially including hex)
			eso := ctx.offset
			for isHex(ch) {
				if radix == 16 {
					ival = (ival << 4) | hexToInt(ch)
				} else if radix == 8 {
					if (ch >= '0') && (ch <= '7') {
						ival = ival<<3 | (int64(ch) - 48)
					} else {
						break
					}
				} else {
					if (ch < '0') || (ch > '9') {
						break
					}
				}

				eso++
				ch = ctx.source[eso]
			}
			if eso == ctx.offset {
				if radix == 16 {
					// Looked like a hexidecimal but it isn't, rollback...
					ctx.offset -= 2
					eso--
					radix = 10
					ch = 'x'
				} else {
					// Just a plain old zero
					ctx.offset--
				}
			}

			if (radix == 8) || (radix == 16) {
				ctx.offset = eso
				lval.parseType = PARSED_LITERAL
				lval.literal = types.IntegerType(ival)
				return GTOK_LITERAL
			}
			if (ch != '.') && (ch != 'e') && (ch != 'E') {
				sval := string(ctx.source[ctx.offset:eso])
				ival, err := strconv.ParseInt(sval, 10, 64)
				if err != nil {
					errstr := "Invalid literal integer: " + err.Error()
					ctx.error = &errstr
					return GTOK_ERROR
				}
				lval.parseType = PARSED_LITERAL
				lval.literal = types.IntegerType(ival)
				ctx.offset = eso
				return GTOK_LITERAL
			}

			// Assume slightly well-behaved code...
			for ((ch >= '0') && (ch <= '9')) || (ch == '.') ||
				(ch == 'e') || (ch == 'E') || (ch == '-') {
				eso++
				ch = ctx.source[eso]
			}
			sval := string(ctx.source[ctx.offset:eso])
			dval, err := strconv.ParseFloat(sval, 64)
			if err != nil {
				errstr := "Invalid literal float: " + err.Error()
				ctx.error = &errstr
				return GTOK_ERROR
			}
			lval.parseType = PARSED_LITERAL
			lval.literal = types.NumberType(dval)
			ctx.offset = eso
			return GTOK_LITERAL
		}

		// String literals
		if (ch == '"') || (ch == '\'') {
			qch := ch
			eso := ctx.offset + 1
			ch = ctx.source[eso]

			str := []byte{}
			for ch != qch {
				if (ch == '\r') || (ch == '\n') || (ch == 0) {
					errstr := "Unterminated string literal"
					ctx.error = &errstr
					return GTOK_ERROR
				}
				if ch == '\\' {
					wch := ' '
					eso++
					ch = ctx.source[eso]
					switch ch {
					case 'b':
						wch = '\b'
						break
					case 'f':
						wch = '\f'
						break
					case 'n':
						wch = '\n'
						break
					case 't':
						wch = '\t'
						break
					case 'v':
						wch = '\v'
						break
					case '\\':
						wch = '\\'
						break
					case '\'':
						wch = '\''
						break
					case '"':
						wch = '"'
						break
					case '\r':
						if ctx.source[eso+1] == '\n' {
							eso++
						}
						fallthrough
					case '\n':
						// Escaped newlines in ECMA are discarded
						ctx.lineNumber++
						break
					case 'x':
						fallthrough
					case 'X':
						if (!isHex(ctx.source[eso+1])) ||
							(!isHex(ctx.source[eso+2])) {
							errstr := "Invalid hex character sequence"
							ctx.error = &errstr
							return GTOK_ERROR
						}
						wch = int32((hexToInt(ctx.source[eso+1]) << 4) |
							hexToInt(ctx.source[eso+2]))
						eso += 2
						break
					case 'u':
						fallthrough
					case 'U':
						if (!isHex(ctx.source[eso+1])) ||
							(!isHex(ctx.source[eso+2])) ||
							(!isHex(ctx.source[eso+3])) ||
							(!isHex(ctx.source[eso+4])) {
							errstr := "Invalid Unicode character sequence"
							ctx.error = &errstr
							return GTOK_ERROR
						}
						wch = int32((hexToInt(ctx.source[eso+1]) << 12) |
							(hexToInt(ctx.source[eso+2]) << 8) |
							(hexToInt(ctx.source[eso+3]) << 4) |
							hexToInt(ctx.source[eso+4]))
						eso += 4
						break
					default:
						if (ch >= '0') && (ch <= '7') {
							wch = int32(ch - 48)
							ch = ctx.source[eso+1]
							if (ch >= '0') && (ch <= '7') {
								eso++
								wch = (wch << 3) | int32(ch-48)
								ch = ctx.source[eso+1]
								if (ch >= '0') && (ch <= '7') {
									eso++
									wch = (wch << 3) | int32(ch-48)
								}
							}
						} else {
							errstr := "Invalid escape character sequence"
							ctx.error = &errstr
							return GTOK_ERROR
						}
					}
					str = append(str, []byte(string(wch))...)
				} else {
					str = append(str, ch)
				}

				eso++
				ch = ctx.source[eso]
			}
			eso++

			lval.parseType = PARSED_LITERAL
			lval.literal = types.StringType(string(str))
			ctx.offset = eso
			return GTOK_LITERAL
		}

		// Embedded regex
		if ctx.regexValid && (ch == '/') {
			ctx.offset++
			eso := ctx.offset
			ch = ctx.source[eso]
			for ch != '/' {
				if (ch == '\r') || (ch == '\n') || (ch == 0) {
					errstr := "Unterminated regex pattern"
					ctx.error = &errstr
					return GTOK_ERROR
				}
				if (ch == '\\') && (ctx.source[eso+1] != 0) {
					eso++
				}
				eso++
				ch = ctx.source[eso]
			}
			eso++

			/* TODO - compile with flags */
			lval.parseType = PARSED_REGEXP
			ctx.offset = eso
			return GTOK_REGEXP
		}

		// Ok, now we are down to the nitty gritty operators
		token := 0
		eso := ctx.offset + 1
		switch ch {
		case '{':
			token = GTOK_LC
			break
		case '}':
			token = GTOK_RC
			break
		case '(':
			token = GTOK_LP
			break
		case ')':
			token = GTOK_RP
			break
		case '[':
			token = GTOK_LB
			break
		case ']':
			token = GTOK_RB
			break
		case '.':
			token = GTOK_DOT
			break
		case ';':
			token = GTOK_SEMI
			break
		case ',':
			token = GTOK_COMMA
			break
		case '~':
			token = GTOK_TILDE
			break
		case '?':
			token = GTOK_QMARK
			break
		case ':':
			token = GTOK_COLON
			break

		case '<':
			if nch == '=' {
				eso++
				token = GTOK_LTEQ
			} else if nch == '<' {
				eso++
				nch = ctx.source[eso]
				if nch == '=' {
					eso++
					lval.assignOp = GTOK_LTLT
					token = GTOK_ASSIGNOP
				} else {
					token = GTOK_LTLT
				}
			} else {
				token = GTOK_LT
			}
			break
		case '>':
			if nch == '=' {
				eso++
				token = GTOK_GTEQ
			} else if nch == '>' {
				eso++
				nch = ctx.source[eso]
				if nch == '>' {
					eso++
					nch = ctx.source[eso]
					if nch == '=' {
						eso++
						lval.assignOp = GTOK_GTGTGT
						token = GTOK_ASSIGNOP
					} else {
						token = GTOK_GTGTGT
					}
				} else if nch == '=' {
					eso++
					lval.assignOp = GTOK_GTGT
					token = GTOK_ASSIGNOP
				} else {
					token = GTOK_GTGT
				}
			} else {
				token = GTOK_GT
			}
			break
		case '=':
			if nch == '=' {
				eso++
				nch = ctx.source[eso]
				if nch == '=' {
					eso++
					token = GTOK_EQEQEQ
				} else {
					token = GTOK_EQEQ
				}
			} else if nch == '~' {
				eso++
				ctx.regexValid = true
				token = GTOK_REGEQ
			} else {
				lval.assignOp = 0
				token = GTOK_ASSIGNOP
			}
			break
		case '!':
			if nch == '=' {
				eso++
				nch = ctx.source[eso]
				if nch == '=' {
					eso++
					token = GTOK_NOTEQEQ
				} else {
					token = GTOK_NOTEQ
				}
			} else if nch == '~' {
				eso++
				ctx.regexValid = true
				token = GTOK_REGNOTEQ
			} else {
				token = GTOK_NOT
			}
			break
		case '+':
			if nch == '+' {
				eso++
				token = GTOK_INCR
			} else if nch == '=' {
				eso++
				lval.assignOp = GTOK_ADD
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_ADD
			}
			break
		case '-':
			if nch == '-' {
				eso++
				token = GTOK_DECR
			} else if nch == '=' {
				eso++
				lval.assignOp = GTOK_SUB
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_SUB
			}
			break
		case '*':
			if nch == '=' {
				eso++
				lval.assignOp = GTOK_MULT
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_MULT
			}
			break
		case '/':
			if nch == '=' {
				eso++
				lval.assignOp = GTOK_DIV
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_DIV
			}
			break
		case '%':
			if nch == '=' {
				eso++
				lval.assignOp = GTOK_MOD
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_MOD
			}
			break
		case '&':
			if nch == '&' {
				eso++
				token = GTOK_ANDAND
			} else if nch == '=' {
				eso++
				lval.assignOp = GTOK_AND
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_AND
			}
			break
		case '|':
			if nch == '|' {
				eso++
				token = GTOK_OROR
			} else if nch == '=' {
				eso++
				lval.assignOp = GTOK_OR
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_OR
			}
			break
		case '^':
			if nch == '=' {
				eso++
				lval.assignOp = GTOK_XOR
				token = GTOK_ASSIGNOP
			} else {
				token = GTOK_XOR
			}
			break
		}
		if token != 0 {
			ctx.offset = eso
			return token
		}

		return GTOK_ERROR
	}

	return 0
}

// Exposed lexer function (for yacc parser) wraps for contextual parsing
func (ctx *gesLex) Lex(lval *gesSymType) int {
	origRegValid := ctx.regexValid
	token := ctx.lex(lval)

	// TODO - label state engine from original code

	// Reset regex parsing state, only valid for next token
	if origRegValid {
		ctx.regexValid = false
	}

	// Apply data from context to current symbol
	lval.ctx = ctx.ctx

	return token
}

// Error recording 'callback' for the yacc parser
func (ctx *gesLex) Error(msg string) {
	ctx.error = &msg
}
