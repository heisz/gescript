/*
 * Definitions of the ECMA data types and values used in the engine.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package types

import "math"

// Note: the data types are openly exposed to support return type checking,
// along with some of the related methods for external convenience.  Also
// provides many of the elements in Section 7.1 (Type Conversion)

// Generic type definition for all underlying ECMA type instances
type DataType interface {
	// Convert the ECMA data value to the 'native' Go data type
	Native() interface{}

	// Type methods encapsulating rules of Section 7.1
	ToPrimitive(pref any) DataType
}

// Natively, undefined and null are 'same', but typed for differentiation
type UndefinedType struct{}

func (undef UndefinedType) Native() interface{} {
	return nil
}
func (undef UndefinedType) ToPrimitive(pref any) DataType {
	return undef
}

type NullType struct{}

func (nval NullType) Native() interface{} {
	return nil
}
func (nval NullType) ToPrimitive(pref any) DataType {
	return nval
}

type BooleanType bool

func (bval BooleanType) Native() interface{} {
	return bool(bval)
}
func (bval BooleanType) ToPrimitive(pref any) DataType {
	return bval
}

type IntegerType int64

func (ival IntegerType) Native() interface{} {
	return int64(ival)
}
func (ival IntegerType) ToPrimitive(pref any) DataType {
	return ival
}

type NumberType float64

func (nval NumberType) Native() interface{} {
	return float64(nval)
}
func (nval NumberType) ToPrimitive(pref any) DataType {
	return nval
}

type StringType string

func (sval StringType) Native() interface{} {
	return string(sval)
}
func (sval StringType) ToPrimitive(pref any) DataType {
	return sval
}

type ArrayType struct {
	Elements []DataType
}

// Native() is found in the conversion elements in util.go

func (arr *ArrayType) ToPrimitive(pref any) DataType {
	// TODO - is this the correct outcome?
	return StringType("[object Array]")
}

// Utility methods to actually work with the array contents
func (arr *ArrayType) Get(index int) DataType {
	if index < 0 || index >= len(arr.Elements) {
		return Undefined
	}
	return arr.Elements[index]
}
func (arr *ArrayType) Set(index int, val DataType) {
	// Automatically extend array if index falls outside of range
	for len(arr.Elements) <= index {
		arr.Elements = append(arr.Elements, Undefined)
	}
	arr.Elements[index] = val
}
func (arr *ArrayType) Length() int {
	return len(arr.Elements)
}
func NewArray(size int) *ArrayType {
	arr := &ArrayType{
		Elements: make([]DataType, size),
	}
	for idx := 0; idx < size; idx++ {
		arr.Elements[idx] = Undefined
	}
	return arr
}

type ObjectType struct {
	Properties map[string]DataType
}

// Native() is found in the conversion elements in util.go

func (obj *ObjectType) ToPrimitive(pref any) DataType {
	// TODO - is this the correct outcome?
	return StringType("[object Object]")
}

// Utility methods to actually work with the object contents
func (obj *ObjectType) Get(propName string) DataType {
	if val, ok := obj.Properties[propName]; ok {
		return val
	}
	return Undefined
}
func (obj *ObjectType) Set(propName string, val DataType) {
	obj.Properties[propName] = val
}
func (obj *ObjectType) Has(propName string) bool {
	_, ok := obj.Properties[propName]
	return ok
}
func NewObject() *ObjectType {
	return &ObjectType{
		Properties: make(map[string]DataType),
	}
}

// In ECMAScript functions are first-class, so we have types for them too
// This is the main interface for native/script functions that can be called
type FunctionType interface {
	DataType
	GetName() string
	Call(args []DataType) (DataType, error)
}

// NativeFn is the signature for Go functions callable from scripts
type NativeFn func(args []DataType) (DataType, error)

// And this is the native function datatype wrapper (implements FunctionType)
type NativeFunction struct {
	Name string
	Fn   NativeFn
}

func (nf *NativeFunction) Native() interface{} {
	return nf.Fn
}
func (nf *NativeFunction) ToPrimitive(pref any) DataType {
	return StringType("function " + nf.Name + "() { [native code] }")
}

func (nf *NativeFunction) GetName() string {
	return nf.Name
}

func (nf *NativeFunction) Call(args []DataType) (DataType, error) {
	return nf.Fn(args)
}

// NativeMethod is a native function bound to an object instance (this)
type NativeMethod struct {
	Target DataType
	Method *NativeFunction
}

func (bm *NativeMethod) Native() interface{} {
	return bm.Method.Fn
}

func (bm *NativeMethod) ToPrimitive(pref any) DataType {
	return StringType("function " + bm.Method.Name + "() { [bound] }")
}

func (bm *NativeMethod) GetName() string {
	return bm.Method.Name
}

func (bm *NativeMethod) Call(args []DataType) (DataType, error) {
	// Prepend target as first argument ("this")
	allArgs := make([]DataType, len(args)+1)
	allArgs[0] = bm.Target
	copy(allArgs[1:], args)
	return bm.Method.Fn(allArgs)
}

// Collection of exposed 'known' constants for internals and external callers
var (
	Undefined DataType = UndefinedType{}
	NaN       DataType = NumberType(math.NaN())
)
