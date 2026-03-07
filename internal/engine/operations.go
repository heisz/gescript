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
	"strconv"

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
	err = prc.push(op.OpData.(types.DataType))
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
	_, lisstr := left.(types.StringType)
	_, risstr := right.(types.StringType)
	if lisstr || risstr {
		leftStr := types.ToString(left)
		rightStr := types.ToString(right)
		return prc.push(types.StringType(leftStr + rightStr))
	}

	// Big sets of switch statements to handle all of the mixes
	var res types.DataType = types.Undefined
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.IntegerType(left.Native().(int64) +
				right.Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64(left.Native().(int64)) +
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.NumberType(left.Native().(float64) +
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.NumberType(left.Native().(float64) +
				right.Native().(float64))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.IntegerType(left.Native().(int64) -
				right.Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64(left.Native().(int64)) -
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.NumberType(left.Native().(float64) -
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.NumberType(left.Native().(float64) -
				right.Native().(float64))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.IntegerType(left.Native().(int64) *
				right.Native().(int64))
		case types.NumberType:
			res = types.NumberType(float64(left.Native().(int64)) *
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.NumberType(left.Native().(float64) *
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.NumberType(left.Native().(float64) *
				right.Native().(float64))
		}
	}

	err = prc.push(res)
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

	switch left.(type) {
	case types.IntegerType:
		lval = float64(left.Native().(int64))
	case types.NumberType:
		lval = left.Native().(float64)
	}

	switch right.(type) {
	case types.IntegerType:
		rval = float64(right.Native().(int64))
	case types.NumberType:
		rval = right.Native().(float64)
	}

	res = types.NumberType(lval / rval)
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			rval := right.Native().(int64)
			if rval != 0 {
				res = types.IntegerType(left.Native().(int64) % rval)
			}
		case types.NumberType:
			lval := float64(left.Native().(int64))
			rval := right.Native().(float64)
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			lval := left.Native().(float64)
			rval := float64(right.Native().(int64))
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		case types.NumberType:
			lval := left.Native().(float64)
			rval := right.Native().(float64)
			if rval != 0 {
				res = types.NumberType(math.Mod(lval, rval))
			}
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// Mask shift amount to 5 bits per specification
	res := types.IntegerType(int32(lval) << (uint32(rval) & 0x1F))
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// Mask shift amount to 5 bits per specification
	res := types.IntegerType(int32(lval) >> (uint32(rval) & 0x1F))
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// Unsigned right shift - convert to uint32 first (still 5 bit mask)
	res := types.IntegerType(uint32(lval) >> (uint32(rval) & 0x1F))
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) <
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) <
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) <
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) <
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) <
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) >
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) >
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) >
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) >
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) >
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) <=
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) <=
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) <=
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) <=
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) <=
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) >=
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) >=
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) >=
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) >=
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) >=
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.UndefinedType:
		switch right.(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(true)
		}
	case types.NullType:
		switch right.(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(true)
		}
	case types.BooleanType:
		switch right.(type) {
		case types.BooleanType:
			res = types.BooleanType(left.Native().(bool) ==
				right.Native().(bool))
		}
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) ==
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) ==
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) ==
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) ==
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) ==
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.UndefinedType:
		switch right.(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(false)
		}
	case types.NullType:
		switch right.(type) {
		case types.UndefinedType, types.NullType:
			res = types.BooleanType(false)
		}
	case types.BooleanType:
		switch right.(type) {
		case types.BooleanType:
			res = types.BooleanType(left.Native().(bool) !=
				right.Native().(bool))
		}
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) !=
				right.Native().(int64))
		case types.NumberType:
			res = types.BooleanType(float64(left.Native().(int64)) !=
				right.Native().(float64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(float64) !=
				float64(right.Native().(int64)))
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) !=
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) !=
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.UndefinedType:
		switch right.(type) {
		case types.UndefinedType:
			res = types.BooleanType(true)
		}
	case types.NullType:
		switch right.(type) {
		case types.NullType:
			res = types.BooleanType(true)
		}
	case types.BooleanType:
		switch right.(type) {
		case types.BooleanType:
			res = types.BooleanType(left.Native().(bool) ==
				right.Native().(bool))
		}
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) ==
				right.Native().(int64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) ==
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) ==
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.UndefinedType:
		switch right.(type) {
		case types.UndefinedType:
			res = types.BooleanType(false)
		}
	case types.NullType:
		switch right.(type) {
		case types.NullType:
			res = types.BooleanType(false)
		}
	case types.BooleanType:
		switch right.(type) {
		case types.BooleanType:
			res = types.BooleanType(left.Native().(bool) !=
				right.Native().(bool))
		}
	case types.IntegerType:
		switch right.(type) {
		case types.IntegerType:
			res = types.BooleanType(left.Native().(int64) !=
				right.Native().(int64))
		}
	case types.NumberType:
		switch right.(type) {
		case types.NumberType:
			res = types.BooleanType(left.Native().(float64) !=
				right.Native().(float64))
		}
	case types.StringType:
		switch right.(type) {
		case types.StringType:
			res = types.BooleanType(left.Native().(string) !=
				right.Native().(string))
		}
	}

	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval & rval)
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval | rval)
	err = prc.push(res)
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
	switch left.(type) {
	case types.IntegerType:
		lval = left.Native().(int64)
	case types.NumberType:
		lval = int64(left.Native().(float64))
	}
	switch right.(type) {
	case types.IntegerType:
		rval = right.Native().(int64)
	case types.NumberType:
		rval = int64(right.Native().(float64))
	}

	// The specification doesn't indicate reduction to 32 bits like shift
	res := types.IntegerType(lval ^ rval)
	err = prc.push(res)
	return
}

func UnaryPlusOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	// ToNumber conversion per specification (TODO - move?)
	var res types.DataType
	switch val.(type) {
	case types.IntegerType:
		res = val
	case types.NumberType:
		res = val
	case types.BooleanType:
		if val.Native().(bool) {
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

	err = prc.push(res)
	return
}

func UnaryMinusOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType
	switch val.(type) {
	case types.IntegerType:
		res = types.IntegerType(-val.Native().(int64))
	case types.NumberType:
		res = types.NumberType(-val.Native().(float64))
	case types.BooleanType:
		if val.Native().(bool) {
			res = types.IntegerType(-1)
		} else {
			res = types.IntegerType(0)
		}
	default:
		res = types.NumberType(0)
	}

	err = prc.push(res)
	return
}

func LogicalNotOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	res := types.DataType(types.BooleanType(!types.IsTruthy(val)))
	err = prc.push(res)
	return
}

func BitwiseNotOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	var ival int64
	switch val.(type) {
	case types.IntegerType:
		ival = val.Native().(int64)
	case types.NumberType:
		ival = int64(val.Native().(float64))
	}

	res := types.DataType(types.IntegerType(^int32(ival)))
	err = prc.push(res)
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

	// Handle closure capture of variable (in capture cell)
	if prc.cells != nil && slotIndex < len(prc.cells) &&
		prc.cells[slotIndex] != nil {
		err = prc.push(*prc.cells[slotIndex].Value)
		return
	}

	err = prc.push(prc.locals[slotIndex])
	return
}

func storeVariable(prc *Process, slot int, val types.DataType) (err error) {
	if slot < 0 || slot >= len(prc.locals) {
		// TODO - proper error type
		return nil
	}

	// Handle closure capture of variable (in capture cell)
	if prc.cells != nil && slot < len(prc.cells) && prc.cells[slot] != nil {
		*prc.cells[slot].Value = val
		return
	}

	// Otherwise it's a local write
	prc.locals[slot] = val
	return
}

func StoreVariableOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	val, err := prc.pop()
	if err != nil {
		return err
	}

	return storeVariable(prc, slotIndex, val)
}

// Like Store above but leave value on stack for assignment chaining
func StoreVariableKeepOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	val, err := prc.peek()
	if err != nil {
		return err
	}

	return storeVariable(prc, slotIndex, val)
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
	prc.exception = &val
	return ErrException
}

