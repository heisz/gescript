/*
 * Parsing methods for various statement types.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/types"
)

/*
 * Section 13 and subsections within
 *
 * StatementListItem:
 *     Statement
 *     | Declaration
 *
 * Statement:
 *     BlockStatement
 *     | VariableStatement
 *     | EmptyStatement
 *     | ExpressionStatement
 *     | IfStatement
 *     | BreakableStatement
 *     | ContinueStatement
 *     | BreakStatement
 *     | ReturnStatement
 *     | WithStatement
 *     | LabelledStatement
 *     | ThrowStatement
 *     | TryStatement
 *     | DebuggerStatement
 *
 * Declaration:
 *     HoistableDeclaration
 *     | ClassDeclaration
 *     | LexicalDeclaration
 *
 * HoistableDeclaration:
 *     FunctionDeclaration
 *     | GeneratorDeclaration
 *
 * BreakableStatement:
 *     IterationStatement
 *     | SwitchStatement
 *
 * Enter: lexer on provided token, exit on last token of statement.
 */
func (prs *parser) parseStatementListItem(token int) {
	switch token {
	case GTOK_LC:
		prs.parseBlockStatement()
		return
	case GTOK_VAR:
		prs.parseVariableDeclaration(DECL_VAR)
		return
	case GTOK_LET:
		prs.parseVariableDeclaration(DECL_LET)
		return
	case GTOK_CONST:
		prs.parseVariableDeclaration(DECL_CONST)
		return
	case GTOK_SEMI:
		// (13.4) Don't really need a separate function to do nothing
		return
	case GTOK_IF:
		prs.parseIfStatement()
		return
	case GTOK_DO:
		prs.parseDoStatement()
		return
	case GTOK_WHILE:
		prs.parseWhileStatement()
		return
	case GTOK_FOR:
		prs.parseForStatement()
		return
	case GTOK_SWITCH:
		prs.parseSwitchStatement()
		return
	case GTOK_RETURN:
		prs.parseReturnStatement()
		return
	case GTOK_CONTINUE:
		prs.parseContinueStatement()
		return
	case GTOK_BREAK:
		prs.parseBreakStatement()
		return
	case GTOK_WITH:
		// TODO - not supported in strict, maybe add
		return
	case GTOK_THROW:
		prs.parseThrowStatement()
		return
	case GTOK_TRY:
		prs.parseTryStatement()
		return
	case GTOK_DEBUGGER:
		// TODO - what to do?
		return
	case GTOK_FUNCTION:
		prs.parseFunctionStatement()
		return
	case GTOK_IDENTIFIER:
		// Check for a labelled statement (colon after identifier)
		identName := prs.ctx.sym.identifier
		nextTok := prs.lex()
		if nextTok == GTOK_COLON {
			// It's a label, store and read ahead to determine context
			prs.pendingLabel = identName
			labelledTok := prs.lex()
			if labelledTok == GTOK_EOF || labelledTok == GTOK_ERROR {
				return
			}

			// Only loops and switch actually do anything with the label
			switch labelledTok {
			case GTOK_DO, GTOK_WHILE, GTOK_FOR, GTOK_SWITCH:
				prs.parseStatementListItem(labelledTok)
			default:
				prs.pendingLabel = ""
				prs.parseStatementListItem(labelledTok)
			}
			return
		}

		// Not a label, use the special method since we've read ahead
		expr := prs.parseExpressionWithIdentifier(0, identName)
		if expr != nil {
			prs.pushEvalExpression(expr)
			// Discard all but top level expression results (stack overflow)
			if prs.blockDepth > 0 {
				prs.pushOpCode(engine.PopOperation, -1)
			}
		}
		return
	}

	// Fallthrough for other expressions (literals, for example)
	expr := prs.parseExpression(0)
	if expr != nil {
		prs.pushEvalExpression(expr)
		// Discard all but top level expression results (stack overflow)
		if prs.blockDepth > 0 {
			prs.pushOpCode(engine.PopOperation, -1)
		}
	}
}

/*
 * Section 13.2
 *
 * BlockStatement:
 *     Block
 *
 * Block:
 *     { StatementList[opt] }
 *
 * Enter: lexer on left brace, exit on right brace.
 */
func (prs *parser) parseBlockStatement() {
	// Push new parser block instance for variable/exit scoping
	prs.block = newBlock(prs.block)
	prs.blockDepth++

	prs.parseStatementList()
	if prs.ctx.sym.token != GTOK_RC {
		prs.addError("Unterminated block statement (missing '}')")
	} else {
		// Advance past the closing brace for the caller
		prs.lex()
	}

	// Return to parent block context on completion
	prs.blockDepth--
	prs.block = prs.block.parent
}

/*
 * Section 13.3
 *
 * VariableStatement:
 *     var VariableDeclarationList ;
 *
 * LexicalDeclaration:
 *     LetOrConst BindingList ;
 *
 * VariableDeclarationList:
 *     VariableDeclaration
 *     | VariableDeclarationList, VariableDeclaration
 *
 * BindingList:
 *     LexicalBinding
 *     | BindingList, LexicalBinding
 *
 * VariableDeclaration/LexicalBinding:
 *     BindingIdentifier Initializer[opt]
 *     | BindingPattern Initializer
 *
 * Enter: lexer on var/let/const, exit on terminating semicolon.
 */
