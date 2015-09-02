package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "encoding/json"
    "os"
    "regexp"
    "strings"
    "time"
)


// This is a debugging tool for me.
func PrintHttpRequest(r *http.Request) {
    fmt.Println("r.Method: " + r.Method)
    fmt.Println("r.URL:")
    fmt.Println("    Scheme: " + r.URL.Scheme)
    fmt.Println("    Opaque: " + r.URL.Opaque)
    fmt.Printf( "    User:   %v\n", r.URL.User)
    fmt.Printf( "    Host: %s\n", r.URL.Host)
    fmt.Printf( "    Host: %s\n", r.URL.Host)
    fmt.Printf( "    Path: %s\n", r.URL.Path)
    fmt.Printf( "    RawPath: %s\n", r.URL.RawPath)
    fmt.Printf( "    RawQuery: %s\n", r.URL.RawQuery)
    fmt.Printf( "    Fragment: %s\n", r.URL.Fragment)
    fmt.Println("r.Proto: " + r.Proto)
    fmt.Printf( "r.ProtoMajor: %d\n", r.ProtoMajor)
    fmt.Printf( "r.ProtoMinor: %d\n", r.ProtoMinor)
    fmt.Printf( "r.Header: %v\n", r.Header)
    fmt.Printf( "r.Body: %v\n", r.Body)
    if nil != r.Body {
        body := r.FormValue("body")
        fmt.Printf("    body = %s\n", body)
    }
    fmt.Printf( "r.ContentLength: %d\n", r.ContentLength)
    fmt.Printf( "r.Close: %t\n", r.Close)
    fmt.Printf( "r.Host: %s\n", r.Host)
    fmt.Printf( "r.Form: %v\n", r.Form)
    fmt.Printf( "r.PostForm: %v\n", r.PostForm)
    fmt.Printf( "r.Trailer: %v\n", r.Trailer)
    fmt.Printf( "r.RemoteAddr: %s\n", r.RemoteAddr)
    fmt.Printf( "r.RequestURI: %s\n", r.RequestURI)
}

//  Data needs for Uhura operating in Master mode
type UhuraMaster struct {
    envDescriptorFname *string
}

//  Data needs for Uhura operating in Slave mode
type UhuraSlave struct {
	masterURL string	// where to contact the master

}

//  The application data structure
type UhuraApp struct {
    port string 	// What port are we listening on
    mode string 	// master or slave
    UhuraMaster		// data unique to master
    UhuraSlave		// data unique to slave

}

type UhuraResponse struct {
    Status string
    Timestamp string
}

var Uhura UhuraApp

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Shutdown Handler")
    log.Println("Normal Shutdown")
    os.Exit(0)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Status Handler")
    m := UhuraResponse{ Status: "OK", Timestamp: time.Now().Format(time.RFC850)}
    str, err := json.Marshal(m)
    if (nil != err) {
        fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n", 
            "encoding error", time.Now().Format(time.RFC850))
    } else {
        fmt.Fprintf(w,string(str))
    }
}

func TestStartHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Test Start Handler")
    m := UhuraResponse{ Status: "OK", Timestamp: time.Now().Format(time.RFC850)}
    str, err := json.Marshal(m)
    if (nil != err) {
        fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n", 
            "encoding error", time.Now().Format(time.RFC850))
    } else {
        fmt.Fprintf(w,string(str))
    }

}

func TestDoneHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Test Done Handler")
    m := UhuraResponse{ Status: "OK", Timestamp: time.Now().Format(time.RFC850)}
    str, err := json.Marshal(m)
    if (nil != err) {
        fmt.Fprintf(w, "{\n\"Status\": \"%s\"\n\"Timestamp:\": \"%s\"\n}\n", 
            "encoding error", time.Now().Format(time.RFC850))
    } else {
        fmt.Fprintf(w,string(str))
    }
}

func makeHandler( fn func (http.ResponseWriter, *http.Request)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
	    PrintHttpRequest(r)
        fn(w, r)
    }
}


func handleCmdLineArgs() {
    portPtr := flag.Int("p", 8080, "port on which uhura listens" )
    modePtr := flag.String("m", "slave", "mode of operation: (master|slave)")
    envdPtr := flag.String("e", "", "environment descriptor, required if mode == master")
    flag.Parse()
    Uhura.port = fmt.Sprintf(":%d", *portPtr)
    match, _ := regexp.MatchString("(master|slave)", strings.ToLower(*modePtr));
    if (!match) {
        log.Printf("*** ERROR *** Mode (-m) must be either 'master' or 'slave'")
        os.Exit(1)
    }
    log.Printf("Uhura starting in %s mode on port %s\n", *modePtr, Uhura.port)

     if (len(*envdPtr) == 0) {
        log.Printf("*** ERROR *** Environment descriptor is required for operation in master mode\n");
        os.Exit(2)
    }
    log.Printf("environment descriptor: %s\n", *envdPtr)
    match, _ = regexp.MatchString("master", strings.ToLower(*modePtr));
    if (match) {
        ParseEnvDescriptor(envdPtr)
    }
}

func UhuruInit() {
	f, err := os.OpenFile("uhura.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
	    log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
    log.Printf("**********   U H U R A   **********")
    handleCmdLineArgs()
}

func main() {
	UhuruInit()
    http.HandleFunc("/shutdown/",   makeHandler(ShutdownHandler))
    http.HandleFunc("/status/",     makeHandler(StatusHandler))
    http.HandleFunc("/test-done/",  makeHandler(TestDoneHandler))
    http.HandleFunc("/test-start/", makeHandler(TestStartHandler))
    err := http.ListenAndServe(Uhura.port, nil)
    if (nil != err) {
        log.Println(err)
    }
}

