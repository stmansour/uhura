// This code handles the incoming http requests. It must communicate with Dispatcher
// in order to read / write shared memory (the EnvDescriptor). It must also communicate
// with the Dispatcher in order to log messages
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// KVMsg is a named structure of key/value pairs for clients to
// add to a status message as needed
type KVMsg struct {
	Name string
	KVs  []KeyVal
}

// StatusReq is the structure of data used by the http code to process
// a status message from tgo
type StatusReq struct {
	State     string              // new status
	InstName  string              // instance name
	UID       string              // uid of app
	Tstamp    string              // when was it sent
	w         http.ResponseWriter // where to write the response
	updateEnv bool                // send this request on to StateOrchestrator and update EnvDescr
	logmsgs   []string            // we'll need to save these until it's safe to print them
	KV        KVMsg               // clients can add messages, like test results, etc.
}

// MapReq is a request from a tgo to respond with
// an updated version of the environment descriptor that
// includes the PublicDNS host names for all instances
// ThisInst identifies which instance is asking for the map
type MapReq struct {
	ThisInst int
}

// UResp is uhura's reply message to the status message from tgo
type UResp struct {
	Status    string
	ReplyCode int
	Timestamp string
}

// RespOK and the rest are response codes to tgo status messages is
// RespNoSuchInstance is the reply code when
// Invalid
const (
	RespOK             = iota // the reply code to a successfully handled status message from Tgo
	RespNoSuchInstance        // instance-uid combination could not be found in the current environment
	InvalidState              // the supplied state is invalid
)

/***********************************************************************************************
 ***********************************************************************************************
 ****   * * * * BEGIN * * * *
 ****   ALL CODE BETWEEN HERE AND THE END MARKER BELOW WILL BLOCK EVERYTHING
 ****   IF THEY MAKE ANY DISPATCHER CHANNEL REQUEST.  DO NOT MAKE CHANNEL REQUESTS OR PRINT
 ****   TO FILES OR THE SCREEN
 ***********************************************************************************************
 ***********************************************************************************************/

// Since we cannot call the logging func in this area
// we'll just save up the strings that the routines were going
// to log and send them later when we use the channels to guarantee
// we're the only function writing.  See
func httplog(s *StatusReq, format string, a ...interface{}) {
	s.logmsgs = append(s.logmsgs, fmt.Sprintf(format, a...))
}

// SendReply is the generic response sender for Uhura
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

// SendOKReply is a convenience routine for sending a reply of OK
func SendOKReply(s *StatusReq) {
	SendReply(s.w, RespOK, "OK")
}

// BadState is a convenience routine for sending a reply to the caller
// indicating that the status message they sent has an invalid state
func BadState(s *StatusReq) {
	r := fmt.Sprintf("BAD STATE: %s", s.State)
	httplog(s, "%s\n", r)
	SendReply(s.w, InvalidState, r)
}

// BadInstUIDCombo is a convenience routine for sending a reply to a
// caller indicating that the instance-name, UID pair provided was not
// found in the currently running environment
func BadInstUIDCombo(s *StatusReq) {
	r := fmt.Sprintf("BAD INSTANCE-UID: %s-%s", s.InstName, s.UID)
	httplog(s, "%s\n", r)
	SendReply(s.w, RespNoSuchInstance, r)
}

// HandleSetStatus does the actual work involved in updating the internal
// data structures based on the status received.
func HandleSetStatus(s *StatusReq, asc *AppStatChg) {
	found := false
	for i := 0; i < len(UEnv.Instances) && !found; i++ {
		if UEnv.Instances[i].InstName != s.InstName {
			continue
		}
		asc.inst = i // found the instance
		for j := 0; j < len(UEnv.Instances[i].Apps) && !found; j++ {
			if s.UID != UEnv.Instances[i].Apps[j].UID {
				continue
			}
			found = true
			asc.app = j // found the app index
			st := stateToInt(s.State)
			if st < 0 {
				s.updateEnv = false
				httplog(s, "Unrecognized State: %s", s.State)
				BadState(s)
			} else {
				asc.state = st
				for k := 0; k < len(s.KV.KVs); k++ {
					asc.KV.KVs = append(asc.KV.KVs, KeyVal{s.KV.KVs[k].Key, s.KV.KVs[k].Val})
				}
				SendOKReply(s)
			}
		}
	}
	if !found {
		s.updateEnv = false
		BadInstUIDCombo(s)
	}
}

