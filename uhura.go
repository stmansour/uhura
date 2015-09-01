package main

import (
    "flag"
    "fmt"
    "io/ioutil"
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
    envDescriptorFname *string;
}

//  Data needs for Uhura operating in Slave mode
type UhuraSlave struct {
	mode string			// operational mode:  master or slave

}

//  The application data structure
type UhuraApp struct {
    port string
    UhuraMaster
    UhuraSlave

}

var Uhura UhuraApp

type UhuraResponse struct {
    Status string
    Timestamp string
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Shutdown Handler")
    fmt.Println("Normal Shutdown")
    os.Exit(0)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Status Handler")
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
    fmt.Println("Test Start Handler")
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
    fmt.Println("Test Done Handler")
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

func readEnvDescriptor(fname *string) (error) {
    fmt.Printf("Loading %s\n", *fname)
    file, e := ioutil.ReadFile(*fname)
    if e != nil {
        fmt.Printf("File error: %v\n", e)
        os.Exit(1)
    }
    fmt.Printf("%s\n", string(file))
    return e;
}

func handleCmdLineArgs() {
    portPtr := flag.Int("p", 8080, "port on which uhura listens" )
    modePtr := flag.String("m", "slave", "mode of operation: (master|slave)")
    envdPtr := flag.String("e", "", "environment descriptor, required if mode == master")
    flag.Parse()

    Uhura.port = fmt.Sprintf(":%d", *portPtr)
    fmt.Printf("portPtr = %d,   port = \"%s\"\n", *portPtr, Uhura.port)

    match, _ := regexp.MatchString("(master|slave)", strings.ToLower(*modePtr));
    if (!match) {
        fmt.Printf("*** ERROR *** Mode (-m) must be either 'master' or 'slave'")
        os.Exit(1)
    }
    fmt.Printf("Uhura starting in %s mode at %s\n", *modePtr, time.Now().Format(time.RFC850))

     if (len(*envdPtr) == 0) {
        fmt.Printf("*** ERROR *** Environment descriptor is required for operation in master mode\n");
        os.Exit(2)
    }
    fmt.Printf("Environment Descriptor: %s\n", *envdPtr)
     match, _ = regexp.MatchString("master", strings.ToLower(*modePtr));
    if (match) {
        _ = readEnvDescriptor(envdPtr)
    }
}

func main() {
    handleCmdLineArgs()
    http.HandleFunc("/shutdown/",   makeHandler(ShutdownHandler))
    http.HandleFunc("/status/",     makeHandler(StatusHandler))
    http.HandleFunc("/test-done/",  makeHandler(TestDoneHandler))
    http.HandleFunc("/test-start/", makeHandler(TestStartHandler))
    err := http.ListenAndServe(Uhura.port, nil)
    if (nil != err) {
        fmt.Println(err)
    }
}

