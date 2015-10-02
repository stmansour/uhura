// This is the code that controls access to the shared memory... the
// environment description. Any other code that wants to read or write
// to the shared memory must com through the dispatcher.  Even access
// to ulog needs to be funneled through the dispatcher.
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// DispatcherCreateChannels creates all the channels used by the different
// go routines that need to access shared resources
func DispatcherCreateChannels() {
	Uhura.HReqMem = make(chan int)
	Uhura.HReqMemAck = make(chan int)
	Uhura.StateChg = make(chan AppStatChg)
	Uhura.StateChgAck = make(chan int)
	Uhura.LogEnvDescr = make(chan int)
	Uhura.LogEnvDescrAck = make(chan int)
	Uhura.LogString = make(chan string)
	Uhura.LogStringAck = make(chan int)
	Uhura.LogStatus = make(chan StatusReq)
	Uhura.LogStatusAck = make(chan int)
	Uhura.ShutdownReq = make(chan int)
	Uhura.ShutdownReqAck = make(chan int)
}

func generateEnvDesrReport() {
	f, err := os.Create("EnvShutdownStatus.json")
	check(err)
	defer f.Close()
	b, err := json.MarshalIndent(&UEnv, "", "    ")
	if err != nil {
		fmt.Printf("Cannot marshal UEnv! Error: %v\n", err)
		os.Exit(2) // no recovery from this
	}
	fileWriteBytes(f, b)
	f.Sync()
}

// This is a go routine, it runs asynchronously.
// It starts a timer to give the last few things in motion a chance to finish
// cleanly before exiting.
// So, before the timer is up, it needs to access the log just like any
// other routine.  When it finishes, execute the shutdown. This
func simpleShutdown() {
	ttl := 5 // seconds
	Uhura.LogString <- fmt.Sprintf("SHUTDOWN will commence in a few seconds\n")
	<-Uhura.LogStringAck
	time.Sleep(time.Duration(rand.Intn(ttl)) * time.Second) // let the dust settle
	AWSTerminateInstances()                                 // terminate the aws instances
	generateEnvDesrReport()                                 // generate env status file
	ulog("Shutdown Handler - Exiting NOW!\n")               // ok, all bets are off now
	ulog("Exiting uhura\n")                                 // just blast the log
	os.Exit(0)                                              // and exit
}

// Dispatcher directs the operation of uhura in steady state. It controls
// the access to shared resources. It handles actions requested by the
// StateOrchestrator.
func Dispatcher() {
	var act int
	startedShutdown := false
	startedTestNow := false
	ulog("Started Dispatcher\n")
	for {
		act = actionNone // don't do anything unless orchestrator tells us
		select {

		case <-Uhura.HReqMem:
			Uhura.HReqMemAck <- 1 // go ahead
			<-Uhura.HReqMemAck    // block until HTTP code is done with mem

		case asc := <-Uhura.StateChg:
			Uhura.StateChgAck <- 1 // got the message and will continue
			act = StateOrchestarator(&asc)

		case sm := <-Uhura.LogStatus:
			Uhura.LogStatusAck <- 1 // let caller continue because http uses memory in Read-Only mode...
			dPrintStatusMsg(&sm)    // any change will need to come through StateOrchestrator via Dispatcher

		case <-Uhura.LogEnvDescr:
			Uhura.LogEnvDescrAck <- 1    // tell caller we're done
			dPrintEnvDescr("Dispatcher") // performing the print

		case s := <-Uhura.LogString:
			Uhura.LogStringAck <- 1
			ulog(s)

		case <-Uhura.ShutdownReq:
			act = actionShutdown // we assume the code only calls this when it should
			Uhura.ShutdownReqAck <- 1
		}

		switch {
		case act == actionNone:
			// nothing to do, just wanted to explicitly show it
		case act == actionTestNow:
			if !startedTestNow {
				go CommsSendTestNow()
				startedTestNow = true
			}
		case act == actionShutdown:
			if !startedShutdown {
				ulog("Normal Shutdown - all TGO instances report DONE\n")
				go simpleShutdown()    // this will give us a few seconds for things to quiesce
				startedShutdown = true // don't call it again
			}
		}
	}
}
