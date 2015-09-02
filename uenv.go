package main

// This package is devoted to handling the environment descriptor
// and making it real.

import (
	"encoding/json"
	"fmt"
    "io/ioutil"
    "log"
    "os"
	// "strings"
)

type AppDescr struct {
	Name string `json: "Name"`	
	Repo string `json: "Repo"`
}

type InstDescr struct {
	InstName string `json:"InstName"`
	OS string `json: "OS"`
	Count int `json: "Count"`
	Apps [] AppDescr `json: "Apps"`
}

type EnvDescr struct {
	EnvName string `json: "EnvName"`
	Instances [] InstDescr
}

//  The main data object for this module
var UEnv EnvDescr

// Execute the descriptor
func ExecuteDescriptor() {
	log.Printf("Executing environment build for: %s\n", UEnv.EnvName)
	log.Printf("Will attempt to build %d Instances\n", len(UEnv.Instances))
	for i := 0; i < len(UEnv.Instances); i++ {
		fmt.Printf("Instance[%d]:  %s,  %s, count=%d\n", 
			i, UEnv.Instances[i].InstName, 
			UEnv.Instances[i].OS, UEnv.Instances[i].Count)
		fmt.Printf("Apps:")
		for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
			fmt.Printf("\t(%s, %s)\n", UEnv.Instances[i].Apps[j].Name, UEnv.Instances[i].Apps[j].Repo)
		}
		fmt.Printf("\n")
	}
	return
}

// Parse the environment
func ParseEnvDescriptor(fname *string) {
	// First, see if we can read the file in
    log.Printf("ParseEnvDescriptor - Loading %s\n", *fname)
    content, e := ioutil.ReadFile(*fname)
    if e != nil {
        log.Printf("File error on Environment Descriptor file: %v\n", e)
        os.Exit(1)		// no recovery from this
    }
    log.Printf("%s\n", string(content))
    
    // OK, now we have the json describing the environment in a string
    // Parse it into an internal data structure...
    err := json.Unmarshal(content, &UEnv)
    if (err != nil) {
    	fmt.Println(err)
    }
    fmt.Println(UEnv)

    // Now that we have the datastructure filled in, we can 
    // begin to execute it.
    ExecuteDescriptor();
}
