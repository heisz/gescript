/*
 * Test methods for the parser implementation (expression-specific)
 *
 * Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package parser

import (
	"reflect"
	"testing"

	"github.com/heisz/gescript/internal/engine"
	"github.com/heisz/gescript/internal/native"
)

func TestEmptyWithComments(tst *testing.T) {
	prg, err := Parse(
		"  // This is a single line comment\r\n" +
			"  /* This is a single line multi-line comment */\n" +
			"\t/*\n" +
			"   * This is a multi-line comment...\r\n" +
			"   */\r\n" +
			"  // Note that there are tabs in here for whitespace scanning...")
	if err != nil {
		tst.Fatalf("Unexpected error parsing comment only file: %v", err)
	}

	// Should be an empty function
	if len(prg.Code) != 0 {
		tst.Fatalf("Unexpected code output from empty program")
	}
}

// Very simple case to check basic expression parsing and execution
func TestBasicExression(tst *testing.T) {
	prg, err := Parse("1 + 2")
	if err != nil {
		tst.Fatalf("Unexpected error parsing basic expression: %v", err)
	}

	if len(prg.Code) != 3 {
		tst.Fatalf("Incorrect operation list for basic expression")
	}
	if (reflect.ValueOf(prg.Code[0].ExecFn).Pointer() !=
		reflect.ValueOf(engine.PushLiteralValue).Pointer()) ||
		(reflect.ValueOf(prg.Code[1].ExecFn).Pointer() !=
			reflect.ValueOf(engine.PushLiteralValue).Pointer()) ||
		(reflect.ValueOf(prg.Code[2].ExecFn).Pointer() !=
			reflect.ValueOf(engine.AdditionOperation).Pointer()) {
		tst.Fatalf("Incorrect operation set for basic expression")
	}

	prc := engine.NewProcess(3, nil, nil, nil)
	res, er := prg.Exec(prc)
	if er != nil {
		tst.Fatalf("Unexpected error running basic expression: %v", err)
	}
	if res.Native().(int64) != 3 {
		tst.Fatalf("Unexpected result from basic expression: %v", res)
	}
}

// Helper function to run an expression and check the expression result
func checkExpr(tst *testing.T, expr string, expected interface{}) {
	prg, errs := Parse(expr)
	if errs != nil && len(errs) > 0 {
		tst.Fatalf("Unexpected error parsing '%s': %v", expr, errs)
	}

	prc := engine.NewProcess(256, nil, nil, native.NativeConstructors)
	res, runErr := prg.Exec(prc)
	if runErr != nil {
		tst.Fatalf("Unexpected error running '%s': %v", expr, runErr)
	}

	actual := res.Native()
	if actual != expected {
		tst.Fatalf("Expression '%s': expected %v (%T), got %v (%T)",
			expr, expected, expected, actual, actual)
	}
}

/* A heck of a lot of different cases to test expression operations */

func TestNullUndefinedLiterals(tst *testing.T) {
	// Both null and undefined return nil as the native value
	checkExpr(tst, "null", nil)
	checkExpr(tst, "undefined", nil)

	// Per specification, equality treats them as the same
	checkExpr(tst, "null == null", true)
	checkExpr(tst, "null != null", false)
	checkExpr(tst, "undefined == undefined", true)
	checkExpr(tst, "undefined != undefined", false)
	checkExpr(tst, "null == undefined", true)
	checkExpr(tst, "null != undefined", false)
	checkExpr(tst, "undefined == null", true)
	checkExpr(tst, "undefined != null", false)

	// While strict equality treats them as different
	checkExpr(tst, "null === null", true)
	checkExpr(tst, "null !== null", false)
	checkExpr(tst, "undefined === undefined", true)
	checkExpr(tst, "undefined !== undefined", false)
	checkExpr(tst, "null === undefined", false)
	checkExpr(tst, "null !== undefined", true)
	checkExpr(tst, "undefined === null", false)
	checkExpr(tst, "undefined !== null", true)

	// Note - not-not case isn't quite what you'd think (see top)
	checkExpr(tst, "!null", true)
	checkExpr(tst, "!undefined", true)
}

