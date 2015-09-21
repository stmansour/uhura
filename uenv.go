package main

// This package is devoted to handling the environment descriptor
// and managing its resources.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

const (
	uINIT = iota
	uREADY
	uTEST
	uDONE

//	uTERM
)

type AppDescr struct {
	UID    string
	Name   string
	Repo   string
	UPort  int
	IsTest bool
	State  int
	RunCmd string
}

type InstDescr struct {
	InstName  string
	OS        string
	HostName  string
	InstAwsID string
	Apps      []AppDescr
}

//  Environment Descriptor. This struct defines a test (or production) environment.
//  EnvName is the name associated with the collection of Instances
//  UhuraURL is the http url where tgo instances should contact uhura
//  UhuraPort (may not be needed) is the port on which uhura listens. Default is 8100
//  ThisInst - when this value is present, it is to inform a tgo instance which instance it is.
//             The value is the index into the instances array.
//  State = the overall state of the environment, one of  INIT, READY, TEST, DONE
//  Instances - an array of instance descriptors that describe each instance in the environment.
type EnvDescr struct {
	EnvName   string
	UhuraURL  string
	UhuraPort int
	ThisInst  int
	State     int // overall state of the environment
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
	n := len(UEnv.Instances[i].Apps)
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
	phoneHome := fmt.Sprintf("$UHURA_MASTER_URL = \"%s\"\n", Uhura.URL)
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
		dirs += fmt.Sprintf("mkdir ~ec2-user/apps/%s\n", UEnv.Instances[i].Apps[j].Name)
		apps += fmt.Sprintf("artf_get %s %s.tar.gz\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)

		//TODO:                       vvvvvv---should be IsController
		if !UEnv.Instances[i].Apps[j].IsTest {
			app := UEnv.Instances[i].Apps[j].Name
			ctrl += fmt.Sprintf("gunzip %s.tar.gz;tar xf %s.tar;cd %s\n", app, app, app)
		}
	}

	// now we have all wwe need to create and write the file
	qmstr := EnvDescrScriptName(i)
	phoneHome := fmt.Sprintf("UHURA_MASTER_URL=%s\n", Uhura.URL)
	phoneHome += fmt.Sprintf("MY_INSTANCE_NAME=\"%s\"\n", UEnv.Instances[i].InstName)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	FileWriteBytes(f, Uhura.QmstrBaseLinux)
	FileWriteString(f, &phoneHome)
	FileWriteString(f, &dirs)
	FileWriteString(f, &apps)
	FileWriteString(f, &ctrl)

	s := "cat >uhura_map.json <<ZZEOF\n"
	FileWriteString(f, &s)

	// content, err := ioutil.ReadFile(Uhura.EnvDescFname)
	// check(err)
	// FileWriteBytes(f, content)
	UEnv.ThisInst = i
	b, err := json.Marshal(&UEnv)
	FileWriteBytes(f, b)

	s = "\nZZEOF\n"
	FileWriteString(f, &s)

	// We want all the files to be owned by ec2-user.  Wait 1 second for everything to get
	// started up, then change the ownership.
	startitup := fmt.Sprint("./activate.sh START > activate.log 2>&1\n")
	FileWriteString(f, &startitup)
	s = fmt.Sprintf("sleep 1;cd ~ec2-user/;chown -R ec2-user:ec2-user *\n")
	FileWriteString(f, &s)

	f.Sync()
}

// Unmarshal the data in descriptor.
// The machine bring-up scripts are created at the same time.
func CreateInstanceScripts() {
	if 0 == UEnv.UhuraPort {
		UEnv.UhuraPort = 8100 // default port for Uhura
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
	app := fmt.Sprintf("%s/%s", path, script)
	arg0 := EnvDescrScriptName(i)
	arg1, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	arg2 := ""
	if Uhura.DryRun {
		arg2 = "-n"
	}

	// Run it
	ulog("exec %s %s %s %s", app, arg0, arg1, arg2)
	if err := exec.Command(app, arg0, arg1, arg2).Run(); err != nil {
		ulog("*** Error *** running %s:  %v\n", app, err.Error())
	}

	// Read in the response
	if !Uhura.DryRun {
		awsinstance := AWSLoadNewInstanceInfo(arg0 + ".json")
		UEnv.Instances[i].InstAwsID = awsinstance.Instances[0].Instanceid
	}
}

func SetInstanceHostNames() {
	ReadAllAwsInstances("descrinst.json")
	for i := 0; i < len(UEnv.Instances); i++ {
		ulog("Search for InstAwsID = %s\n", UEnv.Instances[i].InstAwsID)
		UEnv.Instances[i].HostName = SearchReservationsForPublicDNS(UEnv.Instances[i].InstAwsID)
	}
	DPrintEnvDescr("UEnv after launching all instances:")
}

// Execute the descriptor.  That means create the environment(s).
func ExecuteDescriptor() {
	for i := 0; i < len(UEnv.Instances); i++ {
		ExecScript(i)
	}
	// After all the execs have been done, we need to ask aws for
	// the describe-instances json, then parse it for the public dns names
	// for each of our instances.
	if !Uhura.DryRun {
		// the problem with doing this right away is that it takes
		// aws some time to get all the public dns stuff worked out.
		// So, if we call it immediately, things won't work. We need to
		// wait some (unknown) amount of time.  Let's give it 15 sec and see
		// how we do.
		// This is really bad... we need to figure out something else
		time.Sleep(15 * time.Second)

		args := []string{"ec2", "describe-instances", "--output", "json"}
		cmd := exec.Command("aws", args...)
		outfile, err := os.Create("descrinst.json")
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
		cmd.Stdout = outfile

		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		cmd.Wait()
		SetInstanceHostNames()
	}
}

func WriteEnvDescr() {
	// Now generate the env.json file that we'll send to all the instances
	b, err := json.Marshal(&UEnv)
	f, err := os.Create("uhura_map.json")
	check(err)
	defer f.Close()
	FileWriteBytes(f, b)
	f.Sync()
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

	// Add Uhura's URL to the environment description
	UEnv.UhuraURL = Uhura.URL
	// WriteEnvDescr()   // removed -- I think we'll write it from memory each time because we need to set ThisInst name.

	// Now that we have the datastructure filled in, we can
	// begin to execute it.
	CreateInstanceScripts()
	ExecuteDescriptor()

}
