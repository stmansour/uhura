package main

//
// The logger and some debugging aids...
//

import (
	"fmt"
	"log"
)

// This is uhura's standard logger
func ulog(format string, a ...interface{}) {
	p := fmt.Sprintf(format, a...)
	log.Print(p)
	if Uhura.DebugToScreen {
		fmt.Print(p)
	}
}

func StateToInt(s string) int {
	var i int
	switch {
	case s == "UNKNOWN":
		i = uUNKNOWN
	case s == "INIT":
		i = uINIT
	case s == "READY":
		i = uREADY
	case s == "TEST":
		i = uTEST
	case s == "DONE":
		i = uDONE
	case s == "TERM":
		i = uTERM
	default:
		i = -1
	}
	return i
}

func StateToString(i int) string {
	var s string
	switch {
	case i == uUNKNOWN:
		s = "UNKNOWN"
	case i == uINIT:
		s = "INIT"
	case i == uREADY:
		s = "READY"
	case i == uTEST:
		s = "TEST"
	case i == uDONE:
		s = "DONE"
	case i == uTERM:
		s = "TERM"
	default:
		s = "<<unknown state>>"
	}
	return fmt.Sprintf("%d (%s)", i, s)
}

func PrintEnvInstance(e *InstDescr, i int) {
	ulog("    Instance[%d]:  InstName(%s)\n", i, e.InstName)
	ulog("\tApps:\n")
	for j := 0; j < len(e.Apps); j++ {
		ulog("\t[%d]\tUID         : %s\n", j, e.Apps[j].UID)
		ulog("\t\tName        : %s\n", e.Apps[j].Name)
		ulog("\t\tUPort       : %d\n", e.Apps[j].UPort)
		ulog("\t\tIsTest      : %v\n", e.Apps[j].IsTest)
		ulog("\t\tState       : %s\n", StateToString(e.Apps[j].State))
		ulog("\t\t------------------------------------\n")
	}
}

func DPrintEnvInstance(e *InstDescr, i int) {
	if Uhura.Debug {
		PrintEnvInstance(e, i)
	}
}

func PrintEnvDescr() {
	ulog("----------------------  UEnv  ----------------------\n")
	ulog("EnvName  : %s\n", UEnv.EnvName)
	ulog("State    : %s\n", StateToString(UEnv.State))
	ulog("UhuraPort: %d\n", UEnv.UhuraPort)
	ulog("Instances: %d\n", len(UEnv.Instances))
	for i := 0; i < len(UEnv.Instances); i++ {
		PrintEnvInstance(&UEnv.Instances[i], i)
	}
	ulog("----------------------------------------------------\n")
}
func DPrintEnvDescr(s string) {
	if Uhura.Debug {
		ulog(s)
		PrintEnvDescr()
	}
}

func PrintStatusMsg(s *StatusReq) {
	ulog("##########################################\n")
	ulog("Status Message\n")
	ulog("\tState:		%s\n", s.State)
	ulog("\tInstName:	%s\n", s.InstName)
	ulog("\tUID:		%s\n", s.UID)
	ulog("\tTstamp:		%s\n", s.Tstamp)
	ulog("##########################################\n")
}

func DPrintStatusMsg(s *StatusReq) {
	if Uhura.Debug {
		PrintStatusMsg(s)
	}
}
