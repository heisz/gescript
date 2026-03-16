/*
 * Various structures and other elements to support the parse/execute model.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"strconv"

	"github.com/heisz/gescript/internal/engine"
)

// Originally the tracking context for the Yacc parser, now a hand-built version
type parser struct {
	ctx           *lexer
	body          *engine.Function
	rootBlock     *blockContext
	block         *blockContext
	blockDepth    int
	loopSwitchCtx *loopSwitchContext
	pendingLabel  string
	outerScope    *outerScopeContext
	captures      []captureEntry
	errors        []error
}

// Outer scope context for closure variable resolution
type outerScopeContext struct {
	parent *outerScopeContext
	block  *blockContext
	// This points into the outer functions capture list
	captures *[]captureEntry
}

// Tracking structure for variables to be captured in closure
type captureEntry struct {
	name string
	// This is either the outer scope variable or capture index, based on flag
	slotIndex int
	isCapture bool
}

// Without yacc, can have a richer error instance from the lower elements
// Fully exposed so external context can extract details if needed
type ParserError struct {
	LineNumber int
	ErrorMsg   string
}

func (err *ParserError) Error() string {
	return "[Line " + strconv.Itoa(err.LineNumber) + "] " + err.ErrorMsg
}

func parserError(ctx *lexer, msg string) *ParserError {
	return &ParserError{
		LineNumber: ctx.lineNumber,
		ErrorMsg:   msg,
	}
}

// Convenience method for recording parsing errors with context
func (prs *parser) addError(msg string) {
	prs.errors = append(prs.errors, parserError(prs.ctx, msg))
}

// Variable declaration types for hoisting and mutability rules
type varDeclType int

const (
	DECL_VAR varDeclType = iota
	DECL_LET
	DECL_CONST
)

// Variable definition within a block or function scope (possible outer capture)
type variable struct {
	name        string
	declType    varDeclType
	slotIndex   int
	initialized bool
	isCapture   bool
	captureIdx  int
}

// Declare a variable in the block, checking for (illegal) redeclarations
func (blk *blockContext) defineVariable(prs *parser, name string,
	declType varDeclType) (*variable, bool) {
	if existing, ok := blk.variables[name]; ok {
		if declType != DECL_VAR || existing.declType != DECL_VAR {
			// Cannot redeclare non-var within existing block
			return nil, false
		}
		return existing, true
	}

	vr := &variable{
		name:      name,
		declType:  declType,
		slotIndex: blk.slotBase + blk.slotCount,
		// var declarations are auto-initialized to undefined
		initialized: declType == DECL_VAR,
	}
	blk.variables[name] = vr
	blk.slotCount++

	// Track maximum slot usage for function allocation
	if (blk.slotBase + blk.slotCount) > prs.body.VarCount {
		prs.body.VarCount = blk.slotBase + blk.slotCount
	}

	return vr, true
}

// Resolve a variable by walking up the block scope chain
func (blk *blockContext) resolveVariable(name string) *variable {
	// Could do this recursively but trivial to walk ourselves
	for b := blk; b != nil; b = b.parent {
		if vr, ok := b.variables[name]; ok {
			return vr
		}
	}
	return nil
}

// Ensure captures from a higher/more outer scope are cascaded through parent
func (prs *parser) cascadeCapture(name string, target *outerScopeContext) int {
	if prs.outerScope == nil || prs.outerScope.captures == nil {
		return -1
	}

	// Trivial if immediate parent has already captured it
	for idx, cap := range *prs.outerScope.captures {
		if cap.name == name {
			return idx
		}
	}

	// Need to cascade capture to our immediate parent
	for outer := prs.outerScope; outer != nil; outer = outer.parent {
		if outer == target {
			// Found the scope with the actual variable - capture from locals
			outerVar := outer.block.resolveVariable(name)
			if outerVar != nil {
				capIdx := len(*prs.outerScope.captures)
				*prs.outerScope.captures = append(*prs.outerScope.captures,
					captureEntry{
						name:      name,
						slotIndex: outerVar.slotIndex,
						isCapture: false,
					})
				return capIdx
			}
		}

		// Shortcut if this outer context has already captured the variable
		if outer.captures != nil {
			for outerIdx, outerCap := range *outer.captures {
				if outerCap.name == name {
					if outer == prs.outerScope {
						return outerIdx
					}

					// Chain to the parent capture record
					capIdx := len(*prs.outerScope.captures)
					*prs.outerScope.captures = append(*prs.outerScope.captures,
						captureEntry{
							name:      name,
							slotIndex: outerIdx,
							isCapture: true,
						})
					return capIdx
				}
			}
		}
	}

	return -1
}

// Resolve a variable from outer scope for closure capture (returns idx/-1)
func (prs *parser) resolveCapture(name string) int {
	// Trivial if previously captured
	for idx, cap := range prs.captures {
		if cap.name == name {
			return idx
		}
	}

	// Resolve up the chain of outer scope instances
	for outer := prs.outerScope; outer != nil; outer = outer.parent {
		// First check for local variable instances
		if outer.block != nil {
			outerVar := outer.block.resolveVariable(name)
			if outerVar != nil {
				// Found in outer locals - add to captures with cascade?
				capIdx := len(prs.captures)
				if outer == prs.outerScope {
					// Immediate parent, direct variable capture
					prs.captures = append(prs.captures, captureEntry{
						name:      name,
						slotIndex: outerVar.slotIndex,
						isCapture: false,
					})
				} else {
					// Not immediate, ensure parent has the capture instance
					parentCapIdx := prs.cascadeCapture(name, outer)
					prs.captures = append(prs.captures, captureEntry{
						name:      name,
						slotIndex: parentCapIdx,
						isCapture: true,
					})
				}
				return capIdx
			}
		}

		// Shortcut if this outer context has already captured the variable
		if outer.captures != nil {
			for outerIdx, outerCap := range *outer.captures {
				if outerCap.name == name {
					capIdx := len(prs.captures)
					if outer == prs.outerScope {
						// Direct parent capture instance
						prs.captures = append(prs.captures, captureEntry{
							name:      name,
							slotIndex: outerIdx,
							isCapture: true,
						})
					} else {
						// Need to cascade through the parent instance
						parentCapIdx := prs.cascadeCapture(name, outer)
						prs.captures = append(prs.captures, captureEntry{
							name:      name,
							slotIndex: parentCapIdx,
							isCapture: true,
						})
					}
					return capIdx
				}
			}
		}
	}

	return -1
}

// Blocks are a lexical 'scope', stored as a tree to the root function block
type blockContext struct {
	parent    *blockContext
	variables map[string]*variable
	slotBase  int
	slotCount int
}

// Create a new block with the given parent
func newBlock(parent *blockContext) *blockContext {
	base := 0
	if parent != nil {
		base = parent.slotBase + parent.slotCount
	}

	return &blockContext{
		parent:    parent,
		variables: make(map[string]*variable),
		slotBase:  base,
		slotCount: 0,
	}
}

// Loop/switch context for (labelled) break and continue target tracking
type loopSwitchContext struct {
	parent *loopSwitchContext

	// Note that the label is optional
	label string

	// Track jumps to continue/break once known, target -1 until known
	continueTarget int
	continueJumps  []*engine.OpCode
	breakJumps     []*engine.OpCode
	isSwitch       bool
}

// Push a new loop/switch context, consuming any pending label
func (prs *parser) pushLoopSwitchContext(isSwitch bool) *loopSwitchContext {
	ctx := &loopSwitchContext{
		parent:         prs.loopSwitchCtx,
		label:          prs.pendingLabel,
		continueTarget: -1,
		isSwitch:       isSwitch,
	}
	prs.pendingLabel = ""
	prs.loopSwitchCtx = ctx
	return ctx
}

// Pop the current loop context and patch break jumps to current position
func (prs *parser) popLoopContext() {
	if prs.loopSwitchCtx == nil {
		return
	}
	ctx := prs.loopSwitchCtx
	prs.loopSwitchCtx = ctx.parent

	// All breaks in the context point to the current (exit) position
	exitTarget := len(prs.body.Code)
	for _, jmp := range ctx.breakJumps {
		jmp.OpData = exitTarget
	}
}

// Entry point to the parser, returning a function body or error
// Note: this goes against 'best' practice for an errorlist return but the
// top-level wrapper will do it properly (avoiding circular import)
func Parse(source string) (body *engine.Function, err []error) {
	blk := newBlock(nil)
	prs := parser{
		ctx:       newLexer(source),
		body:      engine.NewFunction("_"),
		rootBlock: blk,
		block:     blk,
	}
	prs.parseStatementList()
	return prs.body, prs.errors
}

// Direct wrapper for regular lex, with automatic error recording
func (prs *parser) lex() (token int) {
	token, err := prs.ctx.lex(&prs.ctx.sym)
	if err != nil {
		prs.errors = append(prs.errors, err)
	}
	return token
}

/*
 * This is used in multiple contexts, from Section 15.1 and 13.2
 *
 * ScriptBody:
 *    StatementList
 *
 * StatementList:
 *    StatementListItem
 *    | StatementList StatementListItem
 *
 * This method is called with the lexer located before the first statement.
 */
