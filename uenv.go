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

const (
	uINIT  = iota
	uREADY = iota
	uTEST  = iota
	uDONE  = iota
)

type AppDescr struct {
	UID       string
	Name      string
	Repo      string
	PublicDNS string
	UPort     int
	IsTest    bool
	State     int
	RunCmd    string
}

type InstDescr struct {
	InstName string
	OS       string
	Apps     []AppDescr
}

type EnvDescr struct {
	EnvName   string
	UhuraPort int
	State     int
	Instances []InstDescr
}

//  The main data object for this module
var UEnv *EnvDescr

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
	var eol *string
	var n int = len(UEnv.Instances[i].Apps)
	for j := 0; j < n; j++ {
		if 1+j == n {
			eol = &empty
		} else {
			eol = &comma
		}
		apps += fmt.Sprintf("\t\t\"%s\"%s\n", UEnv.Instances[i].Apps[j].Name, *eol)
	}
	apps += ")\n"
	qmstr := EnvDescrScriptName(i)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f, Uhura.QmstrHdrWin)
	phoneHome := fmt.Sprintf("$UHURA_MASTER_URL = \"%s\"\n", Uhura.MasterURL)
	phoneHome += fmt.Sprintf("$MY_INSTANCE_NAME = \"%s\"\n", UEnv.Instances[i].InstName)
	FileWriteString(f, &apps)
	FileWriteString(f, &phoneHome)
	FileWriteBytes(f, Uhura.QmstrFtrWin)
	f.Sync()
}

// Make a linux init script for the ith Instance
func MakeLinuxScript(i int) {
	// First, build up a string with all the apps to deploy to this instance
	// Now build a script for each instance
	apps := ""
	dirs := "mkdir ~ec2-user/apps;cd apps\n"
	ctrl := ""
	for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
		dirs += fmt.Sprintf("mkdir ~/apps/%s\n", UEnv.Instances[i].Apps[j].Name)
		apps += fmt.Sprintf("artf_get %s %s.tar.gz\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)

		//TODO:                       vvvvvv---should be IsController
		if !UEnv.Instances[i].Apps[j].IsTest {
			app := UEnv.Instances[i].Apps[j].Name
			ctrl += fmt.Sprintf("gunzip %s.tar.gz;tar xf %s.tar;cd %s;./activate.sh START > activate.log 2>&1\n", app, app, app)
		}
	}

	// now we have all wwe need to create and write the file
	qmstr := EnvDescrScriptName(i)
	phoneHome := fmt.Sprintf("UHURA_MASTER_URL=%s\n", Uhura.MasterURL)
	phoneHome += fmt.Sprintf("MY_INSTANCE_NAME=\"%s\"\n", UEnv.Instances[i].InstName)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f, Uhura.QmstrBaseLinux)
	FileWriteString(f, &phoneHome)
	FileWriteString(f, &dirs)
	FileWriteString(f, &apps)
	FileWriteString(f, &ctrl)
	f.Sync()
}

// Unmarshal the data in descriptor.
// The machine bring-up scripts are created at the same time.
func CreateInstanceScripts() {
	if 0 == UEnv.UhuraPort {
		UEnv.UhuraPort = 8080 // default port for Uhura
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

	// Gather the args...
	arg0 := EnvDescrScriptName(i)
	arg1, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	app := fmt.Sprintf("%s/%s", path, script)
	args := []string{arg0, arg1}

	// Run it
	ulog("exec %s %s %s\n", app, arg0, arg1)
	if err := exec.Command(app, args...).Run(); err != nil {
		ulog("*** Error *** running %s:  %v\n", app, err.Error())
	}
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
	ulog("ParseEnvDescriptor - Loading %s\n", Uhura.EnvDescFname)
	content, e := ioutil.ReadFile(Uhura.EnvDescFname)
	if e != nil {
		ulog("File error on Environment Descriptor file: %v\n", e)
		os.Exit(1) // no recovery from this
	}
	ulog("%s\n", string(content))

	// OK, now we have the json describing the environment in content (a string)
	// Parse it into an internal data structure...
	err := json.Unmarshal(content, &UEnv)
	if err != nil {
		ulog("Error unmarshaling Environment Descriptor json: %s\n", err)
		check(err)
	}
	DPrintEnvDescr("UEnv after initial parse:")

	// Now that we have the datastructure filled in, we can
	// begin to execute it.
	CreateInstanceScripts()
	ExecuteDescriptor()
}
