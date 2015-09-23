package main

import (
	"fmt"
	"testing"
)

type TestStep struct {
	asc   AppStatChg
	state int
	act   int
}

func TestStatusHandler(t *testing.T) {
	// func TestStatusHandler() {
	Uhura.EnvDescFname = "./test/utdata/ut1.json"
	Uhura.DryRun = true
	InitUhura()
	InitEnv()

	// env descr = ./test/utdata/ut1.json
	//            inst, app, state, ENVSTATE, ACTION
	var test1a = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uINIT}, uINIT, actionNone},
		{AppStatChg{0, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uTEST, actionNone},
		{AppStatChg{0, 0, uDONE}, uDONE, actionShutdown},
		{AppStatChg{0, 0, uTERM}, uTERM, actionShutdown},
	}

	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 1\n\n")
	for i := 0; i < len(test1a); i++ {
		test := test1a[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 1a, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                 State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}

	// env descr = ./test/utdata/ut1.json
	//            inst, app, state, ENVSTATE, ACTION
	var test1b = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uTEST, actionNone},
		{AppStatChg{0, 0, uDONE}, uDONE, actionShutdown},
		{AppStatChg{0, 0, uTERM}, uTERM, actionShutdown},
	}
	Uhura.EnvDescFname = "./test/utdata/ut1.json"
	InitEnv()
	UEnv.State = uUNKNOWN

	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 1\n\n")
	for i := 0; i < len(test1b); i++ {
		test := test1b[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 1, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}

	// env descr = ./test/utdata/ut2.json
	//            inst, app, state
	var test2 = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uINIT}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uINIT}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uINIT}, uINIT, actionNone},
		{AppStatChg{0, 0, uREADY}, uINIT, actionNone},
		{AppStatChg{1, 0, uREADY}, uINIT, actionNone},
		{AppStatChg{2, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{1, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{2, 0, uTEST}, uTEST, actionNone},
		{AppStatChg{0, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{1, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{2, 0, uDONE}, uDONE, actionShutdown},
	}
	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 2\n\n")

	Uhura.EnvDescFname = "./test/utdata/ut2.json"
	InitEnv()
	UEnv.State = uUNKNOWN

	for i := 0; i < len(test2); i++ {
		test := test2[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 2, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}

	// env descr = ./test/utdata/ut2.json
	//            inst, app, state,  ENV STATE, action
	var test3 = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uINIT}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uREADY}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uREADY}, uINIT, actionNone},
		{AppStatChg{0, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{1, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{2, 0, uTEST}, uTEST, actionNone},
		{AppStatChg{0, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{1, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{2, 0, uDONE}, uDONE, actionShutdown},
	}
	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 3\n\n")

	Uhura.EnvDescFname = "./test/utdata/ut2.json"
	InitEnv()
	UEnv.State = uUNKNOWN

	for i := 0; i < len(test3); i++ {
		test := test3[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 3, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}

	// env descr = ./test/utdata/ut2.json
	//            inst, app, state,  ENV STATE, action
	var test4 = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uREADY}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uREADY}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{1, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{2, 0, uTEST}, uTEST, actionNone},
		{AppStatChg{0, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{1, 0, uDONE}, uTEST, actionNone},
		{AppStatChg{2, 0, uDONE}, uDONE, actionShutdown},
	}
	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 4\n\n")

	Uhura.EnvDescFname = "./test/utdata/ut2.json"
	InitEnv()
	UEnv.State = uUNKNOWN

	for i := 0; i < len(test4); i++ {
		test := test4[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 4, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}

	// env descr = ./test/utdata/ut2.json
	//            inst, app, state,  ENV STATE, action
	// This case - they go straight from READY to DONE - because they test
	// extremely fast.  Uhura tells them to test. They finish testing before
	// uhura samples their state after informing them to test
	var test5 = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uUNKNOWN}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uREADY}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uREADY}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uREADY}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST}, uREADY, actionNone},
		{AppStatChg{0, 0, uDONE}, uREADY, actionNone},    // uhura moves to DONE right after telling the apps to test
		{AppStatChg{1, 0, uDONE}, uREADY, actionNone},    // when uhura asks, this app is already done testing
		{AppStatChg{2, 0, uDONE}, uDONE, actionShutdown}, // when uhura asks, this app is already done testing
	}
	// fmt.Printf("\n#####################################################\n")
	// fmt.Printf("Test 4\n\n")

	Uhura.EnvDescFname = "./test/utdata/ut2.json"
	InitEnv()
	UEnv.State = uUNKNOWN

	for i := 0; i < len(test5); i++ {
		test := test5[i]
		ulog("Change: Inst=%d, App=%d, State->%s\n", test.asc.inst, test.asc.app, StateToString(test.asc.state))
		act := StateOrchestarator(&test.asc)
		DPrintEnvDescr(fmt.Sprintf("After processing %+v\n", test.asc))
		if test.act != act || test.state != UEnv.State {
			t.Errorf("TEST 5, step %d, Action -  expected: %d, found: %d", i, test.act, act)
			t.Errorf("                State - expected: %s, found: %s", StateToString(test.state), StateToString(UEnv.State))
		}
	}
}