func (prs *parser) parseVariableDeclaration(declType varDeclType) {
	// Loop for multiple declarations
	for {
		// Must start with variable name (identifier)
		tok := prs.lex()
		if tok != GTOK_IDENTIFIER {
			prs.addError("Expected identifier in variable declaration")
			return
		}
		name := prs.ctx.sym.identifier

		// Var hoists to function, let/const stay within current block
		targetBlock := prs.block
		if declType == DECL_VAR {
			targetBlock = prs.rootBlock
		}

		// Define variable in target block, capture disallowed overlap
		varDef, ok := targetBlock.defineVariable(prs, name, declType)
		if !ok {
			prs.addError("Cannot redeclare '" + name + "' in this scope")
			return
		}

		// Check for initializer (assignment)
		tok = prs.lex()
		if tok == GTOK_ASSIGN {
			// Parse initializer expression
			expr := prs.parseExpression(-RBP_NO_COMMA)
			if expr == nil || !prs.pushEvalExpression(expr) {
				return
			}

			// Add the initialization value store operation
			op := prs.pushOpCode(engine.StoreVariableOperation, -1)
			op.OpData = varDef.slotIndex
			varDef.initialized = true
			tok = prs.ctx.sym.token
		} else if declType == DECL_CONST {
			// Const variables must have initializer
			prs.addError("Missing initializer in const declaration")
			return
		}

		// Comma for more declarations, semicolon (implicit) ends declaration
		if tok == GTOK_COMMA {
			continue
		}
		if tok == GTOK_SEMI || tok == GTOK_EOF || tok == GTOK_RC {
			return
		}

		prs.addError("Expected ';' after variable declaration")
		return
	}
}

/*
 * Section 13.6
 *
 * IfStatement:
 *     if ( Expression ) Statement else Statement
 *     | if ( Expression ) Statement
 *
 * Enter: lexer on if, exit after final statement.
 */
func (prs *parser) parseIfStatement() {
	// Require opening parenthesis
	if prs.lex() != GTOK_LP {
		prs.addError("Expected '(' after if statement leader")
		return
	}

	// Parse condition expression
	cond := prs.parseExpression(-1)
	if cond == nil || !prs.pushEvalExpression(cond) {
		return
	}

	// Require closing parenthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after if condition")
		return
	}

	// Jump past the true/then block if condition is false
	jmpFalse := prs.pushOpCode(engine.JumpIfFalseOperation, -1)

	// Parse then statement (add depth to discard expression value)
	tok := prs.lex()
	if tok == GTOK_ERROR || tok == GTOK_EOF {
		return
	}
	prs.blockDepth++
	prs.parseStatementListItem(tok)
	prs.blockDepth--

	// Consume semicolon if present from an expression statement
	if prs.ctx.sym.token == GTOK_SEMI {
		prs.lex()
	}

	// Check for else clause
	tok = prs.ctx.sym.token
	if tok == GTOK_ELSE {
		// Add jump past false/else block to the true block
		jmpEnd := prs.pushOpCode(engine.JumpOperation, -1)
		jmpFalse.OpData = len(prs.body.Code)

		// Parse else statement (again, mark depth for expression discard)
		tok = prs.lex()
		if tok == GTOK_ERROR || tok == GTOK_EOF {
			return
		}
		prs.blockDepth++
		prs.parseStatementListItem(tok)
		prs.blockDepth--

		jmpEnd.OpData = len(prs.body.Code)
	} else {
		// No else, false condition jumps to here
		jmpFalse.OpData = len(prs.body.Code)
	}
}

/*
 * Section 13.7.2
 *
 * IterationStatement:
 *     do Statement while ( Expression ) ;
 *
 * Enter: lexer on do, exit after semicolon.
 */
func (prs *parser) parseDoStatement() {
	// Push loop context for break/continue
	lsCtx := prs.pushLoopSwitchContext(false)

	// Record loop start for jump back to beginning of loop
	loopStart := len(prs.body.Code)

	// Parse loop body (increment depth to discard expression values)
	tok := prs.lex()
	if tok == GTOK_ERROR || tok == GTOK_EOF {
		prs.popLoopContext()
		return
	}
	prs.blockDepth++
	prs.parseStatementListItem(tok)
	prs.blockDepth--

	// Consume semicolon if present from an expression statement
	if prs.ctx.sym.token == GTOK_SEMI {
		prs.lex()
	}

	// Require while specification
	tok = prs.ctx.sym.token
	if tok != GTOK_WHILE {
		prs.addError("Expected 'while' after do statement")
		prs.popLoopContext()
		return
	}

	// Require opening parenthesis
	if prs.lex() != GTOK_LP {
		prs.addError("Expected '(' after while statement")
		prs.popLoopContext()
		return
	}

	// Continue statements in loop jump to condition evaluation
	target := len(prs.body.Code)
	lsCtx.continueTarget = target
	for _, jmp := range lsCtx.continueJumps {
		jmp.OpData = target
	}
	lsCtx.continueJumps = nil

	// Parse condition expression
	cond := prs.parseExpression(-1)
	if cond == nil || !prs.pushEvalExpression(cond) {
		prs.popLoopContext()
		return
	}

	// Require closing parenthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after while condition")
		prs.popLoopContext()
		return
	}
	prs.lex()

	// Jump back to start if condition is true (while)
	jmpLoop := prs.pushOpCode(engine.JumpIfTrueOperation, -1)
	jmpLoop.OpData = loopStart

	// Pop loop context (patches break jumps)
	prs.popLoopContext()
}

/*
 * Section 13.7.3
 *
 * IterationStatement:
 *     while ( Expression ) Statement
 *
 * Enter: lexer on while, exit after statement.
 */