func TestIntegerExpressions(tst *testing.T) {
	checkExpr(tst, "1 + 3", int64(4))
	checkExpr(tst, "100 + 400", int64(500))
	checkExpr(tst, "-5 + 12", int64(7))

	checkExpr(tst, "10 - 4", int64(6))
	checkExpr(tst, "5 - 10", int64(-5))
	checkExpr(tst, "0 - 42", int64(-42))

	checkExpr(tst, "3 * 6", int64(18))
	checkExpr(tst, "12 * 0", int64(0))
	checkExpr(tst, "-3 * 4", int64(-12))

	// Note that division always returns float
	checkExpr(tst, "12 / 2", float64(6.0))
	checkExpr(tst, "7 / 2", float64(3.5))
	checkExpr(tst, "100 / 4", float64(25.0))

	checkExpr(tst, "7 % 3", int64(1))
	checkExpr(tst, "25 % 5", int64(0))
	checkExpr(tst, "-7 % 3", int64(-1))
	checkExpr(tst, "7 % -3", int64(1))
}

func TestFloatExpressions(tst *testing.T) {
	checkExpr(tst, "1.5 + 2.5", float64(4.0))
	checkExpr(tst, "0.1 + 0.3", float64(0.4))

	checkExpr(tst, "6.6 - 3.3", float64(3.3))
	checkExpr(tst, "1.0 - 0.4", float64(0.6))

	checkExpr(tst, "2.5 * 4.0", float64(10))
	checkExpr(tst, "0.5 * 0.5", float64(0.25))

	checkExpr(tst, "7.5 / 2.5", float64(3.0))
	checkExpr(tst, "1.0 / 4.0", float64(0.25))

	checkExpr(tst, "5.5 % 2.5", float64(0.5))
	checkExpr(tst, "10.0 % 3.0", float64(1.0))
	checkExpr(tst, "-7.5 % 4.0", float64(-3.5))
	checkExpr(tst, "7.5 % -4.0", float64(3.5))
}

func TestMixedTypeExpressions(tst *testing.T) {
	checkExpr(tst, "1 + 2.5", float64(3.5))
	checkExpr(tst, "2.5 + 1", float64(3.5))

	checkExpr(tst, "10 - 2.5", float64(7.5))
	checkExpr(tst, "10.5 - 2", float64(8.5))

	checkExpr(tst, "4 * 2.5", float64(10.0))
	checkExpr(tst, "2.5 * 4", float64(10.0))

	checkExpr(tst, "7.5 % 4", float64(3.5))
	checkExpr(tst, "4 % 2.5", float64(1.5))
	checkExpr(tst, "-7.5 % 4", float64(-3.5))
	checkExpr(tst, "7.5 % -4", float64(3.5))
}

// TODO - need error tests for invalid combinations

func TestUnaryExpressions(tst *testing.T) {
	checkExpr(tst, "-5", int64(-5))
	checkExpr(tst, "-(-6)", int64(6))
	checkExpr(tst, "-(3 + 9)", int64(-12))
	checkExpr(tst, "-true", int64(-1))
	checkExpr(tst, "-false", int64(0))

	checkExpr(tst, "+12", int64(12))
	checkExpr(tst, "+(1 + 5)", int64(6))
	checkExpr(tst, "+5.2", float64(5.2))
	checkExpr(tst, "+true", int64(1))
	checkExpr(tst, "+false", int64(0))

	checkExpr(tst, "!0", true)
	checkExpr(tst, "!1", false)
	checkExpr(tst, "!false", true)
	checkExpr(tst, "!true", false)
	checkExpr(tst, "!!true", true)
	checkExpr(tst, "!!1", true)

	checkExpr(tst, "~0", int64(-1))
	checkExpr(tst, "~1", int64(-2))
	checkExpr(tst, "~1.0", int64(-2))
	checkExpr(tst, "~~5", int64(5))
}

