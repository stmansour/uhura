package main

// This package is devoted to handling the environment descriptor
// and making it real.

import (
	"encoding/json"
	"fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
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


// OK, this is a major cop-out, but not sure what else to do...
func check(e error) {
    if e != nil {
        panic(e)
    }
}

// just reduces the lines of code
func FileWriteString(f *os.File, s *string) {
	_, err := f.WriteString(*s)
	check(err)
}

func FileWriteBytes(f *os.File, b []byte) {
	_, err := f.Write(b)
	check(err)
}

// Create a deterministic unique name for each script
func EnvDescrScriptName(i, k int) string {
	var ftype string
	if UEnv.Instances[i].OS == "Windows" {
		ftype = "scr"
	} else {
		ftype = "sh"
	}
	return fmt.Sprintf("qmstr-%s-%d.%s", UEnv.Instances[i].InstName, k, ftype)
}

// Make a Windows init script for the ith Instance, and the kth Count 
func MakeWindowsScript(i, k int) {
	// First, build up a string with all the apps to deploy to this instance
	// Now build a script for each instance.  We assume for Windows that everything
	// is in "ext-tools/utils"
	empty := ""
	comma := ","
	apps := ""
	var eol *string;
	var n int = len(UEnv.Instances[i].Apps)
	for j := 0; j < n; j++ {
		if 1 + j == n { eol = &empty } else { eol = &comma }
		apps += fmt.Sprintf("\t\t\"%s\"%s\n", UEnv.Instances[i].Apps[j].Name, *eol)
	}
	apps += ")\n"
	qmstr := EnvDescrScriptName(i, k)	
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f, Uhura.QmstrHdrWin)
	phoneHome := fmt.Sprintf("$UHURA_MASTER_URL = \"%s\"\n",Uhura.MasterURL)
	phoneHome += fmt.Sprintf("$MY_INSTANCE_NAME = \"%s\"\n$MY_INSTANCE_COUNT = \"%d\"\n",UEnv.Instances[i].InstName, k)
	FileWriteString(f,&apps)
	FileWriteString(f,&phoneHome)
	FileWriteBytes(f,Uhura.QmstrFtrWin)
	f.Sync()
}

// Make a linux init script for the ith Instance, and the kth Count 
func MakeLinuxScript(i, k int) {
	// First, build up a string with all the apps to deploy to this instance
	// Now build a script for each instance
	apps := ""
	for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
		apps += fmt.Sprintf("artf_get %s %s\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)
	}

	// now we have all wwe need to create and write the file
	qmstr := EnvDescrScriptName(i, k)	
	phoneHome := fmt.Sprintf("UHURA_MASTER_URL=%s\n",Uhura.MasterURL)
	phoneHome += fmt.Sprintf("MY_INSTANCE_NAME=\"%s\"\nMY_INSTANCE_COUNT=%d\n",UEnv.Instances[i].InstName, k)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f,Uhura.QmstrBaseLinux)
	FileWriteString(f,&phoneHome)
	FileWriteString(f,&apps)
	f.Sync()
}

// Unmarshal the data in descriptor.
// The machine bring-up scripts are created at the same time.
func CreateInstanceScripts() {
	if 0 == UEnv.UhuraPort {
		UEnv.UhuraPort = 8080		// default port for Uhura
	}
    // Build the quartermaster script to create each environment instance...
	for i := 0; i < len(UEnv.Instances); i++ {
		for j := 0; j < UEnv.Instances[i].Count; j++ {
			if UEnv.Instances[i].OS == "Windows" {
				MakeWindowsScript(i,j)
			} else {
				MakeLinuxScript(i,j)
			}
		}
	}
}

func ExecScript(i,j int) {
	var app string
	if UEnv.Instances[i].OS == "Windows" {
		app = "/c/Accord/bin/cr_win_testenv.sh"
	} else {
		app = "/usr/local/accord/bin/cr_linux_testenv.sh"
	}
	arg0 := EnvDescrScriptName(i, j)
	fmt.Printf("exec.Command(%s, %s)\n", app, arg0)
	cmd := exec.Command(app, arg0)
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("*** Error *** running %s:  %v\n", app, err.Error())
	}
    log.Printf( "exec %s\noutput:\n%s\n", app, string(stdout))
}

// Execute the descriptor.  That means create the environment(s).
func ExecuteDescriptor() {
	for i := 0; i < len(UEnv.Instances); i++ {
		for j := 0; j < UEnv.Instances[i].Count; j++ {
			ExecScript(i,j)
		}
	}
}


// Parse the environment
func ParseEnvDescriptor() {
	// First, see if we can read the file in
    log.Printf("ParseEnvDescriptor - Loading %s\n", *Uhura.EnvDescFname)
    content, e := ioutil.ReadFile(*Uhura.EnvDescFname)
    if e != nil {
        log.Printf("File error on Environment Descriptor file: %v\n", e)
        os.Exit(1)		// no recovery from this
    }
    log.Printf("%s\n", string(content))
    
    // OK, now we have the json describing the environment in content (a string)
    // Parse it into an internal data structure...
    err := json.Unmarshal(content, &UEnv)
    if (err != nil) {
    	log.Printf("Error unmarshaling Environment Descriptor json: %s\n",err)
    	check(err)
    }
 
    // Now that we have the datastructure filled in, we can 
    // begin to execute it.
    CreateInstanceScripts()
    ExecuteDescriptor()
}
