package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//  The application data structure
type UhuraApp struct {
	Port           int             // What port are we listening on
	Debug          bool            // Debug mode -- show ulog messages on screen
	DebugToScreen  bool            // Send logging info to screen too
	DryRun         bool            // when true, scripts it produces skip calls to create new cloud instances
	KeepEnv        bool            // don't terminate environment instances
	InternalTest   bool            // do unit tests...
	StateChg       chan AppStatChg // Http requests mem state update
	StateChgAck    chan int        // acknowledge when done
	LogStatus      chan StatusReq  // http requests to log status message
	LogStatusAck   chan int        // Ack channel
	HReqMem        chan int        // request to access Env Descr memory
	HReqMemAck     chan int        // done with memory
	LogEnvDescr    chan int        // request for EnvDescr to be printed
	LogEnvDescrAck chan int        // EnvDescr ack channel
	LogString      chan string     // log a string to ulog
	LogStringAck   chan int        // log string ack channel
	URL            string          // URL where master can be contacted
	EnvDescFname   string          // The filename of the Environment Descriptor
	LogFile        *os.File        // Uhura's logfile
	QmstrBaseLinux []byte          // data for first part of the Linux shell script
	QmstrHdrWin    []byte          // data for first part of the Windows script
	QmstrFtrWin    []byte          // data for the last part of the Windows script
	// LogEnvDescrDone chan int        // done printing
}

var Uhura UhuraApp

func ProcessCommandLine() {
	dbugPtr := flag.Bool("d", false, "debug mode - includes debug info in logfile")
	dtscPtr := flag.Bool("D", false, "LogToScreen mode - prints log messages to stdout")
	portPtr := flag.Int("p", 8100, "port on which uhura listens")
	dryrPtr := flag.Bool("n", false, "Dry Run - don't actually create new instances on AWS")
	envdPtr := flag.String("e", "", "environment descriptor filename")
	murlPtr := flag.String("t", "", "public dns hostname where master can be contacted")
	keepPtr := flag.Bool("k", false, "Keep environment after tests complete (don't terminate)")
	unitPtr := flag.Bool("u", false, "invoke internal tests")
	flag.Parse()

	Uhura.Port = *portPtr
	Uhura.Debug = *dbugPtr
	Uhura.DebugToScreen = *dtscPtr
	Uhura.DryRun = *dryrPtr
	Uhura.KeepEnv = *keepPtr
	Uhura.InternalTest = *unitPtr

	if Uhura.InternalTest {
		Uhura.DryRun = true // force this for internal tests
	}

	// It's not as straightforward as you might think to set the hostname
	// On a home network (like mine), the host name is not in anyone's dns, so
	// using the hostname as a network address is useless. Here's the logic that
	// seems to work...
	s := ""
	if *murlPtr == "" { // if nothing was specified on the cmd line
		uname, _ := exec.Command("sh", "-c", "uname").Output() // determine the OS name
		sysname := string(uname)
		sysname = strings.TrimRight(sysname, "\n\r")
		if "Darwin" == string(sysname) { // if a Mac, we almost certainly want to use localhost
			s = "localhost" // so use localhost
		} else { // if not, we're probably on AWS, so...
			s, _ = os.Hostname() // just use the host name
		}
		Uhura.URL = fmt.Sprintf("http://%s:%d/", s, Uhura.Port)
	} else {
		Uhura.URL = fmt.Sprintf("http://%s:%d/", *murlPtr, Uhura.Port) // use what was specified on the cmd line
	}
	Uhura.EnvDescFname = fmt.Sprintf("%s", *envdPtr)
	fmt.Printf("Uhura.EnvDescFname = %s\n", Uhura.EnvDescFname)
}

func InitUhura() {
	log.SetOutput(Uhura.LogFile)
	ulog("**********   U H U R A   **********\n")
	ulog("Uhura starting on: %s\n", Uhura.URL)
	if Uhura.Debug {
		ulog("Debug logging enabled\n")
	}
	if Uhura.DebugToScreen {
		ulog("Logging to Screen enabled\n")
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	ulog("Current working directory = %v\n", dir)
	ulog("environment descriptor: %s\n", Uhura.EnvDescFname)
	ulog("Port=%d, Debug=%v, DryRun=%v\n", Uhura.Port, Uhura.Debug, Uhura.DryRun)
	// Pull in the data we need to build the cloud initialization scripts.
	// The data we pull in for Linux is the first half of the script. For
	// Windows we pull in the "header" and the "footer" of the script and we
	// generate a little bit of information to sandwich between them. After
	// we pull in the Environment Description file, we generate a script for
	// each instance we create that makes up the environment we're building.
	// we pass it to the AWS launch command that will run after the OS has
	// started up. We generate these scripts in uenv.go. We're just pulling
	// in the data now.

	// First, validate the directory where we find the files.
	var qmdir string
	qmbasefname := "/usr/local/accord/bin/qmaster.sh" // assume linux name
	if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
		qmbasefname = "/c/Accord/bin/qmaster.sh" // if linux name fails, try windows name
		if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
			fmt.Printf("Cannot find required file qmaster.sh\n")
			os.Exit(3)
		} else {
			qmdir = "/c/Accord/bin"
		}
	} else {
		qmdir = "/usr/local/accord/bin"
	}

	// Linux...
	Uhura.QmstrBaseLinux, err = ioutil.ReadFile(qmbasefname)
	check(err)
	Uhura.QmstrHdrWin, err = ioutil.ReadFile(fmt.Sprintf("%s/qmaster.scr1", qmdir))
	check(err)
	Uhura.QmstrFtrWin, err = ioutil.ReadFile(fmt.Sprintf("%s/qmaster.scr2", qmdir))
	check(err)

}

func InitEnv() {
	ParseEnvDescriptor()
}

// This whole routine is a hack until I work out something better
func UhuraShutdown() {

	time.Sleep(3 * time.Second) // this is a hack until we work out the channel logic
	ulog("Shutdown Handler\n")
	ulog("Normal Shutdown\n")
	ulog("Exiting uhura\n")
	os.Exit(0)
}

func main() {
	// Let's get a log file going first.  If I put this file create in any other call
	// it seems to stop working after the call returns. Must be some sort of a scoping thing
	// that I don't understand. But for now, creating the logfile in the main() routine
	// seems to be the way to make it work.
	var err error
	Uhura.LogFile, err = os.OpenFile("uhura.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer Uhura.LogFile.Close()
	log.SetOutput(Uhura.LogFile)

	// OK, now on with the show...
	ProcessCommandLine()
	InitUhura()
	DispatcherCreateChannels()

	if Uhura.InternalTest {
		BeatOnTheChannelMessaging()
		os.Exit(0)
	}
	InitEnv()
	InitHTTP()
	go Dispatcher() // get the dispatcher going
	err = http.ListenAndServe(fmt.Sprintf(":%d", Uhura.Port), nil)
	if nil != err {
		ulog(string(err.Error()))
		fmt.Printf("*** Error on http.ListenAndServe: %v\n", err)
		os.Exit(4)
	}
}
