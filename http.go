package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type StatusReq struct {
	State    string
	InstName string
	UID      string
	Tstamp   string
	w        http.ResponseWriter
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

//  Check for all states being beyond the supplied state.
//  This can happen during initialization when some instances
//  are in the READY state before others have gotten past
//  the UNKNOWN state.
func AllAppStatesPast(es int) bool {
	past := true
	for i := 0; past && i < len(UEnv.Instances); i++ {
		for j := 0; past && j < len(UEnv.Instances[i].Apps); j++ {
			past = (es < UEnv.Instances[i].Apps[j].State)
		}
	}
	return past
}

func SendOKReply(s *StatusReq) {
	SendReply(s.w, RespOK, "OK")
}

func BadState(s *StatusReq) {
	r := fmt.Sprintf("BAD STATE: %s", s.State)
	ulog("%s\n", r)
	SendReply(s.w, InvalidState, r)
}

func BadInstUidCombo(s *StatusReq) {
	r := fmt.Sprintf("BAD INSTANCE-UID: %s-%s", s.InstName, s.UID)
	ulog("%s\n", r)
	SendReply(s.w, RespNoSuchInstance, r)
}

// Check the state of all environments and see if a state
// change is needed. If so, make the state change and return
// true. Otherwise, return false.
func ChangeState() bool {
	change := false // assume nothing changes

	switch {
	case UEnv.State == uUNKNOWN:
		// This is the exception case.  Some environments may init
		// and transition to READY before other environments have
		// even gotten out of the UNKNONW state.
		if AllAppStatesMatch(uINIT) || AllAppStatesPast(uUNKNOWN) {
			change = true
			UEnv.State = uINIT
		}
	case UEnv.State == uINIT:
		ulog("ChangeState: All instances have reported in\n")
		if AllAppStatesMatch(uREADY) {
			change = true
			UEnv.State = uREADY
		}
	case UEnv.State == uREADY:
		ulog("ChangeState: we're in the READY state\n")
		if AllAppStatesMatch(uTEST) {
			change = true
			UEnv.State = uTEST
		}
	case UEnv.State == uTEST:
		if AllAppStatesMatch(uDONE) {
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
	for change := true; change; {
		change = ChangeState()
		if change {
			switch {
			case UEnv.State == uINIT:
				ulog("All environments have reported in.")
				// nothing else to do until everybody is READY
			case UEnv.State == uREADY:
				CommsSendTestNow()
			case UEnv.State == uTEST:
				ulog("All environments are now in TEST\n")
				// nothing else to do until everybody is DONE
			case UEnv.State == uDONE:
				ulog("All environments DONE\n")
				if Uhura.KeepEnv {
					ulog("Will not terminate environment instances, as uhura was run with -k (keep)\n")
				} else {
					AWSTerminateInstances()
					for i := 0; i < len(UEnv.Instances); i++ {
						for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
							UEnv.Instances[i].Apps[j].State = uTERM
						}
					}
				}
				UEnv.State = uTERM
				DPrintEnvDescr("Terminated All Instances")
				exit_uhura()

			default:
				ulog("State reported: %d\n", UEnv.State)
				fmt.Printf("UEnv.State = %d\n", UEnv.State)
				panic(fmt.Errorf("ProcessStateChanges: Should never happen"))
			}
		}
	}
}

//  When an http handler has an update to make, it pushes
//  the status request onto the channel Uhura.TgoStatus .
//  We read it from the channel and process it here. Other
//  handlers will block until this routine finishes.
//
//	Update internals and make any state change that
//  result from the status update.

//func SetStatus(w http.ResponseWriter, s *StatusReq) error {
func HandleSetStatus(s *StatusReq) {
	DPrintStatusMsg(s) // print what we got
	found := false
	for i := 0; i < len(UEnv.Instances) && !found; i++ {
		if UEnv.Instances[i].InstName != s.InstName {
			continue
		}
		for j := 0; j < len(UEnv.Instances[i].Apps) && !found; j++ {
			if s.UID != UEnv.Instances[i].Apps[j].UID {
				continue
			}
			found = true
			st := StateToInt(s.State)
			if st < 0 {
				_ = fmt.Errorf("Unrecognized State: %s", s.State)
				DPrintEnvDescr("Exiting SetStatus with error\n")
				BadState(s)
			} else {
				UEnv.Instances[i].Apps[j].State = st
				DPrintEnvInstance(&UEnv.Instances[i], i)
				ProcessStateChanges()
				SendOKReply(s)
			}
		}
	}
	if !found {
		_ = fmt.Errorf("NO SUCH INSTANCE-UID: %s-%s", s.InstName, s.UID)
		BadInstUidCombo(s)
	}
}

func SetStatus() {
	ulog("Entering SetStatus\n")
	for {
		// timeout := time.After(1 * time.Minute)
		select {
		case s := <-Uhura.TgoStatus:
			HandleSetStatus(&s)
			// case <-timeout:
			// 	fmt.Printf("SetStatus() - 1 min timeout\n")
			// 	return
		}
	}
}
func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	SendReply(w, RespOK, "OK")
	exit_uhura()
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	ulog("Status Handler\n")
	var s StatusReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		panic(err)
	}

	//  As many errors can occur, we pass in the response writer
	//  and handle the different returns within SetStatus
	s.w = w
	Uhura.TgoStatus <- s
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