func TestIncrDecrExpression(tst *testing.T) {
	checkExpr(tst, "var x = 5; ++x", int64(6))
	checkExpr(tst, "var x = 12; ++x; x", int64(13))
	checkExpr(tst, "var x = 0; ++x; ++x", int64(2))
	checkExpr(tst, "var x = 0; ++x; ++x; x", int64(2))

	checkExpr(tst, "var x = 5; --x", int64(4))
	checkExpr(tst, "var x = 12; --x; x", int64(11))
	checkExpr(tst, "var x = 3; --x; --x", int64(1))
	checkExpr(tst, "var x = 3; --x; --x; x", int64(1))

	checkExpr(tst, "var x = 5; x++", int64(5))
	checkExpr(tst, "var x = 12; x++; x", int64(13))
	checkExpr(tst, "var x = 0; x++; x++", int64(1))
	checkExpr(tst, "var x = 0; x++; x++; x", int64(2))

	checkExpr(tst, "var x = 5; x--", int64(5))
	checkExpr(tst, "var x = 5; x--; x", int64(4))
	checkExpr(tst, "var x = 3; x--; x--", int64(2))
	checkExpr(tst, "var x = 3; x--; x--; x", int64(1))

	checkExpr(tst, "var x = 1.3; ++x", float64(2.3))
	checkExpr(tst, "var x = 1.3; x++; x", float64(2.3))
	checkExpr(tst, "var x = 12.0; --x", float64(11.0))
	checkExpr(tst, "var x = 12.0; x--; x", float64(11.0))

	// This is really just a taste of the possible combinations
	checkExpr(tst, "var x = 5; 10 + ++x", int64(16))
	checkExpr(tst, "var x = 12; 10 + x++", int64(22))
	checkExpr(tst, "var x = 5; ++x + x++", int64(12))

	// Similar style of tests for array element and object property actions
	checkExpr(tst, "var arr = [5, 10, 15]; ++arr[0]", int64(6))
	checkExpr(tst, "var arr = [5, 10, 15]; ++arr[0]; arr[0]", int64(6))
	checkExpr(tst, "var arr = [5, 10, 15]; arr[1]++", int64(10))
	checkExpr(tst, "var arr = [5, 10, 15]; arr[1]++; arr[1]", int64(11))
	checkExpr(tst, "var arr = [5, 10, 15]; --arr[2]", int64(14))
	checkExpr(tst, "var arr = [5, 10, 15]; arr[2]--", int64(15))
	checkExpr(tst, "var arr = [5, 10, 15]; arr[2]--; arr[2]", int64(14))

	checkExpr(tst, "var obj = {x: 10}; ++obj.x", int64(11))
	checkExpr(tst, "var obj = {x: 10}; ++obj.x; obj.x", int64(11))
	checkExpr(tst, "var obj = {x: 10}; obj.x++", int64(10))
	checkExpr(tst, "var obj = {x: 10}; obj.x++; obj.x", int64(11))
	checkExpr(tst, "var obj = {x: 10}; --obj.x", int64(9))
	checkExpr(tst, "var obj = {x: 10}; obj.x--", int64(10))
	checkExpr(tst, "var obj = {x: 10}; obj.x--; obj.x", int64(9))

	checkExpr(tst, "var arr = [{v: 5}]; ++arr[0].v", int64(6))
	checkExpr(tst, "var arr = [{v: 5}]; arr[0].v++; arr[0].v", int64(6))
	checkExpr(tst, "var obj = {arr: [10]}; ++obj.arr[0]", int64(11))
	checkExpr(tst, "var obj = {arr: [10]}; obj.arr[0]++; obj.arr[0]", int64(11))

	checkExpr(tst, "var arr = [5]; 10 + ++arr[0]", int64(16))
	checkExpr(tst, "var arr = [5]; 10 + arr[0]++", int64(15))
	checkExpr(tst, "var obj = {x: 5}; 10 + ++obj.x", int64(16))
	checkExpr(tst, "var obj = {x: 5}; 10 + obj.x++", int64(15))
}

func TestShiftExpressions(tst *testing.T) {
	checkExpr(tst, "1 << 4", int64(16))
	checkExpr(tst, "3 << 2", int64(12))
	checkExpr(tst, "1 << 0", int64(1))

	checkExpr(tst, "16 >> 2", int64(4))
	checkExpr(tst, "100 >> 3", int64(12))
	checkExpr(tst, "-8 >> 2", int64(-2))

	checkExpr(tst, "17 >>> 2", int64(4))
	checkExpr(tst, "-1 >>> 0", int64(4294967295))

	// Float cases don't convert to float in this case (lazy, do l/r together)
	checkExpr(tst, "3.0 << 2.0", int64(12))
	checkExpr(tst, "100.0 >> 3.0", int64(12))
	checkExpr(tst, "17.0 >>> 2.0", int64(4))
}

func TestComparisonExpressions(tst *testing.T) {
	checkExpr(tst, "1 < 2", true)
	checkExpr(tst, "2 < 1", false)
	checkExpr(tst, "1 < 1", false)
	checkExpr(tst, "1.5 < 2", true)
	checkExpr(tst, "1 < 1.5", true)
	checkExpr(tst, "2.0 < 1.5", false)

	checkExpr(tst, "1 > 2", false)
	checkExpr(tst, "2 > 1", true)
	checkExpr(tst, "1 > 1", false)
	checkExpr(tst, "1.5 > 2", false)
	checkExpr(tst, "1 > 1.5", false)
	checkExpr(tst, "2.0 > 1.5", true)

	checkExpr(tst, "1 <= 2", true)
	checkExpr(tst, "2 <= 1", false)
	checkExpr(tst, "1 <= 1", true)
	checkExpr(tst, "1.5 <= 2", true)
	checkExpr(tst, "1 <= 1.5", true)
	checkExpr(tst, "2.0 <= 1.5", false)

	checkExpr(tst, "1 >= 2", false)
	checkExpr(tst, "2 >= 1", true)
	checkExpr(tst, "1 >= 1", true)
	checkExpr(tst, "1.5 >= 2", false)
	checkExpr(tst, "1 >= 1.5", false)
	checkExpr(tst, "2.0 >= 1.5", true)

	checkExpr(tst, `"abc" < "abd"`, true)
	checkExpr(tst, `"abc" < "abc"`, false)
	checkExpr(tst, `"abc" <= "abc"`, true)
	checkExpr(tst, `"z" > "a"`, true)
	checkExpr(tst, `"abc" >= "abc"`, true)
}

