/*
 * Implementations of standard elements for the array type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"fmt"
	"sort"
	"strings"

	"github.com/heisz/gescript/types"
)

// Note: in all methods, args[0] is 'this', aka the array instance

func arrayConcat(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	res := types.NewArray(len(arr.Elements))
	copy(res.Elements, arr.Elements)

	for _, arg := range args[1:] {
		if other, ok := arg.(*types.ArrayType); ok {
			res.Elements = append(res.Elements, other.Elements...)
		} else {
			res.Elements = append(res.Elements, arg)
		}
	}
	return res, nil
}

func arrayEvery(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.BooleanType(true), nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.BooleanType(false),
			fmt.Errorf("Every requires a callback function")
	}

	for idx, entry := range arr.Elements {
		res, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.BooleanType(false), err
		}
		if !types.IsTruthy(res) {
			return types.BooleanType(false), nil
		}
	}
	return types.BooleanType(true), nil
}

func arrayFill(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return arr, nil
	}
	val := args[1]
	length := len(arr.Elements)
	start := 0
	end := length

	if len(args) > 2 {
		start = types.ToInt(args[2])
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
	}
	if len(args) > 3 {
		end = types.ToInt(args[3])
		if end < 0 {
			end = length + end
		}
	}
	if end > length {
		end = length
	}

	for idx := start; idx < end; idx++ {
		arr.Elements[idx] = val
	}
	return arr, nil
}

func arrayFilter(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.NewArray(0), nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.NewArray(0),
			fmt.Errorf("Filter requires a callback function")
	}

	res := types.NewArray(0)
	for idx, entry := range arr.Elements {
		fres, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.NewArray(0), err
		}
		if types.IsTruthy(fres) {
			res.Elements = append(res.Elements, entry)
		}
	}
	return res, nil
}

func arrayFind(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.Undefined, nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.Undefined, fmt.Errorf("Find requires a callback function")
	}

	for idx, entry := range arr.Elements {
		res, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.Undefined, err
		}
		if types.IsTruthy(res) {
			return entry, nil
		}
	}
	return types.Undefined, nil
}

func arrayFindIndex(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.IntegerType(-1), nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.IntegerType(-1),
			fmt.Errorf("FindIndex requires a callback function")
	}

	for idx, entry := range arr.Elements {
		res, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.IntegerType(-1), err
		}
		if types.IsTruthy(res) {
			return types.IntegerType(idx), nil
		}
	}
	return types.IntegerType(-1), nil
}

func flatten(res *types.ArrayType, elements []types.DataType, depth int) {
	for _, elem := range elements {
		if inner, ok := elem.(*types.ArrayType); ok && depth > 0 {
			flatten(res, inner.Elements, depth-1)
		} else {
			res.Elements = append(res.Elements, elem)
		}
	}
}

func arrayFlat(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	depth := 1
	if len(args) > 1 {
		depth = types.ToInt(args[1])
	}

	res := types.NewArray(0)
	flatten(res, arr.Elements, depth)
	return res, nil
}

func arrayForEach(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.Undefined, nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.Undefined,
			fmt.Errorf("ForEach requires a callback function")
	}

	for idx, entry := range arr.Elements {
		_, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.Undefined, err
		}
	}
	return types.Undefined, nil
}

func arrayIncludes(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	searchVal := args[1]
	startIdx := 0
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = len(arr.Elements) + startIdx
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for idx := startIdx; idx < len(arr.Elements); idx++ {
		if types.StrictEquals(arr.Elements[idx], searchVal) {
			return types.BooleanType(true), nil
		}
	}
	return types.BooleanType(false), nil
}

func arrayIndexOf(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.IntegerType(-1), nil
	}
	searchVal := args[1]
	startIdx := 0
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = len(arr.Elements) + startIdx
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for idx := startIdx; idx < len(arr.Elements); idx++ {
		if types.StrictEquals(arr.Elements[idx], searchVal) {
			return types.IntegerType(idx), nil
		}
	}
	return types.IntegerType(-1), nil
}

func arrayJoin(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	sep := ","
	if len(args) > 1 {
		sep = types.ToString(args[1])
	}

	parts := make([]string, len(arr.Elements))
	for idx, entry := range arr.Elements {
		parts[idx] = types.ToString(entry)
	}
	return types.StringType(strings.Join(parts, sep)), nil
}

func arrayLastIndexOf(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.IntegerType(-1), nil
	}
	searchVal := args[1]
	startIdx := len(arr.Elements) - 1
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = len(arr.Elements) + startIdx
		}
	}
	if startIdx >= len(arr.Elements) {
		startIdx = len(arr.Elements) - 1
	}

	for idx := startIdx; idx >= 0; idx-- {
		if types.StrictEquals(arr.Elements[idx], searchVal) {
			return types.IntegerType(idx), nil
		}
	}
	return types.IntegerType(-1), nil
}

func arrayMap(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.NewArray(0), nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.NewArray(0), fmt.Errorf("Map requires a callback function")
	}

	res := types.NewArray(len(arr.Elements))
	for idx, entry := range arr.Elements {
		mres, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.NewArray(0), err
		}
		res.Elements[idx] = mres
	}
	return res, nil
}

func arrayPop(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(arr.Elements) == 0 {
		return types.Undefined, nil
	}
	alen := len(arr.Elements)
	last := arr.Elements[alen-1]
	arr.Elements = arr.Elements[:alen-1]
	return last, nil
}

func arrayPush(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	for _, val := range args[1:] {
		arr.Elements = append(arr.Elements, val)
	}
	return types.IntegerType(len(arr.Elements)), nil
}

func arrayReduce(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.Undefined,
			fmt.Errorf("Reduce requires a callback function")
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.Undefined,
			fmt.Errorf("Reduce requires a callback function")
	}

	startIdx := 0
	var res types.DataType
	if len(args) > 2 {
		res = args[2]
	} else if len(arr.Elements) > 0 {
		res = arr.Elements[0]
		startIdx = 1
	} else {
		return types.Undefined,
			fmt.Errorf("Reduce of empty array with no initial value")
	}

	for idx := startIdx; idx < len(arr.Elements); idx++ {
		rres, err := callback.Call([]types.DataType{
			res, arr.Elements[idx], types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.Undefined, err
		}
		res = rres
	}
	return res, nil
}

func arrayReduceRight(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.Undefined,
			fmt.Errorf("ReduceRight requires a callback function")
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.Undefined,
			fmt.Errorf("ReduceRight requires a callback function")
	}

	endIdx := len(arr.Elements) - 1
	var res types.DataType
	if len(args) > 2 {
		res = args[2]
	} else if len(arr.Elements) > 0 {
		res = arr.Elements[endIdx]
		endIdx--
	} else {
		return types.Undefined,
			fmt.Errorf("ReduceRight of empty array with no initial value")
	}

	for idx := endIdx; idx >= 0; idx-- {
		rres, err := callback.Call([]types.DataType{
			res, arr.Elements[idx], types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.Undefined, err
		}
		res = rres
	}
	return res, nil
}

func arrayReverse(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	// Swap in opposite directions
	for idx, idy := 0, len(arr.Elements)-1; idx < idy; idx, idy = idx+1, idy-1 {
		arr.Elements[idx], arr.Elements[idy] =
			arr.Elements[idy], arr.Elements[idx]
	}
	return arr, nil
}

func arrayShift(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(arr.Elements) == 0 {
		return types.Undefined, nil
	}
	first := arr.Elements[0]
	arr.Elements = arr.Elements[1:]
	return first, nil
}

func arraySlice(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	alen := len(arr.Elements)

	start := 0
	end := alen

	if len(args) > 1 {
		start = types.ToInt(args[1])
		if start < 0 {
			start = alen + start
			if start < 0 {
				start = 0
			}
		}
		if start > alen {
			start = alen
		}
	}
	if len(args) > 2 {
		end = types.ToInt(args[2])
		if end < 0 {
			end = alen + end
		}
		if end > alen {
			end = alen
		}
	}
	if start > end {
		start = end
	}

	res := types.NewArray(end - start)
	copy(res.Elements, arr.Elements[start:end])
	return res, nil
}

func arraySome(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	callback, ok := args[1].(types.FunctionType)
	if !ok {
		return types.BooleanType(false),
			fmt.Errorf("some requires a callback function")
	}

	for idx, entry := range arr.Elements {
		res, err := callback.Call([]types.DataType{
			entry, types.IntegerType(idx), arr,
		})
		if err != nil {
			return types.BooleanType(false), err
		}
		if types.IsTruthy(res) {
			return types.BooleanType(true), nil
		}
	}
	return types.BooleanType(false), nil
}

func arraySort(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	// Default is to sort by string versions of entries
	sort.SliceStable(arr.Elements, func(idx int, idy int) bool {
		return types.ToString(arr.Elements[idx]) <
			types.ToString(arr.Elements[idy])
	})
	return arr, nil
}

func arraySplice(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	alen := len(arr.Elements)

	if len(args) < 2 {
		return types.NewArray(0), nil
	}

	start := types.ToInt(args[1])
	if start < 0 {
		start = alen + start
		if start < 0 {
			start = 0
		}
	}
	if start > alen {
		start = alen
	}

	delCount := alen - start
	if len(args) > 2 {
		delCount = types.ToInt(args[2])
		if delCount < 0 {
			delCount = 0
		}
		if start+delCount > alen {
			delCount = alen - start
		}
	}

	// Remaining arguments are items to splice in
	additions := args[3:]

	// Extract the removed items for return
	removed := types.NewArray(delCount)
	copy(removed.Elements, arr.Elements[start:start+delCount])

	// Splice it all together
	newLen := alen - delCount + len(additions)
	newElems := make([]types.DataType, newLen)
	copy(newElems, arr.Elements[:start])
	copy(newElems[start:], additions)
	copy(newElems[start+len(additions):], arr.Elements[start+delCount:])
	arr.Elements = newElems

	return removed, nil
}

func arrayToString(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	parts := make([]string, len(arr.Elements))
	for idx, entry := range arr.Elements {
		parts[idx] = types.ToString(entry)
	}
	return types.StringType(strings.Join(parts, ",")), nil
}

func arrayUnshift(args []types.DataType) (types.DataType, error) {
	arr := args[0].(*types.ArrayType)
	alen := len(args)
	newElems := make([]types.DataType, alen-1+len(arr.Elements))
	copy(newElems, args[1:])
	copy(newElems[alen-1:], arr.Elements)
	arr.Elements = newElems
	return types.IntegerType(len(arr.Elements)), nil
}

// Resolve properties and methods for the Array type
func arrayMemberResolver(target types.DataType, name string) types.DataType {
	arr, ok := target.(*types.ArrayType)
	if !ok {
		return nil
	}

	// Length is the only dynamic property for arrays
	if name == "length" {
		return types.IntegerType(len(arr.Elements))
	}

	// Otherwise look up the array instance methods
	var method *types.NativeFunction
	switch name {
	case "concat":
		method = &types.NativeFunction{Name: "concat",
			Fn: arrayConcat}
	case "every":
		method = &types.NativeFunction{Name: "every",
			Fn: arrayEvery}
	case "fill":
		method = &types.NativeFunction{Name: "fill",
			Fn: arrayFill}
	case "filter":
		method = &types.NativeFunction{Name: "filter",
			Fn: arrayFilter}
	case "find":
		method = &types.NativeFunction{Name: "find",
			Fn: arrayFind}
	case "findIndex":
		method = &types.NativeFunction{Name: "findIndex",
			Fn: arrayFindIndex}
	case "flat":
		method = &types.NativeFunction{Name: "flat",
			Fn: arrayFlat}
	case "forEach":
		method = &types.NativeFunction{Name: "forEach",
			Fn: arrayForEach}
	case "includes":
		method = &types.NativeFunction{Name: "includes",
			Fn: arrayIncludes}
	case "indexOf":
		method = &types.NativeFunction{Name: "indexOf",
			Fn: arrayIndexOf}
	case "join":
		method = &types.NativeFunction{Name: "join",
			Fn: arrayJoin}
	case "lastIndexOf":
		method = &types.NativeFunction{Name: "lastIndexOf",
			Fn: arrayLastIndexOf}
	case "map":
		method = &types.NativeFunction{Name: "map",
			Fn: arrayMap}
	case "pop":
		method = &types.NativeFunction{Name: "pop",
			Fn: arrayPop}
	case "push":
		method = &types.NativeFunction{Name: "push",
			Fn: arrayPush}
	case "reduce":
		method = &types.NativeFunction{Name: "reduce",
			Fn: arrayReduce}
	case "reduceRight":
		method = &types.NativeFunction{Name: "reduceRight",
			Fn: arrayReduceRight}
	case "reverse":
		method = &types.NativeFunction{Name: "reverse",
			Fn: arrayReverse}
	case "shift":
		method = &types.NativeFunction{Name: "shift",
			Fn: arrayShift}
	case "slice":
		method = &types.NativeFunction{Name: "slice",
			Fn: arraySlice}
	case "some":
		method = &types.NativeFunction{Name: "some",
			Fn: arraySome}
	case "sort":
		method = &types.NativeFunction{Name: "sort",
			Fn: arraySort}
	case "splice":
		method = &types.NativeFunction{Name: "splice",
			Fn: arraySplice}
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: arrayToString}
	case "unshift":
		method = &types.NativeFunction{Name: "unshift",
			Fn: arrayUnshift}
	default:
		return nil
	}
	return &types.NativeMethod{Target: arr, Method: method}
}

func arrayIsArray(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}
	_, isArray := args[0].(*types.ArrayType)
	return types.BooleanType(isArray), nil
}

// Create the Array global constructor with static isArray and member elements
func NewArrayConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("Array",
		func(args []types.DataType) (types.DataType, error) {
			if len(args) == 0 {
				return types.NewArray(0), nil
			}
			if len(args) == 1 {
				// Single arg that is numeric indicates array of length
				if n, ok := args[0].(types.IntegerType); ok {
					return types.NewArray(int(n)), nil
				}
				if n, ok := args[0].(types.NumberType); ok {
					return types.NewArray(int(n)), nil
				}
			}
			// Otherwise, array initializes from arguments
			arr := types.NewArray(len(args))
			copy(arr.Elements, args)
			return arr, nil
		})

	ctor.AddStaticMethod("isArray", arrayIsArray)
	ctor.InstanceMembers = arrayMemberResolver

	return ctor
}