func (prs *parser) parseStatementList() {
	// Lex the first token to start
	token := prs.lex()

	for token != GTOK_EOF {
		if token == GTOK_ERROR || token == GTOK_RC {
			return
		}

		// Track lexer position to detect infinite parse loops
		startOffset := prs.ctx.offset

		prs.parseStatementListItem(token)

		// After statement, check for exit conditions
		token = prs.ctx.sym.token
		if token == GTOK_RC || token == GTOK_EOF {
			return
		}

		// Prevent infinite loops - if no progress was made, skip token
		if prs.ctx.offset == startOffset {
			prs.addError("Unexpected token in statement list")
			token = prs.lex()
			continue
		}

		// Skip semicolon if statement ended on it, all others are on next
		if token == GTOK_SEMI {
			token = prs.lex()
		}
	}
}

// Parser utility method to push an opcode onto the current context body
func (prs *parser) pushOpCode(opFn engine.OpCodeFn,
	stackAdjust int) (op *engine.OpCode) {
	op = &engine.OpCode{
		ExecFn: opFn,
	}
	prs.body.Code = append(prs.body.Code, op)
	return op
}

// Method to translate a $$ expression into the appropriate opcode
func (prs *parser) pushEvalExpression(expr *symType) bool {
	if expr == nil {
		return false
	}
	switch expr.parseType {
	case PARSED_LITERAL:
		op := prs.pushOpCode(engine.PushLiteralValue, 1)
		op.OpData = expr.literal
		return true
	case PARSED_VALUE:
		// Opcodes already generated during parsing, nothing to do
		return true
	case PARSED_IDENTIFIER:
		// Resolve variable (with error) and push load operation
		varDef := prs.block.resolveVariable(expr.identifier)
		if varDef == nil {
			prs.addError("Undefined variable '" + expr.identifier + "'")
			return false
		}
		op := prs.pushOpCode(engine.LoadVariableOperation, 1)
		op.OpData = varDef.slotIndex
		return true
	case PARSED_GLOBAL_REFERENCE:
		// Read value from global context (native/builtin functions)
		op := prs.pushOpCode(engine.LoadGlobalOperation, 1)
		op.OpData = expr.identifier
		return true
	case PARSED_ARRAY_REFERENCE:
		// Target and index already parsed, push the element retrieve op
		prs.pushOpCode(engine.GetElementOperation, -1)
		return true
	case PARSED_MEMBER_REFERENCE:
		// Target already parsed, expression identifier contains property name
		op := prs.pushOpCode(engine.GetPropertyOperation, 0)
		op.OpData = expr.identifier
		return true
	case PARSED_CAPTURE_REFERENCE:
		// Captured variable reference from closure
		op := prs.pushOpCode(engine.LoadCaptureOperation, 1)
		op.OpData = expr.assignOp
		return true
	}

	prs.addError("Parser error, invalid expression type")
	return false
}