func TestEqualityOperations(tst *testing.T) {
	checkExpr(tst, "1 == 1", true)
	checkExpr(tst, "1 == 12", false)
	checkExpr(tst, "1 == 1.0", true)
	checkExpr(tst, "1.0 == 2", false)
	checkExpr(tst, "1.0 == 1.0", true)

	checkExpr(tst, "1 != 12", true)
	checkExpr(tst, "1 != 1", false)
	checkExpr(tst, "1 != 2.0", true)
	checkExpr(tst, "2.0 != 1", true)
	checkExpr(tst, "1.0 != 1.0", false)
	checkExpr(tst, "true == true", true)
	checkExpr(tst, "false == false", true)
	checkExpr(tst, "true != true", false)
	checkExpr(tst, "false != false", false)

	checkExpr(tst, "1 === 1", true)
	checkExpr(tst, "1 === 12", false)
	// Goofed this, in ECMA all numbers are numbers
	checkExpr(tst, "1.0 === 1", true)
	checkExpr(tst, "1 === 1.0", true)
	checkExpr(tst, "2.0 === 2.0", true)
	checkExpr(tst, "true === true", true)
	checkExpr(tst, "false === false", true)

	checkExpr(tst, "1 !== 12", true)
	checkExpr(tst, "1 !== 1", false)
	checkExpr(tst, "1.0 !== 1.0", false)
	checkExpr(tst, "true !== false", true)

	checkExpr(tst, `"hello" == "hello"`, true)
	checkExpr(tst, `"hello" != "hello"`, false)
	checkExpr(tst, `"hello" == "world"`, false)
	checkExpr(tst, `"hello" != "world"`, true)
	checkExpr(tst, `"hello" === "hello"`, true)
	checkExpr(tst, `"hello" !== "world"`, true)

	// Note that null/undefined tests are in the raw literal test above
}

func TestBitwiseExpressions(tst *testing.T) {
	checkExpr(tst, "5 & 3", int64(1))
	checkExpr(tst, "15 & 7", int64(7))
	checkExpr(tst, "12 & 10", int64(8))
	checkExpr(tst, "12.0 & 10.0", int64(8))

	checkExpr(tst, "5 | 3", int64(7))
	checkExpr(tst, "8 | 4", int64(12))
	checkExpr(tst, "11 | 5", int64(15))
	checkExpr(tst, "11.2 | 5.1", int64(15))

	checkExpr(tst, "5 ^ 3", int64(6))
	checkExpr(tst, "15 ^ 15", int64(0))
	checkExpr(tst, "10 ^ 5", int64(15))
	checkExpr(tst, "10.0 ^ 5.0", int64(15))
}

// For now, just test outcome, need other things to test short-circuit
func TestLogicalExpressions(tst *testing.T) {
	checkExpr(tst, "1 && 2", int64(2))
	checkExpr(tst, "0 && 12", int64(0))
	checkExpr(tst, "true && false", false)
	checkExpr(tst, "true && true", true)
	checkExpr(tst, "false && true", false)

	checkExpr(tst, "1 || 12", int64(1))
	checkExpr(tst, "0 || 12", int64(12))
	checkExpr(tst, "false || true", true)
	checkExpr(tst, "true || false", true)
	checkExpr(tst, "false || false", false)

	checkExpr(tst, "1 && 12 && 3", int64(3))
	checkExpr(tst, "11 && 0 && 3", int64(0))
	checkExpr(tst, "0 || 0 || 3", int64(3))
	checkExpr(tst, "1 || 12 || 3", int64(1))
}

// Ditto for ternary (although jumps are clearly implied)
func TestTernaryExpressions(tst *testing.T) {
	checkExpr(tst, "true ? 1 : 12", int64(1))
	checkExpr(tst, "false ? 1 : 12", int64(12))

	checkExpr(tst, "1 ? 12 : 20", int64(12))
	checkExpr(tst, "0 ? 12 : 20", int64(20))

	checkExpr(tst, "true ? (false ? 1 : 2) : 3", int64(2))
	checkExpr(tst, "false ? 1 : (true ? 2 : 3)", int64(2))

	checkExpr(tst, "(1 < 12) ? 100 : 200", int64(100))
	checkExpr(tst, "(12 < 1) ? 100 : 200", int64(200))
	checkExpr(tst, "(5 > 3) ? (10 + 5) : (10 - 5)", int64(15))
}

