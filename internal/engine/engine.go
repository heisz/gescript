/*
 * Primary structures/implementations for the execution of opcode bodies.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package engine

import (
	"github.com/heisz/gescript/types"
)

// Like that other project, process is the execution context of the opcodes
// (in a strict bytecode standard this would be the virtual machine)
type Process struct {
	// Current code body being executed and associated program counter
	body *Function
	pc   int

	// It's a stack-based machine, here is the stack (sp points to next slot)
	stack []*types.DataType
	sp    int
}

// Any stack needs push, peek and pop, push supporting dynamic resizing
// Note that many operations perform direct manipulation and that's ok
func (prc *Process) push(val *types.DataType) (err error) {
	// TODO - resize and overflow!
	prc.stack[prc.sp] = val
	prc.sp++

	return nil
}
func (prc *Process) peek() (val *types.DataType, err error) {
	// TODO - stack underflow
	return prc.stack[prc.sp-1], nil
}
func (prc *Process) pop() (val *types.DataType, err error) {
	// TODO - stack underflow
	prc.sp--
	return prc.stack[prc.sp], nil
}

// The fundamental model in this implementation is that every set of code is
// a function.  For all of the functions that's obvious and the uncontained
// code is compiled into an anonymous function instance.
type Function struct {
	Name string
	Code []*OpCode
}

// Execution 'loop' to run the given function in the associated process
func (body *Function) Exec(prc *Process) (ret types.DataType, err error) {
	// For now, just ram this in
	prc.body = body
	prc.pc = 0
	prc.stack = make([]*types.DataType, 16)
	prc.sp = 0

	// Loop until we run out of runway
	for {
		pc := prc.pc
		if pc < 0 || pc >= len(prc.body.Code) {
			// The first shouldn't happen but just in case...
			break
		}
		op := prc.body.Code[pc]
		err := op.ExecFn(prc, op)
		if err != nil {
			return nil, err
		}
		prc.pc++
	}

	// The last item on the stack is the outer return value
	if prc.sp > 0 {
		return *prc.stack[prc.sp-1], nil
	}
	return types.Undefined, nil
}
