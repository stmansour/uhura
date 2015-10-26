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
	uUNKNOWN = iota
	uINIT
	uREADY
	uTEST
	uDONE
	uTERM
)

// KeyVal is a generic key/value pair struct
type KeyVal struct {
	Key string
	Val string
}

// AppDescr is an application description that provides information
// about an application that we deploy.
type AppDescr struct {
	UID    string
	Name   string
	Repo   string
	UPort  int
	IsTest bool
	State  int
	RunCmd string // if present overrides "activate.sh startr"
	KVs    []KeyVal
}

// ResourceDescr is a structure of data defining what optional resources
// each instance needs.
type ResourceDescr struct {
	MySql     bool   // if true, mysql will be started
	RestoreDB string // name of database to restore. Expects to be found int ext-tools/testing
}

// InstDescr is a structure of data describing every Instance (virtual
// computer) that we deploy in the cloud
type InstDescr struct {
	InstName  string
	OS        string
	HostName  string
	InstAwsID string
	Resources ResourceDescr
	Apps      []AppDescr
}

// EnvDescr is a struct that defines a test (or production) environment.
type EnvDescr struct {
	EnvName   string      // the name associated with the collection of Instances
	UhuraURL  string      // the http url where tgo instances should contact uhura
	UhuraPort int         // (may not be needed) is the port on which uhura listens. Default is 8100
	ThisInst  int         // informs a tgo instance which Instance index it is. Basically ignored within uhura
	State     int         // overall state of the environment
	Instances []InstDescr // the array of InstDescr describing each instance in the env.
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
func fileWriteString(f *os.File, s *string) {
	_, err := f.WriteString(*s)
	check(err)
}

func fileWriteBytes(f *os.File, b []byte) {
	_, err := f.Write(b)
	check(err)
}

// Create a deterministic unique name for each script
func envDescrScriptName(i int) string {
	var ftype string
	if UEnv.Instances[i].OS == "Windows" {
		ftype = "scr"
	} else {
		ftype = "sh"
	}
	return fmt.Sprintf("qmstr-%s.%s", UEnv.Instances[i].InstName, ftype)
}

// Make a Windows init script for the ith Instance
func makeWindowsScript(i int) {
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
	qmstr := envDescrScriptName(i)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	fileWriteBytes(f, Uhura.QmstrHdrWin)
	phoneHome := fmt.Sprintf("$UHURA_MASTER_URL = \"%s\"\n", Uhura.URL)
	phoneHome += fmt.Sprintf("$MY_INSTANCE_NAME = \"%s\"\n", UEnv.Instances[i].InstName)
	fileWriteString(f, &apps)
	fileWriteString(f, &phoneHome)
	fileWriteBytes(f, Uhura.QmstrFtrWin)
	f.Sync()
}

// Make a linux init script for the ith Instance
func makeLinuxScript(i int) {
	// First, install any resources needed (such as a database)
	resources := ""
	if UEnv.Instances[i].Resources.MySql {
		resources += "install_mysql\n"
	}
	if len(UEnv.Instances[i].Resources.RestoreDB) > 0 {
		resources += fmt.Sprintf("restoredb \"%s\"\n", UEnv.Instances[i].Resources.RestoreDB)
	}

	// Next, build up a string with all the apps to deploy to this instance
	// Build a script for each instance
	apps := ""
	dirs := "mkdir ~ec2-user/apps;cd apps\n"
	ctrl := ""
	for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
		dirs += fmt.Sprintf("mkdir ~ec2-user/apps/%s\n", UEnv.Instances[i].Apps[j].Name)
		apps += fmt.Sprintf("artf_get %s %s.tar.gz\n", UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].Name)
		app := UEnv.Instances[i].Apps[j].Name
		ctrl += fmt.Sprintf("gunzip %s.tar.gz;tar xf %s.tar\n", app, app)
	}
	ctrl += fmt.Sprintf("cd tgo\n")

	// now we have all wwe need to create and write the file
	qmstr := envDescrScriptName(i)
	phoneHome := fmt.Sprintf("UHURA_MASTER_URL=%s\n", Uhura.URL)
	phoneHome += fmt.Sprintf("MY_INSTANCE_NAME=\"%s\"\n", UEnv.Instances[i].InstName)
	f, err := os.Create(qmstr)
	check(err)
	defer f.Close()
	fileWriteBytes(f, Uhura.QmstrBaseLinux)
	if len(resources) > 0 {
		fileWriteString(f, &resources)
	}
	fileWriteString(f, &phoneHome)
	fileWriteString(f, &dirs)
	fileWriteString(f, &apps)
	fileWriteString(f, &ctrl)

	s := "cat >uhura_map.json <<ZZEOF\n"
	fileWriteString(f, &s)

	// content, err := ioutil.ReadFile(Uhura.EnvDescFname)
	// check(err)
	// fileWriteBytes(f, content)
	UEnv.ThisInst = i
	b, err := json.Marshal(&UEnv)
	fileWriteBytes(f, b)

	s = "\nZZEOF\n"
	fileWriteString(f, &s)

	// adding sleep 60 (1 min) to address a miming issue. It looks like it takes a while
	// for the ports to be open.  Let's see if this addresses the problem

	// We want all the files to be owned by ec2-user.  Wait 1 second for everything to get
	// started up, then change the ownership.
	startitup := fmt.Sprint("sleep 60;./activate.sh START > activate.log 2>&1\n")
	fileWriteString(f, &startitup)
	s = fmt.Sprintf("sleep 1;cd ~ec2-user/;chown -R ec2-user:ec2-user *\n")
	fileWriteString(f, &s)

	f.Sync()
}