func (prs *parser) parseWhileStatement() {
	// Push loop context for break/continue
	loopSwitchCtx := prs.pushLoopSwitchContext(false)

	// Require opening parenthesis
	if prs.lex() != GTOK_LP {
		prs.addError("Expected '(' after while statement")
		prs.popLoopContext()
		return
	}

	// Record condition start for loop back (continue/loop target)
	condStart := len(prs.body.Code)
	loopSwitchCtx.continueTarget = condStart

	// Parse condition expression
	cond := prs.parseExpression(-1)
	if cond == nil || !prs.pushEvalExpression(cond) {
		prs.popLoopContext()
		return
	}

	// Require closing parenthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after while condition")
		prs.popLoopContext()
		return
	}

	// Jump past body expression to exit if condition false
	jmpExit := prs.pushOpCode(engine.JumpIfFalseOperation, -1)

	// Parse loop statement (add depth to discard expression value)
	tok := prs.lex()
	if tok == GTOK_ERROR || tok == GTOK_EOF {
		prs.popLoopContext()
		return
	}
	prs.blockDepth++
	prs.parseStatementListItem(tok)
	prs.blockDepth--

	// Jump back to condition at start of loop and record exit point
	jmpLoop := prs.pushOpCode(engine.JumpOperation, -1)
	jmpLoop.OpData = condStart
	jmpExit.OpData = len(prs.body.Code)

	// Pop loop context (patches break jumps)
	prs.popLoopContext()
}

/*
 * Section 13.7.4
 *
 * IterationStatement:
 *     for ( Expression[opt] ; Expression[opt] ; Expression[opt] ) Statement
 *     | for ( var VariableDeclarationList ; Expression[opt] ;
 *                                     Expression[opt] ) Statement
 *     | for ( LexicalDeclaration Expression[opt] ; Expression[opt] ) Statement
 *     | for ( ForDeclaration in Expression ) Statement
 *     | for ( ForDeclaration in AssignmentExpression ) Statement
 *
 * Enter: lexer on 'for', exit after statement.
 */
func (prs *parser) parseForStatement() {
	// Push loop context for break/continue
	loopSwitchCtx := prs.pushLoopSwitchContext(false)

	// Require opening parenthesis
	if prs.lex() != GTOK_LP {
		prs.addError("Expected '(' after 'for'")
		prs.popLoopContext()
		return
	}

	// Create wrapper block for let/const declarations in for conditions
	prs.block = newBlock(prs.block)

	// Parse initializer/declaration, capturing in/of form
	tok := prs.lex()
	var forDeclSlot int = -1
	var forDeclName string

	if tok == GTOK_VAR || tok == GTOK_LET || tok == GTOK_CONST {
		// ForDeclaration possible, check for in/of keywords
		declType := DECL_VAR
		if tok == GTOK_LET {
			declType = DECL_LET
		} else if tok == GTOK_CONST {
			declType = DECL_CONST
		}

		// Expect identifier
		if prs.lex() != GTOK_IDENTIFIER {
			prs.addError("Expected identifier in for declaration")
			prs.block = prs.block.parent
			prs.popLoopContext()
			return
		}
		forDeclName = prs.ctx.sym.identifier

		// Check for in/of keywords for that form (TODO - of context keyword?)
		nextTok := prs.lex()
		if nextTok == GTOK_IN || nextTok == GTOK_OF {
			varDef, ok := prs.block.defineVariable(prs, forDeclName, declType)
			if !ok {
				prs.addError("Cannot redeclare '" + forDeclName + "'")
				prs.block = prs.block.parent
				prs.popLoopContext()
				return
			}
			varDef.initialized = true
			forDeclSlot = varDef.slotIndex
			prs.parseForInOfStatement(loopSwitchCtx, forDeclSlot,
				nextTok == GTOK_IN)
			return
		}

		// 'Conventional' for, regular declaration with possible initializer
		varDef, ok := prs.block.defineVariable(prs, forDeclName, declType)
		if !ok {
			prs.addError("Cannot redeclare '" + forDeclName + "'")
			prs.block = prs.block.parent
			prs.popLoopContext()
			return
		}

		// Handle variable initialization if discovered
		if nextTok == GTOK_ASSIGN {
			prs.lex()
			expr := prs.parseExpression(RBP_NO_COMMA)
			if expr != nil {
				prs.pushEvalExpression(expr)
				op := prs.pushOpCode(engine.StoreVariableOperation, -1)
				op.OpData = varDef.slotIndex
				varDef.initialized = true
			}
		}

		// Continue looping for possible multiple declaration/assignments
		for prs.ctx.sym.token == GTOK_COMMA {
			if prs.lex() != GTOK_IDENTIFIER {
				prs.addError("Expected identifier in declaration")
				break
			}
			name := prs.ctx.sym.identifier
			nextVarDef, ok := prs.block.defineVariable(prs, name, declType)
			if !ok {
				prs.addError("Cannot redeclare '" + name + "'")
				break
			}
			if prs.lex() == GTOK_ASSIGN {
				prs.lex()
				expr := prs.parseExpression(RBP_NO_COMMA)
				if expr != nil {
					prs.pushEvalExpression(expr)
					op := prs.pushOpCode(engine.StoreVariableOperation, -1)
					op.OpData = nextVarDef.slotIndex
					nextVarDef.initialized = true
				}
			}
		}
	} else if tok == GTOK_IDENTIFIER {
		// Similar to prior, no declaration but look for in/of form (TODO)
		forDeclName = prs.ctx.sym.identifier
		nextTok := prs.lex()
		if nextTok == GTOK_IN || nextTok == GTOK_OF {
			// Again in/of form but in this case variable must be declared
			varDef := prs.block.resolveVariable(forDeclName)
			if varDef == nil {
				prs.addError("Undefined variable '" + forDeclName + "'")
				prs.block = prs.block.parent
				prs.popLoopContext()
				return
			}
			forDeclSlot = varDef.slotIndex
			prs.parseForInOfStatement(loopSwitchCtx, forDeclSlot,
				nextTok == GTOK_IN)
			return
		}

		// 'Conventional' for, regular expression with possible initializer
		expr := prs.parseExpressionWithIdentifier(0, forDeclName)
		if expr != nil {
			prs.pushEvalExpression(expr)
			prs.pushOpCode(engine.PopOperation, -1)
		}
	} else if tok != GTOK_SEMI {
		// Expression initializer
		expr := prs.parseExpression(0)
		if expr != nil {
			prs.pushEvalExpression(expr)
			// Discard residual expression value from init
			prs.pushOpCode(engine.PopOperation, -1)
		}
	}

	// After initializer, should be on semicolon
	if prs.ctx.sym.token != GTOK_SEMI {
		prs.addError("Expected ';' after for initializer")
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}

	// Record condition start position and parse it (optional)
	condStart := len(prs.body.Code)
	tok = prs.lex()
	var jmpExit *engine.OpCode = nil
	if tok != GTOK_SEMI {
		cond := prs.parseExpression(0)
		if cond == nil || !prs.pushEvalExpression(cond) {
			prs.block = prs.block.parent
			prs.popLoopContext()
			return
		}

		// Condition evaluating to false exits the for loop
		jmpExit = prs.pushOpCode(engine.JumpIfFalseOperation, -1)
	}

	// After condition, should be on semicolon
	if prs.ctx.sym.token != GTOK_SEMI {
		prs.addError("Expected ';' after for condition")
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}

	// Jump past update expression to body (update runs at end)
	jmpBody := prs.pushOpCode(engine.JumpOperation, -1)
	updateStart := len(prs.body.Code)

	// Continue jumps to the update expression
	loopSwitchCtx.continueTarget = updateStart

	// Parse the update expression (optional)
	tok = prs.lex()
	if tok != GTOK_RP {
		update := prs.parseExpression(0)
		if update != nil {
			prs.pushEvalExpression(update)
			// Discard residual update expression value
			prs.pushOpCode(engine.PopOperation, -1)
		}
	}

	// After update evaluation, operations jumps to condition
	jmpCond := prs.pushOpCode(engine.JumpOperation, -1)
	jmpCond.OpData = condStart

	// After all of these optional details, require closing paranthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after for clauses")
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}

	// Update the increment skip jump
	jmpBody.OpData = len(prs.body.Code)

	// Parse loop/body statement (add depth to discard expression value)
	tok = prs.lex()
	if tok == GTOK_ERROR || tok == GTOK_EOF {
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}
	prs.blockDepth++
	prs.parseStatementListItem(tok)
	prs.blockDepth--

	// Jump to update segment on exit from body
	jmpUpd := prs.pushOpCode(engine.JumpOperation, -1)
	jmpUpd.OpData = updateStart

	// Set condition exit jump target, if there was one
	if jmpExit != nil {
		jmpExit.OpData = len(prs.body.Code)
	}

	// Pop loop context (updates break statements) and wrapper block
	prs.popLoopContext()
	prs.block = prs.block.parent
}

