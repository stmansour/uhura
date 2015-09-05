package main

//
// The logger and some debugging aids...
//

import (
    "fmt"
    "log"
    "net/http"
)

func ulog(format string, a ...interface{}) {
    p := fmt.Sprintf(format, a...)
    log.Printf(p)
    if Uhura.Debug {
        fmt.Printf(p)
    }
}

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


func DPrintHttpRequest(r *http.Request) {
    if Uhura.Debug {
        PrintHttpRequest(r)
    }
}

func StateToInt(s string) int {
    var i int
    switch {
        case s == "INIT":
            i = uINIT
        case s == "READY":
            i = uREADY
        case s == "TEST":
            i = uTEST
        case s == "DONE":
            i = uDONE
        default:
            i = 0
    }
    return i
}

func StateToString(i int) string {
    var s string;
    switch {
        case i == uINIT:
            s = "INIT"
        case i == uREADY:
            s = "READY"
        case i == uTEST:
            s = "TEST"
        case i == uDONE:
            s = "DONE"
        default:
            s = "<<unknown state>>"
    }
    return fmt.Sprintf("%d (%s)", i, s)
}

func PrintStatusMsg(s *StatusReq) {
    ulog("Status Message\n")
    ulog("\tState: %s\n", s.State)
    ulog("\tInstName: %s\n", s.InstName)
    ulog("\tUID: %s\n", s.UID)
    ulog("\tTstamp: %s\n", s.Tstamp)
}

func DPrintStatusMsg(s *StatusReq) {
    if Uhura.Debug {
        PrintStatusMsg(s)
    }
}

func PrintEnvInstance( e *InstDescr, i int) {
    ulog("    Instance[%d]\n", i)
    ulog("\tInstName    : %s\n", e.InstName)
    ulog("\tOS          : %s\n", e.OS)
    ulog("\tApps        :\n")
    for j := 0; j < len(e.Apps); j++ {
        ulog("\t\tUID         : %s\n", e.Apps[j].UID )
        ulog("\t\tName        : %s\n", e.Apps[j].Name )
        ulog("\t\tRepo        : %s\n", e.Apps[j].Repo )
        ulog("\t\tPublicDNS   : %s\n", e.Apps[j].PublicDNS )
        ulog("\t\tUPort       : %d\n", e.Apps[j].UPort)
        ulog("\t\tIsTest      : %v\n", e.Apps[j].IsTest)
        ulog("\t\tState       : %s\n", StateToString(e.Apps[j].State))
        ulog("\t\tRunCmd      : %s\n", e.Apps[j].RunCmd)
        ulog("\t\t--------------\n")
    }
}

func DPrintEnvInstance( e *InstDescr, i int) {
    if Uhura.Debug {
        PrintEnvInstance( e, i )
    }
}

func PrintEnvDescr() {
    ulog("----------------------  UEnv  ----------------------\n")
    ulog("EnvName  : %s\n",UEnv.EnvName)
    ulog("State    : %s\n",StateToString(UEnv.State))
    ulog("UhuraPort: %d\n",UEnv.UhuraPort)
    ulog("Instances: %d\n", len(UEnv.Instances))
    for i := 0; i < len(UEnv.Instances); i++ {
        PrintEnvInstance( &UEnv.Instances[i], i )
    }
}

func DPrintEnvDescr(s string) {
    if Uhura.Debug {
        ulog(s)
        PrintEnvDescr()
    }
}