func TestExpressionPrecedence(tst *testing.T) {
	checkExpr(tst, "1 + 12 * 3", int64(37))
	checkExpr(tst, "2 * 3 + 12", int64(18))
	checkExpr(tst, "10 - 2 * 3", int64(4))
	checkExpr(tst, "2 + 3 * 4 + 5", int64(19))

	checkExpr(tst, "10 + 20 / 4", float64(15))
	checkExpr(tst, "20 / 4 + 10", float64(15))

	checkExpr(tst, "10 + 7 % 3", int64(11))

	checkExpr(tst, "1 + 2 << 3", int64(24))
	checkExpr(tst, "1 << 3 + 1", int64(16))

	checkExpr(tst, "1 << 2 < 10", true)

	checkExpr(tst, "1 < 12 == true", true)
	checkExpr(tst, "12 > 1 == true", true)

	checkExpr(tst, "(3 & 1) == 1", true)

	checkExpr(tst, "7 & 3 | 8", int64(11))
	checkExpr(tst, "7 | 3 & 8", int64(7))
	checkExpr(tst, "5 ^ 3 & 7", int64(6))
	checkExpr(tst, "5 & 3 ^ 7", int64(6))

	checkExpr(tst, "0 || 12 && 2", int64(2))
	checkExpr(tst, "12 && 0 || 3", int64(3))

	checkExpr(tst, "1 + 1 ? 10 : 20", int64(10))
	checkExpr(tst, "1 < 2 ? 3 + 4 : 5 + 6", int64(7))
}

func TestExpressionGrouping(tst *testing.T) {
	checkExpr(tst, "(1 + 2) * 3", int64(9))
	checkExpr(tst, "3 * (1 + 2)", int64(9))
	checkExpr(tst, "(10 - 2) * 3", int64(24))

	checkExpr(tst, "(10 + 20) / 5", float64(6))
	checkExpr(tst, "100 / (4 + 1)", float64(20))

	checkExpr(tst, "((1 + 2) * 3) + 4", int64(13))
	checkExpr(tst, "((2 + 3) * (4 + 5))", int64(45))
	checkExpr(tst, "(1 + (2 * (3 + 4)))", int64(15))

	checkExpr(tst, "-(1 + 2)", int64(-3))
	checkExpr(tst, "(-1) + (-2)", int64(-3))
	checkExpr(tst, "!(1 > 2)", true)

	checkExpr(tst, "(6 + 6) == 12", true)
	checkExpr(tst, "(5 - 3) < (12 + 12)", true)

	checkExpr(tst, "(1 && 0) || 5", int64(5))
	checkExpr(tst, "1 && (0 || 5)", int64(5))
}

// This one could technically go on forever...
func TestMixedExpressions(tst *testing.T) {
	checkExpr(tst, "1 + 2 * 3 - 4 / 2", float64(5))

	checkExpr(tst, "(1 < 12) && (3 < 4)", true)
	checkExpr(tst, "(1 > 12) || (3 < 4)", true)
	checkExpr(tst, "(1 > 12) && (3 < 4)", false)

	checkExpr(tst, "1 + (true ? 2 : 3)", int64(3))
	checkExpr(tst, "(1 > 0 ? 10 : 20) * 2", int64(20))

	checkExpr(tst, "(3 + 5) & 7", int64(0))
	checkExpr(tst, "(3 | 5) + 1", int64(8))
	checkExpr(tst, "1 << (2 + 2)", int64(16))

	checkExpr(tst, "((1 + 2) * 3 > 5) ? (10 & 7) : (10 | 7)", int64(2))
}

// Since it's tightly expression related, keep variable tests in here