// Common helper methods for increment and decrement
func incrementValue(val types.DataType) types.DataType {
	switch val.(type) {
	case types.IntegerType:
		return types.IntegerType(val.Native().(int64) + 1)
	case types.NumberType:
		return types.NumberType(val.Native().(float64) + 1)
	}

	// For everything else, result is NaN
	return types.NaN
}

func decrementValue(val types.DataType) types.DataType {
	switch val.(type) {
	case types.IntegerType:
		return types.IntegerType(val.Native().(int64) - 1)
	case types.NumberType:
		return types.NumberType(val.Native().(float64) - 1)
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
	prc.locals[slotIndex] = val
	err = prc.push(val)
	return
}

func PreDecrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	val := decrementValue(prc.locals[slotIndex])
	prc.locals[slotIndex] = val
	err = prc.push(val)
	return
}

func PostIncrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	orig := prc.locals[slotIndex]
	prc.locals[slotIndex] = incrementValue(orig)
	err = prc.push(orig)
	return
}

func PostDecrementOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)
	if slotIndex < 0 || slotIndex >= len(prc.locals) {
		return nil
	}
	orig := prc.locals[slotIndex]
	prc.locals[slotIndex] = decrementValue(orig)
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
	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig := tgt.Get(idx)
		val = incrementValue(orig)
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig := tgt.Get(propName)
		val = incrementValue(orig)
		tgt.Set(propName, val)
	default:
		val = types.NaN
	}

	err = prc.push(val)
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

	var orig types.DataType
	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig = tgt.Get(idx)
		val := incrementValue(orig)
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig = tgt.Get(propName)
		val := incrementValue(orig)
		tgt.Set(propName, val)
	default:
		orig = types.NaN
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
	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig := tgt.Get(idx)
		val = decrementValue(orig)
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig := tgt.Get(propName)
		val = decrementValue(orig)
		tgt.Set(propName, val)
	default:
		val = types.NaN
	}

	err = prc.push(val)
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

	var orig types.DataType
	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		orig = tgt.Get(idx)
		val := decrementValue(orig)
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		}
		orig = tgt.Get(propName)
		val := decrementValue(orig)
		tgt.Set(propName, val)
	default:
		orig = types.NaN
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
	switch tgt := target.(type) {
	case *types.ObjectType:
		orig := tgt.Get(propName)
		val = incrementValue(orig)
		tgt.Set(propName, val)
	default:
		val = types.NaN
	}

	err = prc.push(val)
	return
}

func PostIncrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig types.DataType
	switch tgt := target.(type) {
	case *types.ObjectType:
		orig = tgt.Get(propName)
		val := incrementValue(orig)
		tgt.Set(propName, val)
	default:
		orig = types.NaN
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
	switch tgt := target.(type) {
	case *types.ObjectType:
		orig := tgt.Get(propName)
		val = decrementValue(orig)
		tgt.Set(propName, val)
	default:
		val = types.NaN
	}

	err = prc.push(val)
	return
}

func PostDecrementPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var orig types.DataType
	switch tgt := target.(type) {
	case *types.ObjectType:
		orig = tgt.Get(propName)
		val := decrementValue(orig)
		tgt.Set(propName, val)
	default:
		orig = types.NaN
	}

	err = prc.push(orig)
	return
}

func TypeofOperation(prc *Process, op *OpCode) (err error) {
	val, err := prc.pop()
	if err != nil {
		return err
	}

	var result string
	switch val.(type) {
	case types.UndefinedType:
		result = "undefined"
	case types.NullType:
		// Per spec, typeof null === "object"
		result = "object"
	case types.BooleanType:
		result = "boolean"
	case types.IntegerType, types.NumberType:
		result = "number"
	case types.StringType:
		result = "string"
	case *ScriptFunction, *types.NativeFunction:
		result = "function"
	case *types.ArrayType, *types.ObjectType:
		result = "object"
	default:
		result = "undefined"
	}

	return prc.push(types.StringType(result))
}