// Unmarshal the data in descriptor.
// The machine bring-up scripts are created at the same time.
func createInstanceScripts() {
	if 0 == UEnv.UhuraPort {
		UEnv.UhuraPort = 8100 // default port for Uhura
	}
	// Build the quartermaster script to create each environment instance...
	for i := 0; i < len(UEnv.Instances); i++ {
		if UEnv.Instances[i].OS == "Windows" {
			makeWindowsScript(i)
		} else {
			makeLinuxScript(i)
		}
	}
}

//  Create the environments
//  Note:  the scripts must be in either
func execScript(i int) {
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
	arg0 := envDescrScriptName(i)
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
		ulog("*** Check the 'aws' command, does it need to be configured?\n")
	}

	// Read in the response
	if !Uhura.DryRun {
		awsinstance := AWSLoadNewInstanceInfo(arg0 + ".json")
		UEnv.Instances[i].InstAwsID = awsinstance.Instances[0].Instanceid
	}
}

func setInstanceHostNames() {
	ReadAllAwsInstances("descrinst.json")
	for i := 0; i < len(UEnv.Instances); i++ {
		ulog("Search for InstAwsID = %s\n", UEnv.Instances[i].InstAwsID)
		UEnv.Instances[i].HostName = searchReservationsForPublicDNS(UEnv.Instances[i].InstAwsID)
	}
	dPrintEnvDescr("UEnv after launching all instances:\n")
}

// ExecuteDescriptor - i.e., create the environment.
func ExecuteDescriptor() {
	for i := 0; i < len(UEnv.Instances); i++ {
		execScript(i)
	}
	// After all the execs have been done, we need to ask aws for
	// the describe-instances json, then parse it for the public dns names
	// for each of our instances.
	if Uhura.DryRun {
		// this is a bit of a hack, but it helps with testing
		for i := 0; i < len(UEnv.Instances); i++ {
			UEnv.Instances[i].HostName = "localhost"
		}
	} else {
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
		setInstanceHostNames()
	}
}

func writeEnvDescr() {
	// Now generate the env.json file that we'll send to all the instances
	b, err := json.Marshal(&UEnv)
	f, err := os.Create("uhura_map.json")
	check(err)
	defer f.Close()
	fileWriteBytes(f, b)
	f.Sync()
}

func LoadEnvDescriptor(fname string) {
	content, e := ioutil.ReadFile(fname)
	if e != nil {
		ulog("File error on Environment Descriptor file: %v\n", e)
		os.Exit(1) // no recovery from this
	}

	// OK, now we have the json describing the environment in content (a string)
	// Parse it into an internal data structure...
	err := json.Unmarshal(content, &UEnv)
	if err != nil {
		ulog("Error unmarshaling Environment Descriptor json: %s\n", err)
		check(err)
	}

}

// ParseEnvDescriptor - parse the json file that describes the environment
// we are to build up and manage
func ParseEnvDescriptor() {
	// First, see if we can read the file in
	ulog("ParseEnvDescriptor - Loading %s\n", Uhura.EnvDescFname)
	LoadEnvDescriptor(Uhura.EnvDescFname)
	dPrintEnvDescr("UEnv after initial parse:")

	// Add Uhura's URL to the environment description
	UEnv.UhuraURL = Uhura.URL
	// writeEnvDescr()   // removed -- I think we'll write it from memory each time because we need to set ThisInst name.

	// Now that we have the datastructure filled in, we can
	// begin to execute it.
	createInstanceScripts()
	ExecuteDescriptor()

}
