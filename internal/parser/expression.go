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

func functionExprNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Parse the full script function instance (name is optional)
	fn := prs.parseFunctionDecl(false, false, nil)
	if fn == nil {
		return nil
	}

	// Result of the declaration is a stack value (first-class function)
	fnVal := types.DataType(fn)
	op := prs.pushOpCode(engine.PushFunctionOperation, 1)
	op.OpData = &fnVal

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func arrowLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// This led only occurs for form of single arg: x => expr
	switch left.parseType {
	case PARSED_IDENTIFIER, PARSED_GLOBAL_REFERENCE:
		return prs.parseArrowFunctionBody([]string{left.identifier})
	default:
		prs.addError("Invalid arrow function, expect identifier/arg on left")
		return nil
	}
}

func identifierNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Handle 'undefined' as a global property that returns undefined value
	if sym.identifier == "undefined" {
		rs := *sym
		rs.parseType = PARSED_LITERAL
		rs.literal = types.Undefined
		return &rs
	}

	// Resolve the variable in the block/scope chain
	varDef := prs.block.resolveVariable(sym.identifier)
	if varDef == nil {
		// Variable is not locally declared, find in closure if applicable
		if prs.outerScope != nil {
			capIdx := prs.resolveCapture(sym.identifier)
			if capIdx >= 0 {
				rs := *sym
				rs.parseType = PARSED_CAPTURE_REFERENCE
				rs.assignOp = capIdx
				return &rs
			}
		}

		// Not closure either, final possibility is global/native declaration
		rs := *sym
		rs.parseType = PARSED_GLOBAL_REFERENCE
		return &rs
	}

	// The resolved local variable might actually already be a closure capture
	if varDef.isCapture {
		rs := *sym
		rs.parseType = PARSED_CAPTURE_REFERENCE
		rs.assignOp = varDef.captureIdx
		return &rs
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
	if left == nil {
		prs.addError("Invalid left-hand side in assignment")
		return nil
	}

	// Handle different assignment targets
	switch left.parseType {
	case PARSED_IDENTIFIER:
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

	case PARSED_CAPTURE_REFERENCE:
		// Assignment to a captured variable (closure)
		capIdx := left.assignOp

		// Parse the right-hand side (right associative, use lbp - 1)
		right := prs.parseExpression(prec.lbp - 1)
		if right == nil || !prs.pushEvalExpression(right) {
			return nil
		}

		// Store value but leave on stack (assignment has expression value)
		op := prs.pushOpCode(engine.StoreCaptureKeepOperation, 0)
		op.OpData = capIdx

	case PARSED_ARRAY_REFERENCE:
		// Target/index already handled in prior led, just evaluate and assign
		right := prs.parseExpression(prec.lbp - 1)
		if right == nil || !prs.pushEvalExpression(right) {
			return nil
		}

		prs.pushOpCode(engine.SetElementOperation, -2)

	case PARSED_MEMBER_REFERENCE:
		// Target already handled in prior led, just evaluate and assign
		right := prs.parseExpression(prec.lbp - 1)
		if right == nil || !prs.pushEvalExpression(right) {
			return nil
		}

		op := prs.pushOpCode(engine.SetPropertyOperation, -1)
		op.OpData = left.identifier

	default:
		prs.addError("Invalid left-hand side in assignment")
		return nil
	}

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

// This has lots of cases due to () being possible grouping or argset
func parenNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Quick check for no-arg arrow function
	if prs.ctx.sym.token == GTOK_RP {
		if prs.lex() == GTOK_ERROR {
			return nil
		}
		if prs.ctx.sym.token == GTOK_ARROW {
			// Consume the arrow and parse the function body (no args)
			prs.lex()
			return prs.parseArrowFunctionBody(nil)
		}
		prs.addError("Unexpected empty parentheses")
		return nil
	}

	// Possible arrow paramter set (TODO - group assignment)
	if prs.ctx.sym.token == GTOK_IDENTIFIER {
		ident := prs.ctx.sym.identifier
		tok := prs.lex()

		// Continue to collect variable/parameter list, if applicable
		if tok == GTOK_COMMA || tok == GTOK_RP {
			var varlist []string
			varlist = append(varlist, ident)

			// Repeat for full set of variable identifiers
			for tok == GTOK_COMMA {
				tok = prs.lex()
				if tok != GTOK_IDENTIFIER {
					prs.addError("Expected identifier in variable list")
					return nil
				}
				varlist = append(varlist, prs.ctx.sym.identifier)
				tok = prs.lex()
			}

			// Only valid syntax at this point is a variable list
			if tok != GTOK_RP {
				prs.addError("Expected ')' after variable list")
				return nil
			}

			// Check for this being an arrow function, handle appropriately
			if prs.lex() == GTOK_ERROR {
				return nil
			}
			if prs.ctx.sym.token == GTOK_ARROW {
				// Consume the arrow and parse the function body (with args)
				prs.lex()
				return prs.parseArrowFunctionBody(varlist)
			}

			// TODO - actually support comma expressions
			if len(varlist) > 1 {
				prs.addError("Comma expressions not supported")
				return nil
			}

			// Only one? (x) might just be wrapped identifier, discard ()
			rs := &symType{
				token:      GTOK_IDENTIFIER,
				identifier: ident,
			}
			varDef := prs.block.resolveVariable(ident)
			if varDef == nil {
				rs.parseType = PARSED_GLOBAL_REFERENCE
			} else {
				if !varDef.initialized && varDef.declType != DECL_VAR {
					prs.addError("Cannot access '" + ident +
						"' before initialization")
					return nil
				}
				rs.parseType = PARSED_IDENTIFIER
			}
			return rs
		}

		// Not a varlist, just a grouped expression with leading identifier
		expr := prs.parseExpressionWithIdentifier(0, ident)
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

	// Not starting with identifier - parse as normal grouped expression
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

func prefixIncrDecrNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Parse precedence just below member/element to get correct target
	operand := prs.parseExpression(84)
	if operand == nil {
		return nil
	}

	// Select appropriate operations based on parsed operand
	switch operand.parseType {
	case PARSED_IDENTIFIER:
		// Standard variable increment/decrement
		varDef := prs.block.resolveVariable(operand.identifier)
		if varDef == nil {
			prs.addError("Undefined variable '" + operand.identifier + "'")
			return nil
		}
		if varDef.declType == DECL_CONST {
			prs.addError("Cannot modify constant '" + operand.identifier + "'")
			return nil
		}

		var op *engine.OpCode
		if sym.token == GTOK_INCR {
			op = prs.pushOpCode(engine.PreIncrementOperation, 1)
		} else {
			op = prs.pushOpCode(engine.PreDecrementOperation, 1)
		}
		op.OpData = varDef.slotIndex

	case PARSED_ARRAY_REFERENCE:
		// Element led has already stored target and index operations
		if sym.token == GTOK_INCR {
			prs.pushOpCode(engine.PreIncrementElementOperation, -1)
		} else {
			prs.pushOpCode(engine.PreDecrementElementOperation, -1)
		}

	case PARSED_MEMBER_REFERENCE:
		// Member led has stored target, need to bind property identifier
		var op *engine.OpCode
		if sym.token == GTOK_INCR {
			op = prs.pushOpCode(engine.PreIncrementPropertyOperation, 0)
		} else {
			op = prs.pushOpCode(engine.PreDecrementPropertyOperation, 0)
		}
		op.OpData = operand.identifier

	default:
		prs.addError("Invalid operand for increment/decrement operator")
		return nil
	}

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func postfixIncrDecrLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	if left == nil {
		prs.addError("Invalid operand for postfix operator")
		return nil
	}

	// Select appropriate operations based on parsed lvalue
	switch left.parseType {
	case PARSED_IDENTIFIER:
		// Standard variable increment/decrement
		varDef := prs.block.resolveVariable(left.identifier)
		if varDef == nil {
			prs.addError("Undefined variable '" + left.identifier + "'")
			return nil
		}
		if varDef.declType == DECL_CONST {
			prs.addError("Cannot modify constant '" + left.identifier + "'")
			return nil
		}

		var op *engine.OpCode
		if sym.token == GTOK_INCR {
			op = prs.pushOpCode(engine.PostIncrementOperation, 1)
		} else {
			op = prs.pushOpCode(engine.PostDecrementOperation, 1)
		}
		op.OpData = varDef.slotIndex

	case PARSED_ARRAY_REFERENCE:
		// Element led has already stored target and index operations
		if sym.token == GTOK_INCR {
			prs.pushOpCode(engine.PostIncrementElementOperation, -1)
		} else {
			prs.pushOpCode(engine.PostDecrementElementOperation, -1)
		}

	case PARSED_MEMBER_REFERENCE:
		// Member led has stored target, need to bind property identifier
		var op *engine.OpCode
		if sym.token == GTOK_INCR {
			op = prs.pushOpCode(engine.PostIncrementPropertyOperation, 0)
		} else {
			op = prs.pushOpCode(engine.PostDecrementPropertyOperation, 0)
		}
		op.OpData = left.identifier

	default:
		prs.addError("Invalid operand for postfix operator")
		return nil
	}

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func arrayLiteralNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// Quickly handle the empty array case
	if prs.ctx.sym.token == GTOK_RB {
		// Discard and create an empty array declaration
		if prs.lex() == GTOK_ERROR {
			return nil
		}
		op := prs.pushOpCode(engine.NewArrayOperation, 1)
		op.OpData = 0

		rs := *sym
		rs.parseType = PARSED_VALUE
		return &rs
	}

	// Parse the list of elements onto stack for initializer
	elemCount := 0
	for {
		expr := prs.parseExpression(0)
		if expr == nil || !prs.pushEvalExpression(expr) {
			return nil
		}
		elemCount++

		// Either continuation (comma) or end (right bracket), discard
		if prs.ctx.sym.token == GTOK_COMMA {
			if prs.lex() == GTOK_ERROR {
				return nil
			}
			continue
		}
		if prs.ctx.sym.token == GTOK_RB {
			if prs.lex() == GTOK_ERROR {
				return nil
			}
			break
		}

		prs.addError("Expected ',' or ']' in array literal")
		return nil
	}

	// Push the array operation with the element count (consumes all but one)
	op := prs.pushOpCode(engine.NewArrayOperation, 1-elemCount)
	op.OpData = elemCount

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func objectLiteralNud(prs *parser, prec *precDefn, sym *symType) *symType {
	// For object, track the keyset as a list for the operation
	var keys []string

	// Quickly handle the empty object case
	if prs.ctx.sym.token == GTOK_RC {
		// Discard and create an empty object declaration (empty keys)
		if prs.lex() == GTOK_ERROR {
			return nil
		}
		op := prs.pushOpCode(engine.NewObjectOperation, 1)
		op.OpData = keys

		rs := *sym
		rs.parseType = PARSED_VALUE
		return &rs
	}

	for {
		// Parse key, appears as either identifier or string based on quotes
		var keyName string
		if prs.ctx.sym.token == GTOK_IDENTIFIER {
			keyName = prs.ctx.sym.identifier
		} else if prs.ctx.sym.token == GTOK_LITERAL {
			if str, ok := prs.ctx.sym.literal.(types.StringType); ok {
				keyName = string(str)
			} else {
				prs.addError("Object key must be identifier or string")
				return nil
			}
		} else {
			prs.addError("Expected property name in object literal")
			return nil
		}
		keys = append(keys, keyName)

		// Requires a colon, consume to value expression
		if prs.lex() == GTOK_ERROR {
			return nil
		}
		if prs.ctx.sym.token != GTOK_COLON {
			prs.addError("Expected ':' after property name")
			return nil
		}
		if prs.lex() == GTOK_ERROR {
			return nil
		}

		// Process associated value expression (on stack for create)
		expr := prs.parseExpression(0)
		if expr == nil || !prs.pushEvalExpression(expr) {
			return nil
		}

		// Either continuation (comma) or end (right brace), discard
		if prs.ctx.sym.token == GTOK_COMMA {
			if prs.lex() == GTOK_ERROR {
				return nil
			}
			continue
		}
		if prs.ctx.sym.token == GTOK_RC {
			if prs.lex() == GTOK_ERROR {
				return nil
			}
			break
		}

		prs.addError("Expected ',' or '}' in object literal")
		return nil
	}

	// Push the object operation with the keyset (consumes all but one)
	op := prs.pushOpCode(engine.NewObjectOperation, 1-len(keys))
	op.OpData = keys

	rs := *sym
	rs.parseType = PARSED_VALUE
	return &rs
}

func elementAccessLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Evaluate the associated array/object to read from
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Parse the index or key expression
	indexExpr := prs.parseExpression(0)
	if indexExpr == nil || !prs.pushEvalExpression(indexExpr) {
		return nil
	}

	// Requires the matching closing bracket, discard
	if prs.ctx.sym.token != GTOK_RB {
		prs.addError("Expected ']' after bracket expression")
		return nil
	}
	if prs.lex() == GTOK_ERROR {
		return nil
	}

	// Array reference indicates get/set operation based on context
	rs := *sym
	rs.parseType = PARSED_ARRAY_REFERENCE
	return &rs
}

func memberAccessLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Evaluate the associated object to read from
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Current token must be the property name, save and discard
	if prs.ctx.sym.token != GTOK_IDENTIFIER {
		prs.addError("Expected property name after '.'")
		return nil
	}
	propName := prs.ctx.sym.identifier
	if prs.lex() == GTOK_ERROR {
		return nil
	}

	// Member reference indicates get/set operation based on context
	rs := *sym
	rs.parseType = PARSED_MEMBER_REFERENCE
	rs.identifier = propName
	return &rs
}

