package main

import (
	"fmt"
	"os"
)

// Read in json and validate that we can identify the InstanceId
func TestResourceParsing() {
	LoadEnvDescriptor("./test/sys4/sys4.json")
	if UEnv.Instances[0].Resources.MySql != true {
		fmt.Printf("Resource callout for mysql is wrong, expected true, found %v", UEnv.Instances[0].Resources.MySql)
		os.Exit(1)
	}
	fmt.Printf("Resource load tests: PASS\n")
}