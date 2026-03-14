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
	locals   []types.DataType
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
	stack []types.DataType
	sp    int

	// Local variables for the current function execution frame
	locals []types.DataType

	// Linked list of current exception handler contexts (nested)
	exceptionCtx *ExceptionContext

	// Current exception being propagated - can still be nil for 'none'
	exception *types.DataType

	// Call stack for function invocations
	callStack *CallFrame

	// Cells for local variables that are captured by closures
	cells []*Cell

	// Closure cells from enclosing scopes (for captured variables)
	closure []*Cell

	// Native functions (library and program-provided extensions)
	natives map[string]types.DataType

	// Script globals - functions defined by script
	globals map[string]types.DataType

	// Registered constructors with instance method resolution
	constructors []*types.NativeConstructor
}

// A cell wraps a value by reference for closure sharing
type Cell struct {
	// Does need a pointer in this case for explicit sharing
	Value *types.DataType
}

func NewProcess(depth int, natives map[string]types.DataType,
	globals map[string]types.DataType,
	constructors []*types.NativeConstructor) *Process {
	prc := Process{
		stack:        make([]types.DataType, depth),
		sp:           0,
		natives:      natives,
		globals:      globals,
		constructors: constructors,
	}
	return &prc
}

// Any stack needs push, peek and pop, push supporting dynamic resizing
// Note that many operations perform direct manipulation and that's ok
func (prc *Process) push(val types.DataType) (err error) {
	// TODO - resize and overflow!
	prc.stack[prc.sp] = val
	prc.sp++

	return nil
}
func (prc *Process) peek() (val types.DataType, err error) {
	// TODO - stack underflow
	return prc.stack[prc.sp-1], nil
}
func (prc *Process) pop() (val types.DataType, err error) {
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
	prc.stack = make([]types.DataType, 16)
	prc.sp = 0
	prc.exceptionCtx = nil
	prc.exception = nil

	// Allocate locals array for variable storage (all undefined)
	if body.VarCount > 0 {
		prc.locals = make([]types.DataType, body.VarCount)
		for idx := 0; idx < body.VarCount; idx++ {
			prc.locals[idx] = types.Undefined
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
		return prc.stack[prc.sp-1], nil
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
				prc.locals[ctx.CatchVarSlot] = *prc.exception
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

// Tracking element for captured variables in enclosed function (closure) calls
type CaptureInfo struct {
	Name string
	// Index is to local vars or closure cells based on flag
	SlotIndex int
	IsCapture bool
}

// This structure wraps the core Function for script-defined functions
type ScriptFunction struct {
	Name          string
	ParamNames    []string
	HasRestParam  bool
	Body          *Function
	VarCount      int
	ArgumentsSlot int

	// Lists of variables from enclosing scopes to capture
	Captures []CaptureInfo

	// Populated during runtime, set of cells from enclosing scopes for closures
	Closure []*Cell
}

// Tracking data for a function call with spread arguments
type CallSpreadInfo struct {
	ArgCount   int
	SpreadMask []bool
}

// Tracking data for array literals with spread arguments
type ArraySpreadInfo struct {
	ElemCount  int
	SpreadMask []bool
}

// Implementations for the DataType and FunctionType interfaces
func (sf *ScriptFunction) Native() interface{} {
	return sf
}

func (sf *ScriptFunction) ToPrimitive(pref any) types.DataType {
	return types.StringType("function " + sf.Name + "() { [script code] }")
}

func (sf *ScriptFunction) GetName() string {
	return sf.Name
}

func (sf *ScriptFunction) Call(args []types.DataType) (types.DataType, error) {
	// Create a new process and set up execution context directly
	prc := NewProcess(256, nil, nil, nil)
	prc.body = sf.Body
	prc.pc = 0
	prc.closure = sf.Closure
	prc.cells = nil
	prc.exceptionCtx = nil
	prc.exception = nil

	// Initialize the local variable storage for the function
	if sf.Body.VarCount > 0 {
		prc.locals = make([]types.DataType, sf.Body.VarCount)
		for idx := 0; idx < sf.Body.VarCount; idx++ {
			prc.locals[idx] = types.Undefined
		}
	} else {
		prc.locals = nil
	}

	// Handle rest parameter if specified
	paramCount := len(sf.ParamNames)
	if sf.HasRestParam && paramCount > 0 {
		// All but the last parameter variables get the provided arguments
		for idx := 0; idx < paramCount-1 && idx < len(args); idx++ {
			prc.locals[idx] = args[idx]
		}

		// Remaining arguments assemble into an array for the last parameter
		restStart := paramCount - 1
		if restStart < len(args) {
			restArr := types.NewArray(len(args) - restStart)
			copy(restArr.Elements, args[restStart:])
			prc.locals[restStart] = restArr
		} else {
			prc.locals[restStart] = types.NewArray(0)
		}
	} else {
		// Normal mode, populate the parameter variable values
		for idx := 0; idx < paramCount && idx < len(args); idx++ {
			prc.locals[idx] = args[idx]
		}
	}

	// Create the arguments object if slot is defined (by usage)
	if sf.ArgumentsSlot >= 0 {
		argsArr := types.NewArray(len(args))
		copy(argsArr.Elements, args)
		prc.locals[sf.ArgumentsSlot] = argsArr
	}

	// Run the execution loop directly
	for {
		pc := prc.pc
		if pc < 0 || pc >= len(prc.body.Code) {
			break
		}
		op := prc.body.Code[pc]
		opErr := op.ExecFn(prc, op)
		if opErr != nil {
			if opErr == ErrException {
				if !prc.handleException() {
					return types.Undefined, errors.New("Uncaught exception")
				}
			} else {
				return types.Undefined, opErr
			}
		}
		prc.pc++
	}

	// The last item on the stack is the return value
	if prc.sp > 0 {
		return prc.stack[prc.sp-1], nil
	}
	return types.Undefined, nil
}
