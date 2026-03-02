/*
 * Primary entry point to the gescript module (parse and execute).
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package gescript

import (
	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/internal/native"
	"github.com/heisz/gescript/internal/parser"
	"github.com/heisz/gescript/types"
)

// Exposed container for a parsed script instance
type Script struct {
	body *engine.Function
}

/*
 * ScriptContext defines an execution context for scripts, holding the
 * native/builtin function, the custom native extensions defined by the caller
 * as well as any script-defined globals.
 */
type ScriptContext struct {
	// Map of native (builtin and program defined) functions
	natives map[string]*types.DataType

	// Map of script defined global functions
	globals map[string]*types.DataType
}

// NewScriptContext creates a new execution context with builtin native fns
func NewScriptContext() *ScriptContext {
	ctx := &ScriptContext{
		natives: make(map[string]*types.DataType),
		globals: make(map[string]*types.DataType),
	}
	native.RegisterNatives(ctx.natives)
	return ctx
}

// Clone allows the creation of a core script library and then using it in
// isolation for execution.
func (ctx *ScriptContext) Clone() *ScriptContext {
	res := &ScriptContext{
		natives: make(map[string]*types.DataType, len(ctx.natives)),
		globals: make(map[string]*types.DataType, len(ctx.globals)),
	}
	for key, val := range ctx.natives {
		res.natives[key] = val
	}
	for key, val := range ctx.globals {
		res.globals[key] = val
	}
	return res
}

// Register a native function in the context for script usage
func (ctx *ScriptContext) RegisterFunction(name string, fn types.NativeFn) {
	nativeFunc := &types.NativeFunction{
		Name: name,
		Fn:   fn,
	}
	fnVal := types.DataType(nativeFunc)
	ctx.natives[name] = &fnVal
}

// Parse the source script into an executable Script instance (or error)
func Parse(source string) (prg *Script, err error) {
	prsBody, errs := parser.Parse(source)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &Script{
		body: prsBody,
	}, nil
}

// Run executes the script with a 'standard' context (no additions)
func (prg *Script) Run() (retval types.DataType, err error) {
	return prg.RunWithContext(NewScriptContext())
}

// Run the script with the provided context (native/custom extensions)
func (prg *Script) RunWithContext(ctx *ScriptContext) (retval types.DataType,
	err error) {
	prc := engine.NewProcess(256, ctx.natives, ctx.globals)
	return prg.body.Exec(prc)
}

// Convenience method to parse/execute the script source in one shot (no ctx)
func Run(source string) (retval types.DataType, err error) {
	script, err := Parse(source)
	if err != nil {
		return types.Undefined, err
	}
	return script.Run()
}

// Retrieve a defined function from the context to call() directly
func (ctx *ScriptContext) GetFunction(name string) types.FunctionType {
	if ctx.globals == nil {
		return nil
	}
	fnVal, ok := ctx.globals[name]
	if !ok || fnVal == nil {
		return nil
	}
	fn, ok := (*fnVal).(types.FunctionType)
	if !ok {
		return nil
	}
	return fn
}

// Retrieve a global variable from the context by name (undef if not found)
func (ctx *ScriptContext) GetGlobal(name string) types.DataType {
	if ctx.globals == nil {
		return types.Undefined
	}
	val, ok := ctx.globals[name]
	if !ok || val == nil {
		return types.Undefined
	}
	return *val
}

// Set a global variable by name, which can be referenced by scripts
func (ctx *ScriptContext) SetGlobal(name string, val types.DataType) {
	if ctx.globals == nil {
		ctx.globals = make(map[string]*types.DataType)
	}
	ctx.globals[name] = &val
}