func InstanceofOperation(prc *Process, op *OpCode) (err error) {
	constructor, err := prc.pop()
	if err != nil {
		return err
	}
	obj, err := prc.pop()
	if err != nil {
		return err
	}

	// Get the constructor name for comparison
	var constructorName string
	switch c := constructor.(type) {
	case *types.NativeFunction:
		constructorName = c.Name
	case *ScriptFunction:
		constructorName = c.Name
	default:
		// Not a function, instanceof is false
		return prc.push(types.BooleanType(false))
	}

	// Check object type against constructor name
	var result bool
	switch constructorName {
	case "Function":
		_, isScript := obj.(*ScriptFunction)
		_, isNative := obj.(*types.NativeFunction)
		result = isScript || isNative
	case "Array":
		_, result = obj.(*types.ArrayType)
	case "Object":
		_, result = obj.(*types.ObjectType)
	case "Number":
		_, isInt := obj.(types.IntegerType)
		_, isNum := obj.(types.NumberType)
		result = isInt || isNum
	case "String":
		_, result = obj.(types.StringType)
	case "Boolean":
		_, result = obj.(types.BooleanType)
	default:
		result = false
	}

	return prc.push(types.BooleanType(result))
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

	err = prc.push(types.DataType(arr))
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
	err = prc.push(res)
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
	var res types.DataType
	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		case types.StringType:
			// Handle numerical string indexing (for ... in)
			sidx, convErr := strconv.Atoi(string(ix))
			if convErr != nil {
				res = types.Undefined
				err = prc.push(res)
				return
			}
			idx = sidx
		default:
			res = types.Undefined
			err = prc.push(res)
			return
		}
		res = tgt.Get(idx)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
		case types.StringType:
			propName = string(ix)
		case types.IntegerType:
			propName = fmt.Sprintf("%d", ix)
		default:
			res = types.Undefined
			err = prc.push(res)
			return
		}
		res = tgt.Get(propName)
	case types.StringType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		default:
			undef := types.Undefined
			res = undef
			err = prc.push(res)
			return
		}
		str := string(tgt)
		if idx >= 0 && idx < len(str) {
			ch := types.DataType(types.StringType(str[idx : idx+1]))
			res = ch
		} else {
			res = types.Undefined
		}
	default:
		res = types.Undefined
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

	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		}
		tgt.Set(idx, val)
	case *types.ObjectType:
		var propName string
		switch ix := index.(type) {
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

func DeleteElementOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	target, err := prc.pop()
	if err != nil {
		return err
	}

	switch tgt := target.(type) {
	case *types.ArrayType:
		var idx int
		switch ix := index.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		default:
			return prc.push(types.BooleanType(false))
		}
		if idx >= 0 && idx < len(tgt.Elements) {
			tgt.Elements[idx] = types.Undefined
			return prc.push(types.BooleanType(true))
		}
	case *types.ObjectType:
		propName := types.ToString(index)
		delete(tgt.Properties, propName)
		return prc.push(types.BooleanType(true))
	}

	return prc.push(types.BooleanType(false))
}

func InOperation(prc *Process, op *OpCode) (err error) {
	obj, err := prc.pop()
	if err != nil {
		return err
	}
	prop, err := prc.pop()
	if err != nil {
		return err
	}

	propName := types.ToString(prop)

	switch tgt := obj.(type) {
	case *types.ObjectType:
		_, exists := tgt.Properties[propName]
		return prc.push(types.BooleanType(exists))
	case *types.ArrayType:
		// For arrays, check if index exists
		var idx int
		switch ix := prop.(type) {
		case types.IntegerType:
			idx = int(ix)
		case types.NumberType:
			idx = int(ix)
		default:
			return prc.push(types.BooleanType(false))
		}
		exists := idx >= 0 && idx < len(tgt.Elements)
		return prc.push(types.BooleanType(exists))
	}

	return prc.push(types.BooleanType(false))
}

func GetPropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	target, err := prc.pop()
	if err != nil {
		return err
	}

	var res types.DataType
	switch tgt := target.(type) {
	case *types.ObjectType:
		res = tgt.Get(propName)
	case *types.ArrayType:
		// length is a special member of an array
		if propName == "length" {
			len := types.DataType(types.IntegerType(tgt.Length()))
			res = len
		} else {
			res = types.Undefined
		}
	case types.StringType:
		// length is a special member of a string
		if propName == "length" {
			len := types.DataType(types.IntegerType(len(string(tgt))))
			res = len
		} else {
			res = types.Undefined
		}
	default:
		res = types.Undefined
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

	switch tgt := target.(type) {
	case *types.ObjectType:
		tgt.Set(propName, val)
	}

	// Push the value back onto the stack (residual from assignment)
	err = prc.push(val)
	return
}

func DeletePropertyOperation(prc *Process, op *OpCode) (err error) {
	propName := op.OpData.(string)
	obj, err := prc.pop()
	if err != nil {
		return err
	}

	// Property delete only works on objects
	if objVal, ok := obj.(*types.ObjectType); ok {
		delete(objVal.Properties, propName)
		return prc.push(types.BooleanType(true))
	}

	// Non-object delete returns true (no-op)
	return prc.push(types.BooleanType(true))
}

func LoadGlobalOperation(prc *Process, op *OpCode) (err error) {
	name := op.OpData.(string)

	// Check script globals first (script-defined functions)
	if prc.globals != nil {
		if val, ok := prc.globals[name]; ok {
			return prc.push(val)
		}
	}

	// Then check native functions (library and program extensions)
	if prc.natives != nil {
		if val, ok := prc.natives[name]; ok {
			return prc.push(val)
		}
	}

	// Not found in either, clearly undefined...
	return prc.push(types.Undefined)
}

func StoreGlobalOperation(prc *Process, op *OpCode) (err error) {
	name := op.OpData.(string)
	val, err := prc.peek()
	if err != nil {
		return err
	}

	// Script-defined functions go into the globals map
	if prc.globals == nil {
		prc.globals = make(map[string]types.DataType)
	}
	prc.globals[name] = val
	return
}

func CallOperation(prc *Process, op *OpCode) (err error) {
	argCount := op.OpData.(int)

	// Build array of arguments pulling from stack
	args := make([]types.DataType, argCount)
	for idx := argCount - 1; idx >= 0; idx-- {
		args[idx], err = prc.pop()
		if err != nil {
			return err
		}
	}

	// Pull the function reference to be called
	fnVal, err := prc.pop()
	if err != nil {
		return err
	}

	switch fn := fnVal.(type) {
	case *types.NativeFunction:
		// Native functions are just a direct Go call
		res, callErr := fn.Fn(args)
		if err != nil {
			return callErr
		}
		if res == nil {
			err = prc.push(types.Undefined)
		} else {
			err = prc.push(res)
		}
		return

	case *ScriptFunction:
		// Store the current execution context into the call frame list
		frame := &CallFrame{
			previous: prc.callStack,
			body:     prc.body,
			pc:       prc.pc,
			sp:       prc.sp,
			locals:   prc.locals,
			cells:    prc.cells,
			closure:  prc.closure,
		}
		prc.callStack = frame

		// Set up new execution context for the function
		prc.body = fn.Body
		prc.pc = -1

		// Set up closure/cell data for variable capture (on demand)
		prc.closure = fn.Closure
		prc.cells = nil

		// Initialize the local variable storage for the function
		if prc.body.VarCount > 0 {
			prc.locals = make([]types.DataType, prc.body.VarCount)
			for idx := 0; idx < prc.body.VarCount; idx++ {
				prc.locals[idx] = types.Undefined
			}
		} else {
			prc.locals = nil
		}

		// Populate the parameter variable values
		for idx := 0; idx < len(fn.ParamNames) && idx < len(args); idx++ {
			prc.locals[idx] = args[idx]
		}

		return nil

	default:
		return fmt.Errorf("TypeError: %v is not a function", fnVal)
	}
}