func TestVariableExpressions(tst *testing.T) {
	checkExpr(tst, "var x = 5; x", int64(5))
	checkExpr(tst, "var x = 10; x + 5", int64(15))
	checkExpr(tst, "var x = 3; var y = 4; x + y", int64(7))
	checkExpr(tst, "var x = 2; var y = 3; x * y + 1", int64(7))

	checkExpr(tst, "let x = 5; x", int64(5))
	checkExpr(tst, "let x = 10; x + 5", int64(15))
	checkExpr(tst, "let x = 3; let y = 4; x + y", int64(7))
	checkExpr(tst, "let x = 2; let y = 3; x * y + 1", int64(7))

	checkExpr(tst, "const x = 5; x", int64(5))
	checkExpr(tst, "const x = 10; x + 5", int64(15))
	checkExpr(tst, "const x = 3; const y = 4; x + y", int64(7))
	checkExpr(tst, "const x = 2; const y = 3; x * y + 1", int64(7))

	checkExpr(tst, "var x = 5; x = 10; x", int64(10))
	checkExpr(tst, "var x = 1; x = x + 1; x", int64(2))
	checkExpr(tst, "let x = 5; x = 10; x", int64(10))
	checkExpr(tst, "var x = 1; var y = 2; x = y = 3; x + y", int64(6))

	checkExpr(tst, "var x = 5; (x = 10)", int64(10))
	checkExpr(tst, "var x = 1; (x = 2) + 3", int64(5))

	checkExpr(tst, "var x = 5; x > 3", true)
	checkExpr(tst, "var x = 5; var y = 10; x < y", true)
	checkExpr(tst, "var x = true; x && false", false)
	checkExpr(tst, "var x = 0; x || 5", int64(5))
	checkExpr(tst, "var x = 1; x ? 10 : 20", int64(10))

	checkExpr(tst, "var x = 1; { var x = 2 }; x", int64(2))
	checkExpr(tst, "var x = 1; { var y = 2 }; y + 1", int64(3))
	checkExpr(tst, "var x = 1; { { var x = 3 } }; x", int64(3))
	checkExpr(tst, "var x = 1; { let y = 2 }; x", int64(1))
}

func TestArrayExpressions(tst *testing.T) {
	// No direct test of just array creation, use length member for basics
	checkExpr(tst, "[1, 2, 3].length", int64(3))
	checkExpr(tst, "[].length", int64(0))
	checkExpr(tst, "[1].length", int64(1))
	checkExpr(tst, "[1, 2, 3, 4, 5].length", int64(5))

	checkExpr(tst, "[1, 2, 3][0]", int64(1))
	checkExpr(tst, "[1, 2, 3][1]", int64(2))
	checkExpr(tst, "[1, 2, 3][2]", int64(3))

	checkExpr(tst, "[10, 20, 30][0] + [10, 20, 30][2]", int64(40))

	checkExpr(tst, "var arr = [5, 12, 15]; arr[0]", int64(5))
	checkExpr(tst, "var arr = [5, 12, 15]; arr[1]", int64(12))
	checkExpr(tst, "var arr = [5, 12, 15]; arr[2]", int64(15))

	checkExpr(tst, "var arr = [100, 200, 300]; arr[1 + 1]", int64(300))
	checkExpr(tst, "var idx = 1; [10, 20, 30][idx]", int64(20))

	checkExpr(tst, "var arr = [1, 2, 3]; arr[0] = 12; arr[0]", int64(12))
	checkExpr(tst, "var arr = [1, 2, 3]; arr[1] = 20; arr[1]", int64(20))
	checkExpr(tst, "var arr = [1, 2, 3]; arr[2] = 30; arr[2]", int64(30))
	checkExpr(tst, "var arr = [1, 2, 3]; arr[0] = 100", int64(100))

	checkExpr(tst, `var arr = [0, 0, 0];
                    arr[0] = 1; arr[1] = 12; arr[2] = 3;
                    arr[0] + arr[1] + arr[2]`, int64(16))
	checkExpr(tst, `var arr = [1, 2, 3], idx = 1;
                    arr[idx] = 50; arr[1]`, int64(50))
	checkExpr(tst, `var arr = [1, 2, 3];
                    arr[1 + 1] = 99; arr[2]`, int64(99))

	checkExpr(tst, "var arr = [1]; arr[5] = 100; arr.length", int64(6))
	checkExpr(tst, "var arr = [1]; arr[5] = 100; arr[2]", nil)
	checkExpr(tst, "var arr = [1]; arr[2]", nil)
}

