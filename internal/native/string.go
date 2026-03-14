/*
 * Implementations of standard elements for the string type.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"fmt"
	"strings"

	"github.com/heisz/gescript/types"
)

// Note: in all methods, args[0] is 'this', aka the string instance

func stringCharAt(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	idx := 0
	if len(args) > 1 {
		idx = types.ToInt(args[1])
	}
	if idx < 0 || idx >= len(str) {
		return types.StringType(""), nil
	}
	return types.StringType(str[idx : idx+1]), nil
}

func stringCharCodeAt(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	idx := 0
	if len(args) > 1 {
		idx = types.ToInt(args[1])
	}
	if idx < 0 || idx >= len(str) {
		return types.NaN, nil
	}
	return types.IntegerType(str[idx]), nil
}

func stringConcat(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	var sb strings.Builder
	sb.WriteString(str)
	for _, arg := range args[1:] {
		sb.WriteString(types.ToString(arg))
	}
	return types.StringType(sb.String()), nil
}

func stringEndsWith(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	search := types.ToString(args[1])
	endIdx := len(str)
	if len(args) > 2 {
		endIdx = types.ToInt(args[2])
		if endIdx < 0 {
			endIdx = 0
		}
		if endIdx > len(str) {
			endIdx = len(str)
		}
	}
	if endIdx < len(search) {
		return types.BooleanType(false), nil
	}
	return types.BooleanType(strings.HasSuffix(str[:endIdx], search)), nil
}

func stringIncludes(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	search := types.ToString(args[1])
	startIdx := 0
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > len(str) {
			return types.BooleanType(false), nil
		}
	}
	return types.BooleanType(strings.Contains(str[startIdx:], search)), nil
}

func stringIndexOf(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 2 {
		return types.IntegerType(-1), nil
	}
	search := types.ToString(args[1])
	startIdx := 0
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = 0
		}
	}
	if startIdx >= len(str) {
		if search == "" {
			return types.IntegerType(len(str)), nil
		}
		return types.IntegerType(-1), nil
	}
	idx := strings.Index(str[startIdx:], search)
	if idx == -1 {
		return types.IntegerType(-1), nil
	}
	return types.IntegerType(startIdx + idx), nil
}

func stringLastIndexOf(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 2 {
		return types.IntegerType(-1), nil
	}
	search := types.ToString(args[1])
	endIdx := len(str)
	if len(args) > 2 {
		endIdx = types.ToInt(args[2]) + len(search)
		if endIdx > len(str) {
			endIdx = len(str)
		}
	}
	if endIdx <= 0 {
		if search == "" {
			return types.IntegerType(0), nil
		}
		return types.IntegerType(-1), nil
	}
	idx := strings.LastIndex(str[:endIdx], search)
	return types.IntegerType(idx), nil
}

func stringPadEnd(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	targetLen := len(str)
	padStr := " "
	if len(args) > 1 {
		targetLen = types.ToInt(args[1])
	}
	if len(args) > 2 {
		padStr = types.ToString(args[2])
		if padStr == "" {
			return args[0], nil
		}
	}
	if len(str) >= targetLen {
		return args[0], nil
	}
	var sb strings.Builder
	sb.WriteString(str)
	for sb.Len() < targetLen {
		sb.WriteString(padStr)
	}
	return types.StringType(sb.String()[:targetLen]), nil
}

func stringPadStart(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	targetLen := len(str)
	padStr := " "
	if len(args) > 1 {
		targetLen = types.ToInt(args[1])
	}
	if len(args) > 2 {
		padStr = types.ToString(args[2])
		if padStr == "" {
			return args[0], nil
		}
	}
	if len(str) >= targetLen {
		return args[0], nil
	}
	padNeeded := targetLen - len(str)
	var sb strings.Builder
	for sb.Len() < padNeeded {
		sb.WriteString(padStr)
	}
	return types.StringType(sb.String()[:padNeeded] + str), nil
}

func stringRepeat(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	count := 0
	if len(args) > 1 {
		count = types.ToInt(args[1])
		if count < 0 {
			return types.StringType(""),
				fmt.Errorf("Repeat count must be >= 0")
		}
	}
	return types.StringType(strings.Repeat(str, count)), nil
}

func stringReplace(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 3 {
		return args[0], nil
	}
	search := types.ToString(args[1])
	replace := types.ToString(args[2])
	return types.StringType(strings.Replace(str, search, replace, 1)), nil
}

func stringReplaceAll(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 3 {
		return args[0], nil
	}
	search := types.ToString(args[1])
	replace := types.ToString(args[2])
	return types.StringType(strings.ReplaceAll(str, search, replace)), nil
}

func stringSlice(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	length := len(str)
	start := 0
	end := length

	if len(args) > 1 {
		start = types.ToInt(args[1])
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
		if start > length {
			start = length
		}
	}
	if len(args) > 2 {
		end = types.ToInt(args[2])
		if end < 0 {
			end = length + end
		}
		if end > length {
			end = length
		}
	}
	if start > end {
		return types.StringType(""), nil
	}
	return types.StringType(str[start:end]), nil
}

func stringSplit(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	separator := ""
	limit := -1

	if len(args) > 1 {
		separator = types.ToString(args[1])
	}
	if len(args) > 2 {
		limit = types.ToInt(args[2])
	}

	var parts []string
	if separator == "" {
		// No separator, becomes an array of the characters
		parts = make([]string, len(str))
		for idx, ch := range str {
			parts[idx] = string(ch)
		}
	} else {
		parts = strings.Split(str, separator)
	}

	if limit >= 0 && len(parts) > limit {
		parts = parts[:limit]
	}

	res := types.NewArray(len(parts))
	for i, part := range parts {
		res.Elements[i] = types.StringType(part)
	}
	return res, nil
}

func stringStartsWith(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	if len(args) < 2 {
		return types.BooleanType(false), nil
	}
	search := types.ToString(args[1])
	startIdx := 0
	if len(args) > 2 {
		startIdx = types.ToInt(args[2])
		if startIdx < 0 {
			startIdx = 0
		}
	}
	if startIdx > len(str) {
		return types.BooleanType(false), nil
	}
	return types.BooleanType(strings.HasPrefix(str[startIdx:], search)), nil
}

func stringSubstr(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	length := len(str)
	start := 0
	sublen := length

	if len(args) > 1 {
		start = types.ToInt(args[1])
		if start < 0 {
			start = length + start
			if start < 0 {
				start = 0
			}
		}
	}
	if len(args) > 2 {
		sublen = types.ToInt(args[2])
		if sublen < 0 {
			sublen = 0
		}
	}
	if start >= length {
		return types.StringType(""), nil
	}
	end := start + sublen
	if end > length {
		end = length
	}
	return types.StringType(str[start:end]), nil
}

func stringSubstring(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	length := len(str)
	start := 0
	end := length

	if len(args) > 1 {
		start = types.ToInt(args[1])
		if start < 0 {
			start = 0
		}
		if start > length {
			start = length
		}
	}
	if len(args) > 2 {
		end = types.ToInt(args[2])
		if end < 0 {
			end = 0
		}
		if end > length {
			end = length
		}
	}
	if start > end {
		start, end = end, start
	}
	return types.StringType(str[start:end]), nil
}

func stringToLowerCase(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	return types.StringType(strings.ToLower(str)), nil
}

func stringToString(args []types.DataType) (types.DataType, error) {
	return args[0], nil
}

func stringToUpperCase(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	return types.StringType(strings.ToUpper(str)), nil
}

func stringTrim(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	return types.StringType(strings.TrimSpace(str)), nil
}

func stringTrimEnd(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	return types.StringType(strings.TrimRight(str, " \t\n\r")), nil
}

func stringTrimStart(args []types.DataType) (types.DataType, error) {
	str := string(args[0].(types.StringType))
	return types.StringType(strings.TrimLeft(str, " \t\n\r")), nil
}

// Resolve properties and methods for the String type
func stringMemberResolver(target types.DataType, name string) types.DataType {
	str, ok := target.(types.StringType)
	if !ok {
		return nil
	}

	// Length is the only dynamic property for arrays
	if name == "length" {
		return types.IntegerType(len(string(str)))
	}

	// Otherwise look up the string instance methods
	var method *types.NativeFunction
	switch name {
	case "charAt":
		method = &types.NativeFunction{Name: "charAt",
			Fn: stringCharAt}
	case "charCodeAt":
		method = &types.NativeFunction{Name: "charCodeAt",
			Fn: stringCharCodeAt}
	case "concat":
		method = &types.NativeFunction{Name: "concat",
			Fn: stringConcat}
	case "endsWith":
		method = &types.NativeFunction{Name: "endsWith",
			Fn: stringEndsWith}
	case "includes":
		method = &types.NativeFunction{Name: "includes",
			Fn: stringIncludes}
	case "indexOf":
		method = &types.NativeFunction{Name: "indexOf",
			Fn: stringIndexOf}
	case "lastIndexOf":
		method = &types.NativeFunction{Name: "lastIndexOf",
			Fn: stringLastIndexOf}
	case "padEnd":
		method = &types.NativeFunction{Name: "padEnd",
			Fn: stringPadEnd}
	case "padStart":
		method = &types.NativeFunction{Name: "padStart",
			Fn: stringPadStart}
	case "repeat":
		method = &types.NativeFunction{Name: "repeat",
			Fn: stringRepeat}
	case "replace":
		method = &types.NativeFunction{Name: "replace",
			Fn: stringReplace}
	case "replaceAll":
		method = &types.NativeFunction{Name: "replaceAll",
			Fn: stringReplaceAll}
	case "slice":
		method = &types.NativeFunction{Name: "slice",
			Fn: stringSlice}
	case "split":
		method = &types.NativeFunction{Name: "split",
			Fn: stringSplit}
	case "startsWith":
		method = &types.NativeFunction{Name: "startsWith",
			Fn: stringStartsWith}
	case "substr":
		method = &types.NativeFunction{Name: "substr",
			Fn: stringSubstr}
	case "substring":
		method = &types.NativeFunction{Name: "substring",
			Fn: stringSubstring}
	case "toLowerCase":
		method = &types.NativeFunction{Name: "toLowerCase",
			Fn: stringToLowerCase}
	case "toString":
		method = &types.NativeFunction{Name: "toString",
			Fn: stringToString}
	case "toUpperCase":
		method = &types.NativeFunction{Name: "toUpperCase",
			Fn: stringToUpperCase}
	case "trim":
		method = &types.NativeFunction{Name: "trim",
			Fn: stringTrim}
	case "trimEnd":
		method = &types.NativeFunction{Name: "trimEnd",
			Fn: stringTrimEnd}
	case "trimStart":
		method = &types.NativeFunction{Name: "trimStart",
			Fn: stringTrimStart}
	case "valueOf":
		method = &types.NativeFunction{Name: "valueOf",
			Fn: stringToString}
	default:
		return nil
	}
	return &types.NativeMethod{Target: str, Method: method}
}

func stringFromCharCode(args []types.DataType) (types.DataType, error) {
	var sb strings.Builder
	for _, arg := range args {
		code := types.ToInt(arg)
		sb.WriteByte(byte(code))
	}
	return types.StringType(sb.String()), nil
}

// Create the String global constructor with fromCharCode and member elements
func NewStringConstructor() *types.NativeConstructor {
	ctor := types.NewNativeConstructor("String",
		func(args []types.DataType) (types.DataType, error) {
			if len(args) == 0 {
				return types.StringType(""), nil
			}
			return types.StringType(types.ToString(args[0])), nil
		})

	ctor.AddStaticMethod("fromCharCode", stringFromCharCode)
	ctor.InstanceMembers = stringMemberResolver

	return ctor
}
