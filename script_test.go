/*
 * Test methods for engine as a whole, including the tcNN execution.
 *
 * Copyright (C) 2005-2024 J.M. Heisz.  All Rights Reserved.
 * See the LICENSE file accompanying the distribution your rights to use
 * this software.
 */

package gescript

import (
	"os"
	"testing"
)

// This will be fleshed out...
func TestValueEscape(tst *testing.T) {
	prg, err := Parse("1")
	if err != nil {
		tst.Fatalf("Unexpected parse error for '1'")
	}
	val, err := prg.Run()
	if err != nil {
		tst.Fatalf("Unexpected run error for '1'")
	}
	if val.Native().(int64) != 1 {
		tst.Fatalf("Unexpected return value for '1'")
	}
}

// When someone provides an overall testsuite, you use it!
var passCnt int = 0
var failCnt int = 0

func runTestScript(tst *testing.T, fname string) {
	dat, err := os.ReadFile(fname)
	if err != nil {
		tst.Logf("Error reading test file " + fname + ":" + err.Error())
		return
	}

	_, err = Run(string(dat))
	if err != nil {
		tst.Logf("Error in test execution " + fname + ":" + err.Error())
		failCnt++
	} else {
		tst.Logf("PASS - " + fname)
		passCnt++
	}
}
func scanTestDir(tst *testing.T, dname string) {
	entries, err := os.ReadDir(dname)
	if err != nil {
		tst.Logf("Error reading test directory " + dname + ":" + err.Error())
		return
	}

	for _, ent := range entries {
		if ent.IsDir() {
			scanTestDir(tst, dname+"/"+ent.Name())
		} else {
			runTestScript(tst, dname+"/"+ent.Name())
		}
	}
}
func Test262(tst *testing.T) {
	scanTestDir(tst, "./test262-main/test/language")
	tst.Logf("Test262 results: %d pass, %d fail", passCnt, failCnt)
	if failCnt != 0 {
		tst.Fatalf("Failuser in test262 execution")
	}
}