func TestObjectExpressions(tst *testing.T) {
	checkExpr(tst, "var obj = {a: 1}; obj.a", int64(1))
	checkExpr(tst, "var obj = {a: 10, \"b\": 20}; obj.a + obj.b", int64(30))

	checkExpr(tst, `var obj = {a: 5}; obj["a"]`, int64(5))
	checkExpr(tst, `var obj = {hello: 42}; obj["hello"]`, int64(42))

	checkExpr(tst, "var obj = {x: 1, y: 2}; obj.x * obj.y", int64(2))

	checkExpr(tst, "var arr = [{a: 1}, {a: 12}]; arr[0].a", int64(1))
	checkExpr(tst, "var arr = [{a: 1}, {a: 12}]; arr[1].a", int64(12))

	checkExpr(tst, "var obj = {\"arr\": [10, 20, 30]}; obj.arr[1]", int64(20))
	checkExpr(tst, "var obj = {arr: [10, 20, 30]}; obj.arr.length", int64(3))

	checkExpr(tst, "var obj = {a: 1}; obj.a = 12; obj.a", int64(12))
	checkExpr(tst, "var obj = {a: 1, b: 2}; obj.b = 20; obj.b", int64(20))

	checkExpr(tst, "var obj = {x: 0}; obj.x = 12", int64(12))

	checkExpr(tst, `var obj = {z: 1}; obj["z"] = 100; obj.z`, int64(100))
	checkExpr(tst, `var obj = {}; obj["tst"] = 55; obj.tst`, int64(55))

	checkExpr(tst, "var obj = {}; obj.x = 10; obj.x", int64(10))
	checkExpr(tst, "var obj = {a: 1}; obj.b = 2; obj.a + obj.b", int64(3))

	checkExpr(tst, `var obj = {};
                    obj.x = 1; obj.y = 12; obj.z = 3;
                    obj.x + obj.y + obj.z`, int64(16))

	checkExpr(tst, `var obj = {arr: [1, 2, 3]};
                    obj.arr[0] = 12; obj.arr[0]`, int64(12))

	checkExpr(tst, `var arr = [{x: 1}, {x: 2}];
                    arr[0].x = 12; arr[0].x`, int64(12))

	checkExpr(tst, `var obj = {tag: {val: 0}};
                    obj.tag.val = 42; obj.tag.val`, int64(42))
}

func TestStringExpressions(tst *testing.T) {
	checkExpr(tst, `"".length`, int64(0))
	checkExpr(tst, `"abc".length`, int64(3))
	checkExpr(tst, `"helloworld".length`, int64(10))

	checkExpr(tst, `"helloworld"[0]`, "h")
	checkExpr(tst, `"helloworld"[1]`, "e")
	checkExpr(tst, `"helloworld"[6]`, "o")

	checkExpr(tst, `var s = "helloworld"; s.length`, int64(10))
	checkExpr(tst, `var s = "test"; s[0]`, "t")
}

// With functions as variables, test them in expressions...
func TestFunctionExpressions(tst *testing.T) {
	checkExpr(tst, `function five() {
                        return 5;
                    } five()`, int64(5))

	checkExpr(tst, `function undef() {}
                    undef()`, nil)

	checkExpr(tst, `function add(a, b) {
                        return a + b;
                    } add(1, 12)`, int64(13))

	checkExpr(tst, `function sum(a, b, c) {
                        return a + b + c;
                    } sum(1, 2, 3)`, int64(6))

	checkExpr(tst, `function calc(x) {
                        var y = x * 2;
                        return y + 1;
                    } calc(5)`, int64(11))

	checkExpr(tst, `function inner(y) {
                        return y * 2;
                    }
                    function outer(x) {
                        return inner(x + 1);
                    } outer(5)`, int64(12))

	checkExpr(tst, `var double = function(x) {
                        return x * 2;
                    }; double(5)`, int64(10))

	checkExpr(tst, `var triple = function times3(x) {
                        return x * 3;
                    }; triple(4)`, int64(12))

	checkExpr(tst, `function make(val) {
                        return { x: val };
                    } make(42).x`, int64(42))

	checkExpr(tst, `function make(a, b) {
                        return [a, b];
                    } make(1, 12)[1]`, int64(12))

	checkExpr(tst, `function abs(x) {
                        if (x < 0) return -x;
                        return x;
                    } abs(-5)`, int64(5))

	checkExpr(tst, `function fact(n) {
                        var res = 1;
                        for (var idx = 2; idx <= n; idx++) {
                            res = res * idx;
                        }
                        return res;
                    } fact(5)`, int64(120))

	checkExpr(tst, `function fib(n) {
                        if (n < 2) return n;
                        return fib(n - 1) + fib(n - 2);
                    } fib(10)`, int64(55))

	checkExpr(tst, `(function() {
                         return 42;
                     })()`, int64(42))

	checkExpr(tst, `(function(x) {
                         return x * 2;
                     })(21)`, int64(42))

	checkExpr(tst, `var arr = [ function(x) {
                        return x + 1;
                    } ]; arr[0](11)`, int64(12))

	checkExpr(tst, `var obj = { fn: function(x) {
                        return x * 3;
                    } }; obj.fn(4)`, int64(12))

	// More complex closure tests follow
	checkExpr(tst, `function makeAdder(n) {
	                    return function(x) { return x + n; };
	                }
	                var add5 = makeAdder(5);
	                add5(10)`, int64(15))
}