/*
 * Parse for...in loop body, called from above with lexer after 'in'.
 */
func (prs *parser) parseForInOfStatement(loopSwitchCtx *loopSwitchContext,
	varSlot int, isInLoop bool) {
	// Parse the object expression to iterate over
	prs.lex()
	iterExpr := prs.parseExpression(0)
	if iterExpr == nil || !prs.pushEvalExpression(iterExpr) {
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}

	// Require closing parenthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after for...of/in expression")
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}

	// Add operation to initialize the iteration set
	if isInLoop {
		prs.pushOpCode(engine.ForInKeysOperation, 1)
	} else {
		prs.pushOpCode(engine.ForOfIteratorOperation, 1)
	}

	// Mark loop start for continue and iteration looping
	loopStart := len(prs.body.Code)
	loopSwitchCtx.continueTarget = loopStart

	// And exit condition in this case is the iterator exhaustion
	if isInLoop {
		prs.pushOpCode(engine.ForInHasMoreOperation, 1)
	} else {
		prs.pushOpCode(engine.ForOfHasMoreOperation, 1)
	}
	jmpExit := prs.pushOpCode(engine.JumpIfFalseOperation, -1)

	// Net body starts with retrieval of next iteration value into variable
	var op *engine.OpCode
	if isInLoop {
		op = prs.pushOpCode(engine.ForInNextOperation, 0)
	} else {
		op = prs.pushOpCode(engine.ForOfNextOperation, 0)
	}
	op.OpData = varSlot

	// Parse loop/body statement (add depth to discard expression value)
	tok := prs.lex()
	if tok == GTOK_ERROR || tok == GTOK_EOF {
		prs.block = prs.block.parent
		prs.popLoopContext()
		return
	}
	prs.blockDepth++
	prs.parseStatementListItem(tok)
	prs.blockDepth--

	// Body ends with jump back to start of loop
	jmpLoop := prs.pushOpCode(engine.JumpOperation, -1)
	jmpLoop.OpData = loopStart

	// Exit to here and add cleanup of iteration working data
	jmpExit.OpData = len(prs.body.Code)
	if isInLoop {
		prs.pushOpCode(engine.ForInCleanupOperation, -2)
	} else {
		prs.pushOpCode(engine.ForOfCleanupOperation, -2)
	}

	// Pop loop context (updates break statements) and wrapper block
	prs.popLoopContext()
	prs.block = prs.block.parent
}

