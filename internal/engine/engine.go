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

	"github.com/heisz/gescript/types"
)

// Special errors to indicate a thrown exception and stack underflow
var ErrException = errors.New("exception thrown")
var ErrStackUnderflow = errors.New("stack underflow")

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

	// Indicator for finally handling/rethrow during exception propagation
	finallyRethrow bool

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
	// Classic doubling model for stack expansion
	if prc.sp >= len(prc.stack) {
		newStack := make([]types.DataType, len(prc.stack)*2)
		copy(newStack, prc.stack)
		prc.stack = newStack
	}
	prc.stack[prc.sp] = val
	prc.sp++

	return nil
}
func (prc *Process) peek() (val types.DataType, err error) {
	if prc.sp <= 0 {
		// In theory this never happens but handle it regardless
		return nil, ErrStackUnderflow
	}
	return prc.stack[prc.sp-1], nil
}
func (prc *Process) pop() (val types.DataType, err error) {
	if prc.sp <= 0 {
		// In theory this never happens but handle it regardless
		return nil, ErrStackUnderflow
	}
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

		// No catch, check for finally - run finally then re-throw
		if ctx.FinallyTarget >= 0 {
			prc.finallyRethrow = true
			prc.pc = ctx.FinallyTarget - 1
			return true
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
	ThisSlot      int
	IsArrowFunc   bool

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

// Note that the standard 'call' method is just an undefined this
func (sf *ScriptFunction) Call(args []types.DataType) (types.DataType, error) {
	return sf.CallWithThis(types.Undefined, args)
}

func (sf *ScriptFunction) CallWithThis(thisVal types.DataType,
	args []types.DataType) (types.DataType, error) {
	// Create a new process and set up execution context
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

	bindFunctionParams(prc.locals, sf, thisVal, args)

	// Run the execution loop
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

// Common method to handle this/param/args binding to script variables
func bindFunctionParams(locals []types.DataType, sf *ScriptFunction,
	thisVal types.DataType, args []types.DataType) {
	paramCount := len(sf.ParamNames)

	// Handle rest parameter if applicable
	if sf.HasRestParam && paramCount > 0 {
		//  All but the last parameter variables get the provided arguments
		for idx := 0; idx < paramCount-1 && idx < len(args); idx++ {
			locals[idx] = args[idx]
		}

		// Remaining arguments assemble into an array for the last parameter
		restStart := paramCount - 1
		if restStart < len(args) {
			restArr := types.NewArray(len(args) - restStart)
			copy(restArr.Elements, args[restStart:])
			locals[restStart] = restArr
		} else {
			locals[restStart] = types.NewArray(0)
		}
	} else {
		// Normal mode, populate the parameter variable values
		for idx := 0; idx < paramCount && idx < len(args); idx++ {
			locals[idx] = args[idx]
		}
	}

	// Create arguments object if slot is defined
	if sf.ArgumentsSlot >= 0 {
		argsArr := types.NewArray(len(args))
		copy(argsArr.Elements, args)
		locals[sf.ArgumentsSlot] = argsArr
	}

	// Likewise for the this variable
	if sf.ThisSlot >= 0 {
		locals[sf.ThisSlot] = thisVal
	}
}

// BoundFunction wraps a function with a bound this value and arguments
type BoundFunction struct {
	Target    types.FunctionType
	BoundThis types.DataType
	BoundArgs []types.DataType
}

func (bf *BoundFunction) Native() interface{} {
	return bf.Target.Native()
}

func (bf *BoundFunction) ToPrimitive(pref any) types.DataType {
	return types.StringType("function bound " + bf.Target.GetName() +
		"() { [bound] }")
}

func (bf *BoundFunction) GetName() string {
	return "bound " + bf.Target.GetName()
}

func (bf *BoundFunction) Call(args []types.DataType) (types.DataType, error) {
	// Combine bound args with call args
	fullArgs := make([]types.DataType, len(bf.BoundArgs)+len(args))
	copy(fullArgs, bf.BoundArgs)
	copy(fullArgs[len(bf.BoundArgs):], args)

	// For native functions, prepend bound this as first argument
	if nf, ok := bf.Target.(*types.NativeFunction); ok {
		thisArgs := make([]types.DataType, len(fullArgs)+1)
		thisArgs[0] = bf.BoundThis
		copy(thisArgs[1:], fullArgs)
		return nf.Fn(thisArgs)
	}

	// For script functions, use CallWithThis
	if sf, ok := bf.Target.(*ScriptFunction); ok {
		return sf.CallWithThis(bf.BoundThis, fullArgs)
	}

	// For other function types, use standard call
	return bf.Target.Call(fullArgs)
}

func (bf *BoundFunction) GetBoundThis() types.DataType {
	return bf.BoundThis
}
