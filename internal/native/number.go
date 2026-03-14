/*
 * Implementations of standard elements for the number type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"fmt"
	"math"
	"strconv"

	"github.com/heisz/gescript/types"
)

// Note: in all instance methods, args[0] is 'this', aka the number instance

// Local helper function to retrieve a number from args[0] (this)
func getNumber(args []types.DataType) float64 {
	if len(args) == 0 {
		return 0
	}
	switch v := args[0].(type) {
	case types.IntegerType:
		return float64(v)
	case types.NumberType:
		return float64(v)
	default:
		return math.NaN()
	}
}

func numberToExponential(args []types.DataType) (types.DataType, error) {
	num := getNumber(args)
	digits := -1
	if len(args) > 1 {
		digits = types.ToInt(args[1])
	}

	var result string
	if digits < 0 {
		result = fmt.Sprintf("%e", num)
	} else {
		result = fmt.Sprintf("%.*e", digits, num)
	}
	return types.StringType(result), nil
}

func numberToFixed(args []types.DataType) (types.DataType, error) {
	num := getNumber(args)
	digits := 0
	if len(args) > 1 {
		digits = types.ToInt(args[1])
		if digits < 0 {
			digits = 0
		}
		if digits > 100 {
			digits = 100
		}
	}

	result := fmt.Sprintf("%.*f", digits, num)
	return types.StringType(result), nil
}

func numberToPrecision(args []types.DataType) (types.DataType, error) {
	num := getNumber(args)
	if len(args) < 2 {
		return types.StringType(fmt.Sprintf("%g", num)), nil
	}

	precision := types.ToInt(args[1])
	if precision < 1 {
		precision = 1
	}
	if precision > 100 {
		precision = 100
	}

	result := strconv.FormatFloat(num, 'g', precision, 64)
	return types.StringType(result), nil
}

func numberToString(args []types.DataType) (types.DataType, error) {
	num := getNumber(args)
	radix := 10
	if len(args) > 1 {
		radix = types.ToInt(args[1])
		if radix < 2 || radix > 36 {
			radix = 10
		}
	}

	// For integers, use strconv for radix support
	if num == math.Trunc(num) && !math.IsInf(num, 0) && !math.IsNaN(num) {
		return types.StringType(strconv.FormatInt(int64(num), radix)), nil
	}

	// For floats, only base 10 is supported
	return types.StringType(fmt.Sprintf("%g", num)), nil
}

func numberValueOf(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.IntegerType(0), nil
	}
	return args[0], nil
}

// Resolve the properties and methods for the Number type
func numberMemberResolver(target types.DataType, name string) types.DataType {
	// In this case we have two types of numbers
	var isNumber bool
	switch target.(type) {
	case types.IntegerType, types.NumberType:
		isNumber = true
	}
	if !isNumber {
		return nil
	}

	// Only instance methods in this case
	var method *types.NativeFunction
	switch name {
	case "toExponential":
		method = &types.NativeFunction{Name: "toExponential",
			Fn: numberToExponential}
	case "toFixed":
		method = &types.NativeFunction{Name: "toFixed",
			Fn: numberToFixed}
	case "toPrecision":
		method = &types.NativeFunction{Name: "toPrecision",
			Fn: numberToPrecision}
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: numberToString}
	case "valueOf":
		method = &types.NativeFunction{Name: "valueOf",
			Fn: numberValueOf}
	default:
		return nil
	}
	return &types.NativeMethod{Target: target, Method: method}
}

// But definitely a 'number' of static methods (and properties)

func numberIsFinite(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}

	// Must be an explicit number type (not coerced)
	var num float64
	switch v := args[0].(type) {
	case types.IntegerType:
		num = float64(v)
	case types.NumberType:
		num = float64(v)
	default:
		return types.BooleanType(false), nil
	}

	result := !math.IsNaN(num) && !math.IsInf(num, 0)
	return types.BooleanType(result), nil
}

func numberIsInteger(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}

	switch v := args[0].(type) {
	case types.IntegerType:
		return types.BooleanType(true), nil
	case types.NumberType:
		num := float64(v)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return types.BooleanType(false), nil
		}
		return types.BooleanType(num == math.Trunc(num)), nil
	default:
		return types.BooleanType(false), nil
	}
}

func numberIsNaN(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}

	// Must be an explicit number type (not coerced)
	switch v := args[0].(type) {
	case types.NumberType:
		return types.BooleanType(math.IsNaN(float64(v))), nil
	default:
		return types.BooleanType(false), nil
	}
}

func numberIsSafeInteger(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}

	// Per spec, limits are +/-(2^53 - 1)
	const maxSafeInt = 9007199254740991
	const minSafeInt = -9007199254740991

	switch v := args[0].(type) {
	case types.IntegerType:
		n := int64(v)
		return types.BooleanType(n >= minSafeInt && n <= maxSafeInt), nil
	case types.NumberType:
		num := float64(v)
		if math.IsNaN(num) || math.IsInf(num, 0) {
			return types.BooleanType(false), nil
		}
		if num != math.Trunc(num) {
			return types.BooleanType(false), nil
		}
		return types.BooleanType(num >= minSafeInt && num <= maxSafeInt), nil
	default:
		return types.BooleanType(false), nil
	}
}

func numberParseFloat(args []types.DataType) (types.DataType, error) {
	return ParseFloat(args)
}

func numberParseInt(args []types.DataType) (types.DataType, error) {
	return ParseInt(args)
}

// Create the Number global constructor, lots of static/member elements
func NewNumberConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("Number",
		func(args []types.DataType) (types.DataType, error) {
			if len(args) == 0 {
				return types.NumberType(0), nil
			}
			return types.NumberType(types.ToNumber(args[0])), nil
		})

	ctor.AddStaticMethod("isFinite", numberIsFinite)
	ctor.AddStaticMethod("isInteger", numberIsInteger)
	ctor.AddStaticMethod("isNaN", numberIsNaN)
	ctor.AddStaticMethod("isSafeInteger", numberIsSafeInteger)
	ctor.AddStaticMethod("parseFloat", numberParseFloat)
	ctor.AddStaticMethod("parseInt", numberParseInt)

	ctor.AddStaticProperty("MAX_VALUE",
		types.NumberType(math.MaxFloat64))
	ctor.AddStaticProperty("MIN_VALUE",
		types.NumberType(math.SmallestNonzeroFloat64))
	ctor.AddStaticProperty("NaN",
		types.NumberType(math.NaN()))
	ctor.AddStaticProperty("POSITIVE_INFINITY",
		types.NumberType(math.Inf(1)))
	ctor.AddStaticProperty("NEGATIVE_INFINITY",
		types.NumberType(math.Inf(-1)))
	ctor.AddStaticProperty("MAX_SAFE_INTEGER",
		types.NumberType(9007199254740991))
	ctor.AddStaticProperty("MIN_SAFE_INTEGER",
		types.NumberType(-9007199254740991))
	ctor.AddStaticProperty("EPSILON",
		types.NumberType(math.Nextafter(1, 2)-1))

	ctor.InstanceMembers = numberMemberResolver

	return ctor
}
