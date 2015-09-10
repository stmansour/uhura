package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type StatusReq struct {
	State    string
	InstName string
	UID      string
	Tstamp   string
}

//  Check for the state of all sub-environments. If they
//  have all entered the same state return true. Otherwise
//  return false;
func AllAppStatesMatch(es int) bool {
	same := true
	for i := 0; same && i < len(UEnv.Instances); i++ {
		for j := 0; same && j < len(UEnv.Instances[i].Apps); j++ {
			same = (es == UEnv.Instances[i].Apps[j].State)
		}
	}
	return same
}

func BadState() {
	ulog("Unrecognized state: %d\nTime to panic\n", UEnv.State)
	err := fmt.Errorf("Unrecognized state: %d", UEnv.State)
	check(err)
}

// Check the state of all environments and see if a state
// change is needed. If so, make the state change and return
// true. Otherwise, return false.
func ChangeState() bool {
	change := false // assume nothing changes

	// A state change has occurred. Update the state
	// of the whole environment...
	switch {
	case UEnv.State == uINIT:
		// is everyone in the READY state now???
		if AllAppStatesMatch(uREADY) {
			// Great all environments report READY.
			// Let's transition to the Testing state
			change = true
			UEnv.State = uTEST
			ulog("state changing from INIT to TEST\n")
		}
	case UEnv.State == uREADY:
		// We should never get here. When all
		// sub-environments report ready, we move straight
		// to the testing state
	case UEnv.State == uTEST:
		if AllAppStatesMatch(uDONE) {
			// Great all environments report testing completed.
			// We're done. Transition to the DONE state
			change = true
			UEnv.State = uDONE
			ulog("state change to DONE\n")
		}
	case UEnv.State == uDONE:
		ulog("state change check: %s\n", StateToString(UEnv.State))
	default:
		BadState()
	}
	return change
}

// Perform any state changes needed...
func ProcessStateChanges() {
	if ChangeState() {
		switch {
		case UEnv.State == uTEST:
			ulog("Handle state change to TEST\n")
		case UEnv.State == uDONE:
			ulog("Handle state change to DONE\n")
		default:
			BadState()
		}
	}
}

//  One of the environments has sent status
//  Update internals and make any state change that
//  result from the status update.
func SetStatus(s *StatusReq) {
	DPrintEnvDescr("Entering SetStatus\n")
	// Build the quartermaster script to create each environment instance...
	for i := 0; i < len(UEnv.Instances); i++ {
		if UEnv.Instances[i].InstName == s.InstName {
			for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
				if s.UID == UEnv.Instances[i].Apps[j].UID {
					UEnv.Instances[i].Apps[j].State = StateToInt(s.State)
					DPrintEnvInstance(&UEnv.Instances[i], i)
					ProcessStateChanges()
				}
			}
		}
	}
	DPrintEnvDescr("Exiting SetStatus\n")
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	SendOKReply(w)
	ulog("Shutdown Handler\n")
	ulog("Normal Shutdown\n")
	os.Exit(0)
}

func SendOKReply(w http.ResponseWriter) {
	m := UhuraResponse{Status: "OK", Timestamp: time.Now().Format(time.RFC850)}
	str, err := json.Marshal(m)
	if nil != err {
		fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n",
			"encoding error", time.Now().Format(time.RFC850))
	} else {
		fmt.Fprintf(w, string(str))
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Status Handler\n")
	DPrintHttpRequest(r)

	decoder := json.NewDecoder(r.Body)
	var s StatusReq
	err := decoder.Decode(&s)
	if err != nil {
		panic(err)
	}
	DPrintStatusMsg(&s)

	//  Scan the datastructure, find this instance, mark its status
	SetStatus(&s)
	SendOKReply(w)
}

func MapHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Map Handler\n")
	DPrintHttpRequest(r)
	http.ServeFile(w, r, "/Users/sman/Documents/src/go/src/uhura/test/map.json")
	SendOKReply(w)
}

func TestStartHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Test Start Handler\n")
	DPrintHttpRequest(r)
	SendOKReply(w)
}

func TestDoneHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Test Done Handler\n")
	DPrintHttpRequest(r)
	SendOKReply(w)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//PrintHttpRequest(r)
		fn(w, r)
	}
}
