/*
 * Implementations of standard elements for the boolean type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"github.com/heisz/gescript/types"
)

// Note: in all instance methods, args[0] is 'this', aka the boolean instance

func booleanToString(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("false"), nil
	}

	b, ok := args[0].(types.BooleanType)
	if !ok {
		return types.StringType("false"), nil
	}

	if bool(b) {
		return types.StringType("true"), nil
	}
	return types.StringType("false"), nil
}

func booleanValueOf(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}
	return args[0], nil
}

// Resolve properties and methods for the Boolean type
func booleanMemberResolver(target types.DataType, name string) types.DataType {
	// Check if target is a boolean type
	_, isBoolean := target.(types.BooleanType)
	if !isBoolean {
		return nil
	}

	// Handle instance methods
	var method *types.NativeFunction
	switch name {
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: booleanToString}
	case "valueOf":
		method = &types.NativeFunction{Name: "valueOf",
			Fn: booleanValueOf}
	default:
		return nil
	}
	return &types.NativeMethod{Target: target, Method: method}
}

// Create the Boolean global constructor with member elements
func NewBooleanConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("Boolean",
		func(args []types.DataType) (types.DataType, error) {
			if len(args) == 0 {
				return types.BooleanType(false), nil
			}
			return types.BooleanType(types.IsTruthy(args[0])), nil
		})

	ctor.InstanceMembers = booleanMemberResolver

	return ctor
}
