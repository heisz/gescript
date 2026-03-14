/*
 * Implementations of standard elements for the function type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/types"
)

// Note: in all instance methods, args[0] is 'this', aka the function instance

// Handle all of the different function types calling with 'this'
// Note that this aligns with the callFunctionWithThis in the engine
func callWithThis(ft types.FunctionType, thisArg types.DataType,
	callArgs []types.DataType) (types.DataType, error) {
	switch fn := ft.(type) {
	case *types.NativeFunction:
		// Native functions don't have a this element
		return fn.Fn(callArgs)

	case *types.NativeMethod:
		// Native methods are already tied to a this element
		return fn.Call(callArgs)

	case *engine.BoundFunction:
		// Bound functions loop back with their internal bound target
		return callWithThis(fn.Target, fn.BoundThis,
			append(fn.BoundArgs, callArgs...))

	case *engine.ScriptFunction:
		// Script functions have the direct execution method
		return fn.CallWithThis(thisArg, callArgs)

	default:
        // Fallback is an open this-less function
		return ft.Call(callArgs)
	}
}

func functionApply(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	var thisArg types.DataType = types.Undefined
	if len(args) > 1 {
		thisArg = args[1]
	}

    // For apply the second argument is an array of call arguments
	var callArgs []types.DataType
	if len(args) > 2 {
		if argsArray, ok := args[2].(*types.ArrayType); ok {
			callArgs = argsArray.Elements
		}
	}

	return callWithThis(fn, thisArg, callArgs)
}

func functionBind(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	var boundThis types.DataType = types.Undefined
	if len(args) > 1 {
		boundThis = args[1]
	}

	// Clone the remaining arguments as pre-bound arglist
	var boundArgs []types.DataType
	if len(args) > 2 {
		boundArgs = make([]types.DataType, len(args)-2)
		copy(boundArgs, args[2:])
	}

	// Generate a bound function instance with passed elements
	return &engine.BoundFunction{
		Target:    fn,
		BoundThis: boundThis,
		BoundArgs: boundArgs,
	}, nil
}

func functionCall(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	var thisArg types.DataType = types.Undefined
	if len(args) > 1 {
		thisArg = args[1]
	}

	// Similar to apply but in this case remaining args are call args
	var callArgs []types.DataType
	if len(args) > 2 {
		callArgs = args[2:]
	}

	return callWithThis(fn, thisArg, callArgs)
}

func functionToString(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("function () { }"), nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.StringType("function () { }"), nil
	}

	// ToPrimitive already generates the named descriptor
	return fn.ToPrimitive(nil), nil
}

// Resolve properties and methods for the Function type
func functionMemberResolver(target types.DataType, name string) types.DataType {
	// It needs to be a function instance or return nil
	fn, ok := target.(types.FunctionType)
	if !ok {
		return nil
	}

	// Two regular properties for a function
	switch name {
	case "name":
		return types.StringType(fn.GetName())
	case "length":
		// This should return parameter count but we don't track it, zero...
		return types.IntegerType(0)
	}

	// Otherwise look up the function instance methods
	var method *types.NativeFunction
	switch name {
	case "apply":
		method = &types.NativeFunction{Name: "apply",
			Fn: functionApply}
	case "bind":
		method = &types.NativeFunction{Name: "bind",
			Fn: functionBind}
	case "call":
		method = &types.NativeFunction{Name: "call",
			Fn: functionCall}
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: functionToString}
	default:
		return nil
	}
	return &types.NativeMethod{Target: target, Method: method}
}

// Create the Function global constructor with member methods
func NewFunctionConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("Function",
		func(args []types.DataType) (types.DataType, error) {
			// Return a no-op function, don't have full parse setup (TODO)
			return &types.NativeFunction{
				Name: "anonymous",
				Fn: func(args []types.DataType) (types.DataType, error) {
					return types.Undefined, nil
				},
			}, nil
		})

	ctor.InstanceMembers = functionMemberResolver

	return ctor
}
