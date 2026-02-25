/*
 * Definitions of the ECMA data types and values used in the engine.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package types

import ()

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
func (sval StringType) ToPrimitive(pref any) DataType{
    return sval
}

// Collection of exposed 'known' types for use by internals and external callers
var (
    Undefined DataType = UndefinedType{}
)