func ReturnOperation(prc *Process, op *OpCode) (err error) {
	var retVal types.DataType
	if op.OpData.(bool) {
		retVal, err = prc.pop()
		if err != nil {
			return err
		}
	} else {
		retVal = types.Undefined
	}

	// If call stack is empty, return from the main script (non-compliant)
	if prc.callStack == nil {
		// Empty stack, push return value, jump to end of script execution
		prc.sp = 0
		err = prc.push(retVal)
		prc.pc = len(prc.body.Code)
		return
	}

	// Restore previous execution context from the call stack
	frame := prc.callStack
	prc.callStack = frame.previous
	prc.body = frame.body
	prc.pc = frame.pc
	prc.locals = frame.locals
	prc.cells = frame.cells
	prc.closure = frame.closure
	prc.sp = frame.sp

	// Return value can now be pushed onto the prior call stack
	err = prc.push(retVal)
	return
}

func PushFunctionOperation(prc *Process, op *OpCode) (err error) {
	fnTemplate := op.OpData.(types.DataType)
	sfn, ok := (fnTemplate).(*ScriptFunction)
	if !ok {
		// Not a script function (native), just push the data value
		err = prc.push(fnTemplate)
		return
	}

	// If the function has no closures, can use it directly
	if len(sfn.Captures) == 0 {
		err = prc.push(fnTemplate)
		return
	}

	// Otherwise, create a function from the template with capture cells
	closure := make([]*Cell, len(sfn.Captures))
	for idx, cap := range sfn.Captures {
		if cap.IsCapture {
			if prc.closure != nil && cap.SlotIndex < len(prc.closure) {
				// Capture from the current closure (already attached to a cell)
				closure[idx] = prc.closure[cap.SlotIndex]
			} else {
				// Create a new cell, value is undefined
				closure[idx] = &Cell{Value: &types.Undefined}
			}
		} else {
			// Capture from locals, init cell set if needed and capture value
			if prc.cells == nil {
				prc.cells = make([]*Cell, len(prc.locals))
			}
			for len(prc.cells) <= cap.SlotIndex {
				// Fill intermediate slots with nil if required
				prc.cells = append(prc.cells, nil)
			}
			if prc.cells[cap.SlotIndex] == nil {
				var val types.DataType
				if cap.SlotIndex < len(prc.locals) {
					val = prc.locals[cap.SlotIndex]
				} else {
					val = types.Undefined
				}
				prc.cells[cap.SlotIndex] = &Cell{Value: &val}
			}
			closure[idx] = prc.cells[cap.SlotIndex]
		}
	}

	// Clone the source function with the new capture instance
	fnCopy := &ScriptFunction{
		Name:       sfn.Name,
		ParamNames: sfn.ParamNames,
		Body:       sfn.Body,
		VarCount:   sfn.VarCount,
		Captures:   sfn.Captures,
		Closure:    closure,
	}

	// And that is the value we push onto the stack
	fnVal := types.DataType(fnCopy)
	err = prc.push(fnVal)
	return
}

func LoadCaptureOperation(prc *Process, op *OpCode) (err error) {
	capIdx := op.OpData.(int)
	if prc.closure == nil || capIdx >= len(prc.closure) {
		err = prc.push(types.Undefined)
		return
	}

	// Retrieve the value from the closure cell instance
	cell := prc.closure[capIdx]
	if cell == nil || cell.Value == nil {
		err = prc.push(types.Undefined)
		return
	}
	err = prc.push(*cell.Value)
	return
}

func StoreCaptureOperation(prc *Process, op *OpCode) (err error) {
	capIdx := op.OpData.(int)
	val, err := prc.pop()
	if err != nil {
		return err
	}
	if prc.closure != nil && capIdx < len(prc.closure) {
		// Store the value into the closure cell instance (create if first)
		if prc.closure[capIdx] == nil {
			prc.closure[capIdx] = &Cell{}
		}
		*prc.closure[capIdx].Value = val
	}

	return nil
}

func StoreCaptureKeepOperation(prc *Process, op *OpCode) (err error) {
	capIdx := op.OpData.(int)
	val, err := prc.peek()
	if err != nil {
		return err
	}
	if prc.closure != nil && capIdx < len(prc.closure) {
		// Store the value into the closure cell instance (create if first)
		if prc.closure[capIdx] == nil {
			prc.closure[capIdx] = &Cell{}
		}
		*prc.closure[capIdx].Value = val
	}
	return nil
}

