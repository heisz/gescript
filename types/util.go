/*
 * Collection of utility methods to manipulate type data
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package types

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Determine the 'truthiness' of the data value according to specification
func IsTruthy(val DataType) bool {
	switch val.(type) {
	case UndefinedType, NullType:
		return false
	case BooleanType:
		return val.Native().(bool)
	case IntegerType:
		return val.Native().(int64) != 0
	case NumberType:
		n := val.Native().(float64)
		return n != 0 && n == n // NaN check
	case StringType:
		return len(val.Native().(string)) > 0
	}

	// Objects are truthy
	return true
}

// Convert a data value to its string representation (per ECMA spec)
func ToString(val DataType) string {
	if val == nil {
		return "undefined"
	}
	switch v := val.(type) {
	case UndefinedType:
		return "undefined"
	case NullType:
		return "null"
	case BooleanType:
		if bool(v) {
			return "true"
		}
		return "false"
	case IntegerType:
		return strconv.FormatInt(int64(v), 10)
	case NumberType:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case StringType:
		return string(v)
	case *ArrayType:
		// Arrays stringify as comma-separated values
		parts := make([]string, len(v.Elements))
		for idx, elem := range v.Elements {
			parts[idx] = ToString(elem)
		}
		return strings.Join(parts, ",")
	case *ObjectType:
		return "[object Object]"
	default:
		return val.ToPrimitive(nil).Native().(string)
	}
}

// Convert a data value to an integer (zero if invalid)
func ToInt(val DataType) int {
	num := ToNumber(val)
	if math.IsNaN(num) || math.IsInf(num, 0) {
		return 0
	}
	return int(num)
}

// Convert a data value to a float64 instance (NaN for invalid)
func ToNumber(val DataType) float64 {
	if val == nil {
		return math.NaN()
	}
	switch v := val.(type) {
	case UndefinedType:
		return math.NaN()
	case NullType:
		return 0
	case BooleanType:
		if v {
			return 1
		}
		return 0
	case IntegerType:
		return float64(v)
	case NumberType:
		return float64(v)
	case StringType:
		s := strings.TrimSpace(string(v))
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	default:
		return math.NaN()
	}
}

// Shared function for strict equality (===) comparison
func StrictEquals(val DataType, cmp DataType) bool {
	switch vval := val.(type) {
	case UndefinedType:
		_, ok := cmp.(UndefinedType)
		return ok
	case NullType:
		_, ok := cmp.(NullType)
		return ok
	case BooleanType:
		if cval, ok := cmp.(BooleanType); ok {
			return vval == cval
		}
		return false
	case IntegerType:
		switch cval := cmp.(type) {
		case IntegerType:
			return vval == cval
		case NumberType:
			return float64(vval) == float64(cval)
		}
		return false
	case NumberType:
		switch cval := cmp.(type) {
		case NumberType:
			return float64(vval) == float64(cval)
		case IntegerType:
			return float64(vval) == float64(cval)
		}
		return false
	case StringType:
		if cval, ok := cmp.(StringType); ok {
			return vval == cval
		}
		return false
	case *ArrayType:
		return val == cmp
	case *ObjectType:
		return val == cmp
	default:
		return val == cmp
	}
}

// Translate any Go data type to a gescript datatype using reflection (ges/json)
func NewFromInterface(val interface{}) DataType {
	if val == nil {
		return NullType{}
	}

	rv := reflect.ValueOf(val)
	return fromReflectValue(rv)
}

// Translate a Go array to a gescript array (fully nested translation)
func NewFromSlice(items []interface{}) *ArrayType {
	arr := &ArrayType{
		Elements: make([]DataType, len(items)),
	}
	for idx, item := range items {
		arr.Elements[idx] = NewFromInterface(item)
	}
	return arr
}

// Translate a Go map to a gescript object (fully nested translation)
func NewFromMap(m map[string]interface{}) *ObjectType {
	obj := &ObjectType{
		Properties: make(map[string]DataType),
	}
	for key, val := range m {
		obj.Properties[key] = NewFromInterface(val)
	}
	return obj
}

// Translate a Go structure to a gescript object, using reflection (ges/json)
func NewFromStruct(v interface{}) *ObjectType {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return NewObject()
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return NewObject()
	}
	return objectFromReflectStruct(rv)
}

// Translate a reflection value into the associated gescript datatype
func fromReflectValue(rv reflect.Value) DataType {
	// Dereference pointers and interfaces
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return NullType{}
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		return BooleanType(rv.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return IntegerType(rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		return IntegerType(int64(rv.Uint()))

	case reflect.Float32, reflect.Float64:
		return NumberType(rv.Float())

	case reflect.String:
		return StringType(rv.String())

	case reflect.Slice, reflect.Array:
		return arrayFromReflectValue(rv)

	case reflect.Map:
		return objectFromReflectMap(rv)

	case reflect.Struct:
		return objectFromReflectStruct(rv)

	default:
		// Unsupported type, return undefined
		return UndefinedType{}
	}
}

// Convert a reflection array to gescript array
func arrayFromReflectValue(rv reflect.Value) *ArrayType {
	ln := rv.Len()
	arr := &ArrayType{
		Elements: make([]DataType, ln),
	}
	for idx := 0; idx < ln; idx++ {
		arr.Elements[idx] = fromReflectValue(rv.Index(idx))
	}
	return arr
}

// Convert a reflection map to a gescript object
func objectFromReflectMap(rv reflect.Value) *ObjectType {
	obj := &ObjectType{
		Properties: make(map[string]DataType),
	}
	iter := rv.MapRange()
	for iter.Next() {
		// Convert key to string either directly or through interface
		key := iter.Key()
		var keyStr string
		if key.Kind() == reflect.String {
			keyStr = key.String()
		} else {
			keyStr = reflect.ValueOf(key.Interface()).String()
		}
		obj.Properties[keyStr] = fromReflectValue(iter.Value())
	}
	return obj
}

// Convert a reflection struct to gescript object, using field keys (ges/json)
func objectFromReflectStruct(rv reflect.Value) *ObjectType {
	obj := &ObjectType{
		Properties: make(map[string]DataType),
	}
	rt := rv.Type()

	for idx := 0; idx < rv.NumField(); idx++ {
		field := rt.Field(idx)
		fieldVal := rv.Field(idx)

		// Ignore fields that are not marked for export
		if !field.IsExported() {
			continue
		}

		// Determine the property name from tags
		propName := getFieldName(field)
		if propName == "-" {
			continue
		}

		obj.Properties[propName] = fromReflectValue(fieldVal)
	}
	return obj
}

// Determine the property->field name - in order ges:, json: or field name
func getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("ges"); tag != "" {
		name, _ := parseTag(tag)
		if name != "" {
			return name
		}
	}

	if tag := field.Tag.Get("json"); tag != "" {
		name, _ := parseTag(tag)
		if name != "" {
			return name
		}
	}

	return field.Name
}

// Split the tag options from the base name
func parseTag(tag string) (string, string) {
	idx := strings.Index(tag, ",")
	if idx == -1 {
		return tag, ""
	}
	return tag[:idx], tag[idx+1:]
}

// For array, Native() recursively converts back to Go types
func (arr *ArrayType) Native() interface{} {
	if arr == nil {
		return nil
	}
	result := make([]interface{}, len(arr.Elements))
	for idx, elem := range arr.Elements {
		if elem != nil {
			result[idx] = elem.Native()
		}
	}
	return result
}

// For object, Native() recursively converts back to Go types
func (obj *ObjectType) Native() interface{} {
	if obj == nil {
		return nil
	}
	result := make(map[string]interface{})
	for key, val := range obj.Properties {
		if val != nil {
			result[key] = val.Native()
		}
	}
	return result
}

// Parse a JSON string into a gescript dataset
func ParseJSON(jsonStr string) (DataType, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, err
	}
	return NewFromInterface(raw), nil
}

// Convert a gescript dataset into JSON
func StringifyJSON(dt DataType) (string, error) {
	raw := dt.Native()
	bytes, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