/*
 * Section 13.12
 *
 * SwitchStatement:
 *     switch ( Expression ) CaseBlock
 *
 * CaseBlock:
 *     { CaseClauses[opt] }
 *     | { CaseClauses[opt] DefaultClause CaseClauses[opt] }
 *
 * CaseClauses:
 *     CaseClause
 *     | CaseClauses CaseClause
 *
 * CaseClause:
 *     case Expression : StatementList[opt]
 *
 * DefaultClause:
 *     default : StatementList[opt]
 *
 * Enter: lexer on 'switch', exit after closing brace.
 */
func (prs *parser) parseSwitchStatement() {
	// Push switch context for break (no continue in switch)
	prs.pushLoopSwitchContext(true)

	// Require opening parenthesis
	if prs.lex() != GTOK_LP {
		prs.addError("Expected '(' after switch")
		prs.popLoopContext()
		return
	}

	// Parse switch expression
	expr := prs.parseExpression(-1)
	if expr == nil || !prs.pushEvalExpression(expr) {
		prs.popLoopContext()
		return
	}

	// Require closing parenthesis
	if prs.ctx.sym.token != GTOK_RP {
		prs.addError("Expected ')' after switch expression")
		prs.popLoopContext()
		return
	}

	// Require opening brace
	if prs.lex() != GTOK_LC {
		prs.addError("Expected '{' after switch expression")
		prs.popLoopContext()
		return
	}

	// Track pending jump from failed case comparison and default block
	var caseSkipJmp *engine.OpCode
	var defaultBodyStart = -1

	tok := prs.lex()
	for tok != GTOK_RC && tok != GTOK_EOF && tok != GTOK_ERROR {
		if tok == GTOK_CASE {
			// Update previous case skip jump to here
			if caseSkipJmp != nil {
				caseSkipJmp.OpData = len(prs.body.Code)
				caseSkipJmp = nil
			}

			// Duplicate switch value for comparison operator
			prs.pushOpCode(engine.DupOperation, 1)

			// Parse case expression
			caseExpr := prs.parseExpression(-1)
			if caseExpr == nil || !prs.pushEvalExpression(caseExpr) {
				prs.popLoopContext()
				return
			}

			if prs.ctx.sym.token != GTOK_COLON {
				prs.addError("Expected ':' after case expression")
				prs.popLoopContext()
				return
			}

			// Compare and jump to next case if not equal
			prs.pushOpCode(engine.StrictEqualOperation, -1)
			caseSkipJmp = prs.pushOpCode(engine.JumpIfFalseOperation, -1)

			// Parse case body statements
			tok = prs.lex()
			prs.blockDepth++
			for tok != GTOK_CASE && tok != GTOK_DEFAULT && tok != GTOK_RC &&
				tok != GTOK_EOF && tok != GTOK_ERROR {
				prs.parseStatementListItem(tok)
				if prs.ctx.sym.token == GTOK_SEMI {
					prs.lex()
				}
				tok = prs.ctx.sym.token
			}
			prs.blockDepth--
		} else if tok == GTOK_DEFAULT {
			// Only one default clause allowed per switch
			if defaultBodyStart >= 0 {
				prs.addError("Multiple default clauses in switch")
				prs.popLoopContext()
				return
			}

			// Update previous case skip jump to here (and mark here)
			if caseSkipJmp != nil {
				caseSkipJmp.OpData = len(prs.body.Code)
				caseSkipJmp = nil
			}
			defaultBodyStart = len(prs.body.Code)

			if prs.lex() != GTOK_COLON {
				prs.addError("Expected ':' after default")
				prs.popLoopContext()
				return
			}

			// Parse default body statements
			tok = prs.lex()
			prs.blockDepth++
			for tok != GTOK_CASE && tok != GTOK_DEFAULT && tok != GTOK_RC &&
				tok != GTOK_EOF && tok != GTOK_ERROR {
				prs.parseStatementListItem(tok)
				if prs.ctx.sym.token == GTOK_SEMI {
					prs.lex()
				}
				tok = prs.ctx.sym.token
			}
			prs.blockDepth--

		} else {
			prs.addError("Expected 'case' or 'default' in switch")
			prs.popLoopContext()
			return
		}
	}

	// Update remaining case skip jump to default or end
	if caseSkipJmp != nil {
		if defaultBodyStart >= 0 {
			caseSkipJmp.OpData = defaultBodyStart
		} else {
			caseSkipJmp.OpData = len(prs.body.Code)
		}
	}

	// Pop switch value from stack
	prs.pushOpCode(engine.PopOperation, -1)

	// Need to have ended on the closing brace
	if tok != GTOK_RC {
		prs.addError("Expected '}' at end of switch")
	} else {
		prs.lex()
	}

	// Pop loop context (updates all case break statements)
	prs.popLoopContext()
}

/*
 * Section 13.10
 *
 * ReturnStatement:
 *     return ;
 *     | return Expression ;
 *
 * Enter: lexer on 'return', exit on semicolon.
 */
func (prs *parser) parseReturnStatement() {
	// Check for optional return expression
	tok := prs.lex()
	if tok == GTOK_SEMI || tok == GTOK_RC || tok == GTOK_EOF {
		// No return expression, returns undefined
		op := prs.pushOpCode(engine.ReturnOperation, 0)
		op.OpData = false
		return
	}

	// Parse the return expression
	expr := prs.parseExpression(0)
	if expr == nil || !prs.pushEvalExpression(expr) {
		return
	}

	// Return consuming value (not relevant as return will collapse stack)
	op := prs.pushOpCode(engine.ReturnOperation, -1)
	op.OpData = true
}