/***********************************************************************************************
 ***********************************************************************************************
 ****   * * * * END * * * *
 ****   ALL CODE ABOVE THIS MARKER TO THE BEGIN MARKER COMMENTS BELOW WILL BLOCK EVERYTHING
 ****   IF THEY MAKE ANY DISPATCHER CHANNEL REQUEST.  DO NOT MAKE ANY CHANNEL REQUESTS
 ***********************************************************************************************
 ***********************************************************************************************/

// sendHTTPLogMsgs sends the log messages that were saved during the
// execution area where we had a memory lock.
func sendHTTPLogMsgs(s *StatusReq) {
	for i := 0; i < len(s.logmsgs); i++ {
		Uhura.LogString <- s.logmsgs[i]
		<-Uhura.LogStringAck
	}
}

// StatusHandler is called when the http listener receives a requests
// specifying the "/status/" path. This routine coordinates with the
// Dispatcher to gain access to memory and other shared resources. It
// calls HandleStatus to actually process the status message.
// Steps are as follows
//
//     *) Notify the Dispatcher that we want access to the shared
//        memory. Block until it has been granted.
//     *) Process the request and send the reply, but
//        do not update the shared memory here. Instead, build the
//        structure of data describing the change.
//     *) Send it to the dispatcher, who will, in turn, send it to
//        the state orchestrator.
//
// The idea here is to read enough info to determine the proper
// reply to this status update, then let the StateOrchestrator make
// the changes to the memory and take any appropriate actions.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	var s StatusReq
	var asc AppStatChg
	s.logmsgs = make([]string, 1)
	s.logmsgs[0] = "Status Handler\n"
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		panic(err)
	}
	Uhura.LogStatus <- s      // log status message before we start
	<-Uhura.LogStatusAck      // make sure it was done
	s.w = w                   // send response here
	s.updateEnv = true        // assume we update, set to false if error
	Uhura.HReqMem <- 1        // ask to access the shared mem, blocks until granted
	<-Uhura.HReqMemAck        // make sure we got it
	HandleSetStatus(&s, &asc) // handle the status req
	Uhura.HReqMemAck <- 1     // tell Dispatcher we're done with the data

	sendHTTPLogMsgs(&s)
	if !s.updateEnv {
		return // exit now if we don't update
	}
	Uhura.StateChg <- asc  // otherwise, send the struct describing the update
	<-Uhura.StateChgAck    // wait for confirmation and we're done
	Uhura.LogEnvDescr <- 1 // dump env descr
	<-Uhura.LogEnvDescrAck // make sure it got done
}

// ShutdownHandler handles an http message sent to "/shutdown/"
// TODO: Current implementation is a bit of a hack.  It needs to contact
// the StateOrchestrator for handling this request
func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	SendReply(w, RespOK, "OK")
	Uhura.ShutdownReq <- 1
	<-Uhura.ShutdownReqAck
}

// MapHandler handles an http message sent to "/map/"
// Only one modification is made to the internal datastructure... that is
// the instance that is asking for the map is set in UEnv.ThisInst
func MapHandler(w http.ResponseWriter, r *http.Request) {
	Uhura.LogString <- "Map Handler\n"         // send to log
	<-Uhura.LogStringAck                       // wait for confirmation
	var m MapReq                               // caller identifies "ThisInst"
	decoder := json.NewDecoder(r.Body)         // read it in
	if err := decoder.Decode(&m); err != nil { // if there are problems...
		panic(err)
	}
	Uhura.HReqMem <- 1                                 // ask to access the shared mem, blocks until granted
	<-Uhura.HReqMemAck                                 // make sure we got it
	UEnv.ThisInst = m.ThisInst                         // don't change this tgo's instance pointer
	b, _ := json.Marshal(&UEnv)                        // marshal the environment into bytes
	w.Header().Set("Content-Type", "application/json") // response type is app/json
	w.Write(b)                                         // send the bytes
	Uhura.HReqMemAck <- 1                              // tell Dispatcher we're done with the data
}

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func initHTTP() {
	// Set up the handler functions for our server...
	http.HandleFunc("/shutdown/", makeHandler(ShutdownHandler))
	http.HandleFunc("/status/", makeHandler(StatusHandler))
	http.HandleFunc("/map/", makeHandler(MapHandler))
}
