package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"
)

// ResetUEnv sets the entire internal environment definition back to the
// UKNOWN state. This must only be used for testing.
func ResetUEnv() {
	Uhura.HReqMem <- 1 // ask to access the shared mem, blocks until granted
	<-Uhura.HReqMemAck // make sure we got it
	UEnv.State = uUNKNOWN
	for i := 0; i < len(UEnv.Instances); i++ {
		for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
			UEnv.Instances[i].Apps[j].State = uUNKNOWN
		}
	}
	Uhura.HReqMemAck <- 1 // tell Dispatcher we're done with the data
}

func internalStateTest(inst int, c chan int) {
	// env descr = ./test/utdata/ut2.json
	//            inst, app, state,  ENV STATE, action
	// This case - they go straight from READY to DONE - because they test
	// extremely fast.  Uhura tells them to test. They finish testing before
	// uhura samples their state after informing them to test
	type TestStep struct {
		asc   AppStatChg
		state int
		act   int
	}

	var kvm = KVMsg{"", []KeyVal{}}

	var intTest5 = []TestStep{
		{AppStatChg{0, 0, uUNKNOWN, kvm}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uUNKNOWN, kvm}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uUNKNOWN, kvm}, uUNKNOWN, actionNone},
		{AppStatChg{1, 0, uREADY, kvm}, uUNKNOWN, actionNone},
		{AppStatChg{2, 0, uREADY, kvm}, uUNKNOWN, actionNone},
		{AppStatChg{0, 0, uREADY, kvm}, uREADY, actionTestNow},
		{AppStatChg{0, 0, uTEST, kvm}, uREADY, actionNone},
		{AppStatChg{0, 0, uDONE, kvm}, uREADY, actionNone},    // uhura moves to DONE right after telling the apps to test
		{AppStatChg{1, 0, uDONE, kvm}, uREADY, actionNone},    // when uhura asks, this app is already done testing
		{AppStatChg{2, 0, uDONE, kvm}, uDONE, actionShutdown}, // when uhura asks, this app is already done testing
	}

	for i := 0; i < len(intTest5); i++ {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
		Uhura.LogString <- fmt.Sprintf("internalStateTest[%d]\n", inst)
		<-Uhura.LogStringAck // wait for confirmation
		test := intTest5[i]
		Uhura.StateChg <- test.asc
		reply := <-Uhura.StateChgAck
		if reply != 1 {
			Uhura.LogString <- fmt.Sprintf("internalStateTest[%d] -Curious... dispatcher ack'd with value %d instead of 1\n",
				inst, reply)
			<-Uhura.LogStringAck // wait for confirmation
		}
	}
	c <- 1 // signal that we're done
}

// envDump sends dump the env
func envDump(i int, c chan int) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	Uhura.LogString <- fmt.Sprintf("envDump[%d]\n", i) // update ulog
	<-Uhura.LogStringAck                               // wait for confirmation
	Uhura.LogEnvDescr <- 1                             // request a log dump
	<-Uhura.LogEnvDescrAck                             // wait for it to complete
	c <- 1                                             // signal that we're done
}

// poundTheStatusHandler simulates a thundering-herd of http requestors sending
// status messages.
func poundTheStatusHandler(inst int, c chan int) {
	var kvm = KVMsg{"", []KeyVal{}}
	var intTest5 = []StatusReq{
		StatusReq{"UNKNOWN", "TGO-0", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"UNKNOWN", "TGO-1", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"UNKNOWN", "TGO-2", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"READY", "TGO-0", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"READY", "TGO-1", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"READY", "TGO-2", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"TEST", "TGO-0", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm},
		StatusReq{"DONE", "TGO-1", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm}, // uhura moves to DONE right after telling the apps to test
		StatusReq{"DONE", "TGO-2", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm}, // when uhura asks, this app is already done testing
		StatusReq{"DONE", "TGO-0", "tgo0", time.Now().Format(time.RFC822), nil, false, nil, kvm}, // when uhura asks, this app is already done testing
	}

	ts := httptest.NewServer(http.HandlerFunc(StatusHandler))
	defer ts.Close()
	url := fmt.Sprintf("%s/status/", ts.URL)

	for i := 0; i < len(intTest5); i++ {
		time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
		data, _ := json.Marshal(intTest5[i])
		body := bytes.NewBuffer(data)
		reply, _ := http.Post(url, "application/json", body)
		response, _ := ioutil.ReadAll(reply.Body)
		resp := new(UResp)
		json.Unmarshal(response, resp)
		Uhura.LogString <- fmt.Sprintf("poundTheStatusHandler[%d] - http response: Status: %s, ReplyCode: %d, Timestamp: %s\n",
			inst, resp.Status, resp.ReplyCode, resp.Timestamp)
		<-Uhura.LogStringAck // wait for confirmation
	}
	c <- 1 // signal that we're done
}

// beatOnTheChannelMessaging simulates a thundering-herd of requestors sending
// status messages and requesting envDescr dumps and shared memory changes all
// at once. It validates the operation of the channels that share memory between
// all the different requestors.
func beatOnTheChannelMessaging() {
	Uhura.EnvDescFname = "./test/utdata/ut2.json" // here's the env to load
	initEnv()                                     // make sure we have it before starting dispatcher
	go Dispatcher()                               // get the dispatcher going
	c := make(chan int)                           // comm channel for tests to indicat completion
	n := 5                                        // how many of each go routine?
	m := 0                                        // how many chan reads to do
	for j := 0; j < 10; j++ {
		ResetUEnv()
		for i := 0; i < n; i++ { // blast away at the Orchestrator
			go internalStateTest(i, c)     // update status at random intervals
			m++                            // another read
			go envDump(i, c)               // dump the env descr at random intervals
			m++                            // another read
			go poundTheStatusHandler(i, c) // do a bunch of http stuff at the same time
			m++                            // another read
		}
	}
	for i := 0; i < m; i++ {
		<-c // make sure that all of the go routines completed
	}
}
