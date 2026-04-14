/*
 * Implementations of standard elements for the object type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"github.com/heisz/gescript/types"
)

// Note: in all instance methods, args[0] is 'this', aka the object instance

func objectHasOwnProperty(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	obj := args[0].(*types.ObjectType)
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	propName := types.ToString(args[1])
	return types.BooleanType(obj.Has(propName)), nil
}

func objectToString(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	return types.StringType("[object Object]"), nil
}

func objectValueOf(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	return args[0], nil
}

// Resolve properties and methods for the Object type
func objectMemberResolver(target types.DataType, name string) types.DataType {
	obj, ok := target.(*types.ObjectType)
	if !ok {
		return nil
	}

	// Only instance methods in this case
	var method *types.NativeFunction
	switch name {
	case "hasOwnProperty":
		method = &types.NativeFunction{Name: "hasOwnProperty",
			Fn: objectHasOwnProperty}
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: objectToString}
	case "valueOf":
		method = &types.NativeFunction{Name: "valueOf",
			Fn: objectValueOf}
	default:
		return nil
	}
	return &types.NativeMethod{Target: obj, Method: method}
}

// But plenty of static methods

func objectKeys(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NewArray(0), nil
	}

	obj, ok := args[0].(*types.ObjectType)
	if !ok {
		return types.NewArray(0), nil
	}

	keys := make([]types.DataType, 0, len(obj.Properties))
	for key := range obj.Properties {
		keys = append(keys, types.StringType(key))
	}

	arr := types.NewArray(len(keys))
	copy(arr.Elements, keys)
	return arr, nil
}

func objectValues(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NewArray(0), nil
	}

	obj, ok := args[0].(*types.ObjectType)
	if !ok {
		return types.NewArray(0), nil
	}

	values := make([]types.DataType, 0, len(obj.Properties))
	for _, val := range obj.Properties {
		values = append(values, val)
	}

	arr := types.NewArray(len(values))
	copy(arr.Elements, values)
	return arr, nil
}

func objectEntries(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NewArray(0), nil
	}

	obj, ok := args[0].(*types.ObjectType)
	if !ok {
		return types.NewArray(0), nil
	}

	entries := make([]types.DataType, 0, len(obj.Properties))
	for key, val := range obj.Properties {
		pair := types.NewArray(2)
		pair.Elements[0] = types.StringType(key)
		pair.Elements[1] = val
		entries = append(entries, pair)
	}

	arr := types.NewArray(len(entries))
	copy(arr.Elements, entries)
	return arr, nil
}

func objectFromEntries(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NewObject(), nil
	}

	entries, ok := args[0].(*types.ArrayType)
	if !ok {
		return types.NewObject(), nil
	}

	obj := types.NewObject()
	for _, entry := range entries.Elements {
		pair, ok := entry.(*types.ArrayType)
		if !ok || len(pair.Elements) < 2 {
			continue
		}
		key := types.ToString(pair.Elements[0])
		obj.Properties[key] = pair.Elements[1]
	}

	return obj, nil
}

func objectAssign(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NewObject(), nil
	}

	target, ok := args[0].(*types.ObjectType)
	if !ok {
		return args[0], nil
	}

	// Copy properties from each source
	for _, source := range args[1:] {
		srcObj, ok := source.(*types.ObjectType)
		if !ok {
			continue
		}
		for key, val := range srcObj.Properties {
			target.Properties[key] = val
		}
	}

	return target, nil
}

func objectHasOwn(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}

	obj, ok := args[0].(*types.ObjectType)
	if !ok {
		return types.BooleanType(false), nil
	}

	propName := types.ToString(args[1])
	return types.BooleanType(obj.Has(propName)), nil
}

func objectFreeze(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined, nil
	}
	// We really don't support it, just return the original
	return args[0], nil
}

func objectIsFrozen(prc types.Process,
	args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(true), nil
	}
	_, isObj := args[0].(*types.ObjectType)
	if !isObj {
		return types.BooleanType(true), nil
	}
	// Not implemented, always return false
	return types.BooleanType(false), nil
}

// Create the Object global constructor with static/member elements
func NewObjectConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("Object",
		func(prc types.Process, args []types.DataType) (types.DataType, error) {
			if len(args) == 0 {
				return types.NewObject(), nil
			}

			// For argument that is an object, it's a passthrough
			if obj, ok := args[0].(*types.ObjectType); ok {
				return obj, nil
			}
			if arr, ok := args[0].(*types.ArrayType); ok {
				return arr, nil
			}

			// Otherwise create a new empty object
			return types.NewObject(), nil
		})

	// Add static methods
	ctor.AddStaticMethod("keys", objectKeys)
	ctor.AddStaticMethod("values", objectValues)
	ctor.AddStaticMethod("entries", objectEntries)
	ctor.AddStaticMethod("fromEntries", objectFromEntries)
	ctor.AddStaticMethod("assign", objectAssign)
	ctor.AddStaticMethod("hasOwn", objectHasOwn)
	ctor.AddStaticMethod("freeze", objectFreeze)
	ctor.AddStaticMethod("isFrozen", objectIsFrozen)

	ctor.InstanceMembers = objectMemberResolver

	return ctor
}
