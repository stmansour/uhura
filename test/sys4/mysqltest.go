package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
)
import _ "github.com/go-sql-driver/mysql"

var app struct {
	db   *sql.DB
	user string
}

func errcheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func processCommandLine() {
	uPtr := flag.String("u", "ec2-user", "user name")
	flag.Parse()
	app.user = *uPtr
}

func main() {
	var err error
	processCommandLine()
	s := fmt.Sprintf("%s:@/accord?charset=utf8&parseTime=True", app.user)
	app.db, err = sql.Open("mysql", s)
	if nil != err {
		fmt.Printf("sql.Open: Error = %v\n", err)
	}

	defer app.db.Close()

	err = app.db.Ping()
	if nil != err {
		fmt.Printf("db.Ping: Error = %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("MySql availability test: Passed.\n")
}