func TestArrowExpressions(tst *testing.T) {
	checkExpr(tst, `var five = () => 5;
                    five()`, int64(5))

	checkExpr(tst, `var double = x => x * 2;
                    double(6)`, int64(12))

	checkExpr(tst, `var triple = (x) => x * 3;
                    triple(4)`, int64(12))

	checkExpr(tst, `var add = (a, b) => a + b;
                    add(1, 12)`, int64(13))
	checkExpr(tst, `var sum = (a, b, c) => a + b + c;
                    sum(1, 2, 3)`, int64(6))

	checkExpr(tst, `var square = x => {
                        return x * x;
                    }; square(12)`, int64(144))

	checkExpr(tst, `var mult = (a, b) => {
                        return a * b;
                    }; mult(3, 4)`, int64(12))

	checkExpr(tst, `var calc = (x) => {
                        var y = x * 2;
                        return y + 1;
                    }; calc(5)`, int64(11))

	// Ugh, hate not bracketing the condition but this is a test case...
	checkExpr(tst, `var abs = x => x < 0 ? -x : x; abs(-5)`, int64(5))

	// Second ugh, can't currently handle block ambiguity
	checkExpr(tst, `var make = x => ({val: x}); make(42).val`, int64(42))

	checkExpr(tst, "var pair = (a, b) => [a, b]; pair(1, 2)[0]", int64(1))

	checkExpr(tst, "(() => 12)()", int64(12))
	checkExpr(tst, "((x) => x * 2)(21)", int64(42))

	checkExpr(tst, "var arr = [ x => 2 * x + 2 ]; arr[0](5)", int64(12))

	checkExpr(tst, "var obj = { fn: x => x * 3 }; obj.fn(4)", int64(12))

	checkExpr(tst, `function apply(fn, val) {
                        return fn(val);
                    }
                    apply(x => x * 2, 10)`, int64(20))

	// More complex closures to follow (although this is confusing enough)
	checkExpr(tst, `var outer = x => (y => x + y);
	                outer(5)(3)`, int64(8))
}

func TestClosureExpressions(tst *testing.T) {
	checkExpr(tst, `function makeCounter() {
	                    var count = 0;
	                    return function() {
	                        count = count + 1;
	                        return count;
	                    };
	                }
	                var counter = makeCounter();
	                counter() + counter() + counter()`, int64(6))

	checkExpr(tst, `function multiplier(factor) {
	                    return function(x) { return x * factor; };
	                }
	                var double = multiplier(2);
	                var triple = multiplier(3);
	                double(5) + triple(5)`, int64(25))

	checkExpr(tst, `function makeAdder(n) {
	                    return x => x + n;
	                }
	                var add10 = makeAdder(10);
	                add10(2)`, int64(12))

	checkExpr(tst, `function outer(a) {
	                    return function middle(b) {
	                        return function inner(c) {
	                            return a + b + c;
	                        };
	                    };
	                }
	                outer(1)(2)(3)`, int64(6))

	checkExpr(tst, `var outer = a => b => c => a + b + c;
	                outer(10)(20)(30)`, int64(60))

	checkExpr(tst, `function calc(a, b) {
	                    return function(op) {
	                        if (op === 1) return a + b;
	                        return a * b;
	                    };
	                }
	                var fn = calc(3, 4);
	                fn(1) + fn(2)`, int64(3+4+3*4))

	checkExpr(tst, `function test() {
	                    var x = 10;
	                    var fn = function() { return x; };
	                    x = 20;
	                    return fn();
	                }
	                test()`, int64(20))
}

func TestCommaExpressions(tst *testing.T) {
	checkExpr(tst, "(1, 2, 3)", int64(3))
	checkExpr(tst, "(1 + 2, 3 + 4)", int64(7))

	// Test cases for side effects
	checkExpr(tst, "var x = 0; (x = 2, x + 10)", int64(12))
	checkExpr(tst, "var a = 1, b = 2; (a = 10, b = 20, a + b)", int64(30))

	checkExpr(tst, `var sum = 0;
	                for (var i = 0, j = 10; i < j; i++, j--) {
	                    sum = sum + 1;
	                }
	                sum`, int64(5))

    // This replicates above, make sure RBP break is working correctly
	checkExpr(tst, "var x = 1, y = 2, z = 3; x + y + z", int64(6))

    // Arrow functions is where comma expressions are commonly used
	checkExpr(tst, `var fns = [x => x + 1, x => x * 2];
	                fns[0](5) + fns[1](5)`, int64(16))
}
