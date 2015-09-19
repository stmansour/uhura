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

type UResp struct {
	Status    string
	ReplyCode int
	Timestamp string
}

const (
	RespOK = iota
	RespNoSuchInstance
	InvalidState
)

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

func SendReply(w http.ResponseWriter, rc int, s string) {
	w.Header().Set("Content-Type", "application/json")
	m := UResp{Status: s, ReplyCode: rc, Timestamp: time.Now().Format(time.RFC822)}
	str, err := json.Marshal(m)
	if nil != err {
		fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n",
			"encoding error", time.Now().Format(time.RFC822))
	} else {
		fmt.Fprintf(w, string(str))
	}
}

func SendOKReply(w http.ResponseWriter) {
	SendReply(w, RespOK, "OK")
}

func BadState(w http.ResponseWriter, s *StatusReq) {
	r := fmt.Sprintf("BAD STATE: %s", s.State)
	ulog("%s\n", r)
	SendReply(w, InvalidState, r)
}

func BadInstUidCombo(w http.ResponseWriter, s *StatusReq) {
	r := fmt.Sprintf("BAD INSTANCE-UID: %s-%s", s.InstName, s.UID)
	ulog("%s\n", r)
	SendReply(w, RespNoSuchInstance, r)
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
		panic(fmt.Errorf("ProcessStateChanges: Should never happen"))
	}
	return change
}

// Perform any state changes needed...
func ProcessStateChanges() {
	if ChangeState() {
		switch {
		case UEnv.State == uTEST:
			ulog("Handle state change to TEST\n")
			// send message to all tgos that we move to TEST

		case UEnv.State == uDONE:
			ulog("Handle state change to DONE\n")
			AWSTerminateInstances()
			for i := 0; i < len(UEnv.Instances); i++ {
				for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
					UEnv.Instances[i].Apps[j].State = uTERM
				}
			}
			DPrintEnvDescr("Terminated All Instances")

		default:
			panic(fmt.Errorf("ProcessStateChanges: Should never happen"))
		}
	}
}

//  One of the environments has sent status
//  Update internals and make any state change that
//  result from the status update.
//  the ResponseWriter is passed in to handle the different
//  error cases we might find.
func SetStatus(w http.ResponseWriter, s *StatusReq) error {
	ulog("Entering SetStatus\n")
	found := false
	// Build the quartermaster script to create each environment instance...
	for i := 0; i < len(UEnv.Instances) && !found; i++ {
		if UEnv.Instances[i].InstName == s.InstName {
			for j := 0; j < len(UEnv.Instances[i].Apps) && !found; j++ {
				if s.UID == UEnv.Instances[i].Apps[j].UID {
					found = true
					st := StateToInt(s.State)
					if st < 0 {
						err := fmt.Errorf("Unrecognized State: %s", s.State)
						DPrintEnvDescr("Exiting SetStatus with error\n")
						BadState(w, s)
						return err
					} else {
						UEnv.Instances[i].Apps[j].State = st
						DPrintEnvInstance(&UEnv.Instances[i], i)
						ProcessStateChanges()
						SendOKReply(w)
					}
				}
			}
		}
	}
	if !found {
		err := fmt.Errorf("NO SUCH INSTANCE-UID: %s-%s", s.InstName, s.UID)
		BadInstUidCombo(w, s)
		//DPrintEnvDescr("Exiting SetStatus with error\n")
		return err
	}
	// DPrintEnvDescr("Exiting SetStatus\n")
	ulog("Exiting SetStatus\n")
	return nil
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	SendOKReply(w)
	ulog("Shutdown Handler\n")
	ulog("Normal Shutdown\n")
	os.Exit(0)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Status Handler\n")

	decoder := json.NewDecoder(r.Body)
	var s StatusReq
	err := decoder.Decode(&s)
	if err != nil {
		panic(err)
	}
	DPrintStatusMsg(&s)

	//  Scan the datastructure, find this instance, mark its status
	//  As many errors can occur, we pass in the response writer
	//  and handle the different returns within SetStatus
	err = SetStatus(w, &s)
}

func MapHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Map Handler\n")
	DPrintHttpRequest(r)

	// This is a temporary hack until I can create the real one...
	http.ServeFile(w, r, "test/stateflow_normal/env.json")
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}
