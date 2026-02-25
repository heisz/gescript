/*
 * Processing elements for parsing an expression instance.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"os"
	"strconv"

	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/types"
)

func TestLog(msg string) {
	file, err := os.OpenFile("output",
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(msg + "\n")
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

// Various null and left denotation functions used in the Pratt algorithm below

func literalNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Lexer defined the symbol but need to translate keyword literals
	rs := *sym
	switch sym.token {
	case GTOK_NULL:
		rs.parseType = PARSED_LITERAL
		rs.literal = types.NullType{}
	case GTOK_TRUE:
		rs.parseType = PARSED_LITERAL
		rs.literal = types.BooleanType(true)
	case GTOK_FALSE:
		rs.parseType = PARSED_LITERAL
		rs.literal = types.BooleanType(false)
	}

	return &rs
}

func identifierNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Handle 'undefined' as a global property that returns undefined value
	if sym.identifier == "undefined" {
		rs := *sym
		rs.parseType = PARSED_LITERAL
		rs.literal = types.UndefinedType{}
		return &rs
	}

	// Resolve the variable in the scope chain
	varDef := prs.block.resolveVariable(sym.identifier)
	if varDef == nil {
		prs.addError("Undefined variable '" + sym.identifier + "'")
		return nil
	}

	// Capture uninitialized let/const access (strict)
	if !varDef.initialized && varDef.declType != DECL_VAR {
		prs.addError("Cannot access '" + sym.identifier +
			"' before initialization")
		return nil
	}

	// Leave as identifier for either retrieve or assignment
	rs := *sym
	rs.parseType = PARSED_IDENTIFIER
	return &rs
}

func assignmentLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Left side must be an assignable reference (identifier for now)
	if left == nil || left.parseType != PARSED_IDENTIFIER {
		prs.addError("Invalid left-hand side in assignment")
		return nil
	}

	// Resolve the variable
	varDef := prs.block.resolveVariable(left.identifier)
	if varDef == nil {
		prs.addError("Undefined variable '" + left.identifier + "'")
		return nil
	}

	// Check for const reassignment
	if varDef.declType == DECL_CONST && varDef.initialized {
		prs.addError("Cannot reassign constant '" + left.identifier + "'")
		return nil
	}

	// Parse the right-hand side (right associative, use lbp - 1)
	right := prs.parseExpression(prec.lbp - 1)
	if right == nil || !prs.pushEvalExpression(right) {
		return nil
	}

	// Store value but leave on stack (assignment has expression value)
	op := prs.pushOpCode(engine.StoreVariableKeepOperation, 0)
	op.OpData = varDef.slotIndex

	// Variable has initialization now, mark for use check
	varDef.initialized = true

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func parenNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Parse the grouped expression inside the parentheses (no casting)
	expr := prs.parseExpression(0)
	if expr == nil {
		return nil
	}

	// Consume the closing parenthesis and advance to next token
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected closing parenthesis")
		return nil
	}
	if prs.lex() == GTOK_ERROR {
		return nil
	}

	return expr
}

func unaryNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Parse/push the operand with unary precedence onto the stack
	expr := prs.parseExpression(prec.lbp)
	if expr == nil || !prs.pushEvalExpression(expr) {
		return nil
	}

	// Push the appropriate unary operation
	switch sym.token {
	case GTOK_ADD:
		prs.pushOpCode(engine.UnaryPlusOperation, 0)
	case GTOK_SUB:
		prs.pushOpCode(engine.UnaryMinusOperation, 0)
	case GTOK_NOT:
		prs.pushOpCode(engine.LogicalNotOperation, 0)
	case GTOK_TILDE:
		prs.pushOpCode(engine.BitwiseNotOperation, 0)
	}

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func logicalAndLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Push left operand
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Handle short-circuit jump for false outcome
	jmpEnd := prs.pushOpCode(engine.JumpIfFalseOrPopOperation, -1)

	// Evaluate right side (left was popped from jump skip)
	right := prs.parseExpression(prec.lbp)
	if right == nil || !prs.pushEvalExpression(right) {
		return nil
	}

	// Adjust the jump operation destination
	jmpEnd.OpData = len(prs.body.Code)

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func logicalOrLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Push left operand
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Handle short-circuit jump for true outcome
	jmpEnd := prs.pushOpCode(engine.JumpIfTrueOrPopOperation, -1)

	// Evaluate right side (left was popped from jump skip)
	right := prs.parseExpression(prec.lbp)
	if right == nil || !prs.pushEvalExpression(right) {
		return nil
	}

	// Adjust the jump operation destination
	jmpEnd.OpData = len(prs.body.Code)

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func ternaryLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Evaluate the condition (already parsed)
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Insert jump for false outcome (true follows)
	jmpFalse := prs.pushOpCode(engine.JumpIfFalseOperation, -1)

	// Parse the true/then expression
	thenExpr := prs.parseExpression(0)
	if thenExpr == nil || !prs.pushEvalExpression(thenExpr) {
		return nil
	}

	// Insert jump to skip false outcome, adjust original jump
	jmpEnd := prs.pushOpCode(engine.JumpOperation, -1)
	jmpFalse.OpData = len(prs.body.Code)

	// Consume the colon separator
	if prs.ctx.sym.token != GTOK_COLON {
		prs.addError("Expected ':' in ternary expression")
		return nil
	}
	tok := prs.lex()
	if tok == GTOK_EOF || tok == GTOK_ERROR {
		return nil
	}

	// Parse the false/else expression
	elseExpr := prs.parseExpression(prec.lbp)
	if elseExpr == nil || !prs.pushEvalExpression(elseExpr) {
		return nil
	}

	// Adjust the true exit jump to the end
	jmpEnd.OpData = len(prs.body.Code)

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func infixLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Push left operand onto the stack
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Evaluate the right side, push onto stack
	right := prs.parseExpression(prec.lbp)
	if right == nil || (!prs.pushEvalExpression(right)) {
		return nil
	}

	// Lots of operations to consume the two arguments and leave one result
	switch sym.token {
	case GTOK_ADD:
		prs.pushOpCode(engine.AdditionOperation, -1)
	case GTOK_SUB:
		prs.pushOpCode(engine.SubtractionOperation, -1)
	case GTOK_MULT:
		prs.pushOpCode(engine.MultiplicationOperation, -1)
	case GTOK_DIV:
		prs.pushOpCode(engine.DivisionOperation, -1)
	case GTOK_MOD:
		prs.pushOpCode(engine.ModulusOperation, -1)
	case GTOK_LTLT:
		prs.pushOpCode(engine.LeftShiftOperation, -1)
	case GTOK_GTGT:
		prs.pushOpCode(engine.RightShiftOperation, -1)
	case GTOK_GTGTGT:
		prs.pushOpCode(engine.UnsignedRightShiftOperation, -1)
	case GTOK_LT:
		prs.pushOpCode(engine.LessThanOperation, -1)
	case GTOK_GT:
		prs.pushOpCode(engine.GreaterThanOperation, -1)
	case GTOK_LTEQ:
		prs.pushOpCode(engine.LessThanEqualOperation, -1)
	case GTOK_GTEQ:
		prs.pushOpCode(engine.GreaterThanEqualOperation, -1)
	case GTOK_EQEQ:
		prs.pushOpCode(engine.EqualOperation, -1)
	case GTOK_NOTEQ:
		prs.pushOpCode(engine.NotEqualOperation, -1)
	case GTOK_EQEQEQ:
		prs.pushOpCode(engine.StrictEqualOperation, -1)
	case GTOK_NOTEQEQ:
		prs.pushOpCode(engine.StrictNotEqualOperation, -1)
	case GTOK_AND:
		prs.pushOpCode(engine.BitwiseAndOperation, -1)
	case GTOK_OR:
		prs.pushOpCode(engine.BitwiseOrOperation, -1)
	case GTOK_XOR:
		prs.pushOpCode(engine.BitwiseXorOperation, -1)
	}

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

// Expression parsing uses modified Pratt algorithms (see Crockford)

// Function definitions for the null and left denotations
type (
	nudFn func(prs *parser, prec *precDefn, sym *symType) *symType
	ledFn func(prs *parser, prec *precDefn, sym *symType,
		left *symType) *symType
)

// Combining type to define precedence and associated denotation functions
type precDefn struct {
	lbp int
	nud nudFn
	led ledFn
}

// Translator from token instance to precedence map instance
func prec(token int) *precDefn {
	TestLog("PREC " + strconv.Itoa(token))
	switch token {
	// Purely nud tokens don't have a precedence
	case GTOK_LITERAL, GTOK_NULL, GTOK_TRUE, GTOK_FALSE:
		p := precDefn{lbp: 0, nud: literalNud, led: nil}
		return &p

	// Identifiers for variable access
	case GTOK_IDENTIFIER:
		p := precDefn{lbp: 0, nud: identifierNud, led: nil}
		return &p

	// Grouping has the highest precedence
	case GTOK_LP:
		p := precDefn{lbp: 80, nud: parenNud, led: nil}
		return &p

	// Unary operators - nud only, high precedence
	case GTOK_NOT, GTOK_TILDE:
		p := precDefn{lbp: 70, nud: unaryNud, led: nil}
		return &p

	// Multiplicative operators
	case GTOK_MULT, GTOK_DIV, GTOK_MOD:
		p := precDefn{lbp: 60, nud: nil, led: infixLed}
		return &p

	// Additive operators - also serve as unary +/-
	case GTOK_ADD, GTOK_SUB:
		p := precDefn{lbp: 50, nud: unaryNud, led: infixLed}
		return &p

	// Shift operators
	case GTOK_LTLT, GTOK_GTGT, GTOK_GTGTGT:
		p := precDefn{lbp: 47, nud: nil, led: infixLed}
		return &p

	// Relational operators
	case GTOK_LT, GTOK_GT, GTOK_LTEQ, GTOK_GTEQ:
		p := precDefn{lbp: 45, nud: nil, led: infixLed}
		return &p

	// Equality operators
	case GTOK_EQEQ, GTOK_NOTEQ, GTOK_EQEQEQ, GTOK_NOTEQEQ:
		p := precDefn{lbp: 40, nud: nil, led: infixLed}
		return &p

	// Bitwise AND
	case GTOK_AND:
		p := precDefn{lbp: 38, nud: nil, led: infixLed}
		return &p

	// Bitwise XOR
	case GTOK_XOR:
		p := precDefn{lbp: 37, nud: nil, led: infixLed}
		return &p

	// Bitwise OR
	case GTOK_OR:
		p := precDefn{lbp: 36, nud: nil, led: infixLed}
		return &p

	// Logical AND
	case GTOK_ANDAND:
		p := precDefn{lbp: 35, nud: nil, led: logicalAndLed}
		return &p

	// Logical OR
	case GTOK_OROR:
		p := precDefn{lbp: 30, nud: nil, led: logicalOrLed}
		return &p

	// Ternary conditional
	case GTOK_QMARK:
		p := precDefn{lbp: 20, nud: nil, led: ternaryLed}
		return &p

	// Assignment (lowest precedence, right-associative)
	case GTOK_ASSIGN:
		p := precDefn{lbp: 10, nud: nil, led: assignmentLed}
		return &p
	}

	return nil
}

// Parse an expression when we have read-ahead an identifier
func (prs *parser) parseExpressionWithIdentifier(rbp int,
	identName string) *symType {
	var left *symType

	// Handle 'undefined' as a global property returning undefined value
	if identName == "undefined" {
		left = &symType{
			token:     GTOK_LITERAL,
			parseType: PARSED_LITERAL,
			literal:   types.UndefinedType{},
		}
	} else {
		// Resolve the variable (same action as identifierNud)
		varDef := prs.block.resolveVariable(identName)
		if varDef == nil {
			prs.addError("Undefined variable '" + identName + "'")
			return nil
		}

		if !varDef.initialized && varDef.declType != DECL_VAR {
			prs.addError("Cannot access '" + identName +
				"' before initialization")
			return nil
		}

		// Create the left symType for the identifier
		left = &symType{
			token:      GTOK_IDENTIFIER,
			parseType:  PARSED_IDENTIFIER,
			identifier: identName,
		}
	}

	// Complete the expression (led handling)
	return prs.completeExpression(rbp, left)
}

// Pratt's parsing algorithm, exits with context on next token
func (prs *parser) parseExpression(rbp int) *symType {
	TestLog("Entering parse expression")
	// Handle lookahead if required
	var tok = prs.ctx.sym.token
	if rbp < 0 {
		tok = prs.lex()
		if tok == GTOK_EOF || tok == GTOK_ERROR {
			return nil
		}
	}

	// Determine the led processor for the start and execute it
	tsym := prs.ctx.sym
	tprec := prec(tok)
	if tprec == nil {
		prs.addError("Unexpected expression symbol XXX")
		return nil
	}
	if tprec.nud == nil {
		prs.addError("Expression syntax error (left) near XXX")
		return nil
	}

	// Advance to the next token (opt-semicolon for EOF)
	tok = prs.lex()
	if tok == GTOK_ERROR {
		return nil
	}

	// Process the null/prefix operation
	left := tprec.nud(prs, tprec, &tsym)

	// Complete the expresion (led handling)
	return prs.completeExpression(rbp, left)
}

// Common flow to complete the right binding parsing of the expression
func (prs *parser) completeExpression(rbp int, left *symType) *symType {
	// Loop while right binding power is less than left binding power
	// Note that the precedence lookup returns nil for terminating tokens
	// Use tsym.token not tok since nud/led may have advanced the lexer
	tsym := prs.ctx.sym
	tprec := prec(tsym.token)
	for (tprec != nil) && (rbp < tprec.lbp) {
		if tprec.led == nil {
			// TODO - insert automatic semicolon? or error?
			break
		}

		// Again, advance to the next token (opt-semicolon for EOF)
		tok := prs.lex()
		if tok == GTOK_ERROR {
			return nil
		}

		// Process the left/infix function for the operator
		left = tprec.led(prs, tprec, &tsym, left)
		if left == nil {
			return nil
		}

		tsym = prs.ctx.sym
		tprec = prec(tsym.token)
	}

	return left
}
