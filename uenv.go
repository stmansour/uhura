package main

// This package is devoted to handling the environment descriptor
// and making it real.

import (
	"encoding/json"
	"fmt"
    "io/ioutil"
    "os"
    "os/exec"
)

const (
	uINIT 	= iota
	uREADY 	= iota
	uTEST   = iota
	uDONE   = iota
)

type AppDescr struct {
	UID string 			`json: "UID"`
	Name string 		`json: "Name"`	
	Repo string 		`json: "Repo"`
	PublicDNS string 	`json: "PublicDNS"`
	UPort int   		`json: "UPort"`
	IsTest bool 		`json: "IsTest"`
	State int 			
	RunCmd string 		`jsno: "RunCmd"`
}

type InstDescr struct {
	InstName string `json:"InstName"`
	OS string 		`json: "OS"`
	Apps [] AppDescr `json: "Apps"`
}

type EnvDescr struct {
	EnvName string `json: "EnvName"`
	UhuraPort int `json: "UhuraPort"`
	State int
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
func EnvDescrScriptName(i int) string {
	var ftype string
	if UEnv.Instances[i].OS == "Windows" {
		ftype = "scr"
	} else {
		ftype = "sh"
	}
	return fmt.Sprintf("qmstr-%s.%s", UEnv.Instances[i].InstName, ftype)
}

// Make a Windows init script for the ith Instance
func MakeWindowsScript(i int) {
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
	qmstr := EnvDescrScriptName(i)	
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f, Uhura.QmstrHdrWin)
	phoneHome := fmt.Sprintf("$UHURA_MASTER_URL = \"%s\"\n",Uhura.MasterURL)
	phoneHome += fmt.Sprintf("$MY_INSTANCE_NAME = \"%s\"\n",UEnv.Instances[i].InstName)
	FileWriteString(f,&apps)
	FileWriteString(f,&phoneHome)
	FileWriteBytes(f,Uhura.QmstrFtrWin)
	f.Sync()
}

// Make a linux init script for the ith Instance
func MakeLinuxScript(i int) {
	// First, build up a string with all the apps to deploy to this instance
	// Now build a script for each instance
	apps := ""
	for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
		apps += fmt.Sprintf("artf_get %s %s\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)
	}

	// now we have all wwe need to create and write the file
	qmstr := EnvDescrScriptName(i)	
	phoneHome := fmt.Sprintf("UHURA_MASTER_URL=%s\n",Uhura.MasterURL)
	phoneHome += fmt.Sprintf("MY_INSTANCE_NAME=\"%s\"\n",UEnv.Instances[i].InstName)
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
		if UEnv.Instances[i].OS == "Windows" {
			MakeWindowsScript(i)
		} else {
			MakeLinuxScript(i)
		}
	}
}

//  Create the environments
//  Note:  the scripts must be in either
func ExecScript(i int) {
	var script string
	// determine where the scripts exist
	path := "/c/Accord/bin"

	_, err := os.Stat(path)
	if nil != err {
		path = "/usr/local/accord/bin"
		_, err := os.Stat(path)
		if nil != err {
			ulog("neither /c/Accord/bin nor /usr/local/accord/bin exist!!\nPlease check installation")
			check(err)
		}
	}
	if "Windows" == UEnv.Instances[i].OS {
		script = "cr_win_testenv.sh"
	} else {
		script = "cr_linux_testenv.sh"
	}
	arg0 := EnvDescrScriptName(i)
	app := fmt.Sprintf("%s/%s", path, script )
	cmd := exec.Command(app, arg0)
	stdout, err := cmd.Output()
	if err != nil {
		ulog("*** Error *** running %s:  %v\n", app, err.Error())
	}
    ulog( "exec %s\noutput:\n%s\n", app, string(stdout))
}

// Execute the descriptor.  That means create the environment(s).
func ExecuteDescriptor() {
	for i := 0; i < len(UEnv.Instances); i++ {
		ExecScript(i)
	}
}


// Parse the environment
func ParseEnvDescriptor() {
	// First, see if we can read the file in
    ulog("ParseEnvDescriptor - Loading %s\n", *Uhura.EnvDescFname)
    content, e := ioutil.ReadFile(*Uhura.EnvDescFname)
    if e != nil {
        ulog("File error on Environment Descriptor file: %v\n", e)
        os.Exit(1)		// no recovery from this
    }
    ulog("%s\n", string(content))
    
    // OK, now we have the json describing the environment in content (a string)
    // Parse it into an internal data structure...
    err := json.Unmarshal(content, &UEnv)
    if (err != nil) {
    	ulog("Error unmarshaling Environment Descriptor json: %s\n",err)
    	check(err)
    }
 	DPrintEnvDescr("UEnv after initial parse:");

    // Now that we have the datastructure filled in, we can 
    // begin to execute it.
    CreateInstanceScripts()
    ExecuteDescriptor()
}
