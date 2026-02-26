/*
 * Definition and implementations for operations in the script engine.
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package engine

import (
	"fmt"
	"math"

	"github.com/heisz/gescript/types"
)

// Instead of a set of bytecodes with a big switch, the 'compiler' instead
// maintains an array of opcode instances tied to a function that processes
// the working stack (contained in this process).  Original opcode is provided
// to provide access to context-specific compilation data.  Returned error
// (if non-nil) is either a global failure or an internal catchable exception
// wrapper.

// Note: they aren't really opcodes like a true VM, but stick with the original
// naming/terminology from the C implementation.  However, did switch from
// execution context to a process terminology from that other project...

type OpCodeFn func(prc *Process, op *OpCode) (err error)

type OpCode struct {
	LineNumber int
	ExecFn     OpCodeFn
	OpData     interface{}
}

// All of the various opcode functions appear below

func PushLiteralValue(prc *Process, op *OpCode) (err error) {
	err = prc.push(op.OpData.(*types.DataType))
	return
}

func AdditionOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Per 13.15.3, if either operand is string, result is string concat
	_, lisstr := (*left).(types.StringType)
	_, risstr := (*right).(types.StringType)
	if lisstr || risstr {
		// TODO - need toString() plus string concat
	}

	// Big sets of switch statements to handle all of the mixes
	var res types.DataType = types.Undefined
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.IntegerType((*left).Native().(int64) +
				(*right).Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64((*left).Native().(int64)) +
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.NumberType((*left).Native().(float64) +
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.NumberType((*left).Native().(float64) +
				(*right).Native().(float64))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func SubtractionOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Big sets of switch statements to handle all of the mixes
	var res types.DataType = types.Undefined
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.IntegerType((*left).Native().(int64) -
				(*right).Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64((*left).Native().(int64)) -
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.NumberType((*left).Native().(float64) -
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.NumberType((*left).Native().(float64) -
				(*right).Native().(float64))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func MultiplicationOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Big sets of switch statements to handle all of the mixes
	var res types.DataType = types.Undefined
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.IntegerType((*left).Native().(int64) *
				(*right).Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64((*left).Native().(int64)) *
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.NumberType((*left).Native().(float64) *
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.NumberType((*left).Native().(float64) *
				(*right).Native().(float64))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func DivisionOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Slightly different, division always produces number per specification
	var res types.DataType = types.Undefined
	var lval, rval float64

	switch (*left).(type) {
	case types.IntegerType:
		lval = float64((*left).Native().(int64))
	case types.NumberType:
		lval = (*left).Native().(float64)
	}

	switch (*right).(type) {
	case types.IntegerType:
		rval = float64((*right).Native().(int64))
	case types.NumberType:
		rval = (*right).Native().(float64)
	}

	res = types.NumberType(lval / rval)
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func ModulusOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Per specification, int results in int but float has special rules
	var res types.DataType = types.Undefined
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			rval := (*right).Native().(int64)
			if rval != 0 {
				res = types.IntegerType((*left).Native().(int64) % rval)
			}
		case types.NumberType:
			lval := float64((*left).Native().(int64))
			rval := (*right).Native().(float64)
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			lval := (*left).Native().(float64)
			rval := float64((*right).Native().(int64))
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		case types.NumberType:
			lval := (*left).Native().(float64)
			rval := (*right).Native().(float64)
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func LeftShiftOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Per specification, convert to int for shift operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// Mask shift amount to 5 bits per specification
	res := types.IntegerType(int32(lval) << (uint32(rval) & 0x1F))
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func RightShiftOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Per specification, convert to int for shift operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// Mask shift amount to 5 bits per specification
	res := types.IntegerType(int32(lval) >> (uint32(rval) & 0x1F))
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func UnsignedRightShiftOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Per specification, convert to int for shift operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// Unsigned right shift - convert to uint32 first (still 5 bit mask)
	res := types.IntegerType(uint32(lval) >> (uint32(rval) & 0x1F))
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func LessThanOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) <
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) <
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) <
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) <
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) <
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func GreaterThanOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) >
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) >
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) >
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) >
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) >
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func LessThanEqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) <=
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) <=
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) <=
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) <=
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) <=
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func GreaterThanEqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) >=
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) >=
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) >=
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) >=
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) >=
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func EqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Abstract equality comparison (with type coercion)
	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.UndefinedType:
		switch (*right).(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(true)
		}
	case types.NullType:
		switch (*right).(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(true)
		}
	case types.BooleanType:
		switch (*right).(type) {
		case types.BooleanType:
			res = types.BooleanType((*left).Native().(bool) ==
				(*right).Native().(bool))
		}
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) ==
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) ==
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) ==
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) ==
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) ==
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func NotEqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Abstract inequality comparison (with type coercion, opposite of ==)
	var res types.DataType = types.BooleanType(true)
	switch (*left).(type) {
	case types.UndefinedType:
		switch (*right).(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(false)
		}
	case types.NullType:
		switch (*right).(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(false)
		}
	case types.BooleanType:
		switch (*right).(type) {
		case types.BooleanType:
			res = types.BooleanType((*left).Native().(bool) !=
				(*right).Native().(bool))
		}
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) !=
				(*right).Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64((*left).Native().(int64)) !=
				(*right).Native().(float64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(float64) !=
				float64((*right).Native().(int64)))
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) !=
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) !=
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func StrictEqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Strict equality - types and values must exactly match
	var res types.DataType = types.BooleanType(false)
	switch (*left).(type) {
	case types.UndefinedType:
		switch (*right).(type) {
		case types.UndefinedType:
			res = types.BooleanType(true)
		}
	case types.NullType:
		switch (*right).(type) {
		case types.NullType:
			res = types.BooleanType(true)
		}
	case types.BooleanType:
		switch (*right).(type) {
		case types.BooleanType:
			res = types.BooleanType((*left).Native().(bool) ==
				(*right).Native().(bool))
		}
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) ==
				(*right).Native().(int64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) ==
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) ==
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func StrictNotEqualOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Strict inequality - types must match while values must not
	var res types.DataType = types.BooleanType(true)
	switch (*left).(type) {
	case types.UndefinedType:
		switch (*right).(type) {
		case types.UndefinedType:
			res = types.BooleanType(false)
		}
	case types.NullType:
		switch (*right).(type) {
		case types.NullType:
			res = types.BooleanType(false)
		}
	case types.BooleanType:
		switch (*right).(type) {
		case types.BooleanType:
			res = types.BooleanType((*left).Native().(bool) !=
				(*right).Native().(bool))
		}
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.BooleanType((*left).Native().(int64) !=
				(*right).Native().(int64))
		}
	case types.NumberType:
		switch (*right).(type) {
		case types.NumberType:
			res = types.BooleanType((*left).Native().(float64) !=
				(*right).Native().(float64))
		}
	case types.StringType:
		switch (*right).(type) {
		case types.StringType:
			res = types.BooleanType((*left).Native().(string) !=
				(*right).Native().(string))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func BitwiseAndOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Need integers for bit operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval & rval)
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func BitwiseOrOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Need integers for bit operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval | rval)
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func BitwiseXorOperation(prc *Process, op *OpCode) (err error) {
	// Pull the operands
	right, err := prc.pop()
	if err != nil {
		return err
	}
	left, err := prc.pop()
	if err != nil {
		return err
	}

	// Need integers for bit operations
	var lval, rval int64
	switch (*left).(type) {
	case types.IntegerType:
		lval = (*left).Native().(int64)
	case types.NumberType:
		lval = int64((*left).Native().(float64))
	}
	switch (*right).(type) {
	case types.IntegerType:
		rval = (*right).Native().(int64)
	case types.NumberType:
		rval = int64((*right).Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval ^ rval)
	val := types.DataType(res)
	err = prc.push(&val)
	return
}

func UnaryPlusOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	// ToNumber conversion per specification (TODO - move?)
	var res types.DataType
	switch (*val).(type) {
	case types.IntegerType:
		res = *val
	case types.NumberType:
		res = *val
	case types.BooleanType:
		if (*val).Native().(bool) {
			res = types.IntegerType(1)
		} else {
			res = types.IntegerType(0)
		}
	case types.StringType:
		// TODO - proper string to number parsing
		res = types.NumberType(0)
	default:
		res = types.NumberType(0)
	}

	err = prc.push(&res)
	return
}

func UnaryMinusOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType
	switch (*val).(type) {
	case types.IntegerType:
		res = types.IntegerType(-(*val).Native().(int64))
	case types.NumberType:
		res = types.NumberType(-(*val).Native().(float64))
	case types.BooleanType:
		if (*val).Native().(bool) {
			res = types.IntegerType(-1)
		} else {
			res = types.IntegerType(0)
		}
	default:
		res = types.NumberType(0)
	}

	err = prc.push(&res)
	return
}

func LogicalNotOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	res := types.DataType(types.BooleanType(!types.IsTruthy(val)))
	err = prc.push(&res)
	return
}

func BitwiseNotOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	var ival int64
	switch (*val).(type) {
	case types.IntegerType:
		ival = (*val).Native().(int64)
	case types.NumberType:
		ival = int64((*val).Native().(float64))
	}

	res := types.DataType(types.IntegerType(^int32(ival)))
	err = prc.push(&res)
	return
}

// NOTE: for jump operations (here and elsewhere) always deduct one for pc++

func JumpOperation(prc *Process, op *OpCode) (err error) {
	target := op.OpData.(int)
	prc.pc = target - 1
	return
}

func JumpIfFalseOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	if !types.IsTruthy(val) {
		target := op.OpData.(int)
		prc.pc = target - 1
	}
	return
}

// Short-circuit operation, jump or pop based on truthiness
func JumpIfFalseOrPopOperation(prc *Process, op *OpCode) (err error) {
	// Just peek so we can leave the value
	val, err := prc.peek()
	if err != nil {
		return err
	}

	if !types.IsTruthy(val) {
		// If false, leave the value and jump to target
		target := op.OpData.(int)
		prc.pc = target - 1
	} else {
		// Otherwise, pop/discard the top value and continue
		prc.pop()
	}
	return
}

func JumpIfTrueOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	if types.IsTruthy(val) {
		target := op.OpData.(int)
		prc.pc = target - 1
	}
	return
}

// Short-circuit operation, jump or pop based on truthiness (opposite of prior)
func JumpIfTrueOrPopOperation(prc *Process, op *OpCode) (err error) {
	// Just peek so we can leave the value
	val, err := prc.peek()
	if err != nil {
		return err
	}

	if types.IsTruthy(val) {
		// If true, leave the value and jump to target
		target := op.OpData.(int)
		prc.pc = target - 1
	} else {
		// Otherwise, pop/discard the top value and continue
		prc.pop()
	}
	return
}

func PopOperation(prc *Process, op *OpCode) (err error) {
	_, err = prc.pop()
	return
}

func DupOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.peek()
	if err != nil {
		return err
	}
	err = prc.push(val)
	return
}

func LoadVariableOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		// TODO - proper error type
		return nil
	}
	err = prc.push(prc.locals[slotIndex])
	return
}

func StoreVariableOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		// TODO - proper error type
		return nil
	}
	val, err := prc.pop()
	if err != nil {
		return err
	}
	prc.locals[slotIndex] = val
	return
}

// Like Store above but leave value on stack for assignment chaining
func StoreVariableKeepOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		// TODO - proper error type
		return nil
	}
	val, err := prc.peek()
	if err != nil {
		return err
	}
	prc.locals[slotIndex] = val
	return
}

// Push the associated exception context onto the stack (at try statement)
func PushExceptionContextOperation(prc *Process, op *OpCode) (err error) {
	src := op.OpData.(*ExceptionContext)
	ctx := &ExceptionContext{
		previous:      prc.exceptionCtx,
		CatchTarget:   src.CatchTarget,
		FinallyTarget: src.FinallyTarget,
		EndTarget:     src.EndTarget,
		StackDepth:    prc.sp,
		CatchVarSlot:  src.CatchVarSlot,
	}
	prc.exceptionCtx = ctx
	return
}

func PopExceptionContextOperation(prc *Process, op *OpCode) (err error) {
	if prc.exceptionCtx != nil {
		prc.exceptionCtx = prc.exceptionCtx.previous
	}
	return
}

// Save the exception instance and use the special error to signal try handling
func ThrowOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}
	prc.exception = val
	return ErrException
}

// Common helper methods for increment and decrement
func incrementValue(val *types.DataType) types.DataType {
	switch (*val).(type) {
	case types.IntegerType:
		return types.IntegerType((*val).Native().(int64) + 1)
	case types.NumberType:
		return types.NumberType((*val).Native().(float64) + 1)
	}

	// For everything else, result is NaN
	return types.NaN
}

func decrementValue(val *types.DataType) types.DataType {
	switch (*val).(type) {
	case types.IntegerType:
		return types.IntegerType((*val).Native().(int64) - 1)
	case types.NumberType:
		return types.NumberType((*val).Native().(float64) - 1)
	}

	// For everything else, result is NaN
	return types.NaN
}

func PreIncrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	val := incrementValue(prc.locals[slotIndex])
	prc.locals[slotIndex] = &val
	err = prc.push(&val)
	return
}

func PreDecrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	val := decrementValue(prc.locals[slotIndex])
	prc.locals[slotIndex] = &val
	err = prc.push(&val)
	return
}

func PostIncrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	orig := prc.locals[slotIndex]
	val := incrementValue(orig)
	prc.locals[slotIndex] = &val
	err = prc.push(orig)
	return
}

func PostDecrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	orig := prc.locals[slotIndex]
	val := decrementValue(orig)
	prc.locals[slotIndex] = &val
	err = prc.push(orig)
	return
}

// These are much more complicated because of the different reference types
// And yes, there is a ton of duplicated code.  Trade op binding for dup...
func PreIncrementElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var val types.DataType
	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig := tgt.Get(idx)
		val = incrementValue(orig)
		tgt.Set(idx, &val)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig := tgt.Get(propName)
		val = incrementValue(orig)
		tgt.Set(propName, &val)
	default:
		val = types.NaN
	}

	err = prc.push(&val)
	return
}

func PostIncrementElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig *types.DataType
	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig = tgt.Get(idx)
		val := incrementValue(orig)
		tgt.Set(idx, &val)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig = tgt.Get(propName)
		val := incrementValue(orig)
		tgt.Set(propName, &val)
	default:
		orig = &types.NaN
	}

	err = prc.push(orig)
	return
}

func PreDecrementElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var val types.DataType
	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig := tgt.Get(idx)
		val = decrementValue(orig)
		tgt.Set(idx, &val)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig := tgt.Get(propName)
		val = decrementValue(orig)
		tgt.Set(propName, &val)
	default:
		val = types.NaN
	}

	err = prc.push(&val)
	return
}

func PostDecrementElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig *types.DataType
	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig = tgt.Get(idx)
		val := decrementValue(orig)
		tgt.Set(idx, &val)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig = tgt.Get(propName)
		val := decrementValue(orig)
		tgt.Set(propName, &val)
	default:
		orig = &types.NaN
	}

	err = prc.push(orig)
	return
}

// Ditto for pre/post unary on object member
func PreIncrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var val types.DataType
	switch tgt := (*target).(type) {
	case *types.ObjectType:
		orig := tgt.Get(propName)
		val = incrementValue(orig)
		tgt.Set(propName, &val)
	default:
		val = types.NaN
	}

	err = prc.push(&val)
	return
}

func PostIncrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig *types.DataType
	switch tgt := (*target).(type) {
	case *types.ObjectType:
		orig = tgt.Get(propName)
		val := incrementValue(orig)
		tgt.Set(propName, &val)
	default:
		orig = &types.NaN
	}

	err = prc.push(orig)
	return
}

func PreDecrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var val types.DataType
	switch tgt := (*target).(type) {
	case *types.ObjectType:
		orig := tgt.Get(propName)
		val = decrementValue(orig)
		tgt.Set(propName, &val)
	default:
		val = types.NaN
	}

	err = prc.push(&val)
	return
}

func PostDecrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig *types.DataType
	switch tgt := (*target).(type) {
	case *types.ObjectType:
		orig = tgt.Get(propName)
		val := decrementValue(orig)
		tgt.Set(propName, &val)
	default:
		orig = &types.NaN
	}

	err = prc.push(orig)
	return
}

func NewArrayOperation(prc *Process, op *OpCode) (err error) {
	count := op.OpData.(int)
	arr := types.NewArray(count)

	// Elements are on stack in reverse order
	for idx := count - 1; idx >= 0; idx-- {
		val, err := prc.pop()
		if err != nil {
			return err
		}
		arr.Elements[idx] = val
	}

	res := types.DataType(arr)
	err = prc.push(&res)
	return
}

func NewObjectOperation(prc *Process, op *OpCode) (err error) {
	keys := op.OpData.([]string)
	obj := types.NewObject()

	// Elements are on stack in reverse order
	for idx := len(keys) - 1; idx >= 0; idx-- {
		val, err := prc.pop()
		if err != nil {
			return err
		}
		obj.Set(keys[idx], val)
	}

	res := types.DataType(obj)
	err = prc.push(&res)
	return
}

func GetElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	// Three cases so far depending on target type
	var res *types.DataType
	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		default:
			res = &types.Undefined
			err = prc.push(res)
			return
		}
		res = tgt.Get(idx)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		default:
			res = &types.Undefined
			err = prc.push(res)
			return
		}
		res = tgt.Get(propName)
	case types.StringType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		default:
			undef := types.Undefined
			res = &undef
			err = prc.push(res)
			return
		}
		str := string(tgt)
		if idx >= 0 && idx < len(str) {
			ch := types.DataType(types.StringType(str[idx : idx+1]))
			res = &ch
		} else {
			res = &types.Undefined
		}
	default:
		res = &types.Undefined
	}

	err = prc.push(res)
	return
}

func SetElementOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	switch tgt := (*target).(type) {
	case *types.ArrayType:
		var idx int
		switch ix := (*index).(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := (*index).(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		tgt.Set(propName, val)
	}

	// Push the value back onto the stack (residual from assignment)
	err = prc.push(val)
	return
}

func GetPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var res *types.DataType
	switch tgt := (*target).(type) {
	case *types.ObjectType:
		res = tgt.Get(propName)
	case *types.ArrayType:
		// length is a special member of an array
		if propName == "length" {
			len := types.DataType(types.IntegerType(tgt.Length()))
			res = &len
		} else {
			res = &types.Undefined
		}
	case types.StringType:
		// length is a special member of a string
		if propName == "length" {
			len := types.DataType(types.IntegerType(len(string(tgt))))
			res = &len
		} else {
			res = &types.Undefined
		}
	default:
		res = &types.Undefined
	}

	err = prc.push(res)
	return
}

func SetPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	val, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	switch tgt := (*target).(type) {
	case *types.ObjectType:
		tgt.Set(propName, val)
	}

	// Push the value back onto the stack (residual from assignment)
	err = prc.push(val)
	return
}
