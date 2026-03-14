/*
 * Implementations of standard elements for the function type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"github.com/heisz/gescript/types"
)

// Note: in all instance methods, args[0] is 'this', aka the function instance

// Also note that formal 'this' is still TODO so arg1 is ignored - a lot

func functionApply(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	// Note that we don't support 'this' yet, so we skip it for now
	var callArgs []types.DataType
	if len(args) > 2 {
		if argsArray, ok := args[2].(*types.ArrayType); ok {
			callArgs = argsArray.Elements
		}
	}

	return fn.Call(callArgs)
}

func functionBind(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	// Capture bound arguments for the nested function below
	// Note that we don't support 'this' yet, so we skip it for now
	var boundArgs []types.DataType
	if len(args) > 2 {
		boundArgs = make([]types.DataType, len(args)-2)
		copy(boundArgs, args[2:])
	}

	// Return a new function instance that prepends bound arguments
	boundFn := &types.NativeFunction{
		Name: "bound " + fn.GetName(),
		Fn: func(callArgs []types.DataType) (types.DataType, error) {
			allArgs := make([]types.DataType, len(boundArgs)+len(callArgs))
			copy(allArgs, boundArgs)
			copy(allArgs[len(boundArgs):], callArgs)
			return fn.Call(allArgs)
		},
	}

	return boundFn, nil
}

func functionCall(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}

	fn, ok := args[0].(types.FunctionType)
	if !ok {
		return types.Undefined, nil
	}

	// Note that we don't support 'this' yet, so we skip it for now
	var callArgs []types.DataType
	if len(args) > 2 {
		callArgs = args[2:]
	}

	return fn.Call(callArgs)
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