/*
 * Section 14.1
 *
 * FunctionDeclaration:
 *     function BindingIdentifier ( FormalParameters ) { FunctionBody }
 *
 * FunctionExpression:
 *     function BindingIdentifier[opt] ( FormalParameters ) { FunctionBody }
 *
 * FormalParameters:
 *
 *     | FormalParameterList[?Yield]
 *
 * FormalParameterList:
 *     FunctionRestParameter
 *     | FormalsList
 *     | FormalsList, FunctionRestParameter
 *
 * FormalsList
 *     FormalParameter
 *     | FormalsList, FormalParameter
 *
 * FunctionRestParameter
 *     BindingRestElement
 *
 * FormalParameter:
 *     BindingElement
 *
 * FunctionBody:
 *     FunctionStatementList
 *
 * FunctionStatementList
 *     StatementList[Return][opt]
 *
 * Note: the root method is shared between standard and arrow function parsers.
 *
 * Enter: lexer on name (optional) or body, exit after end of body/expression.
 */
func (prs *parser) parseFunctionDecl(nameReq bool, isArrow bool,
	paramNames []string) *engine.ScriptFunction {
	var fnName string
	var hasRestParam bool

	// Arrow functions have already parsed the preamble
	if !isArrow {
		// Grab function name if available (optional depending on flag)
		tok := prs.ctx.sym.token
		if tok == GTOK_IDENTIFIER {
			fnName = prs.ctx.sym.identifier
			tok = prs.lex()
		} else if nameReq {
			prs.addError("Expected function name")
			return nil
		}

		// Require opening parenthesis for parameters
		if tok != GTOK_LP {
			prs.addError("Expected '(' after function name")
			return nil
		}

		// Parse parameter list
		paramNames = make([]string, 0)
		hasRestParam = false
		tok = prs.lex()
		for tok != GTOK_RP {
			// Check for rest parameter, if found must be the last parameter
			if tok == GTOK_ELLIPSIS {
				hasRestParam = true
				tok = prs.lex()
				if tok != GTOK_IDENTIFIER {
					prs.addError("Expected identifier after '...'")
					return nil
				}
				paramNames = append(paramNames, prs.ctx.sym.identifier)
				tok = prs.lex()
				if tok != GTOK_RP {
					prs.addError("Rest parameter must be last parameter")
					return nil
				}
				break
			}

			if tok != GTOK_IDENTIFIER {
				prs.addError("Expected parameter name")
				return nil
			}
			paramNames = append(paramNames, prs.ctx.sym.identifier)

			tok = prs.lex()
			if tok == GTOK_COMMA {
				tok = prs.lex()
			} else if tok != GTOK_RP {
				prs.addError("Expected ',' or ')' in parameter list")
				return nil
			}
		}

		// Require opening brace for function body
		if prs.lex() != GTOK_LC {
			prs.addError("Expected '{' for function body")
			return nil
		}
	}

	// Save current parser context to handle restore after body parse
	savedCtx := *prs

	// Switch parsing context to a new function instance
	fnBody := engine.NewFunction(fnName)
	fnBlock := newBlock(nil)
	prs.body = fnBody
	prs.rootBlock = fnBlock
	prs.block = fnBlock

	// Create a tracking context for closure variable capture
	var outerCaptures *[]captureEntry
	if savedCtx.captures != nil || savedCtx.outerScope != nil {
		outerCaptures = &savedCtx.captures
	}
	prs.outerScope = &outerScopeContext{
		parent:   savedCtx.outerScope,
		block:    savedCtx.block,
		captures: outerCaptures,
	}
	prs.captures = nil

	// All parameters become defined variables in the function block scope
	for _, paramName := range paramNames {
		varDef, ok := fnBlock.defineVariable(prs, paramName, DECL_VAR)
		if !ok {
			prs.addError("Duplicate parameter name '" + paramName + "'")
			prs.restoreContext(&savedCtx, outerCaptures)
			return nil
		}
		varDef.initialized = true
	}

	// Define this and arguments variable for non-arrow functions
	thisSlot := -1
	argumentsSlot := -1
	if !isArrow {
		thisDef, ok := fnBlock.defineVariable(prs, "this", DECL_CONST)
		if ok {
			thisDef.initialized = true
			thisSlot = thisDef.slotIndex
		}

		argsDef, ok := fnBlock.defineVariable(prs, "arguments", DECL_VAR)
		if ok {
			argsDef.initialized = true
			argumentsSlot = argsDef.slotIndex
		}
	}

	// Parse the function body - differs for regular vs arrow functions
	if isArrow {
		// Arrow can be block body or expression with implicit return
		tok := prs.ctx.sym.token
		if tok == GTOK_LC {
			prs.parseBlockStatement()
			// Not optimized but ensure an undefined is returned if not explicit
			op := prs.pushOpCode(engine.ReturnOperation, 0)
			op.OpData = false
		} else {
			expr := prs.parseExpression(RBP_NO_COMMA)
			if expr == nil || !prs.pushEvalExpression(expr) {
				prs.restoreContext(&savedCtx, outerCaptures)
				return nil
			}
			op := prs.pushOpCode(engine.ReturnOperation, -1)
			op.OpData = true
		}
	} else {
		// Regular function is only a block body
		prs.parseStatementList()
		if prs.ctx.sym.token != GTOK_RC {
			prs.addError("Expected '}' at end of function body")
			// It's broken but we do have a function of sorts, continue
		} else {
			prs.lex()
		}
		// Not optimized but ensure an undefined is returned if not explicit
		op := prs.pushOpCode(engine.ReturnOperation, 0)
		op.OpData = false
	}

	// Collect the captures accumulated in the body for the function
	var captures []engine.CaptureInfo
	for _, cap := range prs.captures {
		captures = append(captures, engine.CaptureInfo{
			Name:      cap.name,
			SlotIndex: cap.slotIndex,
			IsCapture: cap.isCapture,
		})
	}

	// Restore original parser context (captures may have been modified)
	prs.restoreContext(&savedCtx, outerCaptures)

	return &engine.ScriptFunction{
		Name:          fnName,
		ParamNames:    paramNames,
		HasRestParam:  hasRestParam,
		Body:          fnBody,
		VarCount:      fnBody.VarCount,
		ArgumentsSlot: argumentsSlot,
		ThisSlot:      thisSlot,
		IsArrowFunc:   isArrow,
		Captures:      captures,
	}
}

