package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type StatusReply struct {
	Status string
	Tstamp time.Time
}

var tests = []struct {
	State    string
	InstName string
	InstUID  string
	response string
}{
	{"ARGH", "MainTestInstance", "wprog2", "BAD INSTANCE-UID: MainTestInstance-wprog2"},
	{"YUCK", "MainTestInstance", "prog2", "BAD STATE: YUCK"},
	{"INIT", "MainTestInstance", "prog2", "OK"},
	{"INIT", "MainWinInstance", "wprog2", "OK"},
	{"READY", "MainTestInstance", "prog2", "OK"},
	{"READY", "MainWinInstance", "wprog2", "OK"},
	{"TEST", "MainTestInstance", "prog2", "OK"},
	{"TEST", "MainWinInstance", "wprog2", "OK"},
	{"DONE", "MainTestInstance", "prog2", "OK"},
	{"DONE", "MainWinInstance", "wprog2", "OK"},
}

func TestStatusHandler(t *testing.T) {
	Uhura.EnvDescFname = "./test/master_normal_state_flow/env1.json"
	InitUhura()
	SetUpHttpEnv()
	Uhura.Debug = true         // this forces a lot more code to be executed
	Uhura.DebugToScreen = true // ditto
	ts := httptest.NewServer(http.HandlerFunc(StatusHandler))
	defer ts.Close()

	url := fmt.Sprintf("%s/status/", ts.URL)

	for i := 0; i < len(tests); i++ {
		test := tests[i]
		r := StatusReq{test.State, test.InstName, test.InstUID, time.Now().Format(time.RFC822)}
		data, _ := json.Marshal(r)
		body := bytes.NewBuffer(data)
		reply, _ := http.Post(url, "application/json", body)
		response, _ := ioutil.ReadAll(reply.Body)
		resp := new(StatusReply)
		json.Unmarshal(response, resp)
		if test.response != resp.Status {
			t.Errorf("Expected: %s,   Received: %s", test.response, resp.Status)
		}
	}
}

func TestMapHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(MapHandler))
	defer ts.Close()
	url := fmt.Sprintf("%s/map/", ts.URL)
	resp, err := http.Get(url)
	if nil != err {
		t.Errorf("http.Get failed")
	}
	t.Logf("resp = %+v\n", resp)
	b, _ := ioutil.ReadAll(resp.Body)

	// OK, now we have the json describing the environment in content (a string)
	// Parse it into an internal data structure...
	var ed EnvDescr
	err = json.Unmarshal(b, &ed)
	if err != nil {
		t.Errorf("json.Unmarshal failed")
	}
	expect := "My Test Environment"
	if ed.EnvName != expect {
		t.Errorf("expected %s, found %s", expect, ed.EnvName)
	}
}
