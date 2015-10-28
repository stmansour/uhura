package main

import (
	"fmt"
	"os"
)

// Read in json and validate that we can identify the InstanceId
func TestResourceParsing() {
	LoadEnvDescriptor("./test/sys4/sys4.json")
	if UEnv.Instances[0].Resources.MySQL != true {
		fmt.Printf("Resource callout for mysql is wrong, expected true, found %v", UEnv.Instances[0].Resources.MySQL)
		os.Exit(1)
	}
	if UEnv.Instances[0].Apps[1].AppRes.RestoreMySQLdb != "testdb.sql" ||
		UEnv.Instances[0].Apps[1].AppRes.DBname != "accord" {
		fmt.Printf("Resource callout for RestoreMySQLdb is wrong, expected testdb.sql, found %s\n",
			UEnv.Instances[0].Apps[1].AppRes.RestoreMySQLdb)
		fmt.Printf("UEnv = %#v\n", UEnv)
		os.Exit(2)
	}
	fmt.Printf("Resource load tests: PASS\n")
}
