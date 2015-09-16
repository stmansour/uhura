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

type TestStep struct {
	State     string
	InstName  string
	InstUID   string
	ReplyCode int
}

var test1 = []TestStep{
	{"ARGH", "MainTestInstance", "wprog2", RespNoSuchInstance},
	{"YUCK", "MainTestInstance", "prog2", InvalidState},
	{"INIT", "MainTestInstance", "prog2", RespOK},
	{"INIT", "MainWinInstance", "wprog2", RespOK},
	{"READY", "MainTestInstance", "prog2", RespOK},
	{"READY", "MainWinInstance", "wprog2", RespOK},
	{"TEST", "MainTestInstance", "prog2", RespOK},
	{"TEST", "MainWinInstance", "wprog2", RespOK},
	{"DONE", "MainTestInstance", "prog2", RespOK},
	{"DONE", "MainWinInstance", "wprog2", RespOK},
}

func TestStatusHandler(t *testing.T) {
	Uhura.EnvDescFname = "./test/stateflow_normal/env1.json"
	Uhura.DryRun = true
	InitUhura()
	SetUpHttpEnv()

	ts := httptest.NewServer(http.HandlerFunc(StatusHandler))
	defer ts.Close()

	url := fmt.Sprintf("%s/status/", ts.URL)
	for j := 0; j < 2; j++ {
		Uhura.DebugToScreen = (j == 0)    // this forces a lot more code to be executed
		Uhura.Debug = Uhura.DebugToScreen // this forces a lot more code to be executed
		t.Logf("j=%d, Uhura.DebugToScreen=%v, Uhura.Debug=%v", j, Uhura.DebugToScreen, Uhura.Debug)
		for i := 0; i < len(test1); i++ {
			test := test1[i]
			r := StatusReq{test.State, test.InstName, test.InstUID, time.Now().Format(time.RFC822)}
			data, _ := json.Marshal(r)
			body := bytes.NewBuffer(data)
			reply, _ := http.Post(url, "application/json", body)
			response, _ := ioutil.ReadAll(reply.Body)
			resp := new(UResp)
			json.Unmarshal(response, resp)
			if test.ReplyCode != resp.ReplyCode {
				t.Errorf("Expected: %d,   Received: %d", test.ReplyCode, resp.ReplyCode)
			}
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
