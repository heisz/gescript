/*
 * Standard native functions from the ECMA specification.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package native

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/heisz/gescript/types"
)

// Determine whether the argument is a finite number
func IsFinite(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(false), nil
	}

	num := types.ToNumber(args[0])
	result := !math.IsNaN(num) && !math.IsInf(num, 0)
	return types.BooleanType(result), nil
}

// Determine whether the argument is numerically NaN
func IsNaN(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.BooleanType(true), nil
	}

	num := types.ToNumber(args[0])
	return types.BooleanType(math.IsNaN(num)), nil
}

// Parse a string argument into a floating point number
func ParseFloat(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NaN, nil
	}

	str := strings.TrimSpace(types.ToString(args[0]))
	if str == "" {
		return types.NaN, nil
	}

	// Find the longest valid numeric prefix (ECMA is lenient)
	endIdx := 0
	hasDecimal := false
	hasExponent := false
	allowExpSign := false
	for idx, ch := range str {
		if idx == 0 && (ch == '+' || ch == '-') {
			endIdx = idx + 1
			continue
		}
		if ch >= '0' && ch <= '9' {
			allowExpSign = false
			endIdx = idx + 1
			continue
		}
		if ch == '.' && !hasDecimal && !hasExponent {
			hasDecimal = true
			endIdx = idx + 1
			continue
		}
		if (ch == 'e' || ch == 'E') && !hasExponent && endIdx > 0 {
			hasExponent = true
			allowExpSign = true
			endIdx = idx + 1
			continue
		}
		if (ch == '+' || ch == '-') && allowExpSign {
			allowExpSign = false
			endIdx = idx + 1
			continue
		}
		break
	}

	if endIdx == 0 {
		return types.NaN, nil
	}

	numStr := str[:endIdx]
	f, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return types.NaN, nil
	}

	return types.NumberType(f), nil
}

// Parse a string argument into an integer of the specified (optional) radix
func ParseInt(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.NaN, nil
	}

	str := strings.TrimSpace(types.ToString(args[0]))
	if str == "" {
		return types.NaN, nil
	}

	// Default radix is 10, but can be specified
	radix := 10
	if len(args) > 1 {
		r := types.ToNumber(args[1])
		if !math.IsNaN(r) {
			radix = int(r)
		}
	}

	// Directly handle sign for non-standard radix
	negative := false
	if len(str) > 0 && str[0] == '-' {
		negative = true
		str = str[1:]
	} else if len(str) > 0 && str[0] == '+' {
		str = str[1:]
	}

	// Auto-detect radix for 0x prefix (radix 0 or 16)
	if (radix == 0 || radix == 16) && len(str) >= 2 &&
		str[0] == '0' && (str[1] == 'x' || str[1] == 'X') {
		radix = 16
		str = str[2:]
	} else if radix == 0 {
		radix = 10
	}

	// Validate radix
	if radix < 2 || radix > 36 {
		return types.NaN, nil
	}

	// Find valid digits for the radix
	endIdx := 0
	for idx, ch := range str {
		var digit int
		if ch >= '0' && ch <= '9' {
			digit = int(ch - '0')
		} else if ch >= 'a' && ch <= 'z' {
			digit = int(ch-'a') + 10
		} else if ch >= 'A' && ch <= 'Z' {
			digit = int(ch-'A') + 10
		} else {
			break
		}
		if digit >= radix {
			break
		}
		endIdx = idx + 1
	}

	if endIdx == 0 {
		return types.NaN, nil
	}

	numStr := str[:endIdx]
	val, err := strconv.ParseInt(numStr, radix, 64)
	if err != nil {
		return types.NaN, nil
	}

	if negative {
		val = -val
	}

	return types.NumberType(float64(val)), nil
}

// Decode a URI previously encoded by encodeURI (pair)
func DecodeURI(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("undefined"), nil
	}

	str := types.ToString(args[0])

	// decodeURI does not decode: ; / ? : @ & = + $ , #
	var result strings.Builder
	for i := 0; i < len(str); i++ {
		if str[i] == '%' && i+2 < len(str) {
			hex := str[i+1 : i+3]
			if isHexPair(hex) {
				b, _ := strconv.ParseUint(hex, 16, 8)
				ch := byte(b)
				if isURIReserved(rune(ch)) || ch == '#' {
					result.WriteString(str[i : i+3])
				} else {
					result.WriteByte(ch)
				}
				i += 2
				continue
			}
		}
		result.WriteByte(str[i])
	}

	return types.StringType(result.String()), nil
}

// Helper: check if two characters form a valid hex pair
func isHexPair(s string) bool {
	if len(s) != 2 {
		return false
	}
	for _, ch := range s {
		if !((ch >= '0' && ch <= '9') ||
			(ch >= 'a' && ch <= 'f') ||
			(ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}

// Decode a URI component previously encoded by encodeURIComponent (pair)
func DecodeURIComponent(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("undefined"), nil
	}

	str := types.ToString(args[0])

	decoded, err := url.QueryUnescape(str)
	if err != nil {
		// Return the original string on error (lenient)
		return types.StringType(str), nil
	}

	return types.StringType(decoded), nil
}

// Encode a URI using %XX notation for URI safety
func EncodeURI(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("undefined"), nil
	}

	str := types.ToString(args[0])

	// There's a big list of characters not encoded in URI
	var result strings.Builder
	for _, r := range str {
		if isURIUnescaped(r) || isURIReserved(r) || r == '#' {
			result.WriteRune(r)
		} else {
			// Encode the rune as UTF-8 percent-encoded
			for _, b := range []byte(string(r)) {
				result.WriteString(fmt.Sprintf("%%%02X", b))
			}
		}
	}

	return types.StringType(result.String()), nil
}

// Encode a URI component using %XX notation for URI safety (subset)
func EncodeURIComponent(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("undefined"), nil
	}

	str := types.ToString(args[0])

	// In this case, it's a subset of the full URI encoded character set
	var result strings.Builder
	for _, r := range str {
		if isURIUnescaped(r) {
			result.WriteRune(r)
		} else {
			// Encode the rune as UTF-8 percent-encoded
			for _, b := range []byte(string(r)) {
				result.WriteString(fmt.Sprintf("%%%02X", b))
			}
		}
	}

	return types.StringType(result.String()), nil
}

// Set of characters that are never encoded in URI/component
func isURIUnescaped(r rune) bool {
	return (r >= 'A' && r <= 'Z') ||
		(r >= 'a' && r <= 'z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_' || r == '.' || r == '!' ||
		r == '~' || r == '*' || r == '\'' || r == '(' || r == ')'
}

// Set of reserved URI characters (not encoded by encodeURI)
func isURIReserved(r rune) bool {
	return r == ';' || r == ',' || r == '/' || r == '?' ||
		r == ':' || r == '@' || r == '&' || r == '=' ||
		r == '+' || r == '$'
}

// Parse a JSON string into gescript datatypes (JSON.parse)
func JSONParse(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.Undefined,
			fmt.Errorf("JSON.parse requires a string argument")
	}

	str := types.ToString(args[0])
	result, err := types.ParseJSON(str)
	if err != nil {
		return types.Undefined, fmt.Errorf("JSON.parse: %v", err)
	}

	return result, nil
}

// Convert a gescript value to a JSON string (JSON.stringify)
func JSONStringify(args []types.DataType) (types.DataType, error) {
	if len(args) == 0 {
		return types.StringType("undefined"), nil
	}

	str, err := types.StringifyJSON(args[0])
	if err != nil {
		return types.Undefined, fmt.Errorf("JSON.stringify: %v", err)
	}

	return types.StringType(str), nil
}

// Exported native definitions, initialized once at package load
var (
	NativeFunctions    map[string]types.DataType
	NativeConstructors []*types.NativeConstructor
)

func init() {
	NativeFunctions = make(map[string]types.DataType)

	register := func(name string, fn types.NativeFn) {
		NativeFunctions[name] = &types.NativeFunction{Name: name, Fn: fn}
	}

	// Note: eval is registered in scriptcontext to avoid circular import
	register("isFinite", IsFinite)
	register("isNaN", IsNaN)
	register("parseFloat", ParseFloat)
	register("parseInt", ParseInt)
	register("decodeURI", DecodeURI)
	register("decodeURIComponent", DecodeURIComponent)
	register("encodeURI", EncodeURI)
	register("encodeURIComponent", EncodeURIComponent)

	// Create/register the JSON object (static methods)
	jsonObj := types.NewObject()
	jsonObj.Properties["parse"] = &types.NativeFunction{
		Name: "parse", Fn: JSONParse}
	jsonObj.Properties["stringify"] = &types.NativeFunction{
		Name: "stringify", Fn: JSONStringify}
	NativeFunctions["JSON"] = jsonObj

	// Type constructors with full member resolution support
	NativeConstructors = []*types.NativeConstructor{
		NewArrayConstructor(),
		NewStringConstructor(),
		NewObjectConstructor(),
		NewNumberConstructor(),
		NewBooleanConstructor(),
		NewFunctionConstructor(),
	}

	// Register constructors in the natives map by name
	for _, nc := range NativeConstructors {
		NativeFunctions[nc.Name] = nc
	}
}
