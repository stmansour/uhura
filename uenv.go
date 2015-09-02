package main

// This package is devoted to handling the environment descriptor
// and making it real.

import (
	"encoding/json"
	"fmt"
    "io/ioutil"
    "log"
    "os"
)

type AppDescr struct {
	Name string `json: "Name"`	
	Repo string `json: "Repo"`
	IsTest bool `json: "IsTest"`
}

type InstDescr struct {
	InstName string `json:"InstName"`
	OS string `json: "OS"`
	Count int `json: "Count"`
	Apps [] AppDescr `json: "Apps"`
}

type EnvDescr struct {
	EnvName string `json: "EnvName"`
	UhuraPort int `json: "UhuraPort"`
	Instances [] InstDescr
}

//  The main data object for this module
var UEnv EnvDescr

func PrintEnvDescriptor() {
	fmt.Printf("Executing environment build for: %s\n", UEnv.EnvName)
	fmt.Printf("Will attempt to build %d Instances\n", len(UEnv.Instances))
	fmt.Printf("UhuraPort = %d\n", UEnv.UhuraPort)
	for i := 0; i < len(UEnv.Instances); i++ {
		fmt.Printf("Instance[%d]:  %s,  %s, count=%d\n", i, UEnv.Instances[i].InstName, UEnv.Instances[i].OS, UEnv.Instances[i].Count)
		fmt.Printf("Apps:")
		for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
			fmt.Printf("\t(%s, %s, IsTest = %v)\n", UEnv.Instances[i].Apps[j].Name, UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].IsTest)
		}
		fmt.Printf("\n")
	}
}

// OK, this is a major cop-out, but not sure what else to do...
func check(e error) {
    if e != nil {
        panic(e)
    }
}

// Execute the descriptor
func ExecuteDescriptor() {
	if (0 == UEnv.UhuraPort) {
		UEnv.UhuraPort = 8080;		// default port for Uhura
	}
    PrintEnvDescriptor()  // debugging purposes only

    // Read in the basis for each quartermaster script
    qmbasefname := "/usr/local/accord/bin/qmaster.sh"		// assume linux name
    if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
    	qmbasefname = "/c/Accord/bin/qmaster.sh"			// if linux name fails, try windows name
    	if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
    		log.Printf("Cannot find required file qmaster.sh\n")
    		os.Exit(3);
    	}
	}
    qmasterdata, err := ioutil.ReadFile(qmbasefname)
    check(err)

    // Build the quartermaster script to create each environment instance...
    var qmstr string
    var f *os.File
    var apps string
	for i := 0; i < len(UEnv.Instances); i++ {
		// Build up a string with all the apps to deploy to this instance
		apps = "";
		for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
			apps += fmt.Sprintf("artf_get %s %s\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)
		}
		// Now build a script for each instance
		for j := 0; j < UEnv.Instances[i].Count; j++ {
			qmstr = fmt.Sprintf("qmstr-%s-%d", UEnv.Instances[i].InstName, j)	// unique name for each script
			f, err = os.Create(qmstr)
	    	defer f.Close()
	    	fmt.Printf("Created file: %s,  apps = %s\n", qmstr, apps)
	    	_, err = f.Write(qmasterdata)
	    	check(err)
	    	_, err = f.Write([]byte(apps))
	    	check(err)
	    	f.Sync()
		}
	}
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
 
    // Now that we have the datastructure filled in, we can 
    // begin to execute it.
    ExecuteDescriptor();
}
