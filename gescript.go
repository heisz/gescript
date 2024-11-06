/*
 * Primary entry point to the gescript module (parse and execute).
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package gescript

import (
	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/internal/parser"
	"github.com/heisz/gescript/types"
)

// Exposed container for a parsed script instance
type Script struct {
	body *engine.Function
}

// Parse the source script into an executable Script instance (or error)
func Parse(source string) (prg *Script, err error) {
	prsBody, err := parser.Parse(source)
	if err != nil {
		return
	}

	return &Script{
		body: prsBody,
	}, nil
}

// For the parsed script instance, execute it
func (prg *Script) Run() (retval types.DataType, err error) {
	ctx := &engine.Process{}
	return prg.body.Exec(ctx)
}

// Convenience method to parse and execute the script source in one shot
func Run(source string) (retval types.DataType, err error) {
	script, err := Parse(source)
	if err != nil {
		return
	}
	return script.Run()
}