// Helper to restore parser context after function parsing
func (prs *parser) restoreContext(savedCtx *parser, captures *[]captureEntry) {
	prs.body = savedCtx.body
	prs.rootBlock = savedCtx.rootBlock
	prs.block = savedCtx.block
	prs.outerScope = savedCtx.outerScope
	if captures != nil {
		prs.captures = *captures
	} else {
		prs.captures = nil
	}
}

/*
 * Wrapper function declaration statement, parses and stores local/global.
 */
func (prs *parser) parseFunctionStatement() {
	prs.lex()
	fn := prs.parseFunctionDecl(true, false, nil)
	if fn == nil {
		return
	}

	// Functions are first-class variables, in open declaration also global
	varDef, ok := prs.rootBlock.defineVariable(prs, fn.Name, DECL_VAR)
	if !ok {
		prs.addError("Cannot redeclare '" + fn.Name + "' in this scope")
		return
	}
	varDef.initialized = true

	// Generate function as value, store in global table and as local variable
	fnVal := types.DataType(fn)
	op := prs.pushOpCode(engine.PushFunctionOperation, 1)
	op.OpData = fnVal

	globalOp := prs.pushOpCode(engine.StoreGlobalOperation, 0)
	globalOp.OpData = fn.Name

	storeOp := prs.pushOpCode(engine.StoreVariableOperation, -1)
	storeOp.OpData = varDef.slotIndex
}

/*
 * Section 14.2
 *
 * ArrowFunction:
 *     ArrowParameters => ConciseBody
 *
 * ArrowParameters:
 *     BindingIdentifier
 *     | ( FormalParameters )
 *
 * ConciseBody:
 *     ExpressionBody
 *     | { FunctionBody }
 *
 * Called on body (next token after =>) ends on body/expression end.
 */
func (prs *parser) parseArrowFunctionBody(paramNames []string) *symType {
	fn := prs.parseFunctionDecl(false, true, paramNames)
	if fn == nil {
		return nil
	}

	// Push function value onto the stack
	fnVal := types.DataType(fn)
	op := prs.pushOpCode(engine.PushFunctionOperation, 1)
	op.OpData = fnVal

	return &symType{parseType: PARSED_VALUE}
}

/*
 * Section 13.8
 *
 * ContinueStatement:
 *     continue ;
 *     | continue LabelIdentifier ;
 *
 * Enter: lexer on 'continue', exit on semicolon.
 */
func (prs *parser) parseContinueStatement() {
	var lsCtx *loopSwitchContext

	// Check for optional label
	tok := prs.lex()
	if tok == GTOK_IDENTIFIER {
		// Labelled continue, find target loop context
		label := prs.ctx.sym.identifier
		for lsCtx = prs.loopSwitchCtx; lsCtx != nil; lsCtx = lsCtx.parent {
			if lsCtx.label == label && !lsCtx.isSwitch {
				break
			}
		}
		if lsCtx == nil {
			prs.addError("Undefined label '" + label + "' for continue")
			return
		}

		// Consume the label identifier (move to next)
		tok = prs.lex()
	} else {
		// Unlabelled continue, find innermost loop context
		for lsCtx = prs.loopSwitchCtx; lsCtx != nil; lsCtx = lsCtx.parent {
			if !lsCtx.isSwitch {
				break
			}
		}
		if lsCtx == nil {
			prs.addError("Continue encountered outside of loop")
			return
		}
	}

	// Jump to continue target (may need future update depending on context)
	jmp := prs.pushOpCode(engine.JumpOperation, 0)
	if lsCtx.continueTarget >= 0 {
		jmp.OpData = lsCtx.continueTarget
	} else {
		// No target yet, store it for future update
		lsCtx.continueJumps = append(lsCtx.continueJumps, jmp)
	}

	// Already on semicolon or next statement token
}

/*
 * Section 13.9
 *
 * BreakStatement:
 *     break ;
 *     | break LabelIdentifier ;
 *
 * Enter: lexer on 'break', exit on semicolon.
 */
func (prs *parser) parseBreakStatement() {
	var lsCtx *loopSwitchContext

	// Check for optional label
	tok := prs.lex()
	if tok == GTOK_IDENTIFIER {
		// Labelled break, find target loop context
		label := prs.ctx.sym.identifier
		for lsCtx = prs.loopSwitchCtx; lsCtx != nil; lsCtx = lsCtx.parent {
			if lsCtx.label == label {
				break
			}
		}
		if lsCtx == nil {
			prs.addError("Undefined label '" + label + "' for break")
			return
		}
		tok = prs.lex() // Consume the label
	} else {
		// Unlabelled break, find innermost loop or switch context
		lsCtx = prs.loopSwitchCtx
		if lsCtx == nil {
			prs.addError("Break encountered outside of loop or switch")
			return
		}
	}

	// Add jump to be updated when loop/switch ends (never already there)
	jmp := prs.pushOpCode(engine.JumpOperation, 0)
	lsCtx.breakJumps = append(lsCtx.breakJumps, jmp)

	// Already on semicolon or next statement token
}

