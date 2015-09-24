package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// AwsTerm describes the json ouptut from the command 'aws ec2 terminate-instances'
type AwsTerm struct {
	Terminatinginstances []struct {
		Instanceid   string `json:"InstanceId"`
		Currentstate struct {
			Code int    `json:"Code"`
			Name string `json:"Name"`
		} `json:"CurrentState"`
		Previousstate struct {
			Code int    `json:"Code"`
			Name string `json:"Name"`
		} `json:"PreviousState"`
	} `json:"TerminatingInstances"`
}

// AWSTerminateInstances terminates the compute instances that we created
// to make a test/deployment environment.
func AWSTerminateInstances() {
	if !Uhura.DryRun && !Uhura.KeepEnv {
		app := "aws"
		args := []string{"ec2", "terminate-instances", "--output", "json", "--instance-ids"}

		// access shared memory... use the channel
		Uhura.HReqMem <- 1 // ask to access the shared mem, blocks until granted
		<-Uhura.HReqMemAck // make sure we got it
		for i := 0; i < len(UEnv.Instances); i++ {
			args = append(args, UEnv.Instances[i].InstAwsID)
		}
		Uhura.HReqMemAck <- 1 // tell Dispatcher we're done with the data

		cmd := exec.Command(app, args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			fmt.Printf("*** Error *** running aws ec2 terminate-instances:  %v\n", err.Error())
		}

		var term AwsTerm
		err := json.Unmarshal(out.Bytes(), &term)
		if err != nil {
			ulog("Error Unmarshaling output from aws ec2 terminate-instances: %s\n", err)
		}

		// TODO:  check to make sure all instances actually got deleted
	}
}
