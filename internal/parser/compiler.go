/*
 * Various structures and other elements to support the parse/execute model.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"errors"

	"github.com/heisz/gescript/internal/engine"
)

// Carrier structure for tracking context elements of the parsing operations
type parsingContext struct {
	body *engine.Function
}

// Entry point to the yacc-based parser, returning a function body or error
func Parse(source string) (body *engine.Function, err error) {
	lex := newLexer(source)
	if gesParse(lex) != 0 {
		if lex.error != nil {
			return nil, errors.New(*lex.error)
		}
		return nil, errors.New("Unknown parsing error")
	}
	return lex.ctx.body, nil
}

// Parser utility method to push an opcode onto the current context body
func (dd *gesSymType) pushOpCode(opFn engine.OpCodeFn) (op *engine.OpCode) {
	op = &engine.OpCode{
		ExecFn: opFn,
	}
	dd.ctx.body.Code = append(dd.ctx.body.Code, op)
	return op
}

// Method to translate a $$ expression into the appropriate opcode
func (dd *gesSymType) pushEvalExpression(expr *gesSymType) {
	switch expr.parseType {
	case PARSED_LITERAL:
		op := dd.pushOpCode(engine.PushLiteralValue)
		op.OpData = &dd.literal
		break
	}
}
