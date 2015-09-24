package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	cmdTESTNOW = iota // tells Tgo to initiate testing
	cmdSTOP
)

// UCommand is a data structure type containing a command
// that uhura sends to tgo instances
type UCommand struct {
	Command   string
	CmdCode   int
	Timestamp string
}

// SendTgoCommand is used to send an HTTP message to a tgo instance asking
// it to perform some action
func SendTgoCommand(url string, cmd *UCommand, reply *UResp) (int, error) {
	b, err := json.Marshal(cmd)
	if err != nil {
		ulog("Cannot marshal command struct! Error: %v\n", err)
		os.Exit(2) // no recovery from this
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ulog("Cannot Post cmd! Error: %v\n", err)
		return 0, err // ?? maybe there's some retry we can do??
	}
	defer resp.Body.Close()

	ulog("Tgo response received\n")

	// pull out the HTTP response code
	var rc int
	var more string
	fmt.Sscanf(resp.Status, "%d %s", &rc, &more)

	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, reply)
	ulog("Tgo @ %s replied: %v\n", url, reply)
	return rc, err
}

// CommsSendTestNow is invoked by the StateOrchestrator when all instances
// have reported that they are in the READY state. TESTNOW tells the instances
// that they can begin testing.
func CommsSendTestNow() {
	// the UResp struct has what we need to send a command to Tgo
	ulog("Comms: sending TESTNOW\n")
	cmd := UCommand{Command: "TESTNOW", CmdCode: cmdTESTNOW, Timestamp: time.Now().Format(time.RFC822)}
	var reply UResp
	for i := 0; i < len(UEnv.Instances); i++ {
		for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
			if UEnv.Instances[i].Apps[j].Name == "tgo" {
				url := fmt.Sprintf("http://%s:%d/",
					UEnv.Instances[i].HostName,
					UEnv.Instances[i].Apps[j].UPort)
				ulog("Comms: TESTNOW -> tgo @ %s\n", url)
				SendTgoCommand(url, &cmd, &reply)
			}
		}
	}
}
