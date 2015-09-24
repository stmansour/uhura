package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

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

func AWSTerminateInstances() {
	if !Uhura.DryRun && !Uhura.KeepEnv {
		app := "aws"
		args := []string{"ec2", "terminate-instances", "--output", "json", "--instance-ids"}
		for i := 0; i < len(UEnv.Instances); i++ {
			args = append(args, UEnv.Instances[i].InstAwsID)
		}
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