func ForInKeysOperation(prc *Process, op *OpCode) (err error) {
	obj, err := prc.pop()
	if err != nil {
		return err
	}

	var keys []string
	switch tgt := obj.(type) {
	case *types.ObjectType:
		keys = make([]string, 0, len(tgt.Properties))
		for key := range tgt.Properties {
			keys = append(keys, key)
		}
	case *types.ArrayType:
		keys = make([]string, len(tgt.Elements))
		for idx := range tgt.Elements {
			keys[idx] = types.ToString(types.IntegerType(idx))
		}
	default:
		keys = []string{}
	}

	// Push keys (type) array and the starting index of zero
	keysArr := types.NewArray(len(keys))
	for idx, key := range keys {
		keysArr.Elements[idx] = types.StringType(key)
	}
	err = prc.push(keysArr)
	if err != nil {
		return err
	}
	return prc.push(types.IntegerType(0))
}

func ForInHasMoreOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	keys, err := prc.peek()
	if err != nil {
		return err
	}

	arr, ok := keys.(*types.ArrayType)
	if !ok {
		// Uh oh, put the index back and exit the loop
		prc.push(index)
		return prc.push(types.BooleanType(false))
	}

	idx := 0
	if ix, ok := index.(types.IntegerType); ok {
		idx = int(ix)
	}

	// Put the index back and then the boolean hasMore result
	prc.push(index)
	hasMore := idx < len(arr.Elements)
	return prc.push(types.BooleanType(hasMore))
}

func ForInNextOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)

	index, err := prc.pop()
	if err != nil {
		return err
	}
	keys, err := prc.peek()
	if err != nil {
		return err
	}

	arr, ok := keys.(*types.ArrayType)
	if !ok {
        // It's going sideways...
		return prc.push(index)
	}

	idx := 0
	if ix, ok := index.(types.IntegerType); ok {
		idx = int(ix)
	}

	// Store array entry to variable, increment iterator index
	if idx < len(arr.Elements) {
		prc.locals[slotIndex] = arr.Elements[idx]
	}
	return prc.push(types.IntegerType(idx + 1))
}

func ForInCleanupOperation(prc *Process, op *OpCode) (err error) {
    // Just discard the index and original keys array
	prc.pop()
	prc.pop()
	return nil
}

func ForOfIteratorOperation(prc *Process, op *OpCode) (err error) {
    // So much easier than above, already have the iterator instance
	return prc.push(types.IntegerType(0))
}

func ForOfHasMoreOperation(prc *Process, op *OpCode) (err error) {
	index, err := prc.pop()
	if err != nil {
		return err
	}
	iterable, err := prc.peek()
	if err != nil {
		return err
	}

	idx := 0
	if ix, ok := index.(types.IntegerType); ok {
		idx = int(ix)
	}

	// Put the index back (hmmm, peek + 1?)
	prc.push(index)

	// Determine the hasMore condition based on the iterable type
	var hasMore bool
	switch it := iterable.(type) {
	case *types.ArrayType:
		hasMore = idx < len(it.Elements)
	case types.StringType:
		hasMore = idx < len(string(it))
	default:
		hasMore = false
	}

	return prc.push(types.BooleanType(hasMore))
}

func ForOfNextOperation(prc *Process, op *OpCode) (err error) {
	slotIndex := op.OpData.(int)

	index, err := prc.pop()
	if err != nil {
		return err
	}
	iterable, err := prc.peek()
	if err != nil {
		return err
	}

	idx := 0
	if ix, ok := index.(types.IntegerType); ok {
		idx = int(ix)
	}

	// Store appropriate result based on iterable type
	switch it := iterable.(type) {
	case *types.ArrayType:
		if idx < len(it.Elements) {
			prc.locals[slotIndex] = it.Elements[idx]
		}
	case types.StringType:
		if idx < len(string(it)) {
			prc.locals[slotIndex] = types.StringType(string(it)[idx : idx+1])
		}
	}

	// And increment the iterator index
	return prc.push(types.IntegerType(idx + 1))
}

func ForOfCleanupOperation(prc *Process, op *OpCode) (err error) {
    // Just discard the index and the iterator instance
	prc.pop()
	prc.pop()
	return nil
}