func callLed(prs *parser, prec *precDefn, sym *symType,
	left *symType) *symType {
	// Evaluate the function expression (target/this)
	if !prs.pushEvalExpression(left) {
		return nil
	}

	// Parse set of argument expressions
	argCount := 0
	if prs.ctx.sym.token != GTOK_RP {
		for {
			arg := prs.parseExpression(0)
			if arg == nil || !prs.pushEvalExpression(arg) {
				return nil
			}
			argCount++

			// Repeat until argument list is complete
			if prs.ctx.sym.token == GTOK_COMMA {
				if prs.lex() == GTOK_ERROR {
					return nil
				}
				continue
			}
			if prs.ctx.sym.token == GTOK_RP {
				break
			}

			prs.addError("Expected ',' or ')' in argument list")
			return nil
		}
	}

	// Consume closing parenthesis
	if prs.lex() == GTOK_ERROR {
		return nil
	}

	// And call it, consumes args from stack and replaces target with result
	op := prs.pushOpCode(engine.CallOperation, -(argCount))
	op.OpData = argCount

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

	// Object literal
	case GTOK_LC:
		p := precDefn{lbp: 0, nud: objectLiteralNud, led: nil}
		return &p

	// Function expression
	case GTOK_FUNCTION:
		p := precDefn{lbp: 0, nud: functionExprNud, led: nil}
		return &p

	// Grouping/arrow leader (nud) and function call argset (led)
	case GTOK_LP:
		p := precDefn{lbp: 85, nud: parenNud, led: callLed}
		return &p

	// Array literal (nud) or element access (led)
	case GTOK_LB:
		p := precDefn{lbp: 85, nud: arrayLiteralNud, led: elementAccessLed}
		return &p

	// Member access
	case GTOK_DOT:
		p := precDefn{lbp: 85, nud: nil, led: memberAccessLed}
		return &p

	// Unary operators - nud only, high precedence
	case GTOK_NOT, GTOK_TILDE:
		p := precDefn{lbp: 70, nud: unaryNud, led: nil}
		return &p

	// Increment/decrement - both prefix and postfix
	case GTOK_INCR, GTOK_DECR:
		p := precDefn{lbp: 75, nud: prefixIncrDecrNud, led: postfixIncrDecrLed}
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

	// Arrow function (same as assignment, right-associative)
	case GTOK_ARROW:
		p := precDefn{lbp: 10, nud: nil, led: arrowLed}
		return &p
	}

	return nil
}

// Parse an expression when we have read-ahead an identifier
func (prs *parser) parseExpressionWithIdentifier(rbp int,
	identName string) *symType {
	// Create a symType for the identifier and reuse identifierNud logic
	sym := &symType{
		token:      GTOK_IDENTIFIER,
		identifier: identName,
	}
	left := identifierNud(prs, nil, sym)
	if left == nil {
		return nil
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
