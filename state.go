package main

// AppStatChg is a struct of data describing an application status change
// it is sent to the Dispatcher (who in turn sends it to the state orchestrator) via
// a channel.
type AppStatChg struct {
	inst  int // which instance
	app   int // which app
	state int // what state
}

const (
	actionNone = iota
	actionTestNow
	actionShutdown
)

// AllAppStatesPast checks for all states being beyond the supplied state.
// This can happen during initialization when some instances
// are in the READY state before others have gotten past
// the UNKNOWN state.
func AllAppStatesPast(es int) bool {
	past := true
	for i := 0; past && i < len(UEnv.Instances); i++ {
		for j := 0; past && j < len(UEnv.Instances[i].Apps); j++ {
			past = (es < UEnv.Instances[i].Apps[j].State)
		}
	}
	return past
}

// A state change occurs when the states of all apps are greater
// than our current state
func hasStateChanged() bool {
	return AllAppStatesPast(UEnv.State)
}

// StateOrchestarator applies the applications status change to the datastore
// then performs any state change necessary
func StateOrchestarator(asc *AppStatChg) int {
	UEnv.Instances[asc.inst].Apps[asc.app].State = asc.state
	if !hasStateChanged() {
		return actionNone
	}
	switch {
	case UEnv.State == uUNKNOWN:
		if AllAppStatesPast(uINIT) {
			UEnv.State = uREADY
			return actionTestNow
		}
		UEnv.State = uINIT
		return actionNone
	case UEnv.State == uINIT:
		UEnv.State = uREADY
		return actionTestNow // EMAIL ALL TGOs
	case UEnv.State == uREADY:
		if AllAppStatesPast(uTEST) {
			UEnv.State = uDONE
			return actionShutdown // They're all done, shutdown
		}
		UEnv.State = uTEST
		return actionNone
	case UEnv.State == uTEST:
		UEnv.State = uDONE
		return actionShutdown
	case UEnv.State == uDONE:
		UEnv.State = uTERM
		return actionShutdown
	case UEnv.State == uTERM:
		return actionShutdown
	default:
		ulog("ChangeState unhandled: UEnv.State = %s, don't know what the netx state is\n",
			stateToString(UEnv.State))
		return actionNone
	}
}