/*
 * Section 13.14
 *
 * ThrowStatement:
 *     throw Expression ;
 *
 * Enter: lexer on 'throw', exit on semicolon.
 */
func (prs *parser) parseThrowStatement() {
	// Parse the expression to throw
	expr := prs.parseExpression(-1)
	if expr == nil {
		prs.addError("Expected expression after throw")
		return
	}
	if !prs.pushEvalExpression(expr) {
		return
	}

	// If you want your boomerang to come back, first you've got to...
	prs.pushOpCode(engine.ThrowOperation, -1)
}

/*
 * Section 13.15
 *
 * TryStatement:
 *     try Block Catch
 *     | try Block Finally
 *     | try Block Catch Finally
 *
 * Catch:
 *     catch ( CatchParameter ) Block
 *     | catch Block
 *
 * Finally:
 *     finally Block
 *
 * Enter: lexer on 'try', exit after final block.
 */
func (prs *parser) parseTryStatement() {
	// Require opening brace for try block
	if prs.lex() != GTOK_LC {
		prs.addError("Expected '{' after try")
		return
	}

	// Create and push try context - targets will be updated as found
	tryCtx := &engine.ExceptionContext{
		CatchTarget:   -1,
		FinallyTarget: -1,
		EndTarget:     -1,
		CatchVarSlot:  -1,
	}
	tryOp := prs.pushOpCode(engine.PushExceptionContextOperation, 0)
	tryOp.OpData = tryCtx

	// Parse try block (we are on the opening {)
	prs.parseBlockStatement()

	// Pop try context on normal exit and add jump for finally or exit
	prs.pushOpCode(engine.PopExceptionContextOperation, 0)
	jmpFromTry := prs.pushOpCode(engine.JumpOperation, 0)

	// Track if we have catch/finally clauses
	var hasCatch, hasFinally bool
	var jmpFromCatch *engine.OpCode

	// Handle optional catch clause
	tok := prs.ctx.sym.token
	var catchVarName string
	if tok == GTOK_CATCH {
		hasCatch = true
		tryCtx.CatchTarget = len(prs.body.Code)

		// Opening parenthesis indicates option catch variable
		tok = prs.lex()
		if tok == GTOK_LP {
			// Has catch parameter
			tok = prs.lex()
			if tok != GTOK_IDENTIFIER {
				prs.addError("Expected identifier in catch-variable clause")
				return
			}
			catchVarName = prs.ctx.sym.identifier

			// Create a block for the catch variable
			prs.block = newBlock(prs.block)

			// Define catch variable in this block (distinct)
			varDef, ok := prs.block.defineVariable(prs, catchVarName, DECL_LET)
			if !ok {
				prs.addError("Cannot redeclare catch variable")
				prs.block = prs.block.parent
				return
			}
			varDef.initialized = true
			tryCtx.CatchVarSlot = varDef.slotIndex

			if prs.lex() != GTOK_RP {
				prs.addError("Expected ')' after catch variable")
				prs.block = prs.block.parent
				return
			}
			tok = prs.lex()
		}

		// Require opening brace for catch block
		if tok != GTOK_LC {
			prs.addError("Expected '{' after 'catch'")
			if catchVarName != "" {
				prs.block = prs.block.parent
			}
			return
		}

		// Parse catch block (add depth to discard expression value)
		prs.blockDepth++
		prs.parseStatementList()
		prs.blockDepth--

		if prs.ctx.sym.token != GTOK_RC {
			prs.addError("Expected '}' at end of catch block")
		} else {
			prs.lex()
		}

		// Pop catch variable block if we created one
		if catchVarName != "" {
			prs.block = prs.block.parent
		}

		// Add jump from catch to finally or end
		jmpFromCatch = prs.pushOpCode(engine.JumpOperation, 0)

		tok = prs.ctx.sym.token
	}

	// Handle optional finally clause
	finallyStart := len(prs.body.Code)
	if tok == GTOK_FINALLY {
		hasFinally = true
		tryCtx.FinallyTarget = finallyStart

		if prs.lex() != GTOK_LC {
			prs.addError("Expected '{' after finally")
			return
		}

		// Parse finally block (add depth to discard expression value)
		prs.blockDepth++
		prs.parseStatementList()
		prs.blockDepth--

		if prs.ctx.sym.token != GTOK_RC {
			prs.addError("Expected '}' at end of finally block")
		} else {
			prs.lex()
		}

		// Add operation to handle re-throw if exception is propagating
		prs.pushOpCode(engine.FinallyCompleteOperation, 0)
	}

	// Must have at least catch or finally
	if !hasCatch && !hasFinally {
		prs.addError("try without catch or finally")
		return
	}

	// Update all of the jump targets
	endTarget := len(prs.body.Code)
	tryCtx.EndTarget = endTarget

	if hasFinally {
		// Try and catch both jump to finally
		jmpFromTry.OpData = finallyStart
		if jmpFromCatch != nil {
			jmpFromCatch.OpData = finallyStart
		}
	} else {
		// No finally, jump to end instead
		jmpFromTry.OpData = endTarget
		if jmpFromCatch != nil {
			jmpFromCatch.OpData = endTarget
		}
	}
}
