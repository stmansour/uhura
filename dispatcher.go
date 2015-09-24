// This is the code that controls access to the shared memory... the
// environment description. Any other code that wants to read or write
// to the shared memory must com through the dispatcher.  Even access
// to ulog needs to be funneled through the dispatcher.
package main

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
}

// Dispatcher directs the operation of uhura in steady state. It controls
// the access to shared resources. It handles actions requested by the
// StateOrchestrator.
func Dispatcher() {
	var act int
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
		}

		switch {
		case act == actionNone:
			// nothing to do
		case act == actionTestNow:
			ulog("TestNow\n")
		case act == actionShutdown:
			ulog("SHUTDOWN\n")
			AWSTerminateInstances() // calling this here guarantees no one else is accessing the data
			UhuraShutdown()         // TODO: this is an unceremonious shutdown, a hack for now
		}
	}
}
