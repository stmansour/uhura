package main

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

// A state change occurs when the states of all apps are greater
// than our current state
func hasStateChanged() bool {
	return AllAppStatesPast(UEnv.State)
}

// the orchestrator applies the applications status change to the datastore
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
		} else {
			UEnv.State = uINIT
			return actionNone
		}
	case UEnv.State == uINIT:
		UEnv.State = uREADY
		return actionTestNow
	case UEnv.State == uREADY:
		if AllAppStatesPast(uTEST) {
			UEnv.State = uDONE
			return actionShutdown
		} else {
			UEnv.State = uTEST
			return actionNone
		}
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
			StateToString(UEnv.State))
		return actionNone
	}
}
