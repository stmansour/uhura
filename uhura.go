package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "regexp"
    "strings"
)

//  Data needs for Uhura operating in Master mode
type UhuraMaster struct {
    envDescriptorFname *string
}

//  Data needs for Uhura operating in Slave mode
type UhuraSlave struct {
	MasterURL string	// where to contact the master

}

const (
	MODE_SLAVE	 = iota
	MODE_MASTER  = iota
)

//  The application data structure
type UhuraApp struct {
    Port int 				// What port are we listening on
    Mode int 				// master or slave
    Debug bool 				// Debug mode -- show ulog messages on screen
    MasterURL string 		// URL where master can be contacted
    EnvDescFname *string 	// The filename of the Environment Descriptor
    QmstrBaseLinux []byte	// data for first part of the Linux shell script 
    QmstrHdrWin []byte		// data for first part of the Windows script
    QmstrFtrWin []byte		// data for the last part of the Windows script
    UhuraMaster				// data unique to master
    UhuraSlave				// data unique to slave
}

type UhuraResponse struct {
    Status string
    Timestamp string
}

var Uhura UhuraApp



func handleCmdLineArgs() {
	dbugPtr := flag.Bool("d", false, "debug mode - prints log messages to stdout")
    portPtr := flag.Int("p", 8080, "port on which uhura listens" )
    modePtr := flag.String("m", "slave", "mode of operation: (master|slave)")
    envdPtr := flag.String("e", "", "environment descriptor filename, required if mode == master")
    murlPtr := flag.String("t", "localhost", "public dns hostname where master can be contacted")
    flag.Parse()

    Uhura.Port = *portPtr
    Uhura.Debug = *dbugPtr
    ulog("**********   U H U R A   **********\n")

    match, _ := regexp.MatchString("(master|slave)", strings.ToLower(*modePtr));
    if (!match) {
        ulog("*** ERROR *** Mode (-m) must be either 'master' or 'slave'\n")
        os.Exit(1)
    }
    ulog("Uhura starting in %s mode on port %d\n", *modePtr, Uhura.Port)

     if (len(*envdPtr) == 0) {
        ulog("*** ERROR *** Environment descriptor is required for operation in master mode\n");
        os.Exit(2)
    }
    ulog("environment descriptor: %s\n", *envdPtr)
    match, _ = regexp.MatchString("master", strings.ToLower(*modePtr));
    if (match) {
    	Uhura.MasterURL = fmt.Sprintf("http://%s:%d/",*murlPtr,Uhura.Port)
    	Uhura.EnvDescFname = envdPtr
    	Uhura.Mode = MODE_MASTER
    }
}


//  Slog through the minutia of startup.
func UhuruInit() {
	// Let's get a log file going first
	f, err := os.OpenFile("uhura.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

    // Gather the command line info...
    handleCmdLineArgs()
 
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
    qmbasefname := "/usr/local/accord/bin/qmaster.sh"		// assume linux name
    if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
    	qmbasefname = "/c/Accord/bin/qmaster.sh"			// if linux name fails, try windows name
    	if _, err := os.Stat(qmbasefname); os.IsNotExist(err) {
    		ulog("Cannot find required file qmaster.sh\n")
    		os.Exit(3);
    	} else {
    		qmdir = "/c/Accord/bin"
    	}
	} else {
		qmdir = "/usr/local/accord/bin"
	}

	// Linux...
    Uhura.QmstrBaseLinux, err = ioutil.ReadFile(qmbasefname)
    check(err)
    Uhura.QmstrHdrWin, err = ioutil.ReadFile(fmt.Sprintf("%s/qmaster.scr1",qmdir))
    check(err)
    Uhura.QmstrFtrWin, err = ioutil.ReadFile(fmt.Sprintf("%s/qmaster.scr2",qmdir))
    check(err)

	// Very last step in the initialization process.  Now
	// That everything has been pulled in, we can process the
	// Environment Descriptor 
    ParseEnvDescriptor()
 }

func main() {
	UhuruInit()
    http.HandleFunc("/shutdown/",   makeHandler(ShutdownHandler))
    http.HandleFunc("/status/",     makeHandler(StatusHandler))
    http.HandleFunc("/map/",        makeHandler(MapHandler))
    http.HandleFunc("/test-done/",  makeHandler(TestDoneHandler))
    http.HandleFunc("/test-start/", makeHandler(TestStartHandler))
    err := http.ListenAndServe(fmt.Sprintf(":%d",Uhura.Port), nil)
    if (nil != err) {
        ulog(string(err.Error()))
    }
}