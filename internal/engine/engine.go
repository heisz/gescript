/*
 * Primary structures/implementations for the execution of opcode bodies.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package engine

import (
	"errors"
	"os"

	"github.com/heisz/gescript/types"
)

// Special error to indicate a thrown exception
var ErrException = errors.New("exception thrown")

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

// Exception handler context for try blocks, -1 means no target/variable
type ExceptionContext struct {
	previous      *ExceptionContext
	CatchTarget   int
	FinallyTarget int
	EndTarget     int
	StackDepth    int
	CatchVarSlot  int
}

// Storage structure for execution context at a function call boundary
type CallFrame struct {
	previous *CallFrame
	body     *Function
	pc       int
	sp       int
	locals   []*types.DataType
	cells    []*Cell
	closure  []*Cell
}

// Like that other project, process is the execution context of the opcodes
// (in a strict bytecode standard this would be the virtual machine)
type Process struct {
	// Current code body being executed and associated program counter
	body *Function
	pc   int

	// It's a stack-based machine, here is the stack (sp points to next slot)
	stack []*types.DataType
	sp    int

	// Local variables for the current function execution frame
	locals []*types.DataType

	// Linked list of current exception handler contexts (nested)
	exceptionCtx *ExceptionContext

	// Current exception being propagated (nil if none)
	exception *types.DataType

	// Call stack for function invocations
	callStack *CallFrame

	// Cells for local variables that are captured by closures
	cells []*Cell

	// Closure cells from enclosing scopes (for captured variables)
	closure []*Cell

	// Native functions (library and program-provided extensions)
	natives map[string]*types.DataType

	// Script globals - functions defined by script
	globals map[string]*types.DataType
}

func NewProcess(depth int, natives map[string]*types.DataType,
	globals map[string]*types.DataType) *Process {
	prc := Process{
		stack:   make([]*types.DataType, depth),
		sp:      0,
		natives: natives,
		globals: globals,
	}
	return &prc
}

// Any stack needs push, peek and pop, push supporting dynamic resizing
// Note that many operations perform direct manipulation and that's ok
func (prc *Process) push(val *types.DataType) (err error) {
	// TODO - resize and overflow!
	prc.stack[prc.sp] = val
	prc.sp++

	return nil
}
func (prc *Process) peek() (val *types.DataType, err error) {
	// TODO - stack underflow
	return prc.stack[prc.sp-1], nil
}
func (prc *Process) pop() (val *types.DataType, err error) {
	// TODO - stack underflow
	prc.sp--
	return prc.stack[prc.sp], nil
}

// The fundamental model in this implementation is that every set of code is
// a function.  For all of the functions that's obvious and the uncontained
// code is compiled into an anonymous function instance.
type Function struct {
	Name     string
	Code     []*OpCode
	VarCount int
}

func NewFunction(nm string) *Function {
	fn := Function{
		Name: nm,
	}
	return &fn
}

// Execution 'loop' to run the given function in the associated process
func (body *Function) Exec(prc *Process) (ret types.DataType, err error) {
	// For now, just ram this in
	prc.body = body
	prc.pc = 0
	prc.stack = make([]*types.DataType, 16)
	prc.sp = 0
	prc.exceptionCtx = nil
	prc.exception = nil

	// Allocate locals array for variable storage (all undefined)
	if body.VarCount > 0 {
		prc.locals = make([]*types.DataType, body.VarCount)
		// Initialize all slots to undefined
		undef := types.DataType(types.Undefined)
		for idx := 0; idx < body.VarCount; idx++ {
			prc.locals[idx] = &undef
		}
	}

	// Loop until we run out of runway
	for {
		pc := prc.pc
		if pc < 0 || pc >= len(prc.body.Code) {
			// The first shouldn't happen but just in case...
			break
		}
		op := prc.body.Code[pc]
		TestLog("EXECFN!")
		opErr := op.ExecFn(prc, op)
		if opErr != nil {
			if opErr == ErrException {
				if !prc.handleException() {
					// No handler in stack for exception instance
					return nil, errors.New("Uncaught exception")
				}
				// Handler found, frame updated, continue execution
			} else {
				return nil, opErr
			}
		}
		prc.pc++
	}

	// The last item on the stack is the outer return value
	if prc.sp > 0 {
		return *prc.stack[prc.sp-1], nil
	}
	return types.Undefined, nil
}

// Handle an exception by unwinding to context, returns true if 'handled'
func (prc *Process) handleException() bool {
	for prc.exceptionCtx != nil {
		// Pop the topmost try context
		ctx := prc.exceptionCtx
		prc.exceptionCtx = ctx.previous

		// Restore stack depth
		prc.sp = ctx.StackDepth

		// Check for catch handler, store exception if needed and reframe
		if ctx.CatchTarget >= 0 {
			if ctx.CatchVarSlot >= 0 {
				prc.locals[ctx.CatchVarSlot] = prc.exception
			}
			prc.exception = nil
			prc.pc = ctx.CatchTarget - 1
			return true
		}

		// No catch, check for finally
		if ctx.FinallyTarget >= 0 {
			// TODO - need to handle this properly...
		}
	}

	return false
}
