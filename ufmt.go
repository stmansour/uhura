package main

//
// Some debugging aids...
//

import (
    "fmt"
    "net/http"
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

func PrintEnvDescriptor() {
    fmt.Printf("Executing environment build for: %s\n", UEnv.EnvName)
    fmt.Printf("Will attempt to build %d Instances\n", len(UEnv.Instances))
    fmt.Printf("UhuraPort = %d\n", UEnv.UhuraPort)
    for i := 0; i < len(UEnv.Instances); i++ {
        fmt.Printf("Instance[%d]:  %s,  %s, count=%d\n", i, UEnv.Instances[i].InstName, UEnv.Instances[i].OS, UEnv.Instances[i].Count)
        fmt.Printf("Apps:")
        for j := 0; j < len(UEnv.Instances[i].Apps); j++ {
            fmt.Printf("\t(%s, %s, IsTest = %v)\n", UEnv.Instances[i].Apps[j].Name, UEnv.Instances[i].Apps[j].Repo, UEnv.Instances[i].Apps[j].IsTest)
        }
        fmt.Printf("\n")
    }
}
