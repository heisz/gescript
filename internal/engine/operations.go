/*
 * Definition and implementations for operations in the script engine.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package engine

import (
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
	left, err := prc.pop()
	if err != nil {
		return err
	}
	right, err := prc.pop()
	if err != nil {
		return err
	}

	// Really big pair of switch statements to handle all of the mixes
	var res types.DataType = types.UndefinedType{}
	switch (*left).(type) {
	case types.IntegerType:
		switch (*right).(type) {
		case types.IntegerType:
			res = types.IntegerType((*left).Native().(int64) +
				(*right).Native().(int64))
		}
	}

	val := types.DataType(res)
	err = prc.push(&val)
	return
}